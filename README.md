# AutoG

# AutoG is a lightweight, comprehensive, and flexible Agent development framework

- Lightweight: Developed in pure Go language, zero third-party dependencies.
- Comprehensive: Fully-featured, includes a prompt framework, RAG, model interfacing interfaces, supports long-term and short-term memory, planning, action, and reflection capabilities, etc.
- Flexible: A functional + react framework, capable of implementing multi-Agent interactions and dynamic state graphs and control flows through the capabilities of Future functions.

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

# AutoG是一个轻量、完整、灵活的Agent开发框架

- 轻量：纯Go语言开发，零第三方依赖。
- 完整：功能齐全，包含提示工程框架，RAG，模型对接接口，支持长短期记忆、计划、行动和反思能力等。
- 灵活：函数式+响应式框架，可通过Future函数的能力，实现多Agent交互以及动态的状态图和控制流。
