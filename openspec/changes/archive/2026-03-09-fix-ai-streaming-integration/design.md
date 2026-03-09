## Context

The AI module was recently refactored from a custom ActionGraph implementation to use Eino's built-in `react.Agent`. This simplified the architecture but created integration gaps:

1. **Streaming Events Gap**: The `react.Agent` provides streaming via `Stream()` API, but the current implementation in `agent.go` only emits `delta`, `thinking_delta`, and `tool_call` events. The `tool_result` event is missing when tools complete execution.

2. **SecurityAspect Not Wired**: The `SecurityAspect.Middleware()` returns a `compose.ToolMiddleware` that should intercept tool calls for permission checks and approval interruption. However, this middleware is not being passed to the `react.Agent` during construction.

3. **Resume Logic Incomplete**: The `AIAgent.Resume()` method exists but the checkpoint-based resume flow needs to properly handle the `approval_required` interrupt and subsequent approval submission.

4. **Heartbeat Missing from Agent Level**: The heartbeat is only at the Orchestrator level, but long-running tool executions may not trigger heartbeat events.

```
Current Flow (Broken):
┌─────────────┐    Stream()     ┌────────────────┐    Tool Execution    ┌──────────────┐
│ Orchestrator │ ───────────────▶│   react.Agent   │ ────────────────────▶│ ToolsNode    │
└─────────────┘                 └────────────────┘                      └──────────────┘
      │                               │                                       │
      │ emit("delta")                 │ emit("tool_call")                     │ ❌ No tool_result
      │ emit("thinking_delta")        │                                       │ ❌ No SecurityAspect
      │                               │                                       │
      └───────────────────────────────┴───────────────────────────────────────┘
                                      SSE Events Missing: tool_result, approval_required
```

## Goals / Non-Goals

**Goals:**
- Emit complete SSE events: `delta`, `thinking_delta`, `tool_call`, `tool_result`, `approval_required`, `heartbeat`
- Integrate SecurityAspect into tool execution for permission checks and approval interruption
- Complete Resume logic for approval-based checkpoint recovery
- Ensure frontend receives all expected events without modification

**Non-Goals:**
- Frontend changes (frontend already expects these events)
- Changes to tool implementations themselves
- Database schema changes for checkpoint storage

## Decisions

### Decision 1: Use Callbacks for Tool Result Events

**Choice**: Use Eino's callback system to capture tool execution results

**Rationale**: The `react.Agent` executes tools internally. To emit `tool_result` events, we need to hook into the tool execution lifecycle. Eino provides `callbacks.Handler` for this purpose.

**Implementation**:
```go
type streamingCallbackHandler struct {
    callbacks.HandlerBuilder
    emit func(event string, payload map[string]any) bool
}

func (h *streamingCallbackHandler) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
    if info != nil && info.Component == components.ComponentOfTool {
        emit("tool_result", map[string]any{
            "tool_name": info.Name,
            "result":    output.Output,
        })
    }
    return ctx
}
```

**Alternatives Considered**:
- Wrap each tool individually: More invasive, duplicates logic
- Modify react.Agent source: Not maintainable

### Decision 2: Wire SecurityAspect via ToolsNodeConfig

**Choice**: Pass SecurityAspect.Middleware() to the ToolsNodeConfig when creating the react.Agent

**Rationale**: The `compose.ToolsNodeConfig` accepts a `Middlewares` field. This is the canonical way to inject cross-cutting concerns into tool execution.

**Implementation**:
```go
securityAspect := aspect.NewSecurityAspect(registered, checker, handler, logger)
agent, err := react.NewAgent(ctx, &react.AgentConfig{
    ToolCallingModel: model,
    ToolsConfig: compose.ToolsNodeConfig{
        Tools: allTools,
        Middlewares: []compose.ToolMiddleware{
            securityAspect.Middleware(),
        },
    },
    MaxStep: maxStep,
})
```

**Alternatives Considered**:
- Wrap each tool with ApprovableTool: Already done but doesn't integrate permission checks
- Use global middleware: Less fine-grained control

### Decision 3: Checkpoint Store for Resume

**Choice**: Use Redis-backed CheckPointStore for session state persistence

**Rationale**: The `compose.WithCheckPointStore` option allows the Agent to persist state at interruption points. This enables resume after approval.

**Implementation**:
```go
checkpointStore := compose.NewRedisCheckPointStore(redisClient)
agent, err := react.NewAgent(ctx, &react.AgentConfig{
    // ...
}, compose.WithCheckPointStore(checkpointStore))
```

**Alternatives Considered**:
- In-memory store: Not durable across restarts
- Database store: More complex, Redis is already available

### Decision 4: Heartbeat via Context Cancellation

**Choice**: Emit heartbeat events from the streaming loop in agent.Stream()

**Rationale**: The streaming loop already processes chunks. Adding periodic heartbeat emission ensures long-running operations don't timeout.

**Implementation**:
```go
heartbeatTicker := time.NewTicker(10 * time.Second)
defer heartbeatTicker.Stop()

for {
    select {
    case <-heartbeatTicker.C:
        emit("heartbeat", map[string]any{"ts": time.Now().UTC().Format(time.RFC3339)})
    // ... existing chunk processing
    }
}
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| SecurityAspect middleware may reject valid tool calls | Ensure permission checker has correct role mappings |
| Checkpoint storage may fail | Graceful fallback to in-memory for degraded mode |
| Heartbeat may add overhead | Make interval configurable via RunnerConfig |
| Tool result may be large | Truncate or summarize large results |

## Migration Plan

1. **Phase 1**: Add callback handler for tool_result events (no breaking changes)
2. **Phase 2**: Wire SecurityAspect middleware (enhances security)
3. **Phase 3**: Add CheckPointStore for resume capability
4. **Phase 4**: Add heartbeat emission

**Rollback**: Each phase is independent. Revert individual commits if issues arise.
