# Proposal: Eino AI Architecture Optimization

## Summary

基于 eino 框架能力，重构 AI 助手核心架构，实现：
1. **Callbacks 统一事件处理** - 替换散落的 context.Value 传递
2. **Graph 编排替代手动 Orchestrator** - 声明式工作流编排
3. **专家协作机制改进** - Tool 化专家调用，消除正则解析

## Motivation

### 当前问题

**1. 事件处理分散**
```
internal/ai/experts/context.go    → ProgressEmitter (context.Value)
internal/ai/tools/*.go            → ToolEventEmitter (context.Value)
internal/service/ai/chat_handler.go → 多个自定义 emitter
```
- 缺乏统一切面
- 难以追踪和调试
- 无法集成外部监控

**2. Orchestrator 手动编排复杂**
```go
// orchestrator.go:116-180 - 手动管理流式输出
func (o *Orchestrator) StreamExecute(...) {
    if strategy == StrategyPrimaryLed {
        return o.streamPrimaryLed(ctx, req)  // 分支1
    }
    // 分支2: 并行/顺序执行
    sr, sw := schema.Pipe[*schema.Message](64)
    go func() { ... }()  // 手动 goroutine
}
```
- 代码难以维护
- 无可视化能力
- 无中断/恢复支持

**3. 专家协作使用正则解析**
```go
// orchestrator.go:16
var helperRequestPattern = regexp.MustCompile(`\[REQUEST_HELPER:\s*([a-zA-Z0-9_]+):\s*([^\]]+)\]`)
```
- 不可靠的协作机制
- 无法嵌套调用
- 难以调试

### eino 框架能力对比

| 能力 | 当前实现 | eino 提供 |
|------|----------|-----------|
| 事件处理 | context.Value 散落 | Callbacks 统一切面 |
| 流程编排 | 手动 if-else/goroutine | Graph/Chain 声明式 |
| 中断恢复 | 不支持 | Checkpoint 支持 |
| 专家协作 | 正则解析 | Tool 化 / Graph Branch |

## Goals

1. 统一事件处理入口，支持日志、追踪、监控集成
2. 使用 Graph 编排简化 Orchestrator，提升可维护性
3. 专家调用 Tool 化，支持嵌套调用和追踪

## Non-Goals

- 模型提供者扩展（稍后处理）
- Checkpoint 中断恢复（适用场景有限，暂不实现）
- 前端 UI 变更

## Proposed Changes

### Phase 1: Callbacks 统一事件处理

**新建 `internal/ai/callbacks/` 目录**：

```
internal/ai/callbacks/
├── handler.go      # 统一回调处理器
├── events.go       # 事件类型定义
├── context.go      # context 集成
└── handler_test.go
```

**核心接口**：
```go
// handler.go
type AIEventHandler struct {
    emitter EventEmitter
}

func (h *AIEventHandler) OnToolCallStart(ctx context.Context, info *ToolCallInfo) context.Context {
    h.emitter("tool_call", info)
    return ctx
}

func (h *AIEventHandler) OnToolCallEnd(ctx context.Context, info *ToolCallInfo) context.Context {
    h.emitter("tool_result", info)
    return ctx
}

func (h *AIEventHandler) OnExpertStart(ctx context.Context, info *ExpertInfo) context.Context {
    h.emitter("expert_progress", ExpertProgressEvent{Status: "running", ...})
    return ctx
}
```

**迁移现有代码**：
- `experts/context.go` → `callbacks/context.go`
- `tools/*.go` 中的 emitter → 统一到 callbacks

### Phase 2: Graph 编排替代 Orchestrator

**新建 `internal/ai/graph/` 目录**：

```
internal/ai/graph/
├── builder.go      # 图构建器
├── nodes.go        # 节点定义
├── branches.go     # 条件分支
└── builder_test.go
```

**声明式编排**：
```go
// builder.go
func BuildExpertGraph(registry ExpertRegistry) *compose.Graph[*ExecuteRequest, *ExecuteResult] {
    g := compose.NewGraph[*ExecuteRequest, *ExecuteResult]()

    // 节点
    g.AddLambdaNode("route", routeNode)
    g.AddLambdaNode("primary", primaryExpertNode)
    g.AddLambdaNode("helpers", helpersParallelNode)
    g.AddLambdaNode("aggregate", aggregateNode)

    // 边
    g.AddEdge(compose.START, "route")
    g.AddEdge("route", "primary")
    g.AddBranch("primary", helperDecisionBranch)
    g.AddEdge("helpers", "aggregate")
    g.AddEdge("aggregate", compose.END)

    return g
}
```

**节点实现**：
```go
// nodes.go
func primaryExpertNode(ctx context.Context, req *ExecuteRequest) (*ExecuteResult, error) {
    exp, ok := registry.GetExpert(req.Decision.PrimaryExpert)
    if !ok {
        return nil, fmt.Errorf("expert not found")
    }
    return exp.Agent.Generate(ctx, buildMessages(req))
}

func helpersParallelNode(ctx context.Context, req *ExecuteRequest) (*ExecuteResult, error) {
    // 并行执行 helpers
    var wg sync.WaitGroup
    results := make([]ExpertResult, len(req.HelperRequests))
    // ...
}
```

### Phase 3: 专家协作 Tool 化

**专家作为 Tool 暴露**：
```go
// experts/tool_adapter.go
func BuildExpertTool(expert *Expert) tool.InvokableTool {
    return tool.InvokableLambda(func(ctx context.Context, input string) (string, error) {
        req := parseHelperRequest(input)
        resp, err := expert.Agent.Generate(ctx, req.Messages)
        if err != nil {
            return "", err
        }
        return resp.Content, nil
    },
    tool.WithName(expert.Name),
    tool.WithDescription(expert.Persona),
    )
}
```

**主专家通过 Tool 调用助手**：
```go
// 不再需要正则解析 [REQUEST_HELPER:...]
// 主专家直接通过工具调用机制调用助手专家
```

## Impact Analysis

### 改动范围

| 文件 | 改动类型 | 影响 |
|------|----------|------|
| `internal/ai/callbacks/*` | 新建 | 新模块 |
| `internal/ai/graph/*` | 新建 | 新模块 |
| `internal/ai/experts/context.go` | 删除/迁移 | 迁移到 callbacks |
| `internal/ai/experts/orchestrator.go` | 重写 | 使用 Graph |
| `internal/ai/experts/executor.go` | 简化 | 节点实现 |
| `internal/ai/platform_agent.go` | 修改 | 集成新架构 |
| `internal/service/ai/chat_handler.go` | 修改 | 使用 callbacks |
| `internal/ai/tools/*.go` | 修改 | 移除独立 emitter |

### 风险评估

| 风险 | 级别 | 缓解措施 |
|------|------|----------|
| 流式输出兼容性 | 高 | 保留现有 StreamReader 接口 |
| 专家调用语义变化 | 中 | 保持 Tool 调用与现有正则解析语义一致 |
| 测试覆盖 | 中 | 为新模块编写完整测试 |

## Success Criteria

1. **Callbacks**
   - 所有事件通过统一 handler 发射
   - 支持日志、追踪集成
   - 测试覆盖率 > 80%

2. **Graph 编排**
   - Orchestrator 代码量减少 50%
   - 支持可视化调试（可选）
   - 所有现有测试通过

3. **专家协作**
   - 消除正则解析
   - 支持专家嵌套调用
   - 调用路径可追踪

## Timeline

| 阶段 | 内容 | 预估 |
|------|------|------|
| Phase 1 | Callbacks 统一 | 1-2 天 |
| Phase 2 | Graph 编排 | 2-3 天 |
| Phase 3 | 专家 Tool 化 | 1-2 天 |
| 测试 & 集成 | 端到端测试 | 1 天 |

**总计：5-8 天**
