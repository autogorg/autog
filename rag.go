package autog

type Embedding []float64

type LocalDatabase interface {
	AddDocument(doc Document) (path string)
	GetDocument(path string)
	DelDocument(path string)
	GetDocumentChunk(path string, idx int) Chunk
	GetDocumentChunks(path string) []Chunk
	GetDocumentEmbeddings(path string) []Embedding
}

type Chunk struct {
	Index     int       `json:"Index"`
	Query     string    `json:"Query"`
	Text      string    `json:"Text"`
	ByteStart int       `json:"ByteStart"`
	ByteEnd   int       `json:"ByteEnd"`
	Embedding []float64 `json:"Embedding"`
}

type Document struct {
	Path    string  `json:"Path"`
	Title   string  `json:"Title"`
	Desc    string  `json:"Desc"`
	Text    string  `json:"Text"`
	Chunks  []Chunk `json:"Chunks"`
}

type Rag struct {
	Index func()
	Retrieve func(query string) 
}

func (r *Rag) doIndex() {
	if r.Index == nil {
		return
	}
	r.Index()
}

func (r *Rag) doRetrieve(query string) {
	if r.Retrieve == nil {
		return
	}
	r.Retrieve(query)
}