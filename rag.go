package autog

import (
	"fmt"
	"sync"
	"context"
	"strings"
)

const (
    DOCUMENT_PATH_NONE         = ""
	defaultEmbeddingBatch      = 1
	defaultEmbeddingDimensions = 0
)

type Embedding []float64

func (e *Embedding) String(d int) string {
	buf := strings.Builder{}
	buf.WriteString("[")
	for _, f := range *e {
		ff := "%f"
		if d > 0 {
			ff = "%." + fmt.Sprintf("%d", d) + "f"
		}
		buf.WriteString(fmt.Sprintf(ff + ", ", f))
	}
	buf.WriteString("]")
	return buf.String()
}

type Database interface {
	AppendChunks(path string, payload interface{}, chunks []Chunk) error
	SaveChunks(path string, payload interface{}, chunks []Chunk) error
	SearchChunks(path string, embeds []Embedding, topk int) ([]ScoredChunks, error)
}

type ScoredChunk struct {
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
	Embeddings(cxt context.Context, dimensions int, texts []string) ([]Embedding, error)
}

type Rag struct {
	Database Database
	EmbeddingModel EmbeddingModel
	EmbeddingBatch int
	EmbeddingDimensions int
	PostRank func (r *Rag, queries []string, chunks []ScoredChunks) ([]ScoredChunks, error)
}

func (r *Rag) Embeddings(cxt context.Context, texts []string) ([]Embedding, error) {
	var mutex sync.Mutex
	var wg sync.WaitGroup
	var err error
	var batch int

	batch = defaultEmbeddingBatch
	if r.EmbeddingBatch > 0 {
		batch = r.EmbeddingBatch
	}

	dimensions := defaultEmbeddingDimensions
	if r.EmbeddingDimensions > 0 {
		dimensions = r.EmbeddingDimensions
	}

	embeds := make([]Embedding, len(texts))
	for i := 0; i < len(texts); i += batch {
		j := i + batch
		if j > len(texts) {
			j = len(texts)
		}
		qtexts := texts[i:j]

		wg.Add(1)
		go func (i, j int, qtexts []string) {
			defer wg.Done()
			es, eerr := r.EmbeddingModel.Embeddings(cxt, dimensions, qtexts)
			if eerr != nil {
				mutex.Lock()
				defer mutex.Unlock()
				err = eerr
				return
			}
			mutex.Lock()
			defer mutex.Unlock()
			for x := i; x < j; x++ {
				embeds[x] = es[x]
			}
		}(i, j, qtexts)
	}

	wg.Wait()

	return embeds, err
}

func (r *Rag) Indexing(cxt context.Context, path string, payload interface{}, splitter Splitter, overwrite bool) error {
	if path == DOCUMENT_PATH_NONE {
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
	if !overwrite {
		serr = r.Database.AppendChunks(path, payload, chunks)
	} else {
		serr = r.Database.SaveChunks(path, payload, chunks)
	}
	
	return serr
}

func (r *Rag) Retrieval(cxt context.Context, path string, queries []string, topk int) ([]ScoredChunks, error) {
	var scoreds []ScoredChunks
	qembeds, err := r.Embeddings(cxt, queries)
	if err != nil {
		return scoreds, err
	}
	scoreds, err = r.Database.SearchChunks(path, qembeds, topk)
	if err != nil {
		return scoreds, err
	}
	return r.doPostRank(queries, scoreds)
}

func (r *Rag) doPostRank(queries []string, chunks []ScoredChunks) ([]ScoredChunks, error) {
	if r.PostRank == nil {
		return chunks, nil
	}
	return r.PostRank(r, queries, chunks)
}
