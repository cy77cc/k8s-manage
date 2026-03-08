# AI Module Design

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         AI-PaaS Architecture                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   用户输入 / 系统告警                                                        │
│         │                                                                   │
│         ▼                                                                   │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                     IntentRouter (compose.Graph)                    │   │
│   │                                                                     │   │
│   │   PreProcess ──▶ Classify ──▶ Branch ──▶ Domain Graphs             │   │
│   │                                    │                                │   │
│   │                    ┌───────────────┼───────────────┐               │   │
│   │                    ▼               ▼               ▼               │   │
│   │               Infrastructure    Service       Monitor              │   │
│   │                    │               │               │               │   │
│   │                    └───────────────┼───────────────┘               │   │
│   │                                    ▼                                │   │
│   │                            General_Assistance                      │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                        │                                    │
│                                        ▼                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                     ActionGraph (compose.Workflow)                  │   │
│   │                                                                     │   │
│   │   Sanitize ──▶ Reasoning ──▶ Validation ──▶ Execution              │   │
│   │    Lambda        ChatModel       Lambda        ToolsNode           │   │
│   │                                                                     │   │
│   │   ┌─────────────────────────────────────────────────────────────┐  │   │
│   │   │                      State                                   │  │   │
│   │   │   - messages []*schema.Message                               │  │   │
│   │   │   - toolResults map[string]any                               │  │   │
│   │   │   - validationErrors []error                                 │  │   │
│   │   └─────────────────────────────────────────────────────────────┘  │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                        │                                    │
│                                        ▼                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                     SecurityAspect (Callbacks)                      │   │
│   │                                                                     │   │
│   │   - OnToolCallStart: 权限检查 + Dry-run                            │   │
│   │   - Interrupt: Human-in-the-loop for high-risk                     │   │
│   │   - Audit: 操作审计日志                                            │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Components

### 1. IntentRouter

基于 `compose.Graph` 实现的多级意图路由器。

```go
type IntentRouter struct {
    graph      *compose.Graph
    classifier *IntentClassifier
    domains    map[ToolDomain]*ActionGraph
}

type IntentClassifier struct {
    model model.ToolCallingChatModel
}

func (r *IntentRouter) Route(ctx context.Context, input string) (*ActionGraph, error) {
    // 1. PreProcess: 输入清洗
    // 2. Classify: 意图分类
    // 3. Branch: 路由到对应 Domain Graph
    domain, err := r.classifier.Classify(ctx, input)
    if err != nil {
        return r.domains[DomainGeneral], nil // Fallback
    }
    return r.domains[domain], nil
}
```

**路由规则**:

| ToolDomain | 触发条件 | ActionGraph |
|------------|---------|-------------|
| DomainInfrastructure | 主机、集群操作 | InfrastructureActionGraph |
| DomainService | 服务、部署操作 | ServiceActionGraph |
| DomainCICD | 流水线、任务操作 | CICDActionGraph |
| DomainMonitor | 监控、告警操作 | MonitorActionGraph |
| DomainConfig | 配置操作 | ConfigActionGraph |
| DomainGeneral | 其他 / Fallback | GeneralActionGraph |

### 2. ActionGraph

基于 `compose.Workflow` 实现的确定性执行流程。

```go
type ActionGraph struct {
    workflow   *compose.Workflow[ActionInput, ActionOutput]
    chatModel  model.ToolCallingChatModel
    tools      *compose.ToolsNode
    validator  *Validator
}

type ActionInput struct {
    SessionID  string
    Message    string
    UserID     uint64
    Context    map[string]any
}

type ActionOutput struct {
    Response   string
    ToolCalls  []ToolCallResult
    Interrupt  *InterruptInfo // 非空表示需要用户确认
}

type GraphState struct {
    Messages          []*schema.Message
    PendingToolCalls  []ToolCallSpec
    ToolResults       map[string]ToolResult
    ValidationErrors  []ValidationError
}
```

**节点定义**:

```go
// Sanitize Lambda - 输入脱敏
func sanitizeNode(ctx context.Context, input ActionInput) (ActionInput, error) {
    // 移除敏感信息（密码、token等）
    // 注入多租户上下文
    return input, nil
}

// Reasoning Node - LLM 推理
func reasoningNode(ctx context.Context, input ActionInput, state *GraphState) (*schema.Message, error) {
    // 构建增强 Prompt（含 RAG 上下文）
    // 调用 ChatModel 生成回复
    // 返回 Message（可能包含 ToolCalls）
}

// Validation Lambda - 校验层
func validationNode(ctx context.Context, msg *schema.Message, state *GraphState) error {
    if len(msg.ToolCalls) == 0 {
        return nil
    }

    for _, tc := range msg.ToolCalls {
        // Level 1: JSON Schema 校验
        if err := validateJSONSchema(tc); err != nil {
            return err
        }

        // Level 2: K8s OpenAPI 校验 (如果适用)
        if isK8sResource(tc) {
            if err := validateK8sOpenAPI(tc); err != nil {
                return err
            }
        }
    }

    return nil
}

// Execution Node - 工具执行
func executionNode(ctx context.Context, msg *schema.Message, state *GraphState) ([]ToolResult, error) {
    // SecurityAspect 会在工具执行前后注入
    // 包括权限检查、中断处理
    return toolsNode.Invoke(ctx, msg)
}
```

### 3. SecurityAspect

基于 Eino Callbacks 实现的安全切面。

```go
type SecurityAspect struct {
    permissionChecker PermissionChecker
    interruptHandler  InterruptHandler
    auditLogger       AuditLogger
}

func (a *SecurityAspect) OnToolCallStart(ctx context.Context, tc ToolCall) (*ToolCallDecision, error) {
    meta := getToolMeta(tc.Name)

    // 1. 权限检查
    hasPermission, err := a.permissionChecker.Check(ctx, tc)
    if err != nil {
        return nil, err
    }

    if !hasPermission {
        // 无权限: 返回需要创建审批任务
        return &ToolCallDecision{
            Action: DecisionCreateApproval,
            Reason: "用户无权限执行此操作",
        }, nil
    }

    // 2. 风险检查
    if meta.Risk == ToolRiskHigh || meta.Risk == ToolRiskMedium {
        // 有权限但高风险: 返回需要用户确认
        return &ToolCallDecision{
            Action: DecisionInterrupt,
            Reason: a.generateRiskReason(ctx, tc, meta),
        }, nil
    }

    // 3. 低风险直接执行
    return &ToolCallDecision{
        Action: DecisionExecute,
    }, nil
}
```

### 4. Approval System

双轨制审批系统。

```go
// 有权限场景 - LLM Interrupt
type InterruptInfo struct {
    Type        string         // "approval_required"
    ToolName    string
    Risk        ToolRisk
    Reason      string         // AI 说明为什么执行
    Preview     string         // 操作预览
    ToolCall    ToolCallSpec   // 待执行的调用
}

// 无权限场景 - 审批任务
type ApprovalTask struct {
    ID             uint64
    RequesterID    uint64
    Status         string        // pending/approved/rejected/executed

    // 资源路由
    ResourceType   string
    ResourceID     string
    ResourceName   string

    // LLM 生成的详细任务
    TaskDetail     TaskDetail

    // 执行规格
    ToolCalls      []ToolCallSpec
}

type TaskDetail struct {
    Summary        string           // 操作概述
    Steps          []ExecutionStep  // 执行步骤
    RiskAssessment RiskAssessment   // 风险评估
    RollbackPlan   string           // 回滚方案
}
```

**审批人路由**:

```go
type ApprovalRouter interface {
    Route(task *ApprovalTask) (approverID uint64, err error)
}

// 按资源类型路由到负责人
type ResourceOwnerRouter struct {
    db *gorm.DB
}

func (r *ResourceOwnerRouter) Route(task *ApprovalTask) (uint64, error) {
    switch task.ResourceType {
    case "host":
        return r.getHostOwner(task.ResourceID)
    case "cluster":
        return r.getClusterOwner(task.ResourceID)
    case "service":
        return r.getServiceOwner(task.ResourceID)
    default:
        return r.getDefaultApprover()
    }
}
```

**任务执行器**:

```go
type ApprovalExecutor struct {
    tools     *ToolRegistry
    emitter   EventEmitter
}

func (e *ApprovalExecutor) Execute(ctx context.Context, task *ApprovalTask) (*ExecutionResult, error) {
    // 审批通过后直接执行，无需 LLM 参与
    results := make([]ToolResult, 0, len(task.ToolCalls))

    for _, tc := range task.ToolCalls {
        result, err := e.tools.Execute(ctx, tc.ToolName, tc.Params)
        if err != nil {
            return nil, err
        }
        results = append(results, result)
    }

    return &ExecutionResult{Results: results}, nil
}
```

### 5. RAG System

```go
type RAGSystem struct {
    indexer   Indexer
    retriever Retriever
    feedback  FeedbackCollector
}

type KnowledgeSource string

const (
    SourceUserInput    KnowledgeSource = "user_input"    // 用户主动投喂
    SourceFeedback     KnowledgeSource = "feedback"      // 反馈收集
)

type KnowledgeEntry struct {
    ID          string
    Source      KnowledgeSource
    Question    string
    Answer      string
    Embedding   []float32
    Namespace   string        // 多租户隔离
    CreatedAt   time.Time
}

// 反馈收集
func (r *RAGSystem) CollectFeedback(ctx context.Context, sessionID string, feedback Feedback) error {
    if !feedback.IsEffective {
        return nil
    }

    // 提取 Q&A 对
    qa, err := r.extractQA(ctx, sessionID)
    if err != nil {
        return err
    }

    // 向量化并存储
    return r.indexer.Index(ctx, qa)
}
```

## Directory Structure

```
internal/ai/
├── router/                 # IntentRouter
│   ├── router.go          # 主路由器
│   ├── domain_routes.go   # 域路由配置
│   └── classifier.go      # 意图分类 Lambda
│
├── graph/                  # ActionGraph
│   ├── action_graph.go    # 主工作流
│   ├── sanitize.go        # 输入脱敏 Lambda
│   ├── reasoning.go       # LLM 推理节点
│   ├── validation.go      # K8s OpenAPI 校验
│   └── execution.go       # 工具执行
│
├── aspect/                 # SecurityAspect
│   ├── aspect.go          # Aspect 主入口
│   ├── permission.go      # 权限检查
│   ├── interrupt.go       # 风险操作中断
│   └── audit.go           # 审计日志
│
├── approval/               # 审批系统
│   ├── task.go            # ApprovalTask 模型
│   ├── generator.go       # LLM 任务生成
│   ├── executor.go        # 审批通过后执行
│   └── router.go          # 审批人路由
│
├── rag/                    # RAG 系统
│   ├── indexer.go         # 向量索引
│   ├── retriever.go       # 检索器
│   ├── feedback.go        # 反馈收集
│   └── context.go         # 多租户上下文
│
├── state/                  # 状态管理
│   ├── session.go         # 会话状态
│   └── checkpoint.go      # Graph 检查点
│
├── tools/                  # 现有工具（保留）
│   └── ...
│
└── agent.go               # 主入口
```

## Data Contracts

### Tool Call Decision

```json
{
  "tool_name": "host_batch_exec",
  "params": {
    "host_ids": [1, 2, 3],
    "command": "rm -rf /data/logs/*.log"
  },
  "decision": {
    "action": "interrupt",
    "reason": "批量删除操作可能导致数据丢失，建议先备份"
  }
}
```

### SSE Events

```json
// 有权限 - 风险确认
{
  "type": "approval_required",
  "tool": "host_batch_exec",
  "risk": "high",
  "reason": "即将在 3 台主机上执行删除命令...",
  "preview": "rm -rf /data/logs/*.log",
  "session_id": "xxx",
  "checkpoint_id": "yyy"
}

// 无权限 - 审批任务创建
{
  "type": "approval_task_created",
  "task_id": 123,
  "message": "您没有权限执行此操作，已创建审批任务"
}
```

## API Changes

```go
// 现有接口保留
POST /ai/chat              // 聊天入口
POST /ai/chat/respond      // 审批确认 (有权限场景)
GET  /ai/tools             // 工具列表

// 新增
POST /ai/approval/create   // 创建审批任务 (无权限场景)
GET  /ai/approval/list     // 我的审批任务列表
GET  /ai/approval/:id      // 审批任务详情
POST /ai/approval/:id/approve  // 审批通过
POST /ai/approval/:id/reject   // 审批拒绝

POST /ai/feedback          // 对话结束时提交反馈
```
