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
	CanDoAction bool
	DoAction *DoAction
	CanDoReflection bool
	DoReflection *DoReflection
}

func OutputChangeAgentStage(output *Output, stage AgentStage) {
	if output != nil {
		output.ChangeAgentStage(stage)
	}
}

func OutputStreamStart(output *Output) *strings.Builder {
	if output != nil {
		return output.StreamStart()
	}
	return &strings.Builder{}
}

func OutputStreamDelta(output *Output, contentbuf *strings.Builder, delta string) {
	if contentbuf != nil {
		contentbuf.WriteString(delta)
	}
	if output != nil {
		output.StreamDelta(contentbuf, delta)
	}
}

func OutputStreamError(output *Output, contentbuf *strings.Builder, status LLMStatus, errstr string) {
	if output != nil {
		output.StreamError(contentbuf, status, errstr)
	}
}

func OutputStreamEnd(output *Output, contentbuf *strings.Builder) {
	if output != nil {
		output.StreamEnd(contentbuf)
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
	OutputChangeAgentStage(output, AsReadQuestion)
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
	OutputChangeAgentStage(a.Output, AsAskLLM)
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
	a.PromptMessages = msgs
	a.ShortHistoryMessages = append(a.ShortHistoryMessages, msg)
	a.LLM = llm
	a.Stream = stream
	return a
}

func (a *Agent) AskReflection(reflection string) *Agent {
	OutputChangeAgentStage(a.Output, AsAskReflection)
	var contentbuf *strings.Builder
	contentbuf = OutputStreamStart(a.Output)
	OutputStreamDelta(a.Output, contentbuf, reflection)
	OutputStreamEnd(a.Output, contentbuf)
	msg := ChatMessage{ Role:ROLE_USER, Content:reflection }
	a.PromptMessages = append(a.PromptMessages, msg)
	a.ShortHistoryMessages = append(a.ShortHistoryMessages, msg)
	return a
}

func (a *Agent) WaitResponse(cxt context.Context) *Agent {
	OutputChangeAgentStage(a.Output, AsWaitResponse)
	if cxt == nil {
		cxt = context.Background()
	}
	a.Context = cxt
	var sts LLMStatus
	var msg ChatMessage
	var contentbuf *strings.Builder
	if !a.Stream {
		contentbuf = OutputStreamStart(a.Output)
		sts, msg = a.LLM.SendMessages(cxt, a.PromptMessages)
		if sts == LLM_STATUS_OK {
			OutputStreamDelta(a.Output, contentbuf, msg.Content)
		} else {
			OutputStreamError(a.Output, contentbuf, sts, msg.Content)
		}
		OutputStreamEnd(a.Output, contentbuf)
	} else {
		sts, msg = a.LLM.SendMessagesStream(cxt, a.PromptMessages, a.Output)
	}
	a.ResponseStatus  = sts
	a.ResponseMessage = msg
	a.ShortHistoryMessages = append(a.ShortHistoryMessages, a.ResponseMessage)
	a.CanDoAction = a.ResponseStatus == LLM_STATUS_OK
	a.CanDoReflection = false
	return a
}

func (a *Agent) Summarize(cxt context.Context, summary *PromptItem, prefix *PromptItem, force bool) *Agent {
	OutputChangeAgentStage(a.Output, AsSummarize)
	if cxt == nil {
		cxt = context.Background()
	}
	a.Context = cxt
	var contentbuf *strings.Builder
	contentbuf = OutputStreamStart(a.Output)
	smy := &Summary{}
	smy.Cxt = cxt
	smy.LLM = a.LLM
	smy.Output    = a.Output
	smy.OutputBuf = contentbuf
	smy.DisableStream = false
	_, smy.PromptSummary = summary.doGetPrompt(a.Request)
	_, smy.PromptPrefix  = prefix.doGetPrompt(a.Request)

	err := smy.InitSummary()
	if err != nil {
		OutputStreamError(a.Output, contentbuf, LLM_STATUS_BED_MESSAGE, fmt.Sprintf("InitSummary ERROR: %s", err))
		OutputStreamEnd(a.Output, contentbuf)
		return a
	}
	status, smsgs := smy.Summarize(a.LongHistoryMessages, a.ShortHistoryMessages, force)
	if status != LLM_STATUS_OK {
		OutputStreamEnd(a.Output, contentbuf)
		return a
	}
	OutputStreamEnd(a.Output, contentbuf)

	a.LongHistoryMessages  = smsgs
	a.ShortHistoryMessages = []ChatMessage{}

	return a
}

func (a *Agent) Action(doAct *DoAction) *Agent {
	OutputChangeAgentStage(a.Output, AsAction)
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
	OutputChangeAgentStage(a.Output, AsReflection)
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
