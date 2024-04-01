package autog

import (
	"strings"
)

type Input struct {
	ReadContent func() string
}

func (i *Input) doReadContent() string {
	if i.ReadContent == nil {
		return ""
	}
	return i.ReadContent()
}

type Output struct {
	WriteContent func(stage AgentStage, stream StreamStage, buf *strings.Builder, str string)
}

func (o *Output) doWriteContent(stage AgentStage, stream StreamStage, buf *strings.Builder, str string) {
	if o.WriteContent != nil {
		o.WriteContent(stage, stream, buf , str)
	}
}
