package autog_test

import (
	"autog"
	"autog/llm"
)

func ExampleChatAgent() {
	type ChatAgent {
		autog.Agent
	}
	openai := &llm.OpenAi{}
}