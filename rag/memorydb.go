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
	PathToDocument map[string]autog.Document
}


func generateRamdomPath() string {
    currentTime := time.Now()
    timeString := currentTime.Format(time.RFC3339)
	return fmt.Sprintf("md5-%x", md5.Sum([]byte(timeString)))
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


func (md *MemoryDatabase) AddDocument(doc Document) (path string, error) {
	if len(doc.Path) <= 0 {
		doc.Path = generateRamdomPath()
	}
	if _, ok := md.PathToDocument[doc.Path]; ok {
		return doc.Path, fmt.Errorf(ErrDocAreadyExists)
	}
	md.PathToDocument[doc.Path] = doc
	return doc.Path, nil
}

func (md *MemoryDatabase) GetDocument(path string) (Document, error) {
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

func (md *MemoryDatabase) GetDocumentChunk(path string, idx int) (Chunk, error) {
	var chunk Chunk
	doc, ok := md.PathToDocument[path];
	if !ok {
		return chunk, fmt.Errorf(ErrDocNotExists)
	}
	if len(doc.Chunks) <= idx {
		return chunk, fmt.Errorf(ErrChunkNotExists)
	}
	return doc.Chunks[idx], nil
}

func (md *MemoryDatabase) GetDocumentChunks(path string) ([]Chunk, []Embedding, error) {
	var chunks []Chunk
	var embeddings []Embedding
	doc, ok := md.PathToDocument[path];
	if !ok {
		return chunks, embeddings, fmt.Errorf(ErrDocNotExists)
	}
	for _, chunk := range doc.Chunks {
		embeddings = append(embeddings, chunk.Embedding)
	}
	return doc.Chunks, embeddings, nil
}

func (md *MemoryDatabase) GetDatabaseChunks() ([]Chunk, []Embedding, error) {
	var chunks []Chunk
	var embeddings []Embedding
	for path, doc := range md {
		for _, chunk := range doc.Chunks {
			chunks = append(chunks, chunk)
			embeddings = append(embeddings, chunk.Embedding)
		}
	}

	return chunks, embeddings, nil
}

func (md *MemoryDatabase) AddChunks(path string, []Chunk) error {
}

func (md *MemoryDatabase) SearchChunks(path string, embeds []Embedding) {
}


