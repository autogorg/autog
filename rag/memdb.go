package rag

import (
	"fmt"
	"log"
	"os"
	"bytes"
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


func CosSim(qembeds, dbembeds [][]float64, qnorms, dbnorms *[]float64, qi, di int, k int, channel chan<- []ScoredChunks) {

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
				CosSim(embeds[qi:qj], dbembeds[di:dj], &norms, &dbnorms, qi, di, topk, channel)
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


