# AutoG

# AutoG is a lightweight, comprehensive, and flexible Agent development framework

- Lightweight: Developed in pure Go language, zero third-party dependencies.
- Comprehensive: Fully-featured, includes a prompt framework, RAG, model interfacing interfaces, supports long-term and short-term memory, planning, action, and reflection capabilities, etc.
- Flexible: A functional + react framework, capable of implementing multi-Agent interactions and dynamic state graphs and control flows through the capabilities of Future functions.

### Examples

See `ExampleOpenAiChatAgent` in [example_openai_test.go](./example_openai_test.go)
See `ExampleOllamaChatAgent` in [example_ollama_test.go](./example_ollama_test.go)
```go
    // Step 1. A complete agent that supports continuous chat conversations
    &autog.Agent{}.Prompt(system, longHistory, shortHistory).
        ReadQuestion(nil, input, output).
        AskLLM(openai, true). // stream = true
        WaitResponse(nil).
        Action(nil).
        Reflection(nil, 3).
        Summarize(nil, summary, prefix, true) // force = true
```

See `ExampleOpenAiRag` in [example_openai_test.go](./example_openai_test.go)
See `ExampleOllamaRag` in [example_ollama_test.go](./example_ollama_test.go)
```go
    // Step 1. Create a RAG with a memory vector database
    memDB, _ := rag.NewMemDatabase()
    memRag := &autog.Rag{ Database: memDB, EmbeddingModel: openai }

    // Step 2. Split `docstring` into chunks, and save to database
    splitter := &rag.TextSplitter{ChunkSize: 100}
    memRag.Indexing(cxt, "/doc", docstring, splitter, false)

    // Step 3. Search database by question `what is AutoG?`
    scoredss, _ := memRag.Retrieval(cxt, "/doc", []string{"what is AutoG?"}, 3)
    for _, scoreds := range scoredss {
        for _, scored := range scoreds {
            fmt.Printf("Score:%f\n", scored.Score)
            fmt.Printf("Content:[%s]\n", scored.Chunk.GetContent())
        }
    }
```

# AutoG是一个轻量、完整、灵活的Agent开发框架

- 轻量：纯Go语言开发，零第三方依赖。
- 完整：功能齐全，包含提示工程框架，RAG，模型对接接口，支持长短期记忆、计划、行动和反思能力等。
- 灵活：函数式+响应式框架，可通过Future函数的能力，实现多Agent交互以及动态的状态图和控制流。

### 样例

See `ExampleOpenAiChatAgent` in [example_openai_test.go](./example_openai_test.go)
See `ExampleOllamaChatAgent` in [example_ollama_test.go](./example_ollama_test.go)
```go
    // 步骤 1. 一个完整的支持连续聊天对话的智能体
    &autog.Agent{}.Prompt(system, longHistory, shortHistory).
        ReadQuestion(nil, input, output).
        AskLLM(openai, true). // stream = true
        WaitResponse(nil).
        Action(nil).
        Reflection(nil, 3).
        Summarize(nil, summary, prefix, true) // force = true
```

See `ExampleOpenAiRag` in [example_openai_test.go](./example_openai_test.go)
See `ExampleOllamaRag` in [example_ollama_test.go](./example_ollama_test.go)
```go
    // 步骤 1. 创建一个RAG并初始化，使其使用内存向量数据库
    memDB, _ := rag.NewMemDatabase()
    memRag := &autog.Rag{ Database: memDB, EmbeddingModel: openai }

    // 步骤 2. 将 `docstring` 分割成小块块，并保存到数据库
    splitter := &rag.TextSplitter{ChunkSize: 100}
    memRag.Indexing(cxt, "/doc", docstring, splitter, false)

    // 步骤 2. 用问题 `what is AutoG?` 去检索向量数据库
    scoredss, _ := memRag.Retrieval(cxt, "/doc", []string{"what is AutoG?"}, 3)
    for _, scoreds := range scoredss {
        for _, scored := range scoreds {
            fmt.Printf("Score:%f\n", scored.Score)
            fmt.Printf("Content:[%s]\n", scored.Chunk.GetContent())
        }
    }
```
