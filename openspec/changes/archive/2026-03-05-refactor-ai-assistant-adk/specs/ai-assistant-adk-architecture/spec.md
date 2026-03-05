# AI Assistant ADK Architecture

## Capability Overview

定义基于 eino ADK 的 AI 助手架构能力，包括 Plan-Execute-Replan 模式、StatefulInterrupt 审批机制、CheckPointStore 持久化。

## Capability Specification

### 1. Agent Architecture

**Requirement:** AI 助手必须使用 eino ADK 标准架构构建，采用 Plan-Execute-Replan 模式。

```
┌─────────────────────────────────────────────────────────────┐
│                    Platform Agent                            │
│                  (planexecute.New)                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐              │
│  │ Planner  │───▶│ Executor │───▶│Replanner │──┐           │
│  └──────────┘    └──────────┘    └──────────┘  │           │
│       │              │                           │           │
│       │              ▼                           │           │
│       │        ┌──────────┐                      │           │
│       │        │  Tools   │                      │           │
│       │        └──────────┘                      │           │
│       │              │                           │           │
│       └──────────────┴───────────────────────────┘           │
│                      │                                       │
│                      ▼                                       │
│              ┌──────────────┐                               │
│              │ CheckPoint   │                               │
│              │    Store     │                               │
│              └──────────────┘                               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Acceptance Criteria:**
- [ ] 使用 `planexecute.New()` 创建 Agent
- [ ] 支持 Planner/Executor/Replanner 三个组件
- [ ] 最大迭代次数可配置（默认 20）

### 2. Tool Approval Mechanism

**Requirement:** 高风险工具必须使用 `tool.StatefulInterrupt` 实现标准的审批中断流程。

**Approval Flow:**
```
┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│  Tool   │────▶│Interrupt│────▶│  User   │────▶│ Resume  │
│  Call   │     │  Save   │     │ Approve │     │Execute  │
└─────────┘     └─────────┘     └─────────┘     └─────────┘
                      │
                      ▼
               ┌─────────────┐
               │CheckPoint   │
               │   Store     │
               └─────────────┘
```

**Required Types:**
```go
type ApprovalInfo struct {
    ToolName        string         `json:"tool_name"`
    ArgumentsInJSON string         `json:"arguments"`
    Risk            ToolRisk       `json:"risk"`
    Preview         map[string]any `json:"preview"`
}

type ApprovalResult struct {
    Approved         bool    `json:"approved"`
    DisapproveReason *string `json:"disapprove_reason,omitempty"`
}
```

**Acceptance Criteria:**
- [ ] `ApprovalInfo` 和 `ApprovalResult` 已注册 schema
- [ ] 高风险工具使用 `ApprovableTool` 包装
- [ ] 中断状态保存到 CheckPointStore
- [ ] 恢复时正确读取审批结果

### 3. Tool Review Mechanism

**Requirement:** 中风险工具必须使用 `tool.StatefulInterrupt` 实现参数审核编辑流程。

**Required Types:**
```go
type ReviewEditInfo struct {
    ToolName        string `json:"tool_name"`
    ArgumentsInJSON string `json:"arguments"`
}

type ReviewEditResult struct {
    EditedArgumentsInJSON *string `json:"edited_arguments,omitempty"`
    NoNeedToEdit          bool    `json:"no_need_to_edit"`
    Disapproved           bool    `json:"disapproved"`
    DisapproveReason      *string `json:"disapprove_reason,omitempty"`
}
```

**Acceptance Criteria:**
- [ ] `ReviewEditInfo` 和 `ReviewEditResult` 已注册 schema
- [ ] 中风险工具使用 `ReviewableTool` 包装
- [ ] 用户可编辑参数后执行
- [ ] 用户可拒绝执行

### 4. CheckPoint Persistence

**Requirement:** 必须实现 `compose.CheckPointStore` 接口，使用数据库持久化。

**Interface:**
```go
type CheckPointStore interface {
    Set(ctx context.Context, key string, value []byte) error
    Get(ctx context.Context, key string) ([]byte, bool, error)
}
```

**Database Schema:**
```sql
CREATE TABLE ai_checkpoints (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `key` VARCHAR(255) NOT NULL UNIQUE,
    value MEDIUMBLOB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_key (`key`)
);
```

**Acceptance Criteria:**
- [ ] 实现 `DBCheckPointStore` 结构体
- [ ] Set 方法支持 upsert
- [ ] Get 方法正确处理 not found 场景
- [ ] 支持并发访问

### 5. Tool Risk Classification

**Requirement:** 所有工具必须明确风险等级，并应用对应包装器。

**Risk Levels:**
| Level | Wrapper | Behavior |
|-------|---------|----------|
| Low | None | Direct execution |
| Medium | ReviewableTool | User can edit parameters |
| High | ApprovableTool | Requires explicit approval |

**Tool Classification:**
```go
type ToolRisk string

const (
    ToolRiskLow    ToolRisk = "low"
    ToolRiskMedium ToolRisk = "medium"
    ToolRiskHigh   ToolRisk = "high"
)
```

**Acceptance Criteria:**
- [ ] 所有工具定义风险等级
- [ ] 高风险工具有预览函数
- [ ] 构建函数正确应用包装器

### 6. SSE Event Format

**Requirement:** SSE 事件格式必须与现有前端兼容。

**Event Types:**
| Event | Description |
|-------|-------------|
| `delta` | Streaming content chunk |
| `approval_required` | Tool needs approval |
| `review_required` | Tool needs parameter review |
| `tool_call` | Tool invocation |
| `tool_result` | Tool result |
| `error` | Error occurred |
| `done` | Execution complete |

**Approval Event Payload:**
```json
{
    "tool": "host_batch_exec_apply",
    "arguments": "{\"host_ids\":[1,2,3],\"command\":\"df -h\"}",
    "risk": "high",
    "preview": {
        "target_count": 3,
        "resolved_hosts": ["host1", "host2", "host3"]
    }
}
```

**Acceptance Criteria:**
- [ ] 事件格式与现有实现一致
- [ ] 前端无需改动

### 7. HTTP Handler

**Requirement:** HTTP 处理器必须使用 `adk.Runner`，代码量控制在 200 行以内。

**Handler Pattern:**
```go
func (h *handler) chat(c *gin.Context) {
    runner := adk.NewRunner(ctx, adk.RunnerConfig{
        EnableStreaming: true,
        Agent:           h.agent,
    })

    iter := runner.Query(ctx, query)

    for {
        event, ok := iter.Next()
        if !ok {
            break
        }
        h.processEvent(c, flusher, event)
    }
}
```

**Acceptance Criteria:**
- [ ] 使用 `adk.Runner` 处理请求
- [ ] 代码行数 < 200
- [ ] 保持 API 兼容性

## Integration Points

### With Existing Systems

| System | Integration |
|--------|-------------|
| Database | `ai_checkpoints` table, existing session tables |
| RBAC | Tool permission checking |
| Audit | Command execution logging |
| MCP | MCP proxy tools |

### Dependencies

| Dependency | Version | Purpose |
|------------|---------|---------|
| github.com/cloudwego/eino | latest | ADK framework |
| github.com/cloudwego/eino-ext | latest | Model providers |

## Testing Requirements

### Unit Tests
- [ ] `ApprovableTool` 中断恢复逻辑
- [ ] `ReviewableTool` 参数编辑逻辑
- [ ] `DBCheckPointStore` 读写操作
- [ ] 工具风险分级验证

### Integration Tests
- [ ] Plan-Execute-Replan 完整流程
- [ ] 审批中断恢复流程
- [ ] SSE 事件流验证

### End-to-End Tests
- [ ] 用户对话完整流程
- [ ] 审批确认交互流程

## Migration Notes

### Breaking Changes
- 无 API 破坏性变更
- SSE 事件格式保持兼容

### Data Migration
- 新增 `ai_checkpoints` 表
- 现有会话数据无需迁移

### Rollback Strategy
- 功能开关控制新旧实现
- CheckPoint 表保留不影响旧逻辑
