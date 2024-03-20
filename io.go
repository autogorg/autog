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
	WriteAgentStage  func(stage AgentStage)
	WriteStreamStart func() *strings.Builder
	WriteStreamDelta func(contentbuf *strings.Builder, delta string)
	WriteStreamError func(contentbuf *strings.Builder, status LLMStatus, errstr string)
	WriteStreamEnd   func(contentbuf *strings.Builder)
}

func (o *Output) AgentStage(stage AgentStage) {
	if o.WriteAgentStage == nil {
		return
	}
	o.WriteAgentStage(stage)
}

func (o *Output) StreamStart() *strings.Builder {
	if o.WriteStreamStart == nil {
		return nil
	}
	return o.WriteStreamStart()
}

func (o *Output) StreamDelta(contentbuf *strings.Builder, delta string) {
	if o.WriteStreamDelta == nil {
		return
	}
	o.WriteStreamDelta(contentbuf, delta)
}

func (o *Output) StreamError(contentbuf *strings.Builder, status LLMStatus, errstr string) {
	if o.WriteStreamError == nil {
		return
	}
	o.WriteStreamError(contentbuf, status, errstr)
}

func (o *Output) StreamEnd(contentbuf *strings.Builder) {
	if o.WriteStreamEnd == nil {
		return
	}
	o.WriteStreamEnd(contentbuf)
}