package autog

import (
  "math"
  "time"
  "sync"
  "context"
)

const (
    DOCUMENT_PATH_NONE = ""
)

type Embedding []float64

type Database interface {
	SaveDocument(path string, payload interface{}, document Document) error
	SearchChunks(path string, embeds []Embedding, topk int) ([]ScoredChunks, error)
}

type ScoredChunk {
	Chunk Chunk
	Score float64
}

type ScoredChunks []ScoredChunk

type Chunk interface {
	GetIndex() int
	SetIndex(index int)
	GetPath() string
	SetPath(path string)
	GetQuery() string
	SetQuery(query string)
	GetLineStart() int
	SetLineStart(i int)
	GetLineEnd() int
	SetLineEnd(i int)
	GetByteStart() int
	SetByteStart(i int)
	GetByteEnd() int
	SetByteEnd(i int)
	GetContent() string
	SetContent(content string)
	GetPayload() interface{}
	SetPayload(payload interface{})
	GetEmbedding() Embedding
	SetEmbedding(embed Embedding)
}

type Document interface {
	GetPath() string
	SetPath(path string)
	GetPayload() interface{}
	SetPayload(payload interface{})
	GetChunks() []Chunk
	SetChunks(chunks []Chunk)
}

type ParserFunction func (path string, payload interface{}) (Document, error)

type Splitter interface {
	GetParser() ParserFunction
}

type EmbeddingModel interface {
	Embedding(text string) (Embedding, error)
}

type Rag struct {
	Database Database
	EmbeddingModel EmbeddingModel
	PostRank func (r *Rag, queries []string, chunks []ScoredChunks) ([]ScoredChunks, error)
}


func (r *Rag) Embeddings(texts []string) ([]Embedding, error) {
	var mutex sync.Mutex
	var wg sync.WaitGroup
	var err error

	embeds := make([]Embedding, len(texts))
	for i, text := range texts {
		wg.Add(1)
		go func (i int, text string) {
			defer wg.Done()
			embed, eerr := r.EmbeddingModel.Embedding(text)
			if eerr != nil {
				mutex.Lock()
				defer mutex.Unlock()
				err = eerr
				return
			}
			mutex.Lock()
			defer mutex.Unlock()
			embeds[i] = embed
		}(i, text)
	}

	wg.Wait()

	return embeds, err
}

func (r *Rag) Indexing(path string, payload interface{}, splitter Splitter) error {
	if doc.GetPath() == DOCUMENT_PATH_NONE {
		return fmt.Errorf("Document path is empty!")
	}
	parser := splitter.GetParser()
	doc, derr := parser(path, payload)
	if derr != nil {
		return derr
	}

	var qs []string

	for i, chunk := range doc.GetChunks() {
		qs = append(qs, chunk.Query())
	}

	embeds, eerr := r.Embeddings(qs)
	if eerr != nil {
		return eerr
	}
	if len(embeds) != len(doc.GetChunks()) {
		return fmt.Errorf("Embedding Error!")
	}

	for i, chunk := range doc.GetChunks() {
		chunk.SetEmbedding(embeds[i])
	}

	serr = r.Database.SaveDocument(path, payload, doc)
	return serr
}

func (r *Rag) Retrieval(queries []string, path string, topk int) ([]ScoredChunks, error) {
	var scoreds []ScoredChunks
	qembeds, berr := r.Embeddings(queries)
	if berr != nil {
		return scoreds, berr
	}
	return r.Database.SearchChunks(path, qemneds, topk)
}

func (r *Rag) doPostRank(queries []string, chunks []ScoredChunk) ([]ScoredChunks, error) {
	if r.PostRank == nil {
		return chunks, nil
	}
	return r.PostRank(r, queries, chunks)
}
