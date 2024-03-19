package autog

type DoReflection struct {
	Do func(reflection string, retry int)
}

func (d *DoReflection) doDo(reflection string, retry int) {
	if d.Do == nil {
		return
	}
	d.Do(reflection, retry)
}
