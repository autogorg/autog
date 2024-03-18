package autog_test

import (
	"fmt"
	"autog"
	"autog/llm"
)

func ExampleChatAgent() {
	type ChatAgent {
		autog.Agent
	}

	openai := &llm.OpenAi{}
	err := openai.InitLLM()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	chat := &ChatAgent{}

	system := &autog.PromptItem{
		Name : "System",
		GetMessages : func (query string) []ChatMessage {
			return []ChatMessage{
				ChatMessage{
					Role: autog.SYSTEM,
					Content: "你是一个优秀的聊天机器人！"
				}
			}
		}
	}

	longHistory := &autog.PromptItem{
		Name : "LongHistory",
		GetMessages : func (query string) []ChatMessage {
			return chat.GetLongHistory
		}
	}

	shortHistory := &autog.PromptItem{
		Name : "ShortHistory",
		GetMessages : func (query string) []ChatMessage {
			return chat.GetShortHistory
		}
	}

	input := &autog.Input{
		ReadContent: func() string {
			return "你好！"
		}
	}

	output := &autog.Output{
		WriteStreamDelta: func(contentbuf *strings.Builder, delta string) {
			fmt.Print(delta)
		}
	}

	chat.Prompt(
		system,
		longHistory,
		shortHistory,
	)
	.ReadQuestion(nil, input)
	.AskLLM(openai, false)
	.WaitResponse(nil, output)
	.DoAction(nil)
	.DoReflection(nil)
}