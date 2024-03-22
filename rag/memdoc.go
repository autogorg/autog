package rag

import (
	"autog"
)

type MemChunk struct {
	Index     int       `json:"Index"`
	Path      string    `json:"Path"`
	Query     string    `json:"Query"`
	Content   string    `json:"Content"`
	LineStart int       `json:"LineStart"`
	LineEnd   int       `json:"LineEnd"`
	ByteStart int       `json:"ByteStart"`
	ByteEnd   int       `json:"ByteEnd"`
	MetaData  string    `json:"MetaData"`
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

func (chunk *MemChunk) GetLineStart() int {
	return chunk.LineStart
}

func (chunk *MemChunk) SetLineStart(i int) {
	chunk.LineStart = i
}

func (chunk *MemChunk) GetLineEnd() int {
	return chunk.LineEnd
}

func (chunk *MemChunk) SetLineEnd(i int) {
	chunk.LineEnd = i
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

func (chunk *MemChunk) GetMetaData() string {
	return chunk.MetaData
}

func (chunk *MemChunk) SetMetaData(data string) {
	chunk.MetaData = data
}

func (chunk *MemChunk) GetEmbedding() Embedding {
	return chunk.Embedding
}

func (chunk *MemChunk) SetEmbedding(embed Embedding) {
	chunk.Embedding = embed
}

type MemDocument struct {
	Path     string      `json:"Path"`
	Title    string      `json:"Title"`
	Content  string      `json:"Content"`
	MetaData string      `json:"MetaData"`
	Chunks   []*DocChunk `json:"Chunks"`
}

func (doc *MemDocument) GetPath() string {
	return doc.Path
}

func (doc *MemDocument) SetPath(path string) {
	doc.Path = path
}

func (doc *MemDocument) GetTitle() string {
	return doc.Title
}

func (doc *MemDocument) SetTitle(title string) {
	doc.Title = title
}

func (doc *MemDocument) GetContent() string {
	return doc.Content
}

func (doc *MemDocument) SetContent(content string) {
	doc.Content = content
}

func (doc *MemDocument) GetMetaData() string {
	return doc.MetaData
}

func (doc *MemDocument) SetMetaData(data string) {
	doc.MetaData = data
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