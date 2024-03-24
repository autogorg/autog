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
	AppendChunks(path string, payload interface{}, chunks []Chunk]) error
	SaveChunks(path string, payload interface{}, chunks []Chunk]) error
	SearchChunks(path string, embeds []Embedding, topk int) ([]ScoredChunks, error)
}

type ScoredChunk {
	Chunk Chunk
	Score float64
}

type ScoredChunks []*ScoredChunk

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

type ParserFunction func (path string, payload interface{}) ([]Chunk, error)

type Splitter interface {
	GetParser() ParserFunction
}

type EmbeddingModel interface {
	Embedding(cxt context.Context, text string) (Embedding, error)
}

type Rag struct {
	Database Database
	EmbeddingModel EmbeddingModel
	PostRank func (r *Rag, queries []string, chunks []ScoredChunks) ([]ScoredChunks, error)
}

func (r *Rag) Embeddings(cxt context.Context, texts []string) ([]Embedding, error) {
	var mutex sync.Mutex
	var wg sync.WaitGroup
	var err error

	embeds := make([]Embedding, len(texts))
	for i, text := range texts {
		wg.Add(1)
		go func (i int, text string) {
			defer wg.Done()
			embed, eerr := r.EmbeddingModel.Embedding(cxt, text)
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

func (r *Rag) Indexing(cxt context.Context, path string, payload interface{}, splitter Splitter, append bool) error {
	if doc.GetPath() == DOCUMENT_PATH_NONE {
		return fmt.Errorf("Document path is empty!")
	}
	parser := splitter.GetParser()
	chunks, cerr := parser(path, payload)
	if cerr != nil {
		return cerr
	}

	var qs []string

	for _, chunk := range chunks {
		qs = append(qs, chunk.GetQuery())
	}

	embeds, eerr := r.Embeddings(cxt, qs)
	if eerr != nil {
		return eerr
	}
	if len(embeds) != len(chunks) {
		return fmt.Errorf("Embedding Error!")
	}

	for i, chunk := range chunks {
		chunk.SetEmbedding(embeds[i])
	}

	var serr error
	if append {
		serr = r.Database.AppendChunks(path, payload, chunks)
	} else {
		serr = r.Database.SaveChunks(path, payload, chunks)
	}
	
	return serr
}

func (r *Rag) Retrieval(cxt context.Context, queries []string, path string, topk int) ([]ScoredChunks, error) {
	var scoreds []ScoredChunks
	qembeds, berr := r.Embeddings(cxt, queries)
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
