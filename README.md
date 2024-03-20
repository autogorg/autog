# AutoG

AutoG is a lightweight, complete, and open agent development framework

- Lightweight: Developed purely in Go language without any third-party dependencies.
- Complete: Flexible prompt engineering framework, short and long-term memory, planning, acting, and reflecting capabilities.
- Open: Defaults to interfacing with OpenAI's API, but can easily interface with any other model.

### Example

```go
&autog.Agent{}.Prompt(system, longHistory, shortHistory).
    ReadQuestion(nil, input, output).
    AskLLM(openai, true). // stream = true
    WaitResponse(nil).
    Action(nil).
    Reflection(nil, 3).
    Summarize(nil, summary, prefix, true) // force = true
```

# AutoG是一个轻量、完整、开放的代理开发框架

- 轻量：纯Go语言开发，无任何第三方依赖。
- 完整：灵活的提示工程框架，支持长短期记忆、计划、行动和反思能力。
- 开放：默认与 OpenAI 的 API 对接，但可以轻松与任何其他模型对接。