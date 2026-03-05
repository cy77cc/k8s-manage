# Design: Eino AI Architecture Optimization

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         PlatformAgent                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │                    ExpertGraph (compose.Graph)              │   │
│   │                                                             │   │
│   │   START ──▶ route ──▶ primary ──▶ helpers ──▶ aggregate    │   │
│   │                              │              │               │   │
│   │                              └──[branch]────┘               │   │
│   │                                                             │   │
│   └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              ▼                                       │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │                    AICallbacksHandler                       │   │
│   │                                                             │   │
│   │   OnToolCallStart ──▶ OnToolCallEnd ──▶ OnExpertStart ──▶  │   │
│   │                                                          ... │   │
│   └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
│   ┌────────────────────┐    ┌────────────────────┐                  │
│   │   ExpertRegistry   │    │    ToolRegistry    │                  │
│   │   (experts/*.go)   │    │    (tools/*.go)    │                  │
│   └────────────────────┘    └────────────────────┘                  │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

## Component Design

### 1. Callbacks Module

**文件结构**：
```
internal/ai/callbacks/
├── handler.go       # AIEventHandler 实现
├── events.go        # 事件类型定义
├── context.go       # context 集成
├── emitter.go       # EventEmitter 接口
└── handler_test.go  # 单元测试
```

**events.go**：
```go
package callbacks

import "time"

// ToolCallEvent 工具调用事件
type ToolCallEvent struct {
    Tool      string         `json:"tool"`
    CallID    string         `json:"call_id"`
    Arguments map[string]any `json:"arguments,omitempty"`
    Result    any            `json:"result,omitempty"`
    Error     string         `json:"error,omitempty"`
    Timestamp time.Time      `json:"timestamp"`
    Duration  time.Duration  `json:"duration,omitempty"`
}

// ExpertProgressEvent 专家进度事件
type ExpertProgressEvent struct {
    Expert     string `json:"expert"`
    Status     string `json:"status"` // running, done, failed
    Task       string `json:"task,omitempty"`
    DurationMs int64  `json:"duration_ms,omitempty"`
    Error      string `json:"error,omitempty"`
}

// StreamEvent 流式事件
type StreamEvent struct {
    Type      string `json:"type"` // delta, thinking_delta, done, error
    Content   string `json:"content,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}
```

**emitter.go**：
```go
package callbacks

// EventEmitter 事件发射器接口
type EventEmitter interface {
    Emit(event string, payload any) bool
}

// EventEmitterFunc 函数适配器
type EventEmitterFunc func(event string, payload any) bool

func (f EventEmitterFunc) Emit(event string, payload any) bool {
    return f(event, payload)
}

// NopEmitter 空实现
var NopEmitter EventEmitter = EventEmitterFunc(func(string, any) bool { return true })
```

**handler.go**：
```go
package callbacks

import (
    "context"
    "github.com/cloudwego/eino/callbacks"
)

// AIEventHandler 统一事件处理器
type AIEventHandler struct {
    emitter EventEmitter
}

func NewAIEventHandler(emitter EventEmitter) *AIEventHandler {
    if emitter == nil {
        emitter = NopEmitter
    }
    return &AIEventHandler{emitter: emitter}
}

// OnToolCallStart 工具调用开始
func (h *AIEventHandler) OnToolCallStart(ctx context.Context, tool, callID string, args map[string]any) context.Context {
    h.emitter.Emit("tool_call", ToolCallEvent{
        Tool:      tool,
        CallID:    callID,
        Arguments: args,
        Timestamp: time.Now(),
    })
    return ctx
}

// OnToolCallEnd 工具调用结束
func (h *AIEventHandler) OnToolCallEnd(ctx context.Context, tool, callID string, result any, err error, duration time.Duration) context.Context {
    event := ToolCallEvent{
        Tool:      tool,
        CallID:    callID,
        Result:    result,
        Duration:  duration,
        Timestamp: time.Now(),
    }
    if err != nil {
        event.Error = err.Error()
    }
    h.emitter.Emit("tool_result", event)
    return ctx
}

// OnExpertStart 专家执行开始
func (h *AIEventHandler) OnExpertStart(ctx context.Context, expert, task string) context.Context {
    h.emitter.Emit("expert_progress", ExpertProgressEvent{
        Expert: expert,
        Status: "running",
        Task:   task,
    })
    return ctx
}

// OnExpertEnd 专家执行结束
func (h *AIEventHandler) OnExpertEnd(ctx context.Context, expert string, duration time.Duration, err error) context.Context {
    event := ExpertProgressEvent{
        Expert:     expert,
        Status:     "done",
        DurationMs: duration.Milliseconds(),
    }
    if err != nil {
        event.Status = "failed"
        event.Error = err.Error()
    }
    h.emitter.Emit("expert_progress", event)
    return ctx
}
```

**context.go**：
```go
package callbacks

import "context"

type emitterKey struct{}

// WithEmitter 将 emitter 注入 context
func WithEmitter(ctx context.Context, emitter EventEmitter) context.Context {
    return context.WithValue(ctx, emitterKey{}, emitter)
}

// EmitterFromContext 从 context 获取 emitter
func EmitterFromContext(ctx context.Context) EventEmitter {
    if v := ctx.Value(emitterKey{}); v != nil {
        if emitter, ok := v.(EventEmitter); ok {
            return emitter
        }
    }
    return NopEmitter
}

// HandlerFromContext 从 context 构建 handler
func HandlerFromContext(ctx context.Context) *AIEventHandler {
    return NewAIEventHandler(EmitterFromContext(ctx))
}
```

### 2. Graph Module

**文件结构**：
```
internal/ai/graph/
├── builder.go       # 图构建器
├── nodes.go         # 节点实现
├── branches.go      # 条件分支
├── types.go         # 类型定义
└── builder_test.go  # 测试
```

**types.go**：
```go
package graph

import (
    "github.com/cloudwego/eino/schema"
    "github.com/cy77cc/k8s-manage/internal/ai/experts"
)

// GraphInput 图输入
type GraphInput struct {
    Message        string
    History        []*schema.Message
    Decision       *experts.RouteDecision
    RuntimeContext map[string]any
}

// GraphOutput 图输出
type GraphOutput struct {
    Response string
    Traces   []experts.ExpertTrace
    Metadata map[string]any
}
```

**builder.go**：
```go
package graph

import (
    "context"
    "github.com/cloudwego/eino/compose"
    "github.com/cy77cc/k8s-manage/internal/ai/experts"
)

// Builder 图构建器
type Builder struct {
    registry   experts.ExpertRegistry
    aggregator *experts.ResultAggregator
}

func NewBuilder(registry experts.ExpertRegistry, aggregator *experts.ResultAggregator) *Builder {
    return &Builder{
        registry:   registry,
        aggregator: aggregator,
    }
}

// Build 构建专家执行图
func (b *Builder) Build() *compose.Graph[*GraphInput, *GraphOutput] {
    g := compose.NewGraph[*GraphInput, *GraphOutput]()

    // 添加节点
    g.AddLambdaNode("primary", compose.InvokableLambda(b.runPrimary))
    g.AddLambdaNode("helpers_parallel", compose.InvokableLambda(b.runHelpersParallel))
    g.AddLambdaNode("helpers_sequential", compose.InvokableLambda(b.runHelpersSequential))
    g.AddLambdaNode("aggregate", compose.InvokableLambda(b.aggregateResults))

    // 添加边
    g.AddEdge(compose.START, "primary")

    // 条件分支：根据策略选择 helpers 执行方式
    g.AddBranch("primary", b.helperStrategyBranch())

    g.AddEdge("helpers_parallel", "aggregate")
    g.AddEdge("helpers_sequential", "aggregate")
    g.AddEdge("aggregate", compose.END)

    return g
}

// helperStrategyBranch 根据策略选择分支
func (b *Builder) helperStrategyBranch() *compose.GraphBranch {
    return compose.NewGraphBranch(
        func(ctx context.Context, input *GraphInput) (string, error) {
            if input.Decision == nil || len(input.Decision.OptionalHelpers) == 0 {
                return "skip_helpers", nil
            }
            switch input.Decision.Strategy {
            case experts.StrategyParallel:
                return "helpers_parallel", nil
            case experts.StrategySequential:
                return "helpers_sequential", nil
            default:
                return "helpers_parallel", nil
            }
        },
        map[string]string{
            "helpers_parallel":   "helpers_parallel",
            "helpers_sequential": "helpers_sequential",
            "skip_helpers":       "aggregate",
        },
    )
}
```

**nodes.go**：
```go
package graph

import (
    "context"
    "sync"
    "time"

    "github.com/cloudwego/eino/schema"
    "github.com/cy77cc/k8s-manage/internal/ai/callbacks"
    "github.com/cy77cc/k8s-manage/internal/ai/experts"
)

// runPrimary 执行主专家
func (b *Builder) runPrimary(ctx context.Context, input *GraphInput) (*GraphOutput, error) {
    if input.Decision == nil {
        return nil, fmt.Errorf("decision is required")
    }

    handler := callbacks.HandlerFromContext(ctx)
    expertName := input.Decision.PrimaryExpert

    handler.OnExpertStart(ctx, expertName, "primary analysis")
    start := time.Now()

    exp, ok := b.registry.GetExpert(expertName)
    if !ok {
        return nil, fmt.Errorf("expert not found: %s", expertName)
    }

    messages := buildMessages(input.History, input.Message)
    resp, err := exp.Agent.Generate(ctx, messages)

    handler.OnExpertEnd(ctx, expertName, time.Since(start), err)

    if err != nil {
        return &GraphOutput{
            Response: "",
            Traces: []experts.ExpertTrace{{
                ExpertName: expertName,
                Role:       "primary",
                Status:     "failed",
                Error:      err.Error(),
            }},
        }, err
    }

    return &GraphOutput{
        Response: resp.Content,
        Traces: []experts.ExpertTrace{{
            ExpertName: expertName,
            Role:       "primary",
            Status:     "success",
        }},
    }, nil
}

// runHelpersParallel 并行执行助手专家
func (b *Builder) runHelpersParallel(ctx context.Context, input *GraphInput) (*GraphOutput, error) {
    helpers := input.Decision.OptionalHelpers
    if len(helpers) == 0 {
        return &GraphOutput{}, nil
    }

    handler := callbacks.HandlerFromContext(ctx)
    results := make([]experts.ExpertResult, len(helpers))
    var mu sync.Mutex
    var wg sync.WaitGroup

    for i, helperName := range helpers {
        wg.Add(1)
        go func(idx int, name string) {
            defer wg.Done()

            handler.OnExpertStart(ctx, name, "helper analysis")
            start := time.Now()

            exp, ok := b.registry.GetExpert(name)
            if !ok {
                handler.OnExpertEnd(ctx, name, time.Since(start), fmt.Errorf("not found"))
                mu.Lock()
                results[idx] = experts.ExpertResult{
                    ExpertName: name,
                    Error:      fmt.Errorf("expert not found"),
                }
                mu.Unlock()
                return
            }

            messages := buildHelperMessages(input, name)
            resp, err := exp.Agent.Generate(ctx, messages)
            duration := time.Since(start)

            handler.OnExpertEnd(ctx, name, duration, err)

            mu.Lock()
            results[idx] = experts.ExpertResult{
                ExpertName: name,
                Output:     resp.Content,
                Duration:   duration,
                Error:      err,
            }
            mu.Unlock()
        }(i, helperName)
    }

    wg.Wait()

    return &GraphOutput{
        Traces: buildTraces(results, "helper"),
    }, nil
}

// aggregateResults 汇总结果
func (b *Builder) aggregateResults(ctx context.Context, input *GraphOutput) (*GraphOutput, error) {
    if b.aggregator == nil || len(input.Traces) <= 1 {
        return input, nil
    }

    results := make([]experts.ExpertResult, len(input.Traces))
    for i, trace := range input.Traces {
        results[i] = experts.ExpertResult{
            ExpertName: trace.ExpertName,
            Output:     trace.Output,
            Error:      trace.Error,
        }
    }

    aggregated, err := b.aggregator.Aggregate(ctx, results, "")
    if err != nil {
        return input, err
    }

    return &GraphOutput{
        Response: aggregated,
        Traces:   input.Traces,
    }, nil
}
```

### 3. Expert Tool Adapter

**文件结构**：
```
internal/ai/experts/
├── tool_adapter.go  # 专家转 Tool 适配器
```

**tool_adapter.go**：
```go
package experts

import (
    "context"
    "encoding/json"

    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/schema"
)

// ExpertToolInput 专家工具输入
type ExpertToolInput struct {
    Task    string `json:"task"`
    Context string `json:"context,omitempty"`
}

// BuildExpertTool 将专家暴露为可调用的 Tool
func BuildExpertTool(expert *Expert) tool.InvokableTool {
    return tool.InvokableLambda(
        func(ctx context.Context, input string) (string, error) {
            var req ExpertToolInput
            if err := json.Unmarshal([]byte(input), &req); err != nil {
                req.Task = input // fallback to raw input
            }

            messages := []*schema.Message{
                schema.SystemMessage(expert.Persona),
                schema.UserMessage(req.Task),
            }

            if req.Context != "" {
                messages = append(messages[:1],
                    schema.UserMessage("上下文: "+req.Context+"\n\n任务: "+req.Task),
                )
            }

            resp, err := expert.Agent.Generate(ctx, messages)
            if err != nil {
                return "", err
            }

            return resp.Content, nil
        },
        tool.WithName(expert.Name),
        tool.WithDescription(expert.Persona),
    )
}

// BuildExpertTools 为所有专家构建工具集
func BuildExpertTools(registry ExpertRegistry) []tool.BaseTool {
    experts := registry.ListExperts()
    tools := make([]tool.BaseTool, 0, len(experts))

    for _, exp := range experts {
        if exp.Agent != nil {
            tools = append(tools, BuildExpertTool(exp))
        }
    }

    return tools
}
```

## Integration Points

### PlatformAgent 改造

```go
// platform_agent.go
func NewPlatformAgent(ctx context.Context, chatModel model.ToolCallingChatModel, deps tools.PlatformDeps) (*PlatformAgent, error) {
    // ... existing setup ...

    // 新增：构建 Graph
    graphBuilder := graph.NewBuilder(registry, aggregator)
    expertGraph := graphBuilder.Build()
    compiledGraph, err := expertGraph.Compile(ctx)
    if err != nil {
        return nil, err
    }

    return &PlatformAgent{
        // ... existing fields ...
        graph: compiledGraph, // 新增
    }, nil
}

func (p *PlatformAgent) Stream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
    // 使用 callbacks
    handler := callbacks.HandlerFromContext(ctx)
    ctx = callbacks.WithEmitter(ctx, handler)

    // 使用 graph 执行
    if p.graph != nil {
        input := &graph.GraphInput{
            Message:  extractUserMessage(messages),
            History:  messages,
            Decision: p.router.Route(ctx, ...),
        }
        // 流式输出 ...
    }

    // fallback to react.Agent
    return p.Runnable.Stream(ctx, messages)
}
```

### chat_handler.go 改造

```go
// chat_handler.go
func (h *handler) chat(c *gin.Context) {
    // ... existing setup ...

    // 使用统一的 callbacks
    emitter := callbacks.EventEmitterFunc(func(event string, payload any) bool {
        return emit(event, toPayloadMap(payload))
    })
    streamCtx := callbacks.WithEmitter(c.Request.Context(), emitter)

    // ... rest of the handler ...
}
```

## Migration Strategy

### 阶段 1：Callbacks 迁移（低风险）

1. 创建 `internal/ai/callbacks/` 模块
2. 编写单元测试
3. 在 `chat_handler.go` 中集成
4. 迁移 `experts/context.go` 中的 ProgressEmitter
5. 迁移 `tools/*.go` 中的 ToolEventEmitter

### 阶段 2：Graph 编排（中风险）

1. 创建 `internal/ai/graph/` 模块
2. 实现节点和分支
3. 编写集成测试
4. 在 `platform_agent.go` 中并行运行（A/B 测试）
5. 验证输出一致性后切换

### 阶段 3：专家 Tool 化（中风险）

1. 创建 `tool_adapter.go`
2. 注册专家工具到 ToolRegistry
3. 主专家可通过 tool calling 调用助手
4. 移除正则解析逻辑
5. 更新测试

## Testing Strategy

### 单元测试

- `callbacks/handler_test.go` - 事件发射验证
- `graph/builder_test.go` - 图构建和节点执行
- `experts/tool_adapter_test.go` - Tool 调用

### 集成测试

- `experts/integration_test.go` - 端到端专家协作
- `platform_agent_test.go` - 完整流程

### 回归测试

- 确保所有现有测试通过
- 验证 SSE 流式输出格式不变
- 验证前端兼容性
