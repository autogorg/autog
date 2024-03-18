package autog

import (
	"fmt"
	"strings"
	"context"
)

type LLMStatus int

const (
	LLM_STATUS_OK LLMStatus = iota
	LLM_STATUS_USER_CANCELED
	LLM_STATUS_EXCEED_CONTEXT
	LLM_STATUS_BED_REQUEST
	LLM_STATUS_BED_RESPONSE
	LLM_STATUS_BED_MESSAGE
	LLM_STATUS_UNKNOWN_ERROR
)

const (
	SYSTEM    string = "system"
	USER      string = "user"
	ASSISTANT string = "assistant"
)

type ChatMessage struct {
	Role string    `json:"role"`
	Content string `json:"content"`
}

func (cm *ChatMessage) String() string {
	return fmt.Sprintf("{Role:%s, Content:%s}", cm.Role, cm.Content)
}

type LLMStreamer interface {
	StreamStart func() *strings.Builder
	StreamDelta func(contentbuf *strings.Builder, delta string)
	StreamEnd func(contentbuf *strings.Builder)
}

type LLM interface {
	InitLLM() error
	CalcTokens(cxt context.Context, content string) int
	SendMessages(cxt context.Context, msgs []ChatMessage) (LLMStatus, ChatMessage)
	SendMessagesStream(cxt context.Context, msgs []ChatMessage, reader LLMStreamer) (LLMStatus, ChatMessage)
}