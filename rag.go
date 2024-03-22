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
	AddChunks(path string, []Chunk) error
	SearchChunks(path string, embeds []Embedding, topk int) ([]ScoredChunks, error)
}

type ScoredChunk {
	Chunk *Chunk
	Score float64
}

type ScoredChunks []ScoredChunk

type Chunk struct {
	Index     int       `json:"Index"`
	Path      string    `json:"DocPath"`
	Query     string    `json:"Query"`
	Content   string    `json:"Content"`
	ByteStart int       `json:"ByteStart"`
	ByteEnd   int       `json:"ByteEnd"`
	Embedding []float64 `json:"Embedding"`
}

type Splitter interface {
	CreateChunks(path string, content string) ([]Chunk, error)
}

type EmbeddingModel interface {
	Embedding(text string) (Embedding, error)
}

type Rag struct {
	Database Database
	Splitter Splitter
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

func (r *Rag) Indexing(path stribg, content string) ([]Chunk, error) {
	chunks, cerr := r.Splitter.CreateChunks(path, content)
	if cerr != nil {
		return chunks, cerr
	}

	var qs []string

	for i, _ := range chunks {
		qs = append(qs, chunks[i].Query)
	}

	embeds, err := r.Embeddings(qs)
	if err != nil {
		return chunks, err
	}
	if len(embeds) != len(chunks) {
		return doc, fmt.Errorf("Embedding Error!")
	}

	for i, _ := range chunks {
		chunks[i].Embedding = embeds[i]
	}

	err = r.Database.AddChunks(path, chunks)
	return chunks, err
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
