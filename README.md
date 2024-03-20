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