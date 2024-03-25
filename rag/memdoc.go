package rag

import (
	"github.com/autogorg/autog"
)

type MemChunk struct {
	Index     int       `json:"Index"`
	Path      string    `json:"Path"`
	Query     string    `json:"Query"`
	Content   string    `json:"Content"`
	ByteStart int       `json:"ByteStart"`
	ByteEnd   int       `json:"ByteEnd"`
	Payload   string    `json:"Payload"`
	Embedding []float64 `json:"Embedding"`
}

func (chunk *MemChunk) GetIndex() int {
	return chunk.Index
}

func (chunk *MemChunk) SetIndex(index int) {
	chunk.Index = index
}

func (chunk *MemChunk) GetPath() string {
	return chunk.Path
}

func (chunk *MemChunk) SetPath(path string) {
	chunk.Path = path
}

func (chunk *MemChunk) GetQuery() string {
	return chunk.Query
}

func (chunk *MemChunk) SetQuery(query string) {
	chunk.Query = query
}

func (chunk *MemChunk) GetContent() string {
	return chunk.Content
}

func (chunk *MemChunk) SetContent(content string) {
	chunk.Content = content
}

func (chunk *MemChunk) GetByteStart() int {
	return chunk.ByteStart
}

func (chunk *MemChunk) SetByteStart(i int) {
	chunk.ByteStart = i
}

func (chunk *MemChunk) GetByteEnd() int {
	return chunk.ByteEnd
}

func (chunk *MemChunk) SetByteEnd(i int) {
	chunk.ByteEnd = i
}

func (chunk *MemChunk) GetPayload() interface{} {
	return chunk.Payload
}

func (chunk *MemChunk) SetPayload(payload interface{}) {
	str, ok := payload.(string)
	if ok {
		chunk.Payload = str
	}
}

func (chunk *MemChunk) GetEmbedding() autog.Embedding {
	return chunk.Embedding
}

func (chunk *MemChunk) SetEmbedding(embed autog.Embedding) {
	chunk.Embedding = embed
}

type MemDocument struct {
	Path     string      `json:"Path"`
	Payload  string      `json:"Payload"`
	Chunks   []*MemChunk `json:"Chunks"`
}

func (doc *MemDocument) GetPath() string {
	return doc.Path
}

func (doc *MemDocument) SetPath(path string) {
	doc.Path = path
}

func (doc *MemDocument) GetPayload() interface{} {
	return doc.Payload
}

func (doc *MemDocument) SetPayload(payload interface{}) {
	str, ok := payload.(string)
	if ok {
		doc.Payload = str
	}
}

func (doc *MemDocument) GetChunks() []autog.Chunk {
	var chunks []autog.Chunk
	for _, chunk := range doc.Chunks {
		chunks = append(chunks, chunk)
	}
	return chunks
}

func (doc *MemDocument) SetChunks(chunks []autog.Chunk) {
	var memchunks []*MemChunk
	for _, chunk := range chunks {
		if memchunk, ok := chunk.(*MemChunk); ok {
			memchunks = append(memchunks, memchunk)
		}
	}
	doc.Chunks = memchunks
}