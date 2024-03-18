package autog

type DoReflection {
	Do func(reflection string, retry int)
}

func (d *DoReflection) doDo(reflection string, retry int) {
	if d.Do == nil {
		return
	}
	return d.Do(reflection, retry)
}
