package autog_test

import (
	"fmt"
	"strings"
	"context"
	_ "embed"
	"autog"
	"autog/llm"
	"autog/rag"
)

const (
	ApiBase = "https://api.chatpp.org/v1"
	ApiKey  = "sk-***"
)

var (
	//go:embed README.md
	docstring string
) 

func ExampleEmbeddings() {
	openai := &llm.OpenAi{ ApiBase: ApiBase, ApiKey: ApiKey}
	err := openai.InitLLM()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	embedds, err := openai.Embeddings(context.Background(), 5,
		[]string{ 
			"Test Embedding String 1",
			"Test Embedding String 2",
		},
	)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	for _, embed := range embedds {
		fmt.Printf("Embedding: %s\n", embed.String(2))
	}

	// Output:
	// Embedding: [0.52, 0.18, -0.31, 0.11, 0.77, ]
	// Embedding: [0.58, 0.19, -0.37, -0.13, 0.69, ]
}

func ExampleRag() {
	cxt := context.Background()

	openai := &llm.OpenAi{ ApiBase: ApiBase, ApiKey: ApiKey}
	err := openai.InitLLM()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	memDB, err := rag.NewMemDatabase()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	memRag := &autog.Rag{
		Database: memDB,
		EmbeddingModel: openai,
	}

	splitter := &rag.TextSplitter{
		ChunkSize: 100,
	}

	memRag.Indexing(cxt, "/doc", docstring, splitter, false)

	var scoredss []autog.ScoredChunks
	scoredss, err  = memRag.Retrieval(cxt, "/doc", []string{"autog是什么"}, 3)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	for _, scoreds := range scoredss {
		for _, scored := range scoreds {
			fmt.Printf("Score:%f\n", scored.Score)
			fmt.Printf("Content:[%s]\n", scored.Chunk.GetContent())
		}
	}

	// Output:
	// 
}

func ExampleChatAgent() {
	type ChatAgent struct {
		autog.Agent
	}

	openai := &llm.OpenAi{ ApiBase: ApiBase, ApiKey: ApiKey}
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

	output := &autog.Output{}
	output.WriteStreamStart = func() *strings.Builder {
		return &strings.Builder{}
	}
	output.WriteStreamDelta = func(contentbuf *strings.Builder, delta string) {
		if output.AgentStage == autog.AsWaitResponse {
			fmt.Print(delta)
		}
	}
	output.WriteStreamError = func(contentbuf *strings.Builder, status autog.LLMStatus, errstr string) {
		fmt.Print(errstr)
	}
	output.WriteStreamEnd = func(contentbuf *strings.Builder) {
		// You can get whole messsage by contentbuf.String()
		return
	}

    chat.Prompt(system, longHistory, shortHistory).
    ReadQuestion(nil, input, output).
    AskLLM(openai, true). // stream = true
    WaitResponse(nil).
    Action(nil).
    Reflection(nil, 3).
    Summarize(nil, summary, prefix, true) // force = false

    // Output:
    // 你好！
}
