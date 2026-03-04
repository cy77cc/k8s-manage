# Design: 多专家协作机制详细设计

## 架构概览

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         主从协作架构                                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                           配置层                                      │  │
│   │   scene_mappings.yaml                                                │  │
│   │   ┌───────────────────────────────────────────────────────────────┐ │  │
│   │   │ services:detail:                                               │ │  │
│   │   │   primary_expert: service_expert                               │ │  │
│   │   │   optional_helpers:          # 改名，强调可选                    │ │  │
│   │   │     - k8s_workload_expert                                      │ │  │
│   │   │     - topology_expert                                          │ │  │
│   │   │   strategy: primary_led        # 新策略：主专家主导              │ │  │
│   │   └───────────────────────────────────────────────────────────────┘ │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                         执行层                                       │  │
│   │                                                                       │  │
│   │   ┌───────────────────────────────────────────────────────────────┐ │  │
│   │   │                    PrimaryLedOrchestrator                      │ │  │
│   │   │                                                                │ │  │
│   │   │   Phase 1: 决策阶段                                            │ │  │
│   │   │   ├─ 主专家分析用户意图                                         │ │  │
│   │   │   └─ 输出 [REQUEST_HELPER: xxx] 或直接回答                      │ │  │
│   │   │                                                                │ │  │
│   │   │   Phase 2: 助手执行 (可选)                                      │ │  │
│   │   │   ├─ 发送 expert_progress 事件                                  │ │  │
│   │   │   ├─ 助手并行执行 (Generate，更快)                              │ │  │
│   │   │   └─ 静默收集结果                                               │ │  │
│   │   │                                                                │ │  │
│   │   │   Phase 3: 汇总输出                                            │ │  │
│   │   │   └─ 主专家流式输出最终回答                                      │ │  │
│   │   └───────────────────────────────────────────────────────────────┘ │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 详细设计

### 1. 配置结构调整

**scene_mappings.yaml**

```yaml
version: "1.0"

mappings:
  deployment:clusters:
    primary_expert: k8s_expert
    optional_helpers: [monitor_expert]  # 改名
    strategy: primary_led               # 新策略
    context_hints: [cluster_id]
    description: "集群管理场景"

  deployment:hosts:
    primary_expert: host_expert
    optional_helpers: [os_expert]
    strategy: primary_led
    context_hints: [host_id]
    description: "主机管理场景"

  services:detail:
    primary_expert: service_expert
    optional_helpers: [k8s_workload_expert, topology_expert]
    strategy: primary_led
    context_hints: [service_id]
    description: "服务详情场景"

  services:catalog:
    primary_expert: service_expert
    optional_helpers: []                # 无助手，单专家
    strategy: single
    context_hints: [category_id]
    description: "服务目录场景"

  # ... 其他场景
```

### 2. 类型定义

**internal/ai/experts/types.go**

```go
package experts

import (
    "time"
    "github.com/cloudwego/eino/schema"
)

// ExecutionStrategy 执行策略
type ExecutionStrategy string

const (
    StrategySingle     ExecutionStrategy = "single"      // 单专家
    StrategyPrimaryLed ExecutionStrategy = "primary_led" // 主专家主导
    StrategySequential ExecutionStrategy = "sequential"  // 串行 (兼容旧配置)
    StrategyParallel   ExecutionStrategy = "parallel"    // 并行 (兼容旧配置)
)

// RouteDecision 路由决策
type RouteDecision struct {
    PrimaryExpert   string            `json:"primary_expert"`
    OptionalHelpers []string          `json:"optional_helpers"` // 可选助手
    Strategy        ExecutionStrategy `json:"strategy"`
    Confidence      float64           `json:"confidence"`
    Source          string            `json:"source"`
}

// ExecuteRequest 执行请求
type ExecuteRequest struct {
    Message        string             `json:"message"`
    Decision       *RouteDecision     `json:"decision"`
    RuntimeContext map[string]any     `json:"runtime_context"`
    History        []*schema.Message  `json:"history"`        // 完整历史
    EventEmitter   ProgressEmitter    `json:"-"`              // 进度发射器
}

// ExecuteResult 执行结果
type ExecuteResult struct {
    Response string        `json:"response"`
    Traces   []ExpertTrace `json:"traces"`
    Metadata map[string]any `json:"metadata"`
}

// ExpertTrace 专家执行追踪
type ExpertTrace struct {
    ExpertName string        `json:"expert_name"`
    Role       string        `json:"role"`        // "primary" | "helper"
    Status     string        `json:"status"`      // "success" | "skipped" | "failed"
    Duration   time.Duration `json:"duration"`
    Output     string        `json:"output"`      // 仅用于内部传递，不发给前端
}

// ExpertProgressEvent 进度事件 (发送给前端)
type ExpertProgressEvent struct {
    Expert     string `json:"expert"`
    Status     string `json:"status"`     // "running" | "done"
    Task       string `json:"task,omitempty"`
    DurationMs int64  `json:"duration_ms,omitempty"`
}

// ProgressEmitter 进度发射器接口
type ProgressEmitter func(event string, payload any)

// HelperRequest 助手请求 (主专家输出)
type HelperRequest struct {
    ExpertName string `json:"expert_name"`
    Task       string `json:"task"`
}

// PrimaryDecision 主专家决策
type PrimaryDecision struct {
    NeedHelpers    bool            `json:"need_helpers"`
    HelperRequests []HelperRequest `json:"helper_requests,omitempty"`
    DirectAnswer   string          `json:"direct_answer,omitempty"`
}
```

### 3. Orchestrator 重写

**internal/ai/experts/orchestrator.go**

```go
package experts

import (
    "context"
    "errors"
    "fmt"
    "io"
    "strings"
    "sync"
    "time"

    "github.com/cloudwego/eino/schema"
)

type Orchestrator struct {
    registry   ExpertRegistry
    executor   *ExpertExecutor
    aggregator *ResultAggregator
}

func NewOrchestrator(registry ExpertRegistry, aggregator *ResultAggregator) *Orchestrator {
    return &Orchestrator{
        registry:   registry,
        executor:   NewExpertExecutor(registry),
        aggregator: aggregator,
    }
}

// StreamExecute 流式执行 - 主从协作模式
func (o *Orchestrator) StreamExecute(ctx context.Context, req *ExecuteRequest) (*schema.StreamReader[*schema.Message], error) {
    if req == nil || req.Decision == nil {
        return nil, fmt.Errorf("route decision is required")
    }

    // 单专家模式：直接流式输出
    if req.Decision.Strategy == StrategySingle || len(req.Decision.OptionalHelpers) == 0 {
        return o.streamSingleExpert(ctx, req)
    }

    // 主专家主导模式
    return o.streamPrimaryLed(ctx, req)
}

// streamSingleExpert 单专家流式执行
func (o *Orchestrator) streamSingleExpert(ctx context.Context, req *ExecuteRequest) (*schema.StreamReader[*schema.Message], error) {
    exp, ok := o.registry.GetExpert(req.Decision.PrimaryExpert)
    if !ok || exp == nil || exp.Agent == nil {
        return nil, fmt.Errorf("expert not found: %s", req.Decision.PrimaryExpert)
    }

    // 构建消息：历史 + 当前问题
    messages := o.buildMessagesWithHistory(req.History, req.Message)

    return exp.Agent.Stream(ctx, messages)
}

// streamPrimaryLed 主专家主导模式
func (o *Orchestrator) streamPrimaryLed(ctx context.Context, req *ExecuteRequest) (*schema.StreamReader[*schema.Message], error) {
    sr, sw := schema.Pipe[*schema.Message](64)

    go func() {
        defer sw.Close()

        // Phase 1: 主专家决策
        decision, err := o.primaryDecisionPhase(ctx, req, sw)
        if err != nil {
            sw.Send(schema.AssistantMessage(fmt.Sprintf("决策阶段失败: %v", err), nil), nil)
            return
        }

        // 主专家直接回答，不需要助手
        if !decision.NeedHelpers {
            return // 主专家已经在 Phase 1 输出了
        }

        // Phase 2: 助手执行 (静默)
        helperResults, err := o.helperExecutionPhase(ctx, req, decision.HelperRequests)
        if err != nil {
            // 助手失败不阻塞主流程
        }

        // Phase 3: 主专家汇总输出
        o.primarySummaryPhase(ctx, req, helperResults, sw)
    }()

    return sr, nil
}

// primaryDecisionPhase 主专家决策阶段
func (o *Orchestrator) primaryDecisionPhase(ctx context.Context, req *ExecuteRequest, sw *schema.StreamWriter[*schema.Message]) (*PrimaryDecision, error) {
    exp, ok := o.registry.GetExpert(req.Decision.PrimaryExpert)
    if !ok || exp == nil || exp.Agent == nil {
        return nil, fmt.Errorf("primary expert not found: %s", req.Decision.PrimaryExpert)
    }

    // 构建决策 prompt
    decisionPrompt := o.buildDecisionPrompt(req)

    // 使用 Generate 而非 Stream，等待完整决策
    messages := o.buildMessagesWithHistory(req.History, decisionPrompt)
    resp, err := exp.Agent.Generate(ctx, messages)
    if err != nil {
        return nil, err
    }

    content := ""
    if resp != nil {
        content = resp.Content
    }

    // 解析决策
    decision := o.parsePrimaryDecision(content, req.Decision.OptionalHelpers)

    // 如果不需要助手，直接流式输出主专家的回答
    if !decision.NeedHelpers {
        // 重新流式输出（因为上面用的是 Generate）
        streamMsgs := o.buildMessagesWithHistory(req.History, req.Message)
        stream, err := exp.Agent.Stream(ctx, streamMsgs)
        if err == nil {
            for {
                msg, recvErr := stream.Recv()
                if errors.Is(recvErr, io.EOF) {
                    break
                }
                if recvErr != nil {
                    break
                }
                if msg != nil {
                    sw.Send(msg, nil)
                }
            }
            stream.Close()
        }
    }

    return decision, nil
}

// helperExecutionPhase 助手执行阶段
func (o *Orchestrator) helperExecutionPhase(ctx context.Context, req *ExecuteRequest, helperRequests []HelperRequest) ([]ExpertResult, error) {
    if len(helperRequests) == 0 {
        return nil, nil
    }

    results := make([]ExpertResult, 0, len(helperRequests))
    var mu sync.Mutex

    // 发送进度事件
    o.emitProgress(req.EventEmitter, "helper_phase_start", map[string]any{
        "helpers": helperRequests,
    })

    // 并行执行助手
    var wg sync.WaitGroup
    for _, hr := range helperRequests {
        wg.Add(1)
        go func(helperReq HelperRequest) {
            defer wg.Done()

            // 发送开始事件
            o.emitProgress(req.EventEmitter, "expert_progress", ExpertProgressEvent{
                Expert: helperReq.ExpertName,
                Status: "running",
                Task:   helperReq.Task,
            })

            start := time.Now()

            // 执行助手
            result, err := o.executeHelper(ctx, req, helperReq)

            duration := time.Since(start)

            // 发送完成事件
            o.emitProgress(req.EventEmitter, "expert_progress", ExpertProgressEvent{
                Expert:     helperReq.ExpertName,
                Status:     "done",
                DurationMs: duration.Milliseconds(),
            })

            mu.Lock()
            if err != nil {
                results = append(results, ExpertResult{
                    ExpertName: helperReq.ExpertName,
                    Error:      err,
                    Duration:   duration,
                })
            } else {
                results = append(results, *result)
            }
            mu.Unlock()
        }(hr)
    }

    wg.Wait()

    o.emitProgress(req.EventEmitter, "helper_phase_done", map[string]any{
        "count": len(results),
    })

    return results, nil
}

// executeHelper 执行单个助手
func (o *Orchestrator) executeHelper(ctx context.Context, req *ExecuteRequest, helperReq HelperRequest) (*ExpertResult, error) {
    exp, ok := o.registry.GetExpert(helperReq.ExpertName)
    if !ok || exp == nil || exp.Agent == nil {
        return nil, fmt.Errorf("helper expert not found: %s", helperReq.ExpertName)
    }

    start := time.Now()

    // 构建助手消息：历史 + 任务
    taskPrompt := fmt.Sprintf("用户原始请求: %s\n\n你的任务: %s\n\n请执行分析，输出结果供主专家汇总。", req.Message, helperReq.Task)
    messages := o.buildMessagesWithHistory(req.History, taskPrompt)

    // 使用 Generate 而非 Stream，因为：
    // 1. 助手输出不需要流式传输到前端
    // 2. Generate 更快，省去流式开销
    // 3. 结果静默收集，供主专家汇总
    resp, err := exp.Agent.Generate(ctx, messages)
    if err != nil {
        return &ExpertResult{
            ExpertName: helperReq.ExpertName,
            Error:      err,
            Duration:   time.Since(start),
        }, err
    }

    output := ""
    if resp != nil {
        output = resp.Content
    }

    return &ExpertResult{
        ExpertName: helperReq.ExpertName,
        Output:     output,
        Duration:   time.Since(start),
    }, nil
}

// primarySummaryPhase 主专家汇总阶段
func (o *Orchestrator) primarySummaryPhase(ctx context.Context, req *ExecuteRequest, helperResults []ExpertResult, sw *schema.StreamWriter[*schema.Message]) {
    exp, ok := o.registry.GetExpert(req.Decision.PrimaryExpert)
    if !ok || exp == nil || exp.Agent == nil {
        sw.Send(schema.AssistantMessage("主专家不可用", nil), nil)
        return
    }

    // 构建汇总 prompt
    summaryPrompt := o.buildSummaryPrompt(req, helperResults)
    messages := o.buildMessagesWithHistory(req.History, summaryPrompt)

    // 流式输出
    stream, err := exp.Agent.Stream(ctx, messages)
    if err != nil {
        sw.Send(schema.AssistantMessage(fmt.Sprintf("汇总失败: %v", err), nil), nil)
        return
    }
    defer stream.Close()

    for {
        msg, recvErr := stream.Recv()
        if errors.Is(recvErr, io.EOF) {
            break
        }
        if recvErr != nil {
            break
        }
        if msg != nil {
            sw.Send(msg, nil)
        }
    }
}

// buildDecisionPrompt 构建决策 prompt
func (o *Orchestrator) buildDecisionPrompt(req *ExecuteRequest) string {
    var b strings.Builder
    b.WriteString("用户请求: ")
    b.WriteString(req.Message)
    b.WriteString("\n\n")

    if len(req.Decision.OptionalHelpers) > 0 {
        b.WriteString("你可以请求以下助手协助分析（仅在必要时调用）：\n")
        for _, helper := range req.Decision.OptionalHelpers {
            b.WriteString("- ")
            b.WriteString(helper)
            b.WriteString("\n")
        }
        b.WriteString("\n")
        b.WriteString("如果需要助手，请输出：[REQUEST_HELPER: 助手名称: 任务描述]\n")
        b.WriteString("例如：[REQUEST_HELPER: k8s_expert: 检查Pod状态]\n")
        b.WriteString("如果不需要助手，直接回答用户问题即可。\n")
        b.WriteString("请先决策，不要输出其他内容。\n")
    }

    return b.String()
}

// buildSummaryPrompt 构建汇总 prompt
func (o *Orchestrator) buildSummaryPrompt(req *ExecuteRequest, helperResults []ExpertResult) string {
    var b strings.Builder
    b.WriteString("用户请求: ")
    b.WriteString(req.Message)
    b.WriteString("\n\n")

    if len(helperResults) > 0 {
        b.WriteString("助手分析结果：\n")
        for _, result := range helperResults {
            b.WriteString("\n【")
            b.WriteString(result.ExpertName)
            b.WriteString("】\n")
            if result.Error != nil {
                b.WriteString("执行失败: ")
                b.WriteString(result.Error.Error())
            } else {
                b.WriteString(result.Output)
            }
            b.WriteString("\n")
        }
        b.WriteString("\n请基于以上分析结果，给用户一个完整、连贯的回答。\n")
    }

    return b.String()
}

// parsePrimaryDecision 解析主专家决策
func (o *Orchestrator) parsePrimaryDecision(content string, availableHelpers []string) *PrimaryDecision {
    decision := &PrimaryDecision{
        NeedHelpers: false,
    }

    // 解析 [REQUEST_HELPER: expert: task] 格式
    pattern := regexp.MustCompile(`\[REQUEST_HELPER:\s*([a-zA-Z0-9_]+):\s*([^\]]+)\]`)
    matches := pattern.FindAllStringSubmatch(content, -1)

    for _, match := range matches {
        if len(match) >= 3 {
            expertName := strings.TrimSpace(match[1])
            task := strings.TrimSpace(match[2])

            // 验证专家是否在可用列表中
            for _, h := range availableHelpers {
                if h == expertName {
                    decision.NeedHelpers = true
                    decision.HelperRequests = append(decision.HelperRequests, HelperRequest{
                        ExpertName: expertName,
                        Task:       task,
                    })
                    break
                }
            }
        }
    }

    return decision
}

// buildMessagesWithHistory 构建带历史的消息
func (o *Orchestrator) buildMessagesWithHistory(history []*schema.Message, currentMessage string) []*schema.Message {
    messages := make([]*schema.Message, 0, len(history)+1)

    // 添加历史消息（限制最近10轮）
    maxHistory := 10
    start := 0
    if len(history) > maxHistory {
        start = len(history) - maxHistory
    }
    for i := start; i < len(history); i++ {
        if history[i] != nil {
            messages = append(messages, history[i])
        }
    }

    // 添加当前消息
    messages = append(messages, schema.UserMessage(currentMessage))

    return messages
}

// emitProgress 发送进度事件
func (o *Orchestrator) emitProgress(emitter ProgressEmitter, event string, payload any) {
    if emitter != nil {
        emitter(event, payload)
    }
}
```

### 4. Executor 调整

**internal/ai/experts/executor.go**

```go
// StreamStep 流式执行 - 传递历史上下文
func (e *ExpertExecutor) StreamStep(ctx context.Context, step *ExecutionStep, req *ExecuteRequest) (*schema.StreamReader[*schema.Message], error) {
    if step == nil {
        return nil, fmt.Errorf("execution step is nil")
    }

    exp, ok := e.registry.GetExpert(step.ExpertName)
    if !ok || exp == nil {
        return nil, fmt.Errorf("expert not found: %s", step.ExpertName)
    }

    if exp.Agent == nil {
        return schema.StreamReaderFromArray([]*schema.Message{
            schema.AssistantMessage("专家模型未初始化", nil),
        }), nil
    }

    // 关键修改：传递历史上下文
    messages := e.buildExpertMessages(req.History, step, req.Message)

    return exp.Agent.Stream(ctx, messages)
}

// buildExpertMessages 构建专家消息（包含历史）
func (e *ExpertExecutor) buildExpertMessages(history []*schema.Message, step *ExecutionStep, baseMessage string) []*schema.Message {
    messages := make([]*schema.Message, 0, len(history)+1)

    // 添加历史（限制数量）
    maxHistory := 10
    start := 0
    if len(history) > maxHistory {
        start = len(history) - maxHistory
    }
    for i := start; i < len(history); i++ {
        if history[i] != nil {
            messages = append(messages, history[i])
        }
    }

    // 构建当前任务消息
    var taskMsg strings.Builder
    taskMsg.WriteString("用户请求:\n")
    taskMsg.WriteString(baseMessage)
    taskMsg.WriteString("\n")

    if step.Task != "" {
        taskMsg.WriteString("\n当前任务:\n")
        taskMsg.WriteString(step.Task)
        taskMsg.WriteString("\n")
    }

    messages = append(messages, schema.UserMessage(taskMsg.String()))

    return messages
}
```

### 5. Router 调整

**internal/ai/experts/router.go**

```go
func (r *HybridRouter) routeByScene(scene string) *RouteDecision {
    key := normalizeSceneKey(scene)
    if key == "" {
        return nil
    }

    item, ok := r.sceneMappings[key]
    if !ok || strings.TrimSpace(item.PrimaryExpert) == "" {
        return nil
    }

    strategy := item.Strategy
    if strategy == "" {
        strategy = StrategyPrimaryLed  // 默认使用主专家主导
    }

    return &RouteDecision{
        PrimaryExpert:   item.PrimaryExpert,
        OptionalHelpers: append([]string{}, item.OptionalHelpers...), // 改名
        Strategy:        strategy,
        Confidence:      1.0,
        Source:          "scene",
    }
}
```

## SSE 事件流

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         前端事件流示例                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   event: meta                                                               │
│   data: {"sessionId":"sess-xxx"}                                            │
│                                                                             │
│   event: expert_progress          ← 助手开始执行                              │
│   data: {"expert":"k8s_expert","status":"running","task":"检查Pod状态"}      │
│                                                                             │
│   event: expert_progress          ← 助手完成                                 │
│   data: {"expert":"k8s_expert","status":"done","duration_ms":1200}          │
│                                                                             │
│   event: expert_progress                                                     │
│   data: {"expert":"topology_expert","status":"running","task":"分析依赖"}    │
│                                                                             │
│   event: expert_progress                                                     │
│   data: {"expert":"topology_expert","status":"done","duration_ms":800}      │
│                                                                             │
│   event: delta                    ← 主专家流式输出                            │
│   data: {"contentChunk":"## 服务状态分析\n"}                                 │
│                                                                             │
│   event: delta                                                               │
│   data: {"contentChunk":"**payment-api** 服务运行正常..."}                   │
│                                                                             │
│   event: done                                                                │
│   data: {"session":{...},"stream_state":"ok"}                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 前端适配

**ChatInterface.tsx** 新增事件处理：

```tsx
// 处理 expert_progress 事件
case 'expert_progress':
  const progress = payload as ExpertProgressEvent;
  if (progress.status === 'running') {
    setExpertProgress(prev => [...prev, {
      expert: progress.expert,
      task: progress.task,
      status: 'running'
    }]);
  } else if (progress.status === 'done') {
    setExpertProgress(prev => prev.map(p =>
      p.expert === progress.expert
        ? { ...p, status: 'done', durationMs: progress.duration_ms }
        : p
    ));
  }
  break;

// 渲染进度动画
{expertProgress.length > 0 && (
  <div className="expert-progress">
    {expertProgress.map(p => (
      <div key={p.expert} className={p.status}>
        {p.status === 'running' ? (
          <Spinner /> // 动画
        ) : (
          <CheckIcon /> // 完成
        )}
        <span>{p.expert}: {p.task}</span>
      </div>
    ))}
  </div>
)}
```

## 兼容性

### 旧配置兼容

```yaml
# 旧配置仍然支持
services:detail:
  primary_expert: service_expert
  helper_experts: [k8s_expert]    # 自动转为 optional_helpers
  strategy: sequential            # 自动转为 primary_led

# 新配置
services:detail:
  primary_expert: service_expert
  optional_helpers: [k8s_expert]
  strategy: primary_led
```

### 降级策略

1. 主专家不可用 → 使用默认 agent
2. 助手执行失败 → 主专家仍可输出
3. 前端不支持 expert_progress → 正常输出，无动画
