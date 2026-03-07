## MODIFIED Requirements

### Requirement: Agent Architecture

AI 助手 MUST 使用 eino ADK 标准架构构建，采用 Plan-Execute-Replan 模式，并且该编排架构的宿主 MUST 位于 `internal/ai` 中。该架构 MUST 明确区分 Planner、Executor、Replanner 三种角色，并通过 domain executor 路由执行 Host、K8s、Service、Monitor 等运维任务，而不是仅依赖单一平台 agent 处理所有执行场景。

```
┌─────────────────────────────────────────────────────────────┐
│                    AI AIOps Control Plane                   │
│                     (hosted in internal/ai)                 │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────┐    ┌──────────────┐    ┌──────────┐           │
│  │ Planner  │───▶│ Executor     │───▶│Replanner │──┐        │
│  └──────────┘    │ Router       │    └──────────┘  │        │
│       │          └──────┬───────┘                  │        │
│       │                 │                          │        │
│       │      ┌──────────┼──────────┬──────────┐    │        │
│       │      ▼          ▼          ▼          ▼    │        │
│       │   Host Exec   K8s Exec  Service Exec Monitor│       │
│       │                                              │       │
│       └──────────────────────┬───────────────────────┘       │
│                              │                               │
│                              ▼                               │
│                      ┌──────────────┐                        │
│                      │ CheckPoint   │                        │
│                      │ ControlPlane │                        │
│                      └──────────────┘                        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
                     ▲
                     │ delegate
┌─────────────────────────────────────────────────────────────┐
│                     AI Gateway Layer                         │
│       (internal/service/ai routes/handlers/SSE transport)   │
└─────────────────────────────────────────────────────────────┘
```

**Acceptance Criteria:**
- [ ] 使用 ADK 构建明确的 Planner/Executor/Replanner 角色
- [ ] 支持 Planner/Executor/Replanner 三个组件
- [ ] 最大迭代次数可配置（默认 20）
- [ ] `internal/service/ai` 通过委托 AI core 触发 ADK 编排，而不是在 handler 中本地承载编排所有权
- [ ] Host、K8s、Service、Monitor 至少具备正式的 domain executor 宿主

#### Scenario: orchestration ownership lives in the AI core
- **WHEN** reviewers inspect the ADK architecture
- **THEN** the Plan-Execute-Replan orchestration host MUST be defined under `internal/ai`
- **AND** gateway handlers MUST delegate into the AI core instead of acting as the orchestration owner
- **AND** execution routing MUST be able to dispatch work through domain executor boundaries

### Requirement: SSE Event Format

SSE 事件格式 MUST 与现有前端兼容，并且事件语义的来源 MUST 位于 AI core 中。gateway SHALL 仅负责传输封装与兼容序列化，同时 AI core MUST 提供面向 application-card 的平台事件语义。

**Event Types:**
| Event | Description |
|-------|-------------|
| `meta` | Turn/session metadata |
| `delta` | Compatibility streaming content chunk |
| `thinking_delta` | Compatibility reasoning chunk |
| `tool_call` | Tool invocation |
| `tool_result` | Tool result |
| `approval_required` | Tool needs approval |
| `error` | Error occurred |
| `done` | Execution complete |

**Extended Platform Event Family:**
| Event | Description |
|-------|-------------|
| `plan_created` | Plan created for an AIOps task |
| `step_status` | Step state transition |
| `evidence` | Structured operational evidence |
| `ask_user` | Flow-selection or clarification interaction |
| `replan_decision` | Replanning decision |
| `summary` | Final task summary |
| `next_actions` | Recommended next actions |

**Acceptance Criteria:**
- [ ] 现有 SSE 兼容事件仍可投影给当前前端
- [ ] 新的平台事件语义由 `internal/ai` 产生
- [ ] gateway 仅负责传输兼容包装
- [ ] 平台事件可支持 application-card 渲染

#### Scenario: gateway preserves transport compatibility while AI core owns event meaning
- **WHEN** AI execution events are streamed to the frontend
- **THEN** the SSE family MUST remain compatible with the current frontend during rollout
- **AND** the semantic meaning of execution, interrupt, plan, evidence, and outcome events MUST be produced by the AI core

### Requirement: HTTP Handler

HTTP 处理器 MUST 作为 AI gateway 使用 AI control-plane orchestration entrypoint，并且 MUST NOT 继续在 handler 内承载主要编排逻辑。

**Handler Pattern:**
```go
func (h *handler) chat(c *gin.Context) {
    result := h.aiControlPlane.RunTask(ctx, request)
    h.streamResult(c, result)
}
```

**Acceptance Criteria:**
- [ ] handler 负责 request bind、auth/session shell、SSE transport 与错误收尾
- [ ] handler 通过 AI control-plane orchestration entrypoint 执行 chat/resume 流程
- [ ] handler 不直接成为 Planner/Executor/Replanner 的宿主

#### Scenario: handler acts as gateway rather than orchestration owner
- **WHEN** a chat or resume request is processed
- **THEN** the HTTP handler MUST delegate orchestration to the AI core
- **AND** the handler MUST remain limited to gateway and transport responsibilities
