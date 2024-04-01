package autog_test

import (
	"fmt"
	"strings"
	"context"
	_ "embed"
	"github.com/autogorg/autog"
	"github.com/autogorg/autog/llm"
	"github.com/autogorg/autog/rag"
)

const (
	OllamaApiBase    = "http://localhost:11434"
	OllamaModel      = "gemma:2b"
	OllamaModelWeak  = "gemma:2b"
	OllamaModelEmbed = "nomic-embed-text"
)

var (
	//go:embed README.md
	ollamaDocstring string
) 

func ExampleOllamaEmbeddings() {
	ollama := &llm.Ollama{ ApiBase: OllamaApiBase, Model: OllamaModel, ModelWeak: OllamaModelWeak, ModelEmbedding: OllamaModelEmbed }
	err := ollama.InitLLM()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	embedds, err := ollama.Embeddings(context.Background(), 5,
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

func ExampleOllamaRag() {
	cxt := context.Background()

	ollama := &llm.Ollama{ ApiBase: OllamaApiBase, Model: OllamaModel, ModelWeak: OllamaModelWeak, ModelEmbedding: OllamaModelEmbed }
	err := ollama.InitLLM()
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
		EmbeddingModel: ollama,
	}

	splitter := &rag.TextSplitter{
		ChunkSize: 100,
		Overlap: 0.25,
	}

	err = memRag.Indexing(cxt, "/doc", ollamaDocstring, splitter, false)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	var scoredss []autog.ScoredChunks
	scoredss, err  = memRag.Retrieval(cxt, "/doc", []string{"what is AutoG?"}, 3)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	cnt := 0
	for _, scoreds := range scoredss {
		for _, scored := range scoreds {
			// fmt.Printf("Score:%f\n", scored.Score)
			// fmt.Printf("Content:[%s]\n", scored.Chunk.GetContent())
			_ = scored
			cnt++
		}
	}
	fmt.Println(cnt)

	// Output:
	// 3
}

func ExampleOllamaChatAgent() {
	type ChatAgent struct {
		autog.Agent
	}

	ollama := &llm.Ollama{ ApiBase: OllamaApiBase, Model: OllamaModel, ModelWeak: OllamaModelWeak, ModelEmbedding: OllamaModelEmbed }
	err := ollama.InitLLM()
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
		WriteContent: func(stage autog.AgentStage, stream autog.StreamStage, buf *strings.Builder, str string) {
			if stage == autog.AsWaitResponse && stream == autog.StreamStageDelta {
				fmt.Print(str)
			} else if stream == autog.StreamStageError {
				fmt.Print(str)
			}
		},
	}

    chat.Prompt(system, longHistory, shortHistory).
    ReadQuestion(nil, input, output).
    AskLLM(ollama, true). // stream = true
    WaitResponse(nil).
    Action(nil).
    Reflection(nil, 3).
    Summarize(nil, summary, prefix, true) // force = false

    // Output:
    // 你好！
}
