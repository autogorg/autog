package autog

import (
	"strings"
	"context"
	"encoding/json"
)

var (
	defaultMinSummaryTokens int = 1024
	defaultMinSplit int = 4
	defaultMaxDepth int = 3
)

type TokenizedMessage struct {
	Tokens  int
}

type Summary struct {
	Cxt context.Context
	StreamReader StreamReader
	StreamBuffer *strings.Builder
	LLM LLM
	PromptSummary string
	PromptPrefix  string
	DisableStream    bool
	MinSummaryTokens int

	MinSplit    int
	MaxDepth    int
}

func OutputSummaryContent(output StreamReader, contentbuf *strings.Builder, delta string) {
	if contentbuf != nil {
		contentbuf.WriteString(delta)
	}
	if output != nil {
		output.StreamDelta(contentbuf, delta)
	}
}

func OutputSummaryError(output StreamReader, contentbuf *strings.Builder, status LLMStatus, errstr string) {
	if output != nil {
		output.StreamError(contentbuf, status, errstr)
	}
}

func (s *Summary) InitSummary() error {
	if s.MinSummaryTokens <= 0 {
		s.MinSummaryTokens = defaultMinSummaryTokens
	}
	if s.MinSplit <= 0 {
		s.MinSplit = defaultMinSplit
	}
	if s.MaxDepth <= 0 {
		s.MaxDepth = defaultMaxDepth
	}
	return nil
}

func (s *Summary) TokenizeMessages(messages []ChatMessage) ([]TokenizedMessage, int) {
	if s.Cxt == nil || s.LLM == nil {
		return []TokenizedMessage{}, 0
	}
	tokenized := make([]TokenizedMessage, len(messages))
	total := 0
	for i, msg := range messages {
		bytes, _ := json.Marshal(msg)
		tokens  := s.LLM.CalcTokensByWeakModel(s.Cxt, string(bytes))
		tokenized[i] = TokenizedMessage{Tokens: tokens}
		total += tokens
	}
	return tokenized, total
}

func (s *Summary) AskLLM(msgs []ChatMessage) (LLMStatus, ChatMessage) {
	if s.Cxt == nil || s.LLM == nil {
		return LLM_STATUS_BED_REQUEST, ChatMessage{}
	}

	if s.DisableStream {
		return s.LLM.SendMessagesByWeakModel(s.Cxt, msgs)
	}

	return s.LLM.SendMessagesStreamByWeakModel(s.Cxt, msgs, nil)
}

func (s *Summary) SummarizeOnce(msgs []ChatMessage) (LLMStatus, []ChatMessage) {
	if len(s.PromptSummary) <= 0 || len(s.PromptPrefix) <= 0 {
		OutputSummaryError(s.StreamReader, s.StreamBuffer, LLM_STATUS_BED_MESSAGE, "Summary prompt or prefix invalid!")
		return LLM_STATUS_BED_MESSAGE, []ChatMessage{}
	}
	contentbuf := strings.Builder{}
	for _, msg := range msgs {
		if msg.Role != ROLE_USER && msg.Role != ROLE_ASSISTANT {
			continue
		}
		contentbuf.WriteString("# ")
		contentbuf.WriteString(strings.ToUpper(msg.Role))
		contentbuf.WriteString("\n")
		contentbuf.WriteString(msg.Content)
		if !strings.HasSuffix(msg.Content, "\n") {
			contentbuf.WriteString("\n")
		}
	}

	content := contentbuf.String()
	sysmessage  := ChatMessage{ Role: ROLE_SYSTEM, Content: s.PromptSummary}
	usermessage := ChatMessage{ Role: ROLE_USER, Content: content}
	
	status, summarymsg := s.AskLLM([]ChatMessage{sysmessage, usermessage})

	if status != LLM_STATUS_OK {
		OutputSummaryError(s.StreamReader, s.StreamBuffer, status, summarymsg.Content)
		return status, []ChatMessage{summarymsg}
	}
	summaryprefix  := s.PromptPrefix
	summarycontent := summaryprefix + summarymsg.Content

	finalmessage := ChatMessage{ Role: ROLE_USER, Content: summarycontent}
	okmessage    := ChatMessage{ Role: ROLE_ASSISTANT, Content: "OK"}

	OutputSummaryContent(s.StreamReader, s.StreamBuffer, summarycontent)
	return status, []ChatMessage{finalmessage, okmessage}
}

func (s *Summary) SummarizeSplit(force bool, msgs []ChatMessage, depth int) (LLMStatus, []ChatMessage) {
	if s.Cxt == nil || s.LLM == nil {
		OutputSummaryError(s.StreamReader, s.StreamBuffer, LLM_STATUS_BED_MESSAGE, "Summary parameter invalid!")
		return LLM_STATUS_BED_MESSAGE, []ChatMessage{}
	}
	tokenized, total := s.TokenizeMessages(msgs)
	if !force && total <= s.MinSummaryTokens && depth == 0 {
		return LLM_STATUS_OK, msgs
	}

	if len(msgs) <= s.MinSplit || depth > s.MaxDepth {
		return s.SummarizeOnce(msgs)
	}

	tails := 0
	index := len(msgs)
	halfmax := s.MinSummaryTokens / 2

	for i := len(tokenized) - 1; i >= 0; i-- {
		if tails + tokenized[i].Tokens >= halfmax {
			break
		}
		tails += tokenized[i].Tokens
		index = i
	}

	for index > 1 && msgs[index-1].Role != ROLE_ASSISTANT {
		index--
	}

	if index <= s.MinSplit {
		return s.SummarizeOnce(msgs)
	}

	headmsgs := msgs[:index]
	tailmsgs := msgs[index:]
	status, summarymsgs := s.SummarizeOnce(headmsgs)
	if status != LLM_STATUS_OK {
		return status, summarymsgs
	}

	summarybytes, _ := json.Marshal(summarymsgs)
	summarytokens   := s.LLM.CalcTokensByWeakModel(s.Cxt, string(summarybytes))

	tailbytes, _ := json.Marshal(tailmsgs)
	tailtokens   := s.LLM.CalcTokensByWeakModel(s.Cxt, string(tailbytes))

	finalmsgs := append(summarymsgs, tailmsgs...)

	if summarytokens + tailtokens < s.MinSummaryTokens {
		return LLM_STATUS_OK, finalmsgs
	}

	return s.SummarizeSplit(force, finalmsgs, depth + 1)
}

func (s *Summary) Summarize(longHistory []ChatMessage, shortHistory []ChatMessage, force bool) (LLMStatus, []ChatMessage) {
	msgs := longHistory
	msgs =  append(msgs, shortHistory...)
	return s.SummarizeSplit(force, msgs, 0)
}
