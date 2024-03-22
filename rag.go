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
	AddDocument(doc Document) (path string, error)
	GetDocument(path string) (Document, error)
	DelDocument(path string) error
	GetDocumentPaths() ([]string, error)
	GetDocumentChunk(path string, idx int) (Chunk, error)
	GetDocumentChunks(path string) ([]Chunk, []Embedding, error)
	GetDatabaseChunks() ([]Chunk, []Embedding, error)
}

type Database interface {
	AddChunks(path string, []Chunk) error
 SearchChunks(path string, embeds []Embedding) ([]ScoredChunks, error)
}

type ScoredChunk {
	Chunk *Chunk
	Score float64
}

type ScoredChunks []ScoredChunk

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
	Chunks  []Chunk    `json:"Chunks"`
}

type Splitter interface {
	CreateDocument(path string, title string, desc string, content string) (*Document, error)
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

func (r *Rag) Indexing(path stribg, title string, desc string, content string) (*Document, error) {
	doc, derr := r.Splitter.CreateDocument(path, title, desc, content)
	if derr != nil {
		return doc, derr
	}

	var qs []string

	for i, _ := range doc.Chunks {
		qs = append(qs, doc.Chunks[i].Query)
	}

	embeds, err := r.Embeddings(qs)
	if err != nil {
		return doc, err
	}
	if len(embeds) != len(doc.Chunks) {
		return doc, fmt.Errorf("Embedding Error!")
	}

	for i, _ := range doc.Chunks {
		doc.Chunks[i].Embedding = embeds[i]
	}

	return doc, err
}

func (r *Rag) Retrieval(queries []string, docPath string, topk int) ([]ScoredChunks, error) {
	var scoreds []ScoredChunks
	qembeds, berr := r.Embeddings(queries)
	if berr != nil {
		return scoreds, berr
	}
	qnorms := Norms(qembeds)

	var dbchunks []Chunk
	var dbembeds []Embedding
	var dberr error
	if docPath == DOCUMENT_PATH_NONE {
		dbchunks, dbembeds, dberr = r.Database.GetDatabaseChunks()
	} else {
		dbchunks, dbembeds, dberr = r.Database.GetDocumentChunks(docPath)
	}
	if dberr != nil {
		return scoreds, dberr
	}
	dbnorms := Norms(dbembeds)

	qnum  := len(queries)
	channel := make(chan []ScoredChunks)
	scoreds = make([]ScoredChunks, qnum)

	var wg sync.WaitGroup
	const qbatch = 100
	const dbatch = 1000
	for qi := 0; qi < len(qembeds); qi += qbatch {
		for dbi := 0; dbi < len(dbembeds); dbi += dbatch {
			qj := min(qi + qbatch, len(qembeds))
			dj := min(di + dbatch, len(dembeds))

			wg.Add(1)
			go func(qi int, di int, qj int, dj int) {
				wg.Done()
				CosSim(qembeds[qi:qj], dbembeds[di:dj], &qnorms, &dbnorms, qi, di, topk, channel)
			}(qi, di, qj, dj)
		}
	}

	go func() {
		wg.Wait()
		close(channel)
	}()

	cnt := 0
	for cscores := range channel {
		for ci := range cscores {
			idx := ci + cnt * qbatch
			if idx < len(scoreds) {
				scoreds[idx] = cscores[qi]
			}
		}
		cnt++
	}

	return scoreds, nil
}

func (r *Rag) doPostRank(queries []string, chunks []ScoredChunk) ([]ScoredChunks, error) {
	if r.PostRank == nil {
		return chunks, nil
	}
	return r.PostRank(r, queries, chunks)
}
