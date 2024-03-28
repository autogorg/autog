package autog

import (
	"fmt"
	"sync"
	"context"
	"strings"
)

const (
    DOCUMENT_PATH_NONE         = ""
	defaultEmbeddingBatch      = 2
	defaultEmbeddingDimensions = 0
	defaultEmbeddingRoutines   = 5
)

type EmbeddingStage int

const (
	EmbeddingStageIndexing  EmbeddingStage = iota
	EmbeddingStageRetrieval
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
	EmbeddingRoutines int
	EmbeddingDimensions int
	EmbeddingCallback  func (stage EmbeddingStage, texts []string, embeds []Embedding, i, j int, finished, tried int, err error) bool
}

func (r *Rag) Embeddings(cxt context.Context, stage EmbeddingStage, texts []string) ([]Embedding, error) {
	var mutex sync.Mutex
	var wg sync.WaitGroup
	var err error
	var batch int

	batch = defaultEmbeddingBatch
	if r.EmbeddingBatch > 0 {
		batch = r.EmbeddingBatch
	}

	routines := defaultEmbeddingRoutines
	if r.EmbeddingRoutines > 0 {
		routines = r.EmbeddingRoutines
	}

	dimensions := defaultEmbeddingDimensions
	if r.EmbeddingDimensions > 0 {
		dimensions = r.EmbeddingDimensions
	}

	finished := 0
	// Create slots
	concurrents := make(chan struct{}, routines)
	embeds   := make([]Embedding, len(texts))
	for i := 0; i < len(texts); i += batch {
		j := i + batch
		if j > len(texts) {
			j = len(texts)
		}
		qtexts := texts[i:j]

		// Waiting for free slot
		concurrents <- struct{}{}

		wg.Add(1)
		go func (i, j int, qtexts []string) {
			defer wg.Done()
			defer func() {
				// Free a slot
				<-concurrents
			}()

			tried := 0
			retry := false
			for {
				es, eerr := r.EmbeddingModel.Embeddings(cxt, dimensions, qtexts)
				mutex.Lock()
				if eerr != nil {
					err = eerr
				} else {
					for x := i; x < j; x++ {
						embeds[x] = es[x - i]
					}
					finished = finished + (j - i)
				}
				retry = r.doEmbeddingCallback(stage, texts, embeds, i, j, finished, tried, eerr)
				if retry && eerr == nil {
					finished = finished - (j - i)
				}
				mutex.Unlock()
				if retry {
					tried++
					continue
				}
				break
			}
		}(i, j, qtexts)
	}

	wg.Wait()

	return embeds, err
}

func (r *Rag) doEmbeddingCallback(stage EmbeddingStage, texts []string, embeds []Embedding, i, j int, finished, tried int, err error) bool {
	if r.EmbeddingCallback == nil {
		return false
	}
	return r.EmbeddingCallback(stage, texts, embeds, i, j, finished, tried, err)
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

	embeds, eerr := r.Embeddings(cxt, EmbeddingStageIndexing, qs)
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
	qembeds, err := r.Embeddings(cxt, EmbeddingStageRetrieval, queries)
	if err != nil {
		return scoreds, err
	}
	return r.Database.SearchChunks(path, qembeds, topk)
}
