package autog

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