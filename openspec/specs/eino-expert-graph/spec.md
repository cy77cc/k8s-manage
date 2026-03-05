# Spec: Eino Expert Graph

## 概述

使用 eino compose.Graph 实现声明式专家编排，替代手动 Orchestrator。

## 图结构

```
START ──▶ primary ──┬──▶ helpers_parallel ────┬──▶ aggregate ──▶ END
                     │                          │
                     ├──▶ helpers_sequential ───┤
                     │                          │
                     └──▶ [skip] ───────────────┘
```

## 类型定义

### GraphInput

```go
type GraphInput struct {
    Message        string               // 用户消息
    History        []*schema.Message    // 对话历史
    Decision       *RouteDecision       // 路由决策
    RuntimeContext map[string]any       // 运行时上下文
}
```

### GraphOutput

```go
type GraphOutput struct {
    Response string          // 最终响应
    Traces   []ExpertTrace   // 执行追踪
    Metadata map[string]any  // 元数据
}
```

## 节点规范

### primary 节点

- **输入**: GraphInput
- **输出**: GraphOutput (包含主专家响应)
- **职责**: 执行主专家分析

### helpers_parallel 节点

- **输入**: GraphInput
- **输出**: GraphOutput (包含助手追踪)
- **职责**: 并行执行所有助手专家

### helpers_sequential 节点

- **输入**: GraphInput
- **输出**: GraphOutput (包含助手追踪)
- **职责**: 顺序执行助手专家，支持上下文传递

### aggregate 节点

- **输入**: GraphOutput (上游结果)
- **输出**: GraphOutput (汇总结果)
- **职责**: 汇总多个专家结果

## 分支策略

```go
// 条件分支：根据策略选择执行路径
func helperStrategyBranch() *compose.GraphBranch {
    return compose.NewGraphBranch(
        func(ctx context.Context, input *GraphInput) (string, error) {
            if len(input.Decision.OptionalHelpers) == 0 {
                return "skip_helpers", nil
            }
            switch input.Decision.Strategy {
            case StrategyParallel:
                return "helpers_parallel", nil
            case StrategySequential:
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

## 执行策略

| 策略 | 节点选择 | 说明 |
|------|----------|------|
| single | skip_helpers | 仅主专家 |
| primary_led | helpers_parallel | 主专家协调，助手并行 |
| parallel | helpers_parallel | 所有专家并行 |
| sequential | helpers_sequential | 顺序执行 |

## 与现有架构对比

### 旧架构 (Orchestrator)

```go
func (o *Orchestrator) StreamExecute(ctx, req) (*StreamReader, error) {
    if strategy == StrategyPrimaryLed {
        return o.streamPrimaryLed(ctx, req)
    }
    // 手动 goroutine 管理
    sr, sw := schema.Pipe(64)
    go func() {
        defer sw.Close()
        // 手动流程控制...
    }()
    return sr, nil
}
```

### 新架构 (Graph)

```go
func (b *Builder) Build() *compose.Graph {
    g := compose.NewGraph()
    g.AddEdge(compose.START, "primary")
    g.AddBranch("primary", b.helperStrategyBranch())
    g.AddEdge("aggregate", compose.END)
    return g
}

// 执行
runnable, _ := graph.Compile(ctx)
stream, _ := runnable.Stream(ctx, input)
```

## 优势

1. **声明式** - 图结构清晰可见
2. **可测试** - 每个节点可独立测试
3. **可扩展** - 新增节点/分支简单
4. **可观测** - 支持 callbacks 追踪
