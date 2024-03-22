package autog

import (
    "time"
)

const (
    DOCUMENT_PATH_NONE = ""
)

type Embedding []float64


type Database interface {
	AddDocument(doc Document) (path string, error)
	GetDocument(path string) (Document, error)
	DelDocument(path string) error
	GetDocumentPaths() ([]string, error)
	GetDocumentChunk(path string, idx int) (Chunk, error)
	GetDocumentChunks(path string) ([]Chunk, []Embedding, error)
	GetDatabaseChunks() ([]Chunk, []Embedding, error)
}

type ScoredChunk {
	Chunk *Chunk
	Score float64
}

type Chunk struct {
	Index     int       `json:"Index"`
	DocPath   string    `json:"DocPath"`
	Query     string    `json:"Query"`
	Content   string    `json:"Content"`
	ByteStart int       `json:"ByteStart"`
	ByteEnd   int       `json:"ByteEnd"`
	Embedding []float64 `json:"Embedding"`
}

type Document struct {
	Path    string     `json:"Path"`
	Title   string     `json:"Title"`
	Desc    string     `json:"Desc"`
	Content string     `json:"Content"`
	Chunks  []Chunk    `json:"Chunks"`
}

type Splitter interface {
	CreateDocument(path string, title string, desc string, content string) *Document
}

type EmbeddingModel interface {
	Embedding(text string) Embedding
}

type Rag struct {
	Database Database
	Splitter Splitter
	EmbeddingModel EmbeddingModel
	PostRank func (r *Rag, chunks []ScoredChunk) []ScoredChunk
}

func (r *Rag) Indexing(path stribg, title string, desc string, content string) *Document {
	return nil
}

func (r *Rag) Retrieval(queries []string, docPath string, topK int) []ScoredChunk {
	var chunks []Chunk
	return chunks
}

func (r *Rag) doPostRank(chunks []ScoredChunk) []ScoredChunk {
	if r.PostRank == nil {
		return chunks
	}
	return r.PostRank(r, chunks)
}
