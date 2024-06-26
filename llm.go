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
	VerboseNone int = iota
	VerboseShowSending
	VerboseShowReceiving
)

const (
	ROLE_SYSTEM    string = "system"
	ROLE_USER      string = "user"
	ROLE_ASSISTANT string = "assistant"
)

func IsValidRole(role string) bool {
	if role == ROLE_SYSTEM || role == ROLE_USER || role == ROLE_ASSISTANT {
		return true
	}
	return false
}

type ChatMessage struct {
	Role string    `json:"role"`
	Content string `json:"content"`
}

func (cm *ChatMessage) String() string {
	return fmt.Sprintf("{Role:%s, Content:%s}", cm.Role, cm.Content)
}

type StreamReader interface {
	StreamStart() *strings.Builder
	StreamDelta(contentbuf *strings.Builder, delta string)
	StreamError(contentbuf *strings.Builder, status LLMStatus, errstr string)
	StreamEnd(contentbuf *strings.Builder)
}

type LLM interface {
	InitLLM() error

	CalcTokens(cxt context.Context, content string) int
	SendMessages(cxt context.Context, msgs []ChatMessage) (LLMStatus, ChatMessage)
	SendMessagesStream(cxt context.Context, msgs []ChatMessage, reader StreamReader) (LLMStatus, ChatMessage)

	CalcTokensByWeakModel(cxt context.Context, content string) int
	SendMessagesByWeakModel(cxt context.Context, msgs []ChatMessage) (LLMStatus, ChatMessage)
	SendMessagesStreamByWeakModel(cxt context.Context, msgs []ChatMessage, reader StreamReader) (LLMStatus, ChatMessage)
}