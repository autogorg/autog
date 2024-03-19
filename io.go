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
	WriteStreamStart func() *strings.Builder
	WriteStreamDelta func(contentbuf *strings.Builder, delta string)
	WriteStreamEnd func(contentbuf *strings.Builder)
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

func (o *Output) StreamEnd(contentbuf *strings.Builder) {
	if o.WriteStreamEnd == nil {
		return
	}
	o.WriteStreamEnd(contentbuf)
}
