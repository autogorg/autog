package rag

import (
	"fmt"
	"log"
	"os"
	"bytes"
	"container/heap"
	"encoding/json"
	"encoding/gob"
	"crypto/md5"
	"autog"
)

const (
	ErrDocAreadyExists = "Document already exists!"
	ErrDocNotExists    = "Document not exists!"
	ErrChunkNotExists  = "Chunk not exists!"
)

type ScoredChunkIndex {
	Index int
	Score float64
}

type ScoredChunkIndexs []ScoredChunkIndex

func (s ScoredChunkIndexs) Len() int           { return len(s) }
func (s ScoredChunkIndexs) Less(i, j int) bool { return s[i].Score < s[j].Score }
func (s ScoredChunkIndexs) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s *ScoredChunkIndexs) Push(c interface{}) {
	*s = append(*s, c.(ScoredChunkIndex))
}

func (s *ScoredChunkIndexs) Pop() ScoredChunkIndex {
	org := *s
	n := len(org)
	last := org[n-1]
	*s = org[0 : n-1]
	return last
}

func (s *ScoredChunkIndexs) Peek() interface{} {
	return s[0]
}

type MemoryDatabase struct {
	PathToDocument map[string]*MemDocument
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Norm(embed Embedding) float64 {
	n := 0.0
	for _, f := range embed {
		n += f * f
	}
	return math.Sqrt(n)
}

func Norms(embeds []Embedding) []Embedding{
	norms := make([]Embedding, len(embeds))
	for i, embed := range embeds {
		norms[i] = Norm(embed)
	}
	return norms
}

func DotProduct(a, b []float64) float64 {
	result := 0.0
	for i := range a {
		result += a[i] * b[i]
	}
	return result
}

func CosSim(qembeds, dbembeds [][]float64, qnorms, dbnorms *[]float64, qsi, dsi int, topk int, dbchunks *[]autog.Chunk, channel chan<- []autog.ScoredChunks) {
	qn := len(qembeds)
	dn := len(dbembeds)
	scoredChunks := make([]autog.ScoredChunks, qn)

	for qi := 0; qi < qn; qi++ {
		scores := make([]float64, dn)
		for di := 0; di < dn; di++ {
			scores[di] = DotProduct(qembeds[qi], dbembeds[di]) / ((*qnorms)[qi] * (*dbnorms)[di])
		}

		chunkIndexHeap := &autog.ScoredChunkIndexs{}
		heap.Init(chunkIndexHeap)

		for i, score := range scores {
			if pq.Len() < k {
				heap.Push(chunkIndexHeap, ScoredChunkIndex{Index: i, Score: float64(score)})
			} else if score > pq.Peek().Score {
				heap.Pop(chunkIndexHeap)
				heap.Push(chunkIndexHeap, ScoredChunkIndexs{Index: i, Score: float64(score)})
			}
		}

		for chunkIndexHeap.Len() > 0 {
			sci := heap.Pop(chunkIndexHeap).(ScoredChunkIndex)
			sci.Index += dsi
			schunk := &autog.ScoredChunk{
				Chunk : (*dbchunks)[sci.Index],
				Score : sci.Score
			}
			scoredChunks[qi] = append(scoredChunks[qi], schunk)
		}
	}

	channel <- scoredChunks
}

func (md *MemoryDatabase) GetDocument(path string) (*MemDocument, error) {
	doc, ok := md.PathToDocument[path];
	if !ok {
		return doc, return fmt.Errorf(ErrDocNotExists)
	}
	return doc, nil
}

func (md *MemoryDatabase) DelDocument(path string) error {
	if _, ok := md.PathToDocument[path]; !ok {
		return fmt.Errorf(ErrDocNotExists)
	}
	delete(md, path)
	return nil
}

func (md *MemoryDatabase) GetDocumentPaths() ([]string, error) {
	var paths []string
	for path := range md {
		paths = append(paths, path)
	}
	return paths
}

func (md *MemoryDatabase) GetDocumentChunk(path string, idx int) (*MemChunk, error) {
	var chunk *MemChunk
	doc, ok := md.PathToDocument[path];
	if !ok {
		return chunk, fmt.Errorf(ErrDocNotExists)
	}
	if len(doc.Chunks) <= idx {
		return chunk, fmt.Errorf(ErrChunkNotExists)
	}
	return doc.Chunks[idx], nil
}

func (md *MemoryDatabase) GetDocumentChunks(path string) ([]autog.Chunk, []autog.Embedding, error) {
	var chunks []autog.Chunk
	var embeddings []autog.Embedding
	doc, ok := md.PathToDocument[path];
	if !ok {
		return chunks, embeddings, fmt.Errorf(ErrDocNotExists)
	}
	for _, chunk := range doc.Chunks {
		embeddings = append(embeddings, chunk.Embedding)
		chunks = append(chunks, chunk)
	}
	return chunks, embeddings, nil
}

func (md *MemoryDatabase) GetDatabaseChunks() ([]autog.Chunk, []autog.Embedding, error) {
	var chunks []autog.Chunk
	var embeddings []autog.Embedding
	for path, doc := range md {
		for _, chunk := range doc.Chunks {
			chunks = append(chunks, chunk)
			embeddings = append(embeddings, chunk.Embedding)
		}
	}

	return chunks, embeddings, nil
}

func (md *MemoryDatabase) SaveChunks(path string, payload interface{}, chunks []autog.Chunks) error {
    memDoc := &MemDocument{}
	memDoc.SetPath(pat)
	memDoc.SetPayload(payload)
	memDoc.SetChunks(chunks)
	md.PathToDocument[path] = memDoc
    return nil
}

func (md *MemoryDatabase) SearchChunks(path string, embeds []autog.Embedding) ([]autog.ScoredChunks, error) {
	norms := Norms(embeds)
	var scoreds  []autog.ScoredChunks
	var dbchunks []autog.Chunk
	var dbembeds []autog.Embedding
	var dberr error
	if path == autog.DOCUMENT_PATH_NONE {
		dbchunks, dbembeds, dberr = r.Database.GetDatabaseChunks()
	} else {
		dbchunks, dbembeds, dberr = r.Database.GetDocumentChunks(path)
	}
	if dberr != nil {
		return scoreds, dberr
	}
	dbnorms := Norms(dbembeds)

	qnum  := len(queries)
	channel := make(chan []autog.ScoredChunks)
	scoreds = make([]autog.ScoredChunks, qnum)

	var wg sync.WaitGroup
	const qbatch = 100
	const dbatch = 1000
	for qi := 0; qi < len(embeds); qi += qbatch {
		for dbi := 0; dbi < len(dbembeds); dbi += dbatch {
			qj := min(qi + qbatch, len(embeds))
			dj := min(di + dbatch, len(dembeds))

			wg.Add(1)
			go func(qi int, di int, qj int, dj int) {
				wg.Done()
				CosSim(embeds[qi:qj], dbembeds[di:dj], &norms, &dbnorms, qi, di, topk, &dbchunks, channel)
			}(qi, di, qj, dj)
		}
	}

	go func() {
		wg.Wait()
		close(channel)
	}()

	cnt := 0
	for cscores := range channel {
		for ci := range cscores {
			idx := ci + cnt * qbatch
			if idx < len(scoreds) {
				scoreds[idx] = cscores[qi]
			}
		}
		cnt++
	}

	return scoreds, nil
}


