package autog

import (
	"strings"
)

type Input {
	ReadContent func() string
}

func (i *Input) func doReadContent() string {
	return i.ReadContent()
}

type Output {
	WriteStreamStart func() *strings.Builder
	WriteStreamDelta func(contentbuf *strings.Builder, delta string)
	WriteStreamEnd func(contentbuf *strings.Builder)
}

func (o *Output) func StreamStart() *strings.Builder {
	if o.WriteStreamStart == nil {
		return nil
	}
	return o.WriteStreamStart()
}

func (o *Output) func StreamDelta(contentbuf *strings.Builder, delta string) {
	if o.WriteStreamDelta == nil {
		return
	}
	return o.WriteStreamDelta(contentbuf, delta)
}

func (o *Output) func StreamEnd(contentbuf *strings.Builder) {
	if o.WriteStreamEnd == nil {
		return
	}
	return o.WriteStreamEnd(contentbuf)
}
