package autog

import (
	"strings"
	"context"
)

type Agent {
	Prompts []*PromptItem
	Request string
	Context context.Context
	Input  *Input
	Output *Output
	LongHistoryMessages []ChatMessage
	ShortHistoryMessages []ChatMessage
	PendingMessages []ChatMessage
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
	return a
}

func (a *Agent) WaitRequest(cxt context.Context, input *Input) *Agent {
	a.Context = cxt
	a.Input   = input
	a.Request = input.doReadInput()
	return a
}

func (a *Agent) AskLLM(llm LLM, stream bool) *Agent {
	var msgs []ChatMessage
	msg := ChatMessage{ Role:USER, Content:a.Request }
	for pmt := range a.Prompts {
		pms := pmt.doGenMessages(a.Request)
		if len(pms) > 0 {
			msgs = append(msgs, pms)
		}
	}
	a.PendingMessages = msgs
	a.ShortHistoryMessages = append(a.ShortHistoryMessages, msg)
	a.LLM = llm
	a.Stream = stream
	return a
}

func (a *Agent) AskReflection(reflection string) *Agent {
	msg := ChatMessage{ Role:USER, Content:reflection }
	a.PendingMessages = append(a.PendingMessages, msg)
	a.ShortHistoryMessages = append(a.ShortHistoryMessages, msg)
	return a
}

func (a *Agent) WaitResponse(cxt context.Context, output *Output) *Agent {
	a.Context = cxt
	a.Output  = output
	if !a.Stream {
		buf := output.StreamStart()
		sts, msg := a.LLM.SendMessages(cxt, a.PendingMessages)
		a.ResponseStatus  = sts
		a.ResponseMessage = msg
		if sts == LLM_STATUS_OK {
			output.StreamDelta(buf, msg.Content)
		}
		output.StreamEnd(buf)
		a.ShortHistoryMessages = append(a.ShortHistoryMessages, msg)
	} else {
		sts, msg := a.LLM.SendMessagesStream(cxt, a.PendingMessages, output)
		a.ResponseStatus  = sts
		a.ResponseMessage = msg
		a.ShortHistoryMessages = append(a.ShortHistoryMessages, msg)
	}
	a.CanDoAction = a.ResponseStatus == LLM_STATUS_OK
	a.CanDoReflection = false
	return a
}


