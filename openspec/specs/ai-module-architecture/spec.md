# AI Module Architecture Spec

## Overview

AI 模块基于 Eino 框架构建，实现确定性、安全性与云原生基础设施深度集成的 AI-PaaS 平台核心能力。

## Core Principles

1. **确定性** - LLM 推理结果需经过校验才能执行
2. **安全性** - 风险操作需人工审批，权限校验前置
3. **可观测性** - 完整的追踪与审计能力

## Architecture

### IntentRouter

**职责**: 解析用户输入，路由至对应的业务 ActionGraph。

**实现**: `compose.Graph` + Branch

**路由规则**:

| Domain | 工具前缀 | 说明 |
|--------|---------|------|
| `infrastructure` | `host_*`, `cluster_*` | 主机、集群操作 |
| `service` | `service_*`, `deploy_*` | 服务、部署操作 |
| `cicd` | `cicd_*`, `job_*` | CI/CD 操作 |
| `monitor` | `monitor_*`, `alert_*` | 监控、告警操作 |
| `config` | `config_*` | 配置操作 |
| `general` | 其他 | 通用问答 |

### ActionGraph

**职责**: 确定性工作流执行。

**实现**: `compose.Workflow`

**节点**:

1. `sanitize` - 输入脱敏，注入多租户上下文
2. `reasoning` - LLM 推理，生成回复或工具调用
3. `validation` - JSON Schema + K8s OpenAPI 校验
4. `execution` - 工具执行（含 SecurityAspect）

**状态**:

```go
type GraphState struct {
    Messages          []*schema.Message
    PendingToolCalls  []ToolCallSpec
    ToolResults       map[string]ToolResult
    ValidationErrors  []ValidationError
}
```

### SecurityAspect

**职责**: 工具调用的安全切面。

**实现**: Eino Callbacks + Tool Middleware

**决策**:

| 条件 | 决策 | 后续动作 |
|------|------|---------|
| 无权限 | `DecisionCreateApproval` | 创建审批任务，中断流程 |
| 有权限 + 高风险 | `DecisionInterrupt` | LLM 说明原因，等待用户确认 |
| 有权限 + 低风险 | `DecisionExecute` | 直接执行 |

### Approval System

**双轨制**:

1. **有权限场景**: LLM Interrupt → 用户确认 → Resume Graph
2. **无权限场景**: 创建审批任务 → 审批人审批 → 执行器执行

**审批人路由**: 按资源类型路由到对应负责人。

### RAG System

**数据来源**:

1. 用户主动投喂（文档、FAQ）
2. 反馈收集（有效解决方案自动入库）

**多租户隔离**: 通过 Namespace 字段实现。

## Data Models

### ApprovalTask

```go
type ApprovalTask struct {
    ID             uint64
    RequesterID    uint64
    Status         string  // pending/approved/rejected/executed
    ResourceType   string
    ResourceID     string
    TaskDetail     TaskDetail
    ToolCalls      []ToolCallSpec
    ApproverID     uint64
    ApprovedAt     *time.Time
    ExecutedAt     *time.Time
}
```

### TaskDetail

```go
type TaskDetail struct {
    Summary        string
    Steps          []ExecutionStep
    RiskAssessment RiskAssessment
    RollbackPlan   string
}
```

## API

### Chat

```
POST /ai/chat
Request:  { "session_id": "xxx", "message": "..." }
Response: SSE Stream
```

### Approval Response (有权限)

```
POST /ai/chat/respond
Request:  { "session_id": "xxx", "approved": true }
Response: SSE Stream (Resume Graph)
```

### Approval Task (无权限)

```
POST /ai/approval/create
GET  /ai/approval/list
GET  /ai/approval/:id
POST /ai/approval/:id/approve
POST /ai/approval/:id/reject
```

### Feedback

```
POST /ai/feedback
Request:  { "session_id": "xxx", "effective": true }
```

## Error Handling

| 错误类型 | 处理方式 |
|---------|---------|
| 意图分类失败 | 路由到 General_Assistance |
| 校验失败 | 返回错误信息，不执行 |
| 权限不足 | 创建审批任务 |
| 执行失败 | 记录日志，返回错误信息 |

## Observability

- **TraceID**: 每个 Graph 执行生成唯一 TraceID
- **审计日志**: 所有工具调用记录到 audit_log 表
- **指标**: 工具调用次数、延迟、成功率
