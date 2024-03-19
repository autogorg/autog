package autog


type Action struct {
	Name string
	Desc string
	NeedRun func (content string) (need bool)
	Check func (content string) (ok bool, err string, payload interface{})
	Run func (content string, payload interface{}) (ok bool, err string)
}

func (a *Action) doNeedRun(content string) (need bool) {
	return a.NeedRun(content)
}

func (a *Action) doCheck(content string) (ok bool, err string, payload interface{}) {
	return a.Check(content)
}

func (a *Action) doRun(content string, payload interface{}) (ok bool, err string) {
	return a.Run(content, payload)
}

type DoAction struct {
	Do func(content string) (ok bool, reflection string)
}

func (d *DoAction) doDo(content string) (ok bool, reflection string) {
	return d.Do(content)
}


