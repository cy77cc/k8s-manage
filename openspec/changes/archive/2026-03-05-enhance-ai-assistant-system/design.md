# Design: AI助手系统优化

## 1. Eino框架废弃方法迁移

### 1.1 当前问题分析

```
当前架构:
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│   PlatformAgent                                                             │
│   ├── Runnable (*react.Agent)                                               │
│   ├── registry (ExpertRegistry)                                             │
│   ├── router (*HybridRouter)                                                │
│   ├── orchestrator (*Orchestrator) ◀── DEPRECATED                           │
│   ├── graphRunner (compose.Runnable) ◀── 新版，但未被完全使用               │
│   └── tools (map[string]tool.InvokableTool)                                 │
│                                                                             │
│   问题:                                                                      │
│   1. orchestrator.go 整个文件标记为 Deprecated                              │
│   2. platform_agent.go:150-155 仍使用 orchestrator 作为回退                │
│   3. graphRunner 虽然创建但功能不完整                                        │
│   4. react.NewAgent 使用旧的 AgentConfig 方式                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 迁移方案

#### 1.2.1 Orchestrator → Graph Runner

**移除的代码:**
- `internal/ai/experts/orchestrator.go`
- `internal/ai/experts/orchestrator_test.go`
- `internal/ai/experts/orchestrator_primary_led_test.go`

**增强的代码:**
- `internal/ai/graph/builder.go` - 增加流式执行支持
- `internal/ai/graph/runners.go` - 完善执行器逻辑
- `internal/ai/platform_agent.go` - 移除 orchestrator 字段，统一使用 graphRunner

**Graph Runner 架构:**

```go
// internal/ai/graph/builder.go

type Builder struct {
    primaryRunner PrimaryRunner
    helperRunner  HelperRunner
    // 新增: 流式执行支持
    streamPrimaryRunner StreamPrimaryRunner
    streamHelperRunner  StreamHelperRunner
}

// BuildStream 构建支持流式输出的图
func (b *Builder) BuildStream(ctx context.Context) (*compose.Graph[*GraphInput, *schema.StreamReader[*schema.Message]], error) {
    g := compose.NewGraph[*GraphInput, *schema.StreamReader[*schema.Message]]()

    // 路由节点
    route := compose.InvokableLambda(func(_ context.Context, in *GraphInput) (*GraphInput, error) {
        return in, nil
    })

    // 主专家流式节点
    primary := compose.InvokableLambda(func(ctx context.Context, in *GraphInput) (*schema.StreamReader[*schema.Message], error) {
        return b.streamPrimaryRunner.Run(ctx, in)
    })

    // 助手并行流式节点
    parallelHelpers := compose.InvokableLambda(func(ctx context.Context, in *GraphInput) (*schema.StreamReader[*schema.Message], error) {
        return b.streamHelperRunner.RunParallel(ctx, in)
    })

    // ... 构建图结构

    return g, nil
}
```

#### 1.2.2 react.NewAgent 配置优化

**当前方式:**
```go
agent, err := react.NewAgent(r.ctx, &react.AgentConfig{
    ToolCallingModel: r.chatModel,
    ToolsConfig:      compose.ToolsNodeConfig{Tools: baseTools},
    MaxStep:          20,
    MessageModifier:  react.NewPersonaModifier(persona),
})
```

**推荐方式 (v0.8+ 风格):**
```go
agent, err := react.NewAgent(r.ctx,
    react.WithTools(r.ctx, baseTools...),
    react.WithChatModelOptions(/* model options */),
    react.WithMaxStep(20),
    react.WithMessageModifier(react.NewPersonaModifier(persona)),
)
```

### 1.3 文件变更清单

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `internal/ai/experts/orchestrator.go` | 删除 | 废弃的编排器 |
| `internal/ai/experts/orchestrator_*.go` | 删除 | 相关测试文件 |
| `internal/ai/graph/builder.go` | 修改 | 增加流式支持 |
| `internal/ai/graph/runners.go` | 修改 | 完善执行逻辑 |
| `internal/ai/graph/types.go` | 修改 | 增加流式类型 |
| `internal/ai/platform_agent.go` | 修改 | 移除 orchestrator |
| `internal/ai/experts/registry.go` | 修改 | 更新 Agent 创建方式 |

---

## 2. 确认-执行审批机制

### 2.1 设计理念

采用类似 Claude Code 的"确认-执行"模式：**无论用户是否有权限，变更类操作都必须先展示预览并请求确认**。

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                     "确认-执行"模式核心理念                                   │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   传统模式（废弃）:                                                          │
│   ┌────────────────────────────────────────────────────────────────────┐    │
│   │ 有权限 ──▶ 直接执行                                                 │    │
│   │ 无权限 ──▶ 发起审批                                                 │    │
│   └────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│   新模式（采用）:                                                            │
│   ┌────────────────────────────────────────────────────────────────────┐    │
│   │ 变更操作 ──▶ 展示预览 ──▶ 请求用户确认                              │    │
│   │                           │                                         │    │
│   │              ┌────────────┼────────────┐                            │    │
│   │              ▼            ▼            ▼                            │    │
│   │          [确认]        [取消]      [超时]                           │    │
│   │              │            │            │                            │    │
│   │              ▼            ▼            ▼                            │    │
│   │         风险判断       取消操作      自动取消                        │    │
│   │              │                                                        │    │
│   │    ┌─────────┼─────────┐                                              │    │
│   │    ▼         ▼         ▼                                              │    │
│   │ [低风险]  [中风险]  [高风险]                                           │    │
│   │    │         │         │                                              │    │
│   │    ▼         ▼         ▼                                              │    │
│   │ 直接执行  权限检查  第三方审批                                         │    │
│   └────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 完整审批流程

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                     确认-执行流程状态机                                       │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   AI 准备执行工具                                                            │
│        │                                                                     │
│        ▼                                                                     │
│   ┌─────────────────────────────────────┐                                   │
│   │ 工具类型检查                        │                                   │
│   │ mode == readonly ?                  │                                   │
│   └──────────────┬──────────────────────┘                                   │
│                  │                                                           │
│         ┌────────┴────────┐                                                 │
│         ▼                 ▼                                                 │
│     [readonly]        [mutating]                                            │
│         │                 │                                                 │
│         ▼                 ▼                                                 │
│   ┌───────────┐    ┌───────────────────┐                                    │
│   │ 直接执行  │    │ 生成操作预览      │                                    │
│   │ (可配置)  │    │ PreviewBuilder    │                                    │
│   └───────────┘    └─────────┬─────────┘                                    │
│                              │                                               │
│                              ▼                                               │
│                    ┌───────────────────┐                                    │
│                    │ 推送确认请求      │                                    │
│                    │ SSE: confirm_     │                                    │
│                    │ required          │                                    │
│                    └─────────┬─────────┘                                    │
│                              │                                               │
│               ┌──────────────┼──────────────┐                                │
│               ▼              ▼              ▼                                │
│           [确认]          [取消]         [超时]                              │
│               │              │              │                               │
│               ▼              ▼              ▼                               │
│        ┌─────────────┐  ┌─────────┐  ┌─────────────┐                        │
│        │ 风险分级判断│  │ 返回取消│  │ 自动取消    │                        │
│        │ RiskLevel   │  │         │  │ 返回超时    │                        │
│        └──────┬──────┘  └─────────┘  └─────────────┘                        │
│               │                                                              │
│    ┌──────────┼──────────┐                                                   │
│    ▼          ▼          ▼                                                   │
│ [low]      [medium]    [high]                                                │
│    │          │          │                                                   │
│    ▼          ▼          ▼                                                   │
│ 直接执行   权限检查   创建审批工单                                            │
│    │          │          │                                                   │
│    │    ┌─────┴─────┐    │                                                   │
│    │    ▼           ▼    │                                                   │
│    │ [有权限]    [无权限]│                                                   │
│    │    │           │    │                                                   │
│    │    ▼           └────┼───▶ 合并到审批流程                                │
│    │ 直接执行            │                                                   │
│    │                     ▼                                                   │
│    │           ┌───────────────────┐                                        │
│    │           │ WebSocket 通知    │                                        │
│    │           │ 审批人            │                                        │
│    │           └─────────┬─────────┘                                        │
│    │                     │                                                   │
│    │          ┌──────────┼──────────┐                                        │
│    │          ▼          ▼          ▼                                        │
│    │      [通过]      [拒绝]     [超时]                                      │
│    │          │          │          │                                        │
│    │          ▼          ▼          ▼                                        │
│    │      approved   rejected  expired                                       │
│    │          │          │          │                                        │
│    │          ▼          ▼          ▼                                        │
│    │      执行工具   返回错误  返回超时                                       │
│    │          │          │          │                                        │
│    └──────────┴──────────┴──────────┘                                        │
│               │                                                              │
│               ▼                                                              │
│        ┌─────────────┐                                                       │
│        │ 返回结果    │                                                       │
│        │ 记录审计    │                                                       │
│        └─────────────┘                                                       │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

### 2.3 操作预览生成

```go
// internal/service/ai/preview_builder.go

type PreviewBuilder struct {
    db             *gorm.DB
    toolRegistry   *tools.Registry
}

type OperationPreview struct {
    ToolName        string         `json:"tool_name"`
    ToolDescription string         `json:"tool_description"`
    Params          map[string]any `json:"params"`
    RiskLevel       string         `json:"risk_level"`      // low/medium/high
    Mode            string         `json:"mode"`            // readonly/mutating
    TargetResources []TargetResource `json:"target_resources"`
    ImpactScope     string         `json:"impact_scope"`
    PreviewDiff     string         `json:"preview_diff,omitempty"`
    Timeout         time.Duration  `json:"timeout"`
}

type TargetResource struct {
    Type string `json:"type"`     // host/service/cluster
    ID   int64  `json:"id"`
    Name string `json:"name"`
}

// BuildPreview 为变更操作生成预览
func (b *PreviewBuilder) BuildPreview(ctx context.Context, toolName string, params map[string]any) (*OperationPreview, error) {
    meta, ok := b.toolRegistry.GetMeta(toolName)
    if !ok {
        return nil, fmt.Errorf("tool not found: %s", toolName)
    }

    preview := &OperationPreview{
        ToolName:        toolName,
        ToolDescription: meta.Description,
        Params:          params,
        RiskLevel:       string(meta.Risk),
        Mode:            string(meta.Mode),
        Timeout:         5 * time.Minute, // 默认确认超时
    }

    // 提取目标资源
    preview.TargetResources = b.extractTargetResources(toolName, params)

    // 生成影响范围描述
    preview.ImpactScope = b.generateImpactScope(toolName, params, preview.TargetResources)

    // 对于部署类操作，生成差异预览
    if strings.Contains(toolName, "deploy") || strings.Contains(toolName, "apply") {
        preview.PreviewDiff = b.generateDeployDiff(ctx, toolName, params)
    }

    return preview, nil
}

// generateImpactScope 生成影响范围描述
func (b *PreviewBuilder) generateImpactScope(toolName string, params map[string]any, resources []TargetResource) string {
    switch {
    case strings.Contains(toolName, "batch_exec"):
        return fmt.Sprintf("将在 %d 台主机上执行命令", len(params["host_ids"].([]int)))
    case strings.Contains(toolName, "deploy"):
        return fmt.Sprintf("将更新服务配置并重新部署")
    case strings.Contains(toolName, "delete"):
        return "⚠️ 此操作将删除资源，不可恢复"
    default:
        if len(resources) > 0 {
            return fmt.Sprintf("影响 %d 个资源", len(resources))
        }
        return "常规操作"
    }
}
```

### 2.4 用户确认服务

```go
// internal/service/ai/confirmation_service.go

type ConfirmationService struct {
    store    *confirmationStore
    notifier *WebSocketNotifier
    timeout  time.Duration
}

type ConfirmationRequest struct {
    ID        string            `json:"id"`
    Preview   *OperationPreview `json:"preview"`
    Status    string            `json:"status"`     // pending/confirmed/cancelled/expired
    CreatedAt time.Time         `json:"created_at"`
    ExpiresAt time.Time         `json:"expires_at"`
}

// RequestConfirmation 请求用户确认
func (s *ConfirmationService) RequestConfirmation(ctx context.Context, uid uint64, preview *OperationPreview) (*ConfirmationRequest, error) {
    req := &ConfirmationRequest{
        ID:        fmt.Sprintf("confirm-%d", time.Now().UnixNano()),
        Preview:   preview,
        Status:    "pending",
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(s.timeout),
    }

    // 存储确认请求
    s.store.Save(req)

    // 通过 SSE 推送确认请求给用户
    s.notifier.NotifyUser(uid, "confirmation_required", map[string]any{
        "confirmation_id": req.ID,
        "preview":         preview,
        "expires_at":      req.ExpiresAt,
        "message":         fmt.Sprintf("即将执行 %s，请确认操作", preview.ToolName),
    })

    // 启动超时监控
    go s.watchConfirmationTimeout(req, uid)

    return req, nil
}

// WaitForConfirmation 等待用户确认
func (s *ConfirmationService) WaitForConfirmation(ctx context.Context, confirmationID string) (*ConfirmationResult, error) {
    // 阻塞等待或通过 channel 通知
    resultCh := make(chan *ConfirmationResult, 1)
    s.store.RegisterResultChannel(confirmationID, resultCh)

    select {
    case result := <-resultCh:
        return result, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

// Confirm 确认操作
func (s *ConfirmationService) Confirm(confirmationID string, uid uint64) error {
    req, ok := s.store.Get(confirmationID)
    if !ok {
        return fmt.Errorf("confirmation not found")
    }
    if req.Status != "pending" {
        return fmt.Errorf("confirmation already processed: %s", req.Status)
    }

    req.Status = "confirmed"
    s.store.Save(req)

    // 通知等待的协程
    s.store.NotifyResult(confirmationID, &ConfirmationResult{
        Confirmed: true,
        UserID:    uid,
    })

    return nil
}

// Cancel 取消操作
func (s *ConfirmationService) Cancel(confirmationID string, uid uint64) error {
    req, ok := s.store.Get(confirmationID)
    if !ok {
        return fmt.Errorf("confirmation not found")
    }

    req.Status = "cancelled"
    s.store.Save(req)

    s.store.NotifyResult(confirmationID, &ConfirmationResult{
        Confirmed: false,
        UserID:    uid,
        Reason:    "user cancelled",
    })

    return nil
}
```

### 2.5 对象权限检查

```go
// internal/service/ai/permission_checker.go

type PermissionChecker struct {
    casbinEnforcer *casbin.Enforcer
    db             *gorm.DB
}

// ToolResourceMapping 定义工具到资源的映射
var ToolResourceMapping = map[string]ResourceExtractor{
    "host_ssh_exec_readonly": {
        ResourceType: "host",
        IDParam:      "host_id",
        Action:       "execute",
    },
    "host_batch_exec_preview": {
        ResourceType: "host",
        IDParam:      "host_ids", // 数组类型
        Action:       "execute",
        IsArray:      true,
    },
    "service_deploy_preview": {
        ResourceType: "service",
        IDParam:      "service_id",
        Action:       "deploy",
    },
    "k8s_list_pods": {
        ResourceType: "cluster",
        IDParam:      "cluster_id",
        Action:       "read",
    },
    // ... 其他工具映射
}

type ResourceExtractor struct {
    ResourceType string
    IDParam      string
    Action       string
    IsArray      bool
}

// CheckPermission 检查用户对资源的操作权限
func (p *PermissionChecker) CheckPermission(ctx context.Context, uid uint64, toolName string, params map[string]any) (bool, error) {
    mapping, ok := ToolResourceMapping[toolName]
    if !ok {
        // 未定义映射的工具，使用默认权限检查
        return p.checkDefaultPermission(uid, toolName)
    }

    subject := fmt.Sprintf("user:%d", uid)

    if mapping.IsArray {
        // 数组类型资源ID，检查每一个
        ids := extractArrayParam(params, mapping.IDParam)
        for _, id := range ids {
            resource := fmt.Sprintf("%s:%v", mapping.ResourceType, id)
            allowed, err := p.casbinEnforcer.Enforce(subject, resource, mapping.Action)
            if err != nil || !allowed {
                return false, err
            }
        }
        return true, nil
    }

    // 单一资源ID
    id := params[mapping.IDParam]
    resource := fmt.Sprintf("%s:%v", mapping.ResourceType, id)
    return p.casbinEnforcer.Enforce(subject, resource, mapping.Action)
}

// FindApprovers 查找有权限审批的用户
func (p *PermissionChecker) FindApprovers(resourceType string, resourceID any) []uint64 {
    // 查找对该资源有 admin 或 approve 权限的用户
    var approvers []uint64

    // 通过 Casbin policy 查询
    // 或通过数据库查询拥有该资源管理权限的用户
    resource := fmt.Sprintf("%s:%v", resourceType, resourceID)

    // 获取所有有权限的 subject
    subjects := p.casbinEnforcer.GetSubjectsForDomain(resource)
    for _, sub := range subjects {
        if uid := parseUserID(sub); uid > 0 {
            approvers = append(approvers, uid)
        }
    }

    return approvers
}
```

### 2.3 审批工单数据模型

```go
// internal/model/ai_approval.go

type AIApprovalTicket struct {
    ID           string    `gorm:"column:id;type:varchar(64);primaryKey"`
    Tool         string    `gorm:"column:tool;type:varchar(128);index"`
    Params       string    `gorm:"column:params;type:json"`
    ResourceType string    `gorm:"column:resource_type;type:varchar(64);index"`
    ResourceID   string    `gorm:"column:resource_id;type:varchar(128);index"`
    Risk         string    `gorm:"column:risk;type:varchar(32)"`
    Mode         string    `gorm:"column:mode;type:varchar(32)"`
    Status       string    `gorm:"column:status;type:varchar(32);index"` // pending/approved/rejected/expired
    RequestUID   uint64    `gorm:"column:request_uid;index"`
    ReviewUID    uint64    `gorm:"column:review_uid"`
    ExpiresAt    time.Time `gorm:"column:expires_at;index"`
    CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
    ReviewedAt   *time.Time `gorm:"column:reviewed_at"`
    Reason       string    `gorm:"column:reason;type:text"` // 审批理由
}

func (AIApprovalTicket) TableName() string { return "ai_approval_tickets" }
```

### 2.4 WebSocket 实时推送

```go
// internal/service/ai/approval_notifier.go

type ApprovalNotifier struct {
    hub      *websocket.Hub
    store    *approvalStore
    timeout  time.Duration
}

// NotifyApprovers 向审批人推送审批请求
func (n *ApprovalNotifier) NotifyApprovers(ctx context.Context, ticket *AIApprovalTicket, approverUIDs []uint64) error {
    notification := ApprovalNotification{
        Type:        "approval_required",
        TicketID:    ticket.ID,
        Tool:        ticket.Tool,
        ResourceType: ticket.ResourceType,
        ResourceID:  ticket.ResourceID,
        Risk:        ticket.Risk,
        RequesterID: ticket.RequestUID,
        ExpiresAt:   ticket.ExpiresAt,
        Message:     fmt.Sprintf("需要审批: %s 操作", ticket.Tool),
    }

    // 向所有审批人推送
    for _, uid := range approverUIDs {
        if err := n.hub.SendToUser(uid, notification); err != nil {
            log.Warn("failed to notify approver", "uid", uid, "err", err)
        }
    }

    // 启动超时检查
    go n.watchTimeout(ticket)

    return nil
}

// watchTimeout 监控审批超时
func (n *ApprovalNotifier) watchTimeout(ticket *AIApprovalTicket) {
    timer := time.NewTimer(time.Until(ticket.ExpiresAt))
    defer timer.Stop()

    <-timer.C

    // 检查是否仍未处理
    current, _ := n.store.Get(ticket.ID)
    if current.Status == "pending" {
        n.store.UpdateStatus(ticket.ID, "expired", 0)
        // 通知请求方
        n.hub.SendToUser(ticket.RequestUID, ApprovalNotification{
            Type:     "approval_expired",
            TicketID: ticket.ID,
            Message:  "审批请求已超时",
        })
    }
}
```

### 2.6 审批配置

```yaml
# configs/approval_config.yaml

# 确认-执行模式配置
confirmation:
  default_timeout_minutes: 5    # 用户确认默认超时
  skip_for_readonly: true       # 只读操作是否跳过确认

# 风险分级配置
risk_levels:
  low:
    require_confirmation: true   # 需要用户确认
    require_approval: false      # 不需要第三方审批
  medium:
    require_confirmation: true
    require_approval: false      # 有权限即可执行
    check_permission: true       # 需要检查权限
  high:
    require_confirmation: true
    require_approval: true       # 必须第三方审批
    approval_timeout_minutes: 30

# 工具级别配置（覆盖默认风险分级）
tools:
  # 只读工具 - 可跳过确认
  host_list_inventory:
    risk_level: low
    skip_confirmation: true

  k8s_list_pods:
    risk_level: low
    skip_confirmation: true

  # 变更工具 - 需要确认
  host_ssh_exec_readonly:
    risk_level: low
    require_confirmation: true
    description: "SSH只读命令执行"

  host_batch_exec_preview:
    risk_level: medium
    require_confirmation: true
    check_permission: true
    description: "批量命令预览"

  host_batch_exec_apply:
    risk_level: high
    require_confirmation: true
    require_approval: true
    approval_timeout_minutes: 60
    description: "批量命令执行"
    approver_permissions:
      - "host:admin"
      - "host:execute:approve"

  service_deploy_preview:
    risk_level: medium
    require_confirmation: true
    check_permission: true
    description: "服务部署预览"

  service_deploy_apply:
    risk_level: high
    require_confirmation: true
    require_approval: true
    approval_timeout_minutes: 30
    description: "服务部署执行"
    approver_permissions:
      - "service:admin"
      - "deployment:approve"

  k8s_delete_pod:
    risk_level: high
    require_confirmation: true
    require_approval: true
    approval_timeout_minutes: 15
    description: "删除Pod"
    approver_permissions:
      - "cluster:admin"
```

---

## 3. RAG知识库

### 3.1 系统架构

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                          RAG知识库架构                                        │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌────────────────────────────────────────────────────────────────────┐    │
│   │                        数据摄入层                                   │    │
│   │                                                                    │    │
│   │   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │    │
│   │   │ 定时任务调度 │  │ 数据提取器   │  │ 向量化处理器 │           │    │
│   │   │ (cron)       │  │ (extractors) │  │ (embedder)   │           │    │
│   │   └──────┬───────┘  └──────┬───────┘  └──────┬───────┘           │    │
│   │          │                 │                 │                    │    │
│   │          └─────────────────┼─────────────────┘                    │    │
│   │                            ▼                                      │    │
│   │                   ┌────────────────┐                              │    │
│   │                   │   Milvus       │                              │    │
│   │                   │   Upsert       │                              │    │
│   │                   └────────────────┘                              │    │
│   └────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│   ┌────────────────────────────────────────────────────────────────────┐    │
│   │                        检索服务层                                   │    │
│   │                                                                    │    │
│   │   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │    │
│   │   │ Query        │  │ Reranker     │  │ Context      │           │    │
│   │   │ Embedder     │  │ (可选)       │  │ Builder      │           │    │
│   │   └──────┬───────┘  └──────┬───────┘  └──────┬───────┘           │    │
│   │          │                 │                 │                    │    │
│   │          └─────────────────┼─────────────────┘                    │    │
│   │                            ▼                                      │    │
│   │                   增强后的 Prompt                                  │    │
│   └────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│   ┌────────────────────────────────────────────────────────────────────┐    │
│   │                        Milvus Collections                          │    │
│   │                                                                    │    │
│   │   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐            │    │
│   │   │ tool_       │   │ platform_   │   │ trouble_    │            │    │
│   │   │ examples    │   │ assets      │   │ shooting    │            │    │
│   │   │             │   │             │   │             │            │    │
│   │   │ 向量维度:   │   │ 向量维度:   │   │ 向量维度:   │            │    │
│   │   │ 1536        │   │ 1536        │   │ 1536        │            │    │
│   │   │             │   │             │   │             │            │    │
│   │   │ 索引类型:   │   │ 索引类型:   │   │ 索引类型:   │            │    │
│   │   │ IVF_FLAT    │   │ IVF_FLAT    │   │ IVF_FLAT    │            │    │
│   │   └─────────────┘   └─────────────┘   └─────────────┘            │    │
│   └────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Collection Schema 设计

```go
// internal/rag/milvus_schema.go

// ToolExampleCollection 工具调用示例
var ToolExampleSchema = &entity.Schema{
    CollectionName: "tool_examples",
    Description:    "工具调用成功案例和参数示例",
    AutoID:         true,
    Fields: []*entity.Field{
        {Name: "id", DataType: entity.FieldTypeInt64, PrimaryKey: true, AutoID: true},
        {Name: "tool_name", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "128"}},
        {Name: "intent", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "512"}},
        {Name: "params_json", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "2048"}},
        {Name: "result_summary", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "1024"}},
        {Name: "success", DataType: entity.FieldTypeBool},
        {Name: "scene", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "64"}},
        {Name: "embedding", DataType: entity.FieldTypeFloatVector, TypeParams: map[string]string{"dim": "1536"}},
        {Name: "created_at", DataType: entity.FieldTypeInt64},
    },
}

// PlatformAssetCollection 平台资产索引
var PlatformAssetSchema = &entity.Schema{
    CollectionName: "platform_assets",
    Description:    "平台资源元数据索引",
    AutoID:         true,
    Fields: []*entity.Field{
        {Name: "id", DataType: entity.FieldTypeInt64, PrimaryKey: true, AutoID: true},
        {Name: "asset_type", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "32"}}, // host/cluster/service/deployment
        {Name: "asset_id", DataType: entity.FieldTypeInt64},
        {Name: "name", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "256"}},
        {Name: "description", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "1024"}},
        {Name: "metadata", DataType: entity.FieldTypeJSON},
        {Name: "status", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "32"}},
        {Name: "embedding", DataType: entity.FieldTypeFloatVector, TypeParams: map[string]string{"dim": "1536"}},
        {Name: "updated_at", DataType: entity.FieldTypeInt64},
    },
}

// TroubleshootingCaseCollection 故障排查案例
var TroubleshootingCaseSchema = &entity.Schema{
    CollectionName: "troubleshooting_cases",
    Description:    "故障诊断和解决方案案例",
    AutoID:         true,
    Fields: []*entity.Field{
        {Name: "id", DataType: entity.FieldTypeInt64, PrimaryKey: true, AutoID: true},
        {Name: "title", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "256"}},
        {Name: "symptoms", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "2048"}},
        {Name: "diagnosis", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "4096"}},
        {Name: "solution", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "4096"}},
        {Name: "related_tools", DataType: entity.FieldTypeJSON},
        {Name: "severity", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "16"}},
        {Name: "embedding", DataType: entity.FieldTypeFloatVector, TypeParams: map[string]string{"dim": "1536"}},
        {Name: "created_at", DataType: entity.FieldTypeInt64},
    },
}
```

### 3.3 数据摄入服务

```go
// internal/rag/ingestion.go

type IngestionService struct {
    milvusClient *milvus.Client
    embedder     Embedder
    db           *gorm.DB
}

// IngestToolExamples 从命令执行记录摄入工具示例
func (s *IngestionService) IngestToolExamples(ctx context.Context, since time.Time) error {
    var executions []model.AICommandExecution
    err := s.db.Where("created_at > ?", since).
        Where("status = ?", "succeeded").
        Find(&executions).Error
    if err != nil {
        return err
    }

    var records []ToolExampleRecord
    for _, exec := range executions {
        intent := exec.CommandText
        params := parseParams(exec.ParamsJSON)

        // 构建用于向量化的文本
        textForEmbedding := fmt.Sprintf("%s %s", exec.Intent, exec.CommandText)
        embedding, err := s.embedder.Embed(ctx, textForEmbedding)
        if err != nil {
            continue
        }

        records = append(records, ToolExampleRecord{
            ToolName:      extractToolName(exec),
            Intent:        intent,
            ParamsJSON:    exec.ParamsJSON,
            ResultSummary: exec.ExecutionSummary,
            Success:       exec.Status == "succeeded",
            Scene:         exec.Scene,
            Embedding:     embedding,
            CreatedAt:     exec.CreatedAt.Unix(),
        })
    }

    return s.milvusClient.Insert(ctx, "tool_examples", records...)
}

// IngestPlatformAssets 全量同步平台资产
func (s *IngestionService) IngestPlatformAssets(ctx context.Context) error {
    // 主机资产
    var hosts []model.Host
    s.db.Find(&hosts)
    for _, host := range hosts {
        text := fmt.Sprintf("%s %s %s", host.Name, host.IP, host.Hostname)
        embedding, _ := s.embedder.Embed(ctx, text)
        // ... 插入到 Milvus
    }

    // 服务资产
    var services []model.Service
    s.db.Find(&services)
    // ...

    // 集群资产
    var clusters []model.Cluster
    s.db.Find(&clusters)
    // ...

    return nil
}
```

### 3.4 检索增强

```go
// internal/rag/retriever.go

type RAGRetriever struct {
    milvusClient *milvus.Client
    embedder     Embedder
}

// Retrieve 检索相关上下文
func (r *RAGRetriever) Retrieve(ctx context.Context, query string, topK int) (*RAGContext, error) {
    // 1. 查询向量化
    queryEmbedding, err := r.embedder.Embed(ctx, query)
    if err != nil {
        return nil, err
    }

    // 2. 多 Collection 并行检索
    var wg sync.WaitGroup
    var toolExamples []ToolExample
    var assets []PlatformAsset
    var cases []TroubleshootingCase

    wg.Add(3)
    go func() {
        defer wg.Done()
        toolExamples, _ = r.searchToolExamples(ctx, queryEmbedding, topK/3)
    }()
    go func() {
        defer wg.Done()
        assets, _ = r.searchAssets(ctx, queryEmbedding, topK/3)
    }()
    go func() {
        defer wg.Done()
        cases, _ = r.searchCases(ctx, queryEmbedding, topK/3)
    }()
    wg.Wait()

    // 3. 构建增强上下文
    return &RAGContext{
        ToolExamples:        toolExamples,
        RelatedAssets:       assets,
        TroubleshootingCases: cases,
    }, nil
}

// BuildAugmentedPrompt 构建增强后的 Prompt
func (r *RAGRetriever) BuildAugmentedPrompt(query string, context *RAGContext) string {
    var sb strings.Builder

    if len(context.ToolExamples) > 0 {
        sb.WriteString("\n【相关工具调用示例】\n")
        for _, ex := range context.ToolExamples {
            sb.WriteString(fmt.Sprintf("- 工具: %s\n  意图: %s\n  参数: %s\n",
                ex.ToolName, ex.Intent, ex.ParamsJSON))
        }
    }

    if len(context.RelatedAssets) > 0 {
        sb.WriteString("\n【相关平台资源】\n")
        for _, asset := range context.RelatedAssets {
            sb.WriteString(fmt.Sprintf("- %s: %s (ID: %d)\n",
                asset.AssetType, asset.Name, asset.AssetID))
        }
    }

    if len(context.TroubleshootingCases) > 0 {
        sb.WriteString("\n【相关故障案例】\n")
        for _, c := range context.TroubleshootingCases {
            sb.WriteString(fmt.Sprintf("- %s\n  解决方案: %s\n",
                c.Title, c.Solution))
        }
    }

    return fmt.Sprintf("%s\n\n用户问题:\n%s", sb.String(), query)
}
```

---

## 4. SKILLS支持

### 4.1 配置文件格式

```yaml
# configs/skills.yaml

version: "1.0"

skills:
  # 服务部署技能
  - name: deploy_service
    display_name: "部署服务"
    description: "将服务部署到指定集群，包含预览、审批、执行三步"
    trigger_patterns:
      - "部署服务"
      - "发布服务"
      - "上线服务"
      - "deploy service"
    risk_level: high
    required_permissions:
      - "service:deploy"
    timeout_minutes: 60
    parameters:
      - name: service_id
        type: int
        required: true
        description: "要部署的服务ID"
        enum_source: "services"
      - name: cluster_id
        type: int
        required: true
        description: "目标集群ID"
        enum_source: "clusters"
    steps:
      - name: preview
        description: "预览部署变更"
        tool: service_deploy_preview
        params_template: |
          service_id: ${service_id}
          cluster_id: ${cluster_id}
      - name: approval
        type: approval
        message: "服务部署需要审批，请确认变更内容"
        timeout_minutes: 30
      - name: execute
        description: "执行部署"
        tool: service_deploy_apply
        params_template: |
          service_id: ${service_id}
          cluster_id: ${cluster_id}
          approval_token: ${approval.token}

  # 主机诊断技能
  - name: diagnose_host
    display_name: "主机诊断"
    description: "对主机进行全面的系统诊断，包括CPU、内存、磁盘、进程检查"
    trigger_patterns:
      - "诊断主机"
      - "主机检查"
      - "服务器体检"
      - "diagnose host"
    risk_level: low
    parameters:
      - name: host_id
        type: int
        required: true
        description: "目标主机ID"
        enum_source: "hosts"
      - name: host_keyword
        type: string
        required: false
        description: "主机名称关键词（作为备选）"
    steps:
      - name: resolve_host
        type: resolver
        description: "解析主机ID"
        resolver: resolve_host_id
        params:
          host_id: ${host_id}
          keyword: ${host_keyword}
      - name: cpu_memory
        description: "检查CPU和内存使用"
        tool: os_cpu_mem
        params_template: |
          target: ${resolved_host_id}
      - name: disk
        description: "检查磁盘使用"
        tool: os_disk
        params_template: |
          target: ${resolved_host_id}
      - name: processes
        description: "检查资源占用最高的进程"
        tool: os_process_top
        params_template: |
          target: ${resolved_host_id}
          limit: 10

  # 批量命令执行技能
  - name: batch_exec
    display_name: "批量执行命令"
    description: "在多台主机上批量执行命令"
    trigger_patterns:
      - "批量执行"
      - "批量运行命令"
      - "多台主机执行"
    risk_level: high
    required_permissions:
      - "host:execute"
    parameters:
      - name: host_ids
        type: "[]int"
        required: true
        description: "目标主机ID列表"
        enum_source: "hosts"
        multiple: true
      - name: command
        type: string
        required: true
        description: "要执行的命令"
      - name: reason
        type: string
        required: false
        description: "执行原因（用于审计）"
    steps:
      - name: preview
        tool: host_batch_exec_preview
        params_template: |
          host_ids: ${host_ids}
          command: ${command}
          reason: ${reason}
      - name: approval
        type: approval
        message: "批量命令执行需要审批"
        timeout_minutes: 30
      - name: execute
        tool: host_batch_exec_apply
        params_template: |
          host_ids: ${host_ids}
          command: ${command}
          reason: ${reason}
          approval_token: ${approval.token}
```

### 4.2 核心组件实现

```go
// internal/ai/skills/registry.go

type SkillRegistry struct {
    skills map[string]*Skill
    mu     sync.RWMutex
}

type Skill struct {
    Name               string
    DisplayName        string
    Description        string
    TriggerPatterns    []string
    RiskLevel          string
    RequiredPermissions []string
    TimeoutMinutes     int
    Parameters         []SkillParameter
    Steps              []SkillStep
}

type SkillParameter struct {
    Name        string
    Type        string
    Required    bool
    Description string
    EnumSource  string
    Multiple    bool
}

type SkillStep struct {
    Name           string
    Type          string // tool/approval/resolver
    Description   string
    Tool          string
    ParamsTemplate string
    Message       string
    TimeoutMinutes int
}

func LoadSkills(path string) (*SkillRegistry, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var config SkillsConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }

    registry := &SkillRegistry{skills: make(map[string]*Skill)}
    for _, s := range config.Skills {
        registry.skills[s.Name] = s
    }

    return registry, nil
}

func (r *SkillRegistry) MatchSkill(message string) (*Skill, float64) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    messageLower := strings.ToLower(message)

    for _, skill := range r.skills {
        for _, pattern := range skill.TriggerPatterns {
            if strings.Contains(messageLower, strings.ToLower(pattern)) {
                return skill, 1.0
            }
        }
    }

    return nil, 0
}
```

```go
// internal/ai/skills/executor.go

type SkillExecutor struct {
    registry     *SkillRegistry
    toolRegistry tools.ToolRegistry
    approver     *ApprovalService
    ragRetriever *rag.Retriever
}

type ExecutionContext struct {
    Skill         *Skill
    Params        map[string]any
    ResolvedParams map[string]any
    ApprovalToken string
    StepResults   map[string]any
}

func (e *SkillExecutor) Execute(ctx context.Context, skill *Skill, params map[string]any) (*SkillResult, error) {
    // 1. 参数验证
    if err := e.validateParams(skill, params); err != nil {
        return nil, err
    }

    execCtx := &ExecutionContext{
        Skill:       skill,
        Params:      params,
        StepResults: make(map[string]any),
    }

    // 2. 逐步执行
    for _, step := range skill.Steps {
        result, err := e.executeStep(ctx, step, execCtx)
        if err != nil {
            return nil, err
        }
        execCtx.StepResults[step.Name] = result

        // 处理审批步骤
        if step.Type == "approval" {
            approval, ok := result.(*ApprovalResult)
            if !ok || !approval.Approved {
                return nil, fmt.Errorf("审批未通过: %s", approval.Message)
            }
            execCtx.ApprovalToken = approval.Token
        }
    }

    return &SkillResult{
        SkillName: skill.Name,
        Success:   true,
        Results:   execCtx.StepResults,
    }, nil
}

func (e *SkillExecutor) executeStep(ctx context.Context, step SkillStep, execCtx *ExecutionContext) (any, error) {
    switch step.Type {
    case "tool":
        return e.executeToolStep(ctx, step, execCtx)
    case "approval":
        return e.executeApprovalStep(ctx, step, execCtx)
    case "resolver":
        return e.executeResolverStep(ctx, step, execCtx)
    default:
        return nil, fmt.Errorf("unknown step type: %s", step.Type)
    }
}

func (e *SkillExecutor) executeToolStep(ctx context.Context, step SkillStep, execCtx *ExecutionContext) (*ToolResult, error) {
    // 解析参数模板
    params := e.renderParamsTemplate(step.ParamsTemplate, execCtx)

    // 执行工具
    return e.toolRegistry.Execute(ctx, step.Tool, params)
}

func (e *SkillExecutor) executeApprovalStep(ctx context.Context, step SkillStep, execCtx *ExecutionContext) (*ApprovalResult, error) {
    // 创建审批请求
    approval, err := e.approver.CreateApproval(ctx, &ApprovalRequest{
        Skill:    execCtx.Skill.Name,
        Message:  step.Message,
        Params:   execCtx.Params,
        Timeout:  time.Duration(step.TimeoutMinutes) * time.Minute,
    })
    if err != nil {
        return nil, err
    }

    // 等待审批结果（阻塞或通过 WebSocket 通知）
    result, err := e.approver.WaitForApproval(ctx, approval.ID)
    if err != nil {
        return nil, err
    }

    return result, nil
}
```

### 4.3 与 PlatformAgent 集成

```go
// internal/ai/platform_agent.go (修改后)

type PlatformAgent struct {
    Runnable     *react.Agent
    Model        model.ToolCallingChatModel
    registry     experts.ExpertRegistry
    router       *experts.HybridRouter
    graphRunner  compose.Runnable[*aigraph.GraphInput, *aigraph.GraphOutput]
    streamRunner compose.Runnable[*aigraph.GraphInput, *schema.StreamReader[*schema.Message]]
    tools        map[string]tool.InvokableTool
    metas        map[string]tools.ToolMeta
    mcp          *tools.MCPClientManager

    // 新增: Skills 支持
    skillRegistry *skills.SkillRegistry
    skillExecutor *skills.SkillExecutor
    ragRetriever  *rag.Retriever
}

func (p *PlatformAgent) Stream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
    if p == nil {
        return nil, fmt.Errorf("agent not initialized")
    }

    // 1. 尝试匹配 Skill
    lastMessage := p.lastUserMessage(messages)
    if skill, confidence := p.skillRegistry.MatchSkill(lastMessage); skill != nil && confidence > 0.8 {
        params := p.extractSkillParams(skill, lastMessage, ctx)
        return p.executeSkillStream(ctx, skill, params)
    }

    // 2. 尝试 Expert 路由
    req := p.buildExecuteRequest(ctx, messages)
    if req != nil && req.Decision != nil {
        if p.streamRunner != nil {
            return p.streamRunner.Stream(ctx, p.buildGraphInput(req))
        }
    }

    // 3. 回退到默认 Agent
    return p.Runnable.Stream(ctx, messages)
}

func (p *PlatformAgent) executeSkillStream(ctx context.Context, skill *skills.Skill, params map[string]any) (*schema.StreamReader[*schema.Message], error) {
    sr, sw := schema.Pipe[*schema.Message](64)

    go func() {
        defer sw.Close()

        // 执行技能
        result, err := p.skillExecutor.Execute(ctx, skill, params)
        if err != nil {
            sw.Send(schema.AssistantMessage(fmt.Sprintf("技能执行失败: %v", err), nil), nil)
            return
        }

        // 格式化输出
        output := p.formatSkillResult(result)
        sw.Send(schema.AssistantMessage(output, nil), nil)
    }()

    return sr, nil
}
```

---

## 5. 文件变更总览

| 文件路径 | 变更类型 | 功能模块 |
|----------|----------|----------|
| `internal/ai/experts/orchestrator.go` | 删除 | Eino迁移 |
| `internal/ai/experts/orchestrator_test.go` | 删除 | Eino迁移 |
| `internal/ai/experts/orchestrator_primary_led_test.go` | 删除 | Eino迁移 |
| `internal/ai/graph/builder.go` | 修改 | Eino迁移 |
| `internal/ai/graph/runners.go` | 修改 | Eino迁移 |
| `internal/ai/graph/types.go` | 修改 | Eino迁移 |
| `internal/ai/platform_agent.go` | 修改 | 全部 |
| `internal/ai/experts/registry.go` | 修改 | Eino迁移 |
| `internal/service/ai/policy.go` | 修改 | 权限审批 |
| `internal/service/ai/permission_checker.go` | 新增 | 权限审批 |
| `internal/service/ai/approval_notifier.go` | 新增 | 权限审批 |
| `internal/model/ai_approval.go` | 新增 | 权限审批 |
| `internal/rag/milvus_client.go` | 新增 | RAG |
| `internal/rag/milvus_schema.go` | 新增 | RAG |
| `internal/rag/ingestion.go` | 新增 | RAG |
| `internal/rag/retriever.go` | 新增 | RAG |
| `internal/rag/embedder.go` | 新增 | RAG |
| `internal/ai/skills/registry.go` | 新增 | SKILLS |
| `internal/ai/skills/executor.go` | 新增 | SKILLS |
| `internal/ai/skills/router.go` | 新增 | SKILLS |
| `configs/skills.yaml` | 新增 | SKILLS |
| `configs/approval_config.yaml` | 新增 | 权限审批 |
| `go.mod` | 修改 | 依赖更新 |
