## Context

当前 AI 助手模块存在架构与 eino ADK 不兼容的问题，需要完整重构。该变更涉及：

- 删除 3 个自建框架目录（experts、graph、callbacks）
- 重写核心 Agent 构建逻辑
- 改造工具系统使用 StatefulInterrupt
- 简化 HTTP 处理层

属于高复杂度、高风险的重构变更，但对系统长期可维护性至关重要。

## Goals / Non-Goals

**Goals:**
- 完全拥抱 eino ADK 标准架构，删除所有自建框架代码
- 使用 Plan-Execute-Replan 模式实现多步骤任务编排
- 使用 StatefulInterrupt 实现标准的审批/确认流程
- 使用 adk.Runner 简化事件处理
- 实现 CheckPointStore 数据库持久化
- 保持现有 API 和 SSE 事件格式的向后兼容

**Non-Goals:**
- 不新增工具功能（保留现有 50+ 工具）
- 不改变前端交互
- 不引入新的 LLM 提供商
- 不实现多租户隔离（保持现有单租户模式）

## Architecture

### Before (Current)

```
internal/ai/
├── client.go                    # 模型创建
├── platform_agent.go            # 主 Agent (聚合太多职责)
├── experts/                     # 自建专家系统 (应删除)
│   ├── registry.go              # 专家注册
│   ├── router.go                # 专家路由
│   ├── executor.go              # 专家执行
│   ├── aggregator.go            # 结果聚合
│   └── ...
├── graph/                       # 自建图编排 (应删除)
│   ├── builder.go               # 图构建器
│   ├── runners.go               # 运行器
│   └── ...
├── callbacks/                   # 自建回调 (应删除)
│   ├── handler.go               # 回调处理器
│   ├── emitter.go               # 事件发射器
│   └── ...
└── tools/                       # 工具系统 (保留改造)
    ├── tools_registry.go        # 800+ 行注册代码
    └── ...

internal/service/ai/
├── chat_handler.go              # 600+ 行 (应简化)
├── events_sse.go                # SSE 处理
└── store.go                     # 会话存储
```

### After (Target)

```
internal/ai/
├── agent.go                     # PlatformAgent (使用 ADK)
├── model.go                     # 模型创建 (简化)
├── store.go                     # CheckPointStore 实现
└── tools/
    ├── wrapper.go               # StatefulInterrupt 包装器
    ├── ops.go                   # OS 工具集
    ├── k8s.go                   # K8s 工具集
    ├── service.go               # 服务管理工具集
    ├── deployment.go            # 部署工具集
    ├── cicd.go                  # CI/CD 工具集
    ├── monitor.go               # 监控工具集
    ├── config.go                # 配置管理工具集
    ├── governance.go            # 治理工具集
    ├── inventory.go             # 资产清单工具集
    ├── contracts.go             # 工具类型定义
    └── mcp.go                   # MCP 代理工具

internal/service/ai/
├── handler.go                   # 简化的 HTTP 处理 (~150 行)
├── events.go                    # SSE 事件格式定义
└── store.go                     # 会话存储 (保留)
```

## Decisions

### Decision 1: 使用 Plan-Execute-Replan 而非 Supervisor

**Choice:** 使用 `planexecute.New()` 构建 Agent

**Rationale:**
- Plan-Execute-Replan 更适合运维场景，可以动态调整执行计划
- Supervisor 模式适合固定角色的多 Agent 协作，运维场景更偏向任务编排
- eino-examples 中的 `plan-execute-replan` 示例提供了完整的参考实现

**Alternative considered:** 使用 `supervisor.New()`
- 缺点：需要预先定义角色分工，不够灵活
- 缺点：动态任务编排能力弱于 Plan-Execute-Replan

### Decision 2: 使用 StatefulInterrupt 而非自定义审批票据

**Choice:** 使用 `tool.StatefulInterrupt` + `schema.Register` 实现审批

**Rationale:**
- 这是 eino 标准的中断恢复机制
- 框架自动处理状态保存和恢复
- 与 ADK Runner 完美集成
- 审批结果类型安全

**Code Example:**
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

func init() {
    schema.Register[*ApprovalInfo]()
    schema.Register[*ApprovalResult]()
}

func (t *ApprovableTool) InvokableRun(ctx context.Context, args string, opts ...tool.Option) (string, error) {
    wasInterrupted, _, storedArgs := tool.GetInterruptState[string](ctx)

    if !wasInterrupted {
        // 首次调用 → 中断请求审批
        return "", tool.StatefulInterrupt(ctx, &ApprovalInfo{
            ToolName:        t.info.Name,
            ArgumentsInJSON: args,
            Risk:            t.risk,
            Preview:         t.previewFn(ctx, args),
        }, args)
    }

    // 检查审批结果
    _, hasData, result := tool.GetResumeContext[*ApprovalResult](ctx)
    if !hasData || !result.Approved {
        return "审批已拒绝", nil
    }

    // 审批通过 → 执行
    return t.tool.InvokableRun(ctx, storedArgs, opts...)
}
```

**Alternative considered:** 保持现有审批票据机制
- 缺点：与 ADK Runner 不兼容
- 缺点：状态管理不标准，需要手动处理恢复逻辑

### Decision 3: 工具分级包装策略

**Choice:** 按风险等级使用不同的包装器

| 风险等级 | 包装器 | 行为 |
|---------|--------|------|
| Low | 无包装 | 直接执行 |
| Medium | ReviewEditTool | 用户可编辑参数后执行 |
| High | ApprovableTool | 需要明确审批才能执行 |

**Rationale:**
- 分级处理避免过度打扰用户
- 高风险操作必须审批，符合安全合规要求
- 中风险操作允许用户修正参数

**Code Example:**
```go
func BuildAllTools(ctx context.Context, deps PlatformDeps) ([]tool.BaseTool, error) {
    var allTools []tool.BaseTool

    // Low risk - 直接使用
    for _, t := range buildLowRiskTools(deps) {
        allTools = append(allTools, t)
    }

    // Medium risk - ReviewEdit 包装
    for _, t := range buildMediumRiskTools(deps) {
        allTools = append(allTools, &ReviewableTool{InvokableTool: t})
    }

    // High risk - Approval 包装
    for _, t := range buildHighRiskTools(deps) {
        allTools = append(allTools, &ApprovableTool{
            InvokableTool: t,
            previewFn:     buildPreviewFn(t, deps),
        })
    }

    return allTools, nil
}
```

### Decision 4: CheckPointStore 数据库实现

**Choice:** 实现 `compose.CheckPointStore` 接口，使用数据库持久化

**Rationale:**
- 支持跨请求的状态恢复（如审批等待）
- 与 ADK Runner 集成，自动保存执行状态
- 可用于实现长期运行任务的恢复

**Schema:**
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

**Alternative considered:** 使用内存存储
- 缺点：服务重启后状态丢失，审批恢复失败
- 缺点：无法支持分布式部署

### Decision 5: SSE 事件格式适配

**Choice:** 保持现有 SSE 事件格式，适配 `adk.AgentEvent`

**Mapping:**
| adk.AgentEvent | SSE Event Name |
|----------------|----------------|
| `event.Output.MessageStream` | `delta` |
| `event.Action.Interrupted` (ApprovalInfo) | `approval_required` |
| `event.Action.Interrupted` (ReviewEditInfo) | `review_required` |
| `event.Action.ToolCalls` | `tool_call` |
| `event.Err` | `error` |
| `event.Action.Exit` | `done` |

**Rationale:**
- 前端无需改动
- 向后兼容现有 SSE 消费逻辑

## Risks / Trade-offs

### Risk 1: 功能回归
- **Risk:** 重构可能导致部分功能缺失或行为变化
- **Mitigation:**
  - 保留现有测试用例
  - 增加集成测试覆盖
  - 灰度发布，监控错误率

### Risk 2: 审批流程变更
- **Risk:** 新的 StatefulInterrupt 流程与现有审批票据不兼容
- **Mitigation:**
  - 保持 SSE 事件格式不变
  - 前端无需感知后端实现变化
  - 提供平滑迁移路径

### Risk 3: 性能影响
- **Risk:** Plan-Execute-Replan 模式可能引入额外延迟
- **Mitigation:**
  - 简单查询使用单步执行
  - 复杂任务才启用完整流程
  - 监控执行延迟指标

### Risk 4: CheckPoint 存储开销
- **Risk:** 频繁的 CheckPoint 写入可能影响数据库
- **Mitigation:**
  - 使用 MEDIUMBLOB 存储压缩数据
  - 设置合理的过期清理策略
  - 可选的 Redis 缓存层

## Migration Plan

### Phase 1: 基础设施 (Week 1)
1. 创建 `ai_checkpoints` 表
2. 实现 `DBCheckPointStore`
3. 创建 `tools/wrapper.go`（ApprovableTool、ReviewableTool）
4. 单元测试覆盖

### Phase 2: Agent 重构 (Week 2)
1. 创建新的 `agent.go`（使用 planexecute）
2. 拆分 `tools_registry.go` 为多个工具集文件
3. 为工具添加正确的包装器
4. 集成测试

### Phase 3: HTTP 层简化 (Week 3)
1. 重写 `handler.go`（使用 adk.Runner）
2. 实现 SSE 事件适配器
3. 保持 API 兼容性测试
4. 端到端测试

### Phase 4: 清理与验证 (Week 4)
1. 删除 `experts/`、`graph/`、`callbacks/` 目录
2. 删除旧的 `platform_agent.go`
3. 更新依赖引用
4. 全量回归测试
5. 性能基准测试

### Rollback Strategy
- 通过功能开关回退到旧实现
- 保留旧代码在 `legacy/` 目录 2 个版本周期
- CheckPoint 表保留，不影响旧逻辑

## Testing Strategy

### Unit Tests
- `ApprovableTool` 中断恢复逻辑
- `ReviewableTool` 参数编辑逻辑
- `DBCheckPointStore` 读写操作
- 工具注册与包装逻辑

### Integration Tests
- Plan-Execute-Replan 完整流程
- 审批中断恢复流程
- SSE 事件流验证
- CheckPoint 持久化验证

### End-to-End Tests
- 用户对话完整流程
- 审批确认交互流程
- 多步骤任务执行流程
- 错误恢复流程

### Performance Tests
- 单次查询延迟基准
- 并发请求处理能力
- CheckPoint 写入性能
- 内存占用监控

## Open Questions

1. **是否需要保留 RAG 增强？**
   - 现有 `ragRetriever` 在重构中如何处理？
   - 是否需要作为独立的工具实现？

2. **MCP 集成如何适配？**
   - 现有 `mcp_client.go` 是否需要改造？
   - MCP 工具是否需要审批包装？

3. **会话历史如何迁移？**
   - 现有会话历史是否需要转换格式？
   - CheckPoint 与会话历史的关系？

4. **是否需要反思循环？**
   - 对于复杂诊断任务是否增加反思环节？
   - 如何判断何时启用反思？
