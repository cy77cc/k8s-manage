# Multi-Domain Agent Architecture Design

## Goals

1. **领域隔离**：每个领域的 Planner 只关注自己的工具，减少选择压力
2. **并行执行**：多个领域的 Plan 可以并行生成，Executor 按依赖顺序执行
3. **职责分离**：Planner 只规划不执行，Executor 统一执行
4. **工具分层**：Discovery 工具用于参数补全，Action 工具用于真正操作
5. **跨领域协作**：通过变量引用语法实现领域间数据传递

## Non-Goals

1. **不改变前端交互**：SSE 事件类型保持兼容，前端无感知
2. **不改变 API 接口**：内部架构重构，对外接口不变
3. **不改变工具定义**：现有工具只需添加元数据标记，无需重写
4. **不支持嵌套领域**：领域是扁平的，不支持子领域

## Architecture Overview

```
┌────────────────────────────────────────────────────────────────────────┐
│                              Entry Point                                │
│                          HybridAgent.Query()                            │
└───────────────────────────────┬────────────────────────────────────────┘
                                │
                                ▼
┌────────────────────────────────────────────────────────────────────────┐
│                         Orchestrator Planner                            │
│                                                                         │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                │
│  │ Intent      │───▶│ Domain      │───▶│ Dispatch    │                │
│  │ Analysis    │    │ Selection   │    │ Plan        │                │
│  └─────────────┘    └─────────────┘    └─────────────┘                │
│                                                                         │
│  Input:  UserMessage                                                   │
│  Output: []DomainRequest                                               │
└───────────────────────────────┬────────────────────────────────────────┘
                                │
          ┌─────────────────────┼─────────────────────┐
          │                     │                     │
          ▼                     ▼                     ▼
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│ InfraPlanner     │  │ ServicePlanner   │  │ MonitorPlanner   │
│                  │  │                  │  │                  │
│ Discovery Tools: │  │ Discovery Tools: │  │ Discovery Tools: │
│ - host_list      │  │ - service_list   │  │ - (无)           │
│ - cluster_list   │  │ - deploy_target  │  │                  │
│ - k8s_query      │  │ - credential_list│  │                  │
│                  │  │                  │  │                  │
│ Output:          │  │ Output:          │  │ Output:          │
│ DomainPlan       │  │ DomainPlan       │  │ DomainPlan       │
└────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘
         │                     │                     │
         └─────────────────────┼─────────────────────┘
                               │
                               ▼
┌────────────────────────────────────────────────────────────────────────┐
│                              Executor                                   │
│                                                                         │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                │
│  │ Plan        │───▶│ DAG         │───▶│ Execution   │                │
│  │ Merger      │    │ Builder     │    │ Engine      │                │
│  └─────────────┘    └─────────────┘    └─────────────┘                │
│                                                                         │
│  All Tools (Discovery + Action)                                        │
│                                                                         │
│  Input:  []DomainPlan                                                  │
│  Output: ExecutionResult                                               │
└───────────────────────────────┬────────────────────────────────────────┘
                                │
                                ▼
┌────────────────────────────────────────────────────────────────────────┐
│                              Replanner                                  │
│                                                                         │
│  ┌─────────────┐    ┌─────────────┐                                    │
│  │ Result      │───▶│ Decision    │────────▶ [END / Re-orchestrate]    │
│  │ Validator   │    │ Maker       │                                    │
│  └─────────────┘    └─────────────┘                                    │
│                                                                         │
│  Input:  ExecutionResult                                               │
│  Output: ReplanDecision                                                │
└────────────────────────────────────────────────────────────────────────┘
```

## Component Design

### 1. Orchestrator Planner

**职责**：
- 分析用户意图
- 识别涉及的领域
- 构造领域请求并分发

**实现方式**：使用 LLM + 领域描述 Prompt

```go
// internal/ai/orchestrator/planner.go

type OrchestratorPlanner struct {
    model       model.ToolCallingChatModel
    domains     map[string]DomainDescriptor
}

type DomainDescriptor struct {
    Name        string
    Description string
    Keywords    []string
    Tools       []string
}

type DomainRequest struct {
    Domain      string
    UserIntent  string
    Context     map[string]any
}

func (p *OrchestratorPlanner) Plan(ctx context.Context, message string) ([]DomainRequest, error) {
    // 1. 构造 Prompt，包含所有领域描述
    prompt := p.buildDomainSelectionPrompt(message)

    // 2. 调用 LLM 进行领域选择
    response, err := p.model.Generate(ctx, prompt)

    // 3. 解析输出，提取涉及的领域
    domains := p.parseDomainSelection(response)

    // 4. 构造 DomainRequest
    requests := make([]DomainRequest, len(domains))
    for i, d := range domains {
        requests[i] = DomainRequest{
            Domain:     d.Name,
            UserIntent: message,
            Context:    d.Context,
        }
    }
    return requests, nil
}
```

**Prompt 示例**：

```
你是一个任务分发器。分析用户请求，识别涉及哪些领域。

可用领域：
- infrastructure: 主机管理、K8s集群、容器运行时、操作系统状态
- service: 服务管理、部署目标、服务目录、服务部署
- cicd: 流水线管理、构建任务、发布流程
- monitor: 监控指标、告警管理、服务拓扑
- config: 配置管理、配置项、配置对比
- user: 用户管理、角色权限、审计日志

用户请求：{message}

输出 JSON 格式：
{
  "domains": [
    {"name": "infrastructure", "context": {"focus": "主机状态"}},
    {"name": "monitor", "context": {"focus": "告警检查"}}
  ]
}
```

### 2. Domain Planner

**职责**：
- 接收领域请求
- 规划执行步骤
- 声明步骤依赖
- 输出变量引用

**实现方式**：使用 eino 的 ReAct Agent 或自定义 Graph

```go
// internal/ai/planner/domain_planner.go

type DomainPlanner interface {
    Domain() string
    Plan(ctx context.Context, req DomainRequest) (*DomainPlan, error)
}

type DomainPlan struct {
    Domain string
    Steps  []PlanStep
}

type PlanStep struct {
    ID        string         `json:"id"`
    Tool      string         `json:"tool"`
    Params    map[string]any `json:"params"`
    DependsOn []string       `json:"depends_on,omitempty"`
    Produces  []string       `json:"produces,omitempty"`
    Requires  []string       `json:"requires,omitempty"`
}

// ============ 具体实现 ============

type ServicePlanner struct {
    model          model.ToolCallingChatModel
    discoveryTools map[string]tool.InvokableTool
}

func (p *ServicePlanner) Plan(ctx context.Context, req DomainRequest) (*DomainPlan, error) {
    // 1. 调用 Discovery 工具补全参数（如果需要）
    enrichedContext := p.enrichContext(ctx, req)

    // 2. 使用 LLM 生成计划
    prompt := p.buildPlanningPrompt(req.UserIntent, enrichedContext)
    response, err := p.model.Generate(ctx, prompt)

    // 3. 解析为 DomainPlan
    plan := p.parsePlan(response)

    return plan, nil
}

func (p *ServicePlanner) enrichContext(ctx context.Context, req DomainRequest) map[string]any {
    // 调用 Discovery 工具获取 service_id, cluster_id 等
    // 例如：service_list_inventory(keyword="支付")
    // 这些调用发生在规划阶段，由 Planner 执行
}
```

**Plan 输出示例**：

```json
{
  "domain": "service",
  "steps": [
    {
      "id": "get_service",
      "tool": "service_list_inventory",
      "params": {"keyword": "支付服务"},
      "produces": ["service_id"]
    },
    {
      "id": "get_cluster",
      "tool": "cluster_list_inventory",
      "params": {"env": "production"},
      "produces": ["cluster_id"]
    },
    {
      "id": "deploy",
      "tool": "service_deploy_apply",
      "params": {
        "service_id": {"$ref": "service.get_service.service_id"},
        "cluster_id": {"$ref": "service.get_cluster.cluster_id"}
      },
      "depends_on": ["get_service", "get_cluster"],
      "produces": ["deployment_id"]
    }
  ]
}
```

### 3. Executor

**职责**：
- 合并所有 DomainPlan
- 构建全局 DAG
- 按拓扑顺序执行
- 处理变量引用

**实现方式**：自定义执行引擎

```go
// internal/ai/executor/executor.go

type Executor struct {
    tools     map[string]tool.InvokableTool
    registry  *ToolRegistry
}

type ExecutionContext struct {
    Plans       []DomainPlan
    StepResults map[string]StepResult  // step_id -> result
    DAG         *ExecutionDAG
}

type StepResult struct {
    ID     string
    Output map[string]any
    Error  error
}

func (e *Executor) Execute(ctx context.Context, plans []DomainPlan) (*ExecutionResult, error) {
    // 1. 合并所有步骤
    allSteps := e.mergeSteps(plans)

    // 2. 构建 DAG
    dag := e.buildDAG(allSteps)

    // 3. 拓扑排序
    order := dag.TopologicalSort()

    // 4. 按顺序执行
    execCtx := &ExecutionContext{
        Plans:       plans,
        StepResults: make(map[string]StepResult),
        DAG:         dag,
    }

    for _, stepID := range order {
        step := dag.GetStep(stepID)

        // 解析变量引用
        resolvedParams := e.resolveParams(step.Params, execCtx)

        // 执行工具
        result, err := e.executeStep(ctx, step, resolvedParams)
        execCtx.StepResults[stepID] = StepResult{
            ID:     stepID,
            Output: result,
            Error:  err,
        }

        if err != nil {
            // 错误处理策略：停止 / 跳过 / 重试
            if e.shouldStop(step, err) {
                break
            }
        }
    }

    return e.buildResult(execCtx), nil
}

func (e *Executor) resolveParams(params map[string]any, ctx *ExecutionContext) map[string]any {
    resolved := make(map[string]any)
    for k, v := range params {
        if ref, ok := v.(map[string]any); ok && ref["$ref"] != nil {
            // 解析 {$ref: "domain.step_id.field"}
            refPath := ref["$ref"].(string)
            resolved[k] = e.resolveRef(refPath, ctx)
        } else {
            resolved[k] = v
        }
    }
    return resolved
}

func (e *Executor) resolveRef(refPath string, ctx *ExecutionContext) any {
    // refPath 格式: "domain.step_id.field"
    // 例如: "service.get_service.service_id"
    parts := strings.Split(refPath, ".")
    if len(parts) < 3 {
        return nil
    }

    stepID := parts[1]  // get_service
    field := parts[2]   // service_id

    result, ok := ctx.StepResults[stepID]
    if !ok {
        return nil
    }

    return result.Output[field]
}
```

**DAG 构建逻辑**：

```go
// internal/ai/executor/dag.go

type ExecutionDAG struct {
    steps map[string]*PlanStep
    edges map[string][]string  // step_id -> dependent step_ids
}

func (e *Executor) buildDAG(steps []PlanStep) *ExecutionDAG {
    dag := &ExecutionDAG{
        steps: make(map[string]*PlanStep),
        edges: make(map[string][]string),
    }

    // 1. 收集所有步骤
    for i := range steps {
        dag.steps[steps[i].ID] = &steps[i]
    }

    // 2. 建立依赖边
    for _, step := range steps {
        // 领域内依赖
        for _, dep := range step.DependsOn {
            dag.edges[dep] = append(dag.edges[dep], step.ID)
        }

        // 跨领域依赖（通过 $ref 解析）
        for _, v := range step.Params {
            if ref, ok := v.(map[string]any); ok && ref["$ref"] != nil {
                refPath := ref["$ref"].(string)
                depStepID := extractStepID(refPath)  // 解析出 step_id
                if depStepID != step.ID {
                    dag.edges[depStepID] = append(dag.edges[depStepID], step.ID)
                }
            }
        }
    }

    return dag
}
```

### 4. Replanner

**职责**：
- 验证执行结果完整性
- 决定是否需要重规划
- 提供重规划建议

```go
// internal/ai/replanner/replanner.go

type Replanner struct {
    model model.ToolCallingChatModel
}

type ReplanDecision struct {
    NeedReplan bool
    Reason     string
    Suggestions []string
}

func (r *Replanner) Evaluate(ctx context.Context, result *ExecutionResult) (*ReplanDecision, error) {
    // 1. 检查是否有失败步骤
    if result.HasErrors() {
        return &ReplanDecision{
            NeedReplan: true,
            Reason:     "部分步骤执行失败",
            Suggestions: r.analyzeFailures(result),
        }, nil
    }

    // 2. 检查结果是否符合预期
    prompt := r.buildValidationPrompt(result)
    response, err := r.model.Generate(ctx, prompt)

    // 3. 解析 LLM 判断
    decision := r.parseDecision(response)

    return decision, nil
}
```

## Tool Layering

### Discovery vs Action 分类

```go
// internal/ai/tools/classification.go

type ToolCategory string

const (
    ToolCategoryDiscovery ToolCategory = "discovery"  // 只读查询
    ToolCategoryAction    ToolCategory = "action"     // 执行操作
)

var ToolClassifications = map[string]ToolCategory{
    // Discovery - 只读查询
    "host_list_inventory":       ToolCategoryDiscovery,
    "cluster_list_inventory":    ToolCategoryDiscovery,
    "service_list_inventory":    ToolCategoryDiscovery,
    "service_catalog_list":      ToolCategoryDiscovery,
    "deployment_target_list":    ToolCategoryDiscovery,
    "credential_list":           ToolCategoryDiscovery,
    "cicd_pipeline_list":        ToolCategoryDiscovery,
    "job_list":                  ToolCategoryDiscovery,
    "user_list":                 ToolCategoryDiscovery,
    "role_list":                 ToolCategoryDiscovery,
    "k8s_query":                 ToolCategoryDiscovery,
    "k8s_list_resources":        ToolCategoryDiscovery,
    "k8s_events":                ToolCategoryDiscovery,
    "k8s_logs":                  ToolCategoryDiscovery,
    "monitor_alert":             ToolCategoryDiscovery,
    "monitor_metric":            ToolCategoryDiscovery,
    "config_app_list":           ToolCategoryDiscovery,
    "config_item_get":           ToolCategoryDiscovery,
    "audit_log_search":          ToolCategoryDiscovery,
    "topology_get":              ToolCategoryDiscovery,

    // Action - 执行操作
    "service_deploy_apply":      ToolCategoryAction,
    "host_exec":                 ToolCategoryAction,
    "host_batch_exec_apply":     ToolCategoryAction,
    "cicd_pipeline_trigger":     ToolCategoryAction,
    "job_run":                   ToolCategoryAction,
    "host_batch_status_update":  ToolCategoryAction,
}
```

### Tool Registry 按领域组织

```go
// internal/ai/tools/registry.go

type ToolRegistry struct {
    byDomain map[string][]ToolInfo
    byName   map[string]ToolInfo
}

type ToolInfo struct {
    Name       string
    Domain     string
    Category   ToolCategory
    Tool       tool.InvokableTool
    Meta       ToolMeta
}

func NewToolRegistry(tools []RegisteredTool) *ToolRegistry {
    r := &ToolRegistry{
        byDomain: make(map[string][]ToolInfo),
        byName:   make(map[string]ToolInfo),
    }

    for _, t := range tools {
        domain := inferDomain(t.Meta.Name)
        category := ToolClassifications[t.Meta.Name]
        if category == "" {
            category = ToolCategoryAction  // 默认为 Action
        }

        info := ToolInfo{
            Name:     t.Meta.Name,
            Domain:   domain,
            Category: category,
            Tool:     t.Tool,
            Meta:     t.Meta,
        }

        r.byDomain[domain] = append(r.byDomain[domain], info)
        r.byName[t.Meta.Name] = info
    }

    return r
}

func inferDomain(toolName string) string {
    switch {
    case strings.HasPrefix(toolName, "host_") ||
         strings.HasPrefix(toolName, "k8s_") ||
         strings.HasPrefix(toolName, "os_"):
        return "infrastructure"
    case strings.HasPrefix(toolName, "service_") ||
         strings.HasPrefix(toolName, "deployment_") ||
         strings.HasPrefix(toolName, "credential_"):
        return "service"
    case strings.HasPrefix(toolName, "cicd_") ||
         strings.HasPrefix(toolName, "job_"):
        return "cicd"
    case strings.HasPrefix(toolName, "monitor_") ||
         strings.HasPrefix(toolName, "topology_"):
        return "monitor"
    case strings.HasPrefix(toolName, "config_"):
        return "config"
    case strings.HasPrefix(toolName, "user_") ||
         strings.HasPrefix(toolName, "role_") ||
         strings.HasPrefix(toolName, "permission_") ||
         strings.HasPrefix(toolName, "audit_"):
        return "user"
    default:
        return "general"
    }
}
```

## eino Graph Integration

### Graph Structure

```go
// internal/ai/graph/orchestrator_graph.go

func BuildOrchestratorGraph(ctx context.Context, cfg *GraphConfig) (*compose.Graph[UserInput, ExecutionResult], error) {
    g := compose.NewGraph[UserInput, ExecutionResult]()

    // Node 1: Orchestrator Planner
    orchestrator := NewOrchestratorPlanner(cfg.OrchestratorModel, cfg.Domains)
    g.AddLambdaNode("orchestrator", compose.InvokableLambdaWithOption(
        func(ctx context.Context, input UserInput, opts ...any) ([]DomainRequest, error) {
            return orchestrator.Plan(ctx, input.Message)
        },
    ))

    // Node 2: Domain Planners (并行)
    plannersNode := NewDomainPlannersNode(cfg.Planners)
    g.AddLambdaNode("planners", compose.InvokableLambdaWithOption(
        plannersNode.Execute,
    ))

    // Node 3: Executor
    executor := NewExecutor(cfg.Tools)
    g.AddLambdaNode("executor", compose.InvokableLambdaWithOption(
        func(ctx context.Context, plans []DomainPlan, opts ...any) (*ExecutionResult, error) {
            return executor.Execute(ctx, plans)
        },
    ))

    // Node 4: Replanner
    replanner := NewReplanner(cfg.ReplannerModel)
    g.AddLambdaNode("replanner", compose.InvokableLambdaWithOption(
        func(ctx context.Context, result *ExecutionResult, opts ...any) (*ReplanDecision, error) {
            return replanner.Evaluate(ctx, result)
        },
    ))

    // Edges
    g.AddEdge(compose.START, "orchestrator")
    g.AddEdge("orchestrator", "planners")
    g.AddEdge("planners", "executor")
    g.AddEdge("executor", "replanner")
    g.AddEdge("replanner", compose.END)

    return g, nil
}

// DomainPlannersNode 并行执行多个 Planner
type DomainPlannersNode struct {
    planners map[string]DomainPlanner
}

func (n *DomainPlannersNode) Execute(ctx context.Context, requests []DomainRequest) ([]DomainPlan, error) {
    plans := make([]DomainPlan, len(requests))

    var wg sync.WaitGroup
    var mu sync.Mutex
    errs := make([]error, len(requests))

    for i, req := range requests {
        wg.Add(1)
        go func(idx int, r DomainRequest) {
            defer wg.Done()

            planner, ok := n.planners[r.Domain]
            if !ok {
                errs[idx] = fmt.Errorf("no planner for domain: %s", r.Domain)
                return
            }

            plan, err := planner.Plan(ctx, r)
            if err != nil {
                errs[idx] = err
                return
            }

            mu.Lock()
            plans[idx] = *plan
            mu.Unlock()
        }(i, req)
    }

    wg.Wait()

    for _, err := range errs {
        if err != nil {
            return nil, err
        }
    }

    return plans, nil
}
```

## File Structure

```
internal/ai/
├── orchestrator/
│   ├── planner.go           # Orchestrator Planner 实现
│   ├── prompt.go            # 领域选择 Prompt
│   └── types.go             # DomainRequest, DomainDescriptor
│
├── planner/
│   ├── interface.go         # DomainPlanner 接口
│   ├── registry.go          # Planner 注册表
│   ├── infrastructure.go    # Infra Planner 实现
│   ├── service.go           # Service Planner 实现
│   ├── cicd.go              # CICD Planner 实现
│   ├── monitor.go           # Monitor Planner 实现
│   ├── config.go            # Config Planner 实现
│   ├── user.go              # User Planner 实现
│   └── prompt.go            # 规划 Prompt 模板
│
├── executor/
│   ├── executor.go          # Executor 主入口
│   ├── dag.go               # DAG 构建与拓扑排序
│   ├── resolver.go          # 变量引用解析
│   └── types.go             # ExecutionContext, StepResult
│
├── replanner/
│   ├── replanner.go         # Replanner 实现
│   ├── validator.go         # 结果验证逻辑
│   └── types.go             # ReplanDecision
│
├── graph/
│   ├── orchestrator_graph.go # 主 Graph 构建
│   └── planners_node.go      # 并行 Planner 节点
│
├── tools/
│   ├── registry.go          # 按领域组织的 Tool Registry
│   ├── classification.go    # Discovery/Action 分类
│   └── ... (现有工具实现)
│
├── modes/
│   └── agentic.go           # 适配新架构入口
│
└── hybrid.go                # HybridAgent 调用链调整
```

## Migration Strategy

### Phase 1: 并行运行

```go
// 内部配置开关
type AgentConfig struct {
    UseMultiDomainArch bool `yaml:"use_multi_domain_arch"`
}

func (a *HybridAgent) Query(ctx context.Context, sessionID, message string) *adk.AsyncIterator[*types.AgentResult] {
    if a.cfg.UseMultiDomainArch {
        return a.queryMultiDomain(ctx, sessionID, message)
    }
    return a.queryLegacy(ctx, sessionID, message)
}
```

### Phase 2: 逐步迁移

1. 先实现 Orchestrator + Executor，使用现有 Planner 逻辑
2. 逐步添加 Domain Planner
3. 移除旧架构

### Rollback

通过配置开关随时切回旧架构。

## Decisions

| Decision | Choice | Rationale | Alternatives |
|----------|--------|-----------|--------------|
| Planner 是否执行工具 | 否，只规划 | 保持 Planner 无状态，易测试 | Planner 内部执行 Discovery |
| 领域间数据传递 | 变量引用语法 `$ref` | 统一处理，Executor 负责 | Orchestrator 负责传递 |
| 并行策略 | 领域级并行 | 粒度适中，实现简单 | 步骤级并行（更复杂） |
| Graph 结构 | 线性流水线 | 简单可控 | 带循环的 Graph（复杂） |

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Planner 并行延迟 | 高 | 使用较小模型，Prompt 优化 |
| 变量引用解析失败 | 中 | 执行前校验 DAG，提前报错 |
| 跨领域依赖复杂 | 中 | 限制跨领域依赖层级，文档规范 |
| 迁移期间功能回退 | 高 | 配置开关，并行运行验证 |
