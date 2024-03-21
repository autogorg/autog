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

func (md *MemoryDatabase) GetDocumentChunks(path string) ([]Chunk, error) {
	var chunks []Chunk
	doc, ok := md.PathToDocument[path];
	if !ok {
		return chunks, fmt.Errorf(ErrDocNotExists)
	}
	return doc.Chunks, nil
}

func (md *MemoryDatabase) GetDocumentEmbeddings(path string) ([]Embedding, error) {
	var embeddings []Embedding
	doc, ok := md.PathToDocument[path];
	if !ok {
		return embeddings, fmt.Errorf(ErrDocNotExists)
	}
	for _, chunk := range doc.Chunks {
		embeddings = append(embeddings, chunk.Embedding)
	}
	return embeddings, nil
}
