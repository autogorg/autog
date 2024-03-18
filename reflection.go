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

func (a *Agent) DoReflection(doRef *DoReflection, retry int) *Agent {
	if !a.CanDoReflection {
		return a
	}
	a.DoReflection = doRef
	retry -= 1
	if retry <= 0 {
		return a
	}
	if doRef == nil {
		defRef := &DoReflection {
			Do : func (reflection string, retry int) {
				a.AskReflection(reflection)
				a.WaitResponse(nil, a.Output)
				a.DoAction()
				a.DoReflection(defRef)
			}
		}
		defRef.doDo(a.ReflectionContent, retry)
	} else {
		doRef.doDo(a.ReflectionContent, retry)
	}
	return a
}
