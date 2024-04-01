package autog

import (
	"fmt"
	"strings"
	"context"
)

type AgentStage int

const (
	AsReadQuestion AgentStage = iota
	AsAskLLM
	AsAskReflection
	AsWaitResponse
	AsAction
	AsReflection
	AsSummarize
)

type StreamStage int

const (
	StreamStageStart StreamStage = iota
	StreamStageDelta
	StreamStageError
	StreamStageEnd
)

type Agent struct {
	Prompts []*PromptItem
	Request string
	Context context.Context
	Input  *Input
	Output *Output
	LongHistoryMessages []ChatMessage
	ShortHistoryMessages []ChatMessage
	PromptMessages []ChatMessage
	LLM LLM
	Stream bool
	ResponseStatus  LLMStatus
	ResponseMessage ChatMessage
	ReflectionContent string
	AgentStage AgentStage
	CanDoAction bool
	DoAction *DoAction
	CanDoReflection bool
	DoReflection *DoReflection
}

func (a *Agent) StreamStart() *strings.Builder {
	contentbuf := &strings.Builder{}
	if a.Output != nil && a.Output.WriteContent != nil{
		a.Output.WriteContent(a.AgentStage, StreamStageStart, contentbuf, "")
	}
	return contentbuf
}

func (a *Agent) StreamDelta(contentbuf *strings.Builder, delta string) {
	if a.Output != nil && a.Output.WriteContent != nil{
		a.Output.WriteContent(a.AgentStage, StreamStageDelta, contentbuf, delta)
	}
}

func (a *Agent) StreamError(contentbuf *strings.Builder, status LLMStatus, errstr string) {
	if a.Output != nil && a.Output.WriteContent != nil{
		a.Output.WriteContent(a.AgentStage, StreamStageError, contentbuf, errstr)
	}
}

func (a *Agent) StreamEnd(contentbuf *strings.Builder) {
	if a.Output != nil && a.Output.WriteContent != nil{
		a.Output.WriteContent(a.AgentStage, StreamStageEnd, contentbuf, "")
	}
}


func (a *Agent) GetLongHistory() []ChatMessage {
	return a.LongHistoryMessages
}

func (a *Agent) GetShortHistory() []ChatMessage {
	return a.ShortHistoryMessages
}

func (a *Agent) Prompt(prompts ...*PromptItem) *Agent {
	a.Prompts = prompts
	a.CanDoAction = false
	a.CanDoReflection = false
	a.ReflectionContent = ""
	return a
}

func (a *Agent) ReadQuestion(cxt context.Context, input *Input, output *Output) *Agent {
	a.AgentStage = AsReadQuestion
	if cxt == nil {
		cxt = context.Background()
	}
	a.Context = cxt
	a.Input   = input
	a.Output  = output
	a.Request = input.doReadContent()
	return a
}

func (a *Agent) AskLLM(llm LLM, stream bool) *Agent {
	a.AgentStage = AsAskLLM
	var msgs []ChatMessage
	msg := ChatMessage{ Role:ROLE_USER, Content:a.Request }
	for _, pmt := range a.Prompts {
		pms := pmt.doGetMessages(a.Request)
		if len(pms) > 0 {
			msgs = append(msgs, pms...)
		}
		role, prompt := pmt.doGetPrompt(a.Request)
		if IsValidRole(role) && len(prompt) > 0 {
			m := ChatMessage{ Role:role, Content:prompt }
			msgs = append(msgs, m)
		}
	}
	msgs = append(msgs, msg)
	a.PromptMessages = msgs
	a.ShortHistoryMessages = append(a.ShortHistoryMessages, msg)
	a.LLM = llm
	a.Stream = stream
	return a
}

func (a *Agent) AskReflection(reflection string) *Agent {
	a.AgentStage = AsAskReflection
	var contentbuf *strings.Builder
	contentbuf = a.StreamStart()
	a.StreamDelta(contentbuf, reflection)
	a.StreamEnd(contentbuf)
	msg := ChatMessage{ Role:ROLE_USER, Content:reflection }
	a.PromptMessages = append(a.PromptMessages, msg)
	a.ShortHistoryMessages = append(a.ShortHistoryMessages, msg)
	return a
}

func (a *Agent) WaitResponse(cxt context.Context) *Agent {
	a.AgentStage = AsWaitResponse
	if cxt == nil {
		cxt = context.Background()
	}
	a.Context = cxt
	var sts LLMStatus
	var msg ChatMessage
	var contentbuf *strings.Builder
	if !a.Stream {
		contentbuf = a.StreamStart()
		sts, msg = a.LLM.SendMessages(cxt, a.PromptMessages)
		if sts == LLM_STATUS_OK {
			a.StreamDelta(contentbuf, msg.Content)
		} else {
			a.StreamError(contentbuf, sts, msg.Content)
		}
		a.StreamEnd(contentbuf)
	} else {
		sts, msg = a.LLM.SendMessagesStream(cxt, a.PromptMessages, a)
	}
	a.ResponseStatus  = sts
	a.ResponseMessage = msg
	a.ShortHistoryMessages = append(a.ShortHistoryMessages, a.ResponseMessage)
	a.CanDoAction = a.ResponseStatus == LLM_STATUS_OK
	a.CanDoReflection = false
	return a
}

func (a *Agent) Summarize(cxt context.Context, summary *PromptItem, prefix *PromptItem, force bool) *Agent {
	a.AgentStage = AsSummarize
	if cxt == nil {
		cxt = context.Background()
	}
	a.Context = cxt
	var contentbuf *strings.Builder
	contentbuf = a.StreamStart()
	smy := &Summary{}
	smy.Cxt = cxt
	smy.LLM = a.LLM
	smy.StreamReader = a
	smy.StreamBuffer = contentbuf
	smy.DisableStream = false
	_, smy.PromptSummary = summary.doGetPrompt(a.Request)
	_, smy.PromptPrefix  = prefix.doGetPrompt(a.Request)

	err := smy.InitSummary()
	if err != nil {
		a.StreamError(contentbuf, LLM_STATUS_BED_MESSAGE, fmt.Sprintf("InitSummary ERROR: %s", err))
		a.StreamEnd(contentbuf)
		return a
	}
	status, smsgs := smy.Summarize(a.LongHistoryMessages, a.ShortHistoryMessages, force)
	if status != LLM_STATUS_OK {
		a.StreamEnd(contentbuf)
		return a
	}
	a.StreamEnd(contentbuf)

	a.LongHistoryMessages  = smsgs
	a.ShortHistoryMessages = []ChatMessage{}

	return a
}

func (a *Agent) Action(doAct *DoAction) *Agent {
	a.AgentStage = AsAction
	a.DoAction = doAct
	if !a.CanDoAction {
		return a
	}
	a.CanDoAction = false
	a.CanDoReflection = false
	a.ReflectionContent = ""
	if a.DoAction == nil {
		return a
	}
	ok, react := a.DoAction.doDo(a.ResponseMessage.Content)
	a.ReflectionContent = react
	a.CanDoReflection = !ok
	return a
}

func (a *Agent) Reflection(doRef *DoReflection, retry int) *Agent {
	a.AgentStage = AsReflection
	if doRef == nil {
		doRef = &DoReflection {
			Do : func (reflection string, retry int) {
				a.AskReflection(reflection)
				a.WaitResponse(a.Context)
				a.Action(a.DoAction)
				a.Reflection(doRef, retry)
			},
		}
	}
	a.DoReflection = doRef
	if !a.CanDoReflection {
		return a
	}
	react := a.ReflectionContent
	a.CanDoAction = false
	a.CanDoReflection = false
	a.ReflectionContent = ""
	retry -= 1
	if retry <= 0 {
		return a
	}
	if len(react) > 0 {
		a.DoReflection.doDo(react, retry)
	}
	return a
}
