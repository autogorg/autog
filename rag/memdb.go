package rag

import (
	"fmt"
	"math"
	"sync"
	"container/heap"
	"github.com/autogorg/autog"
)

const (
	ErrDocAreadyExists = "Document already exists!"
	ErrDocNotExists    = "Document not exists!"
	ErrChunkNotExists  = "Chunk not exists!"
)

type ScoredChunkIndex struct {
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

func (s *ScoredChunkIndexs) Pop() interface{} {
	org := *s
	n := len(org)
	last := org[n-1]
	*s = org[0 : n-1]
	return last
}

func (s ScoredChunkIndexs) Peek() interface{} {
	return s[0]
}

type MemDocuments []*MemDocument

func (m *MemDocuments) Append(doc *MemDocument) {
	*m = append(*m, doc)
}

type MemoryDatabase struct {
	PathToDocuments map[string]*MemDocuments
}

func NewMemDatabase() (*MemoryDatabase, error) {
	md  := &MemoryDatabase{}
	err := md.InitDatabase()
	return md, err 
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Norm(embed autog.Embedding) float64 {
	n := 0.0
	for _, f := range embed {
		n += f * f
	}
	return math.Sqrt(n)
}

func Norms(embeds []autog.Embedding) autog.Embedding{
	norms := make(autog.Embedding, len(embeds))
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

func CosSim(qembeds, dbembeds []autog.Embedding, qnorms, dbnorms *autog.Embedding, qsi, dsi int, topk int, dbchunks *[]autog.Chunk, channel chan<- []autog.ScoredChunks) {
	qn := len(qembeds)
	dn := len(dbembeds)
	scoredChunks := make([]autog.ScoredChunks, qn)

	for qi := 0; qi < qn; qi++ {
		scores := make([]float64, dn)
		for di := 0; di < dn; di++ {
			scores[di] = DotProduct(qembeds[qi], dbembeds[di]) / ((*qnorms)[qi] * (*dbnorms)[di])
		}

		chunkIndexHeap := &ScoredChunkIndexs{}
		heap.Init(chunkIndexHeap)

		for i, score := range scores {
			if chunkIndexHeap.Len() < topk {
				heap.Push(chunkIndexHeap, ScoredChunkIndex{Index: i, Score: float64(score)})
			} else {
				scoreidx, ok := chunkIndexHeap.Peek().(ScoredChunkIndex)
				if ok && score > scoreidx.Score {
					heap.Pop(chunkIndexHeap)
					heap.Push(chunkIndexHeap, ScoredChunkIndex{Index: i, Score: float64(score)})
				}
			}
		}

		for chunkIndexHeap.Len() > 0 {
			sci := heap.Pop(chunkIndexHeap).(ScoredChunkIndex)
			sci.Index += dsi
			schunk := &autog.ScoredChunk{
				Chunk : (*dbchunks)[sci.Index],
				Score : sci.Score,
			}
			scoredChunks[qi] = append(scoredChunks[qi], schunk)
		}
	}

	channel <- scoredChunks
}

func (md *MemoryDatabase) InitDatabase() error {
	md.PathToDocuments = make(map[string]*MemDocuments)
	return nil
}

func (md *MemoryDatabase) GetDocuments(path string) (*MemDocuments, error) {
	docs, ok := md.PathToDocuments[path];
	if !ok {
		return docs, fmt.Errorf("GetDocuments by [" + path + "] ERROR: " + ErrDocNotExists)
	}
	return docs, nil
}

func (md *MemoryDatabase) DelDocuments(path string) error {
	if _, ok := md.PathToDocuments[path]; !ok {
		return fmt.Errorf("DelDocuments by [" + path + "] ERROR: " + ErrDocNotExists)
	}
	delete(md.PathToDocuments, path)
	return nil
}

func (md *MemoryDatabase) GetPaths() ([]string, error) {
	var paths []string
	for path := range md.PathToDocuments {
		paths = append(paths, path)
	}
	return paths, nil
}

func (md *MemoryDatabase) GetPathChunks(path string) ([]autog.Chunk, []autog.Embedding, error) {
	var chunks []autog.Chunk
	var embeddings []autog.Embedding
	docs, ok := md.PathToDocuments[path]
	if !ok {
		return chunks, embeddings, fmt.Errorf("GetPathChunks by [" + path + "] ERROR: " + ErrDocNotExists)
	}
	for _, doc := range *docs {
		for _, chunk := range doc.Chunks {
			embeddings = append(embeddings, chunk.Embedding)
			chunks = append(chunks, chunk)
		}
	}
	return chunks, embeddings, nil
}

func (md *MemoryDatabase) GetChunks() ([]autog.Chunk, []autog.Embedding, error) {
	var chunks []autog.Chunk
	var embeddings []autog.Embedding
	for _, docs := range md.PathToDocuments {
		for _, doc := range *docs {
			for _, chunk := range doc.Chunks {
				chunks = append(chunks, chunk)
				embeddings = append(embeddings, chunk.Embedding)
			}
		}
	}

	return chunks, embeddings, nil
}

func (md *MemoryDatabase) AppendChunks(path string, payload interface{}, chunks []autog.Chunk) error {
	if _, ok := md.PathToDocuments[path]; !ok {
		return md.SaveChunks(path, payload, chunks)
	}
    memDoc := &MemDocument{}
	memDoc.SetPath(path)
	memDoc.SetPayload(payload)
	memDoc.SetChunks(chunks)
	p2docs := md.PathToDocuments[path]
	p2docs.Append(memDoc)
    return nil
}

func (md *MemoryDatabase) SaveChunks(path string, payload interface{}, chunks []autog.Chunk) error {
    memDoc := &MemDocument{}
	memDoc.SetPath(path)
	memDoc.SetPayload(payload)
	memDoc.SetChunks(chunks)
	md.PathToDocuments[path] = &MemDocuments{memDoc}
    return nil
}

func (md *MemoryDatabase) SearchChunks(path string, embeds []autog.Embedding, topk int) ([]autog.ScoredChunks, error) {
	norms := Norms(embeds)
	var scoreds  []autog.ScoredChunks
	var dbchunks []autog.Chunk
	var dbembeds []autog.Embedding
	var dberr error
	if path == autog.DOCUMENT_PATH_NONE {
		dbchunks, dbembeds, dberr = md.GetChunks()
	} else {
		dbchunks, dbembeds, dberr = md.GetPathChunks(path)
	}
	if dberr != nil {
		return scoreds, dberr
	}
	dbnorms := Norms(dbembeds)

	qnum  := len(embeds)
	channel := make(chan []autog.ScoredChunks)
	scoreds = make([]autog.ScoredChunks, qnum)

	var wg sync.WaitGroup
	const qbatch = 100
	const dbatch = 1000
	for qi := 0; qi < len(embeds); qi += qbatch {
		for di := 0; di < len(dbembeds); di += dbatch {
			qj := min(qi + qbatch, len(embeds))
			dj := min(di + dbatch, len(dbembeds))

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
				scoreds[idx] = cscores[ci]
			}
		}
		cnt++
	}

	return scoreds, nil
}


