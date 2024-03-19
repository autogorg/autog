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
		Name : "System",
		GetMessages : func (query string) []autog.ChatMessage {
			return []autog.ChatMessage{
				autog.ChatMessage{
					Role: autog.SYSTEM,
					Content: "你是一个优秀的聊天机器人！",
				},
			}
		},
	}

	longHistory := &autog.PromptItem{
		Name : "LongHistory",
		GetMessages : func (query string) []autog.ChatMessage {
			return chat.GetLongHistory()
		},
	}

	shortHistory := &autog.PromptItem{
		Name : "ShortHistory",
		GetMessages : func (query string) []autog.ChatMessage {
			return chat.GetShortHistory()
		},
	}

	input := &autog.Input{
		ReadContent: func() string {
			return "你好！"
		},
	}

	output := &autog.Output{
		WriteStreamDelta: func(contentbuf *strings.Builder, delta string) {
			fmt.Print(delta)
		},
	}

	chat.Prompt(system, longHistory, shortHistory).
    ReadQuestion(nil, input).
    AskLLM(openai, false).
    WaitResponse(nil, output).
    Action(nil).
    Reflection(nil, 3)
}