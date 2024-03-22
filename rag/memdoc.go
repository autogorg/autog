package rag

import (
	"autog"
)

type MemChunk struct {
	Index     int       `json:"Index"`
	Path      string    `json:"DocPath"`
	Query     string    `json:"Query"`
	Content   string    `json:"Content"`
	ByteStart int       `json:"ByteStart"`
	ByteEnd   int       `json:"ByteEnd"`
	Embedding []float64 `json:"Embedding"`
}

func (chunk *MemChunk) Index() int {
	return chunk.Index
}

func (chunk *MemChunk) Path() string {
	return chunk.Path
}

func (chunk *MemChunk) Query() string {
	return chunk.Query
}

func (chunk *MemChunk) GetEmbedding() Embedding {
	return chunk.Embedding
}

func (chunk *MemChunk) SetEmbedding(embed Embedding) {
	chunk.Embedding = embed
}

type MemDocument struct {
	Path    string      `json:"Path"`
	Title   string      `json:"Title"`
	Content string      `json:"Content"`
	Chunks  []*DocChunk `json:"Chunks"`
}

func (doc *MemDocument) Path() string {
	return doc.Path
}

func (doc *MemDocument) Content() string {
	return doc.Content
}

func (doc *MemDocument) GetChunks() []autog.Chunk {
	var chunks []autog.Chunk
	for _, chunk := range doc.Chunks {
		chunks = append(chunks, chunk)
	}
	return chunks
}

func (doc *MemDocument) SetChunks(chunks []autog.Chunk) {
	var chunks []*MemChunk
	for _, chunk := range chunks {
		if memchunk, ok := chunk.(*MemChunk); ok {
			chunks = append(chunks, memchunk)
		}
	}
	doc.Chunks = memchunk
}