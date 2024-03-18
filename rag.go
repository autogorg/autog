package autog

type Rag {
	Index func()
	Retrieve func(query string) 
}

func (r *Rag) doIndex() {
	return r.Index()
}

func (r *Rag) doRetrieve(query string) {
	return r.Retrieve(query)
}