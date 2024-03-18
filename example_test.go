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
	
}