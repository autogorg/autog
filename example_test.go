package autog_test

import (
	"fmt"
	"strings"
	"autog"
	"autog/llm"
)

func ExampleChatAgent() {
	type ChatAgent struct {
		autog.Agent
	}

	openai := &llm.OpenAi{
		ApiBase : "https://api.chatpp.org/v1",
		ApiKey  : "sk-ae32368ec577de764f25ca39daac4fbd",
	}
	err := openai.InitLLM()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	chat := &ChatAgent{}

	system := &autog.PromptItem{
		GetPrompt : func (query string) (role string, prompt string) {
			return autog.ROLE_SYSTEM, `你是一个echo机器人，总是原文返回我的问题，例如我的问题是："你好！"，你回答也必须是："你好！"`
		},
	}

	longHistory := &autog.PromptItem{
		GetMessages : func (query string) []autog.ChatMessage {
			return chat.GetLongHistory()
		},
	}

	shortHistory := &autog.PromptItem{
		GetMessages : func (query string) []autog.ChatMessage {
			return chat.GetShortHistory()
		},
	}

	summary := &autog.PromptItem{
		GetPrompt : func (query string) (role string, prompt string) {
			return "", "用500字以内总计一下我们的历史对话！"
		},
	}

	prefix := &autog.PromptItem{
		GetPrompt : func (query string) (role string, prompt string) {
			return "", "我们的历史对话总结如下："
		},
	}

	input := &autog.Input{
		ReadContent: func() string {
			return "你好！"
		},
	}

	output := &autog.Output{
		WriteStreamStart: func() *strings.Builder {
			return &strings.Builder{}
		},
		WriteStreamDelta: func(contentbuf *strings.Builder, delta string) {
			fmt.Print(delta)
		},
		WriteStreamError: func(contentbuf *strings.Builder, status autog.LLMStatus, errstr string) {
			fmt.Print(errstr)
		},
		WriteStreamEnd: func(contentbuf *strings.Builder) {
			// You can get whole messsage by contentbuf.String()
			return
		},
	}

	chat.Prompt(system, longHistory, shortHistory).
    ReadQuestion(nil, input, output).
    AskLLM(openai, false).
    WaitResponse(nil).
    Action(nil).
    Reflection(nil, 3).
	Summarize(nil, summary, prefix, false, output)

	// Output:
	// 你好！
}