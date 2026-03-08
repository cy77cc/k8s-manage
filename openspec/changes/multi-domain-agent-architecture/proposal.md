## Why

当前 AI Agent 架构使用单一的 `planexecute` 模式，所有领域工具（K8s、主机、服务、监控等 50+ 工具）混在一起交给 Planner 选择。这导致：

1. **Planner 选择压力大**：需要在 50+ 工具中定位正确的工具，容易选错或遗漏
2. **跨领域任务支持弱**：无法优雅处理涉及多个领域的复杂任务
3. **工具职责不清**：Discovery 工具（查 ID）和 Action 工具（执行操作）混杂，Executor 承担了过多职责
4. **扩展性差**：新增领域工具会进一步增加 Planner 负担

需要重构为多领域并行的 Agent 架构，实现领域隔离、职责分离、并行执行。

## What Changes

- 新增 `Orchestrator Planner`：负责意图分析、领域识别、任务分发，不关心执行细节
- 新增多个 `Domain Planner`：每个领域一个 Planner，只做规划不执行，输出带依赖关系的 Action Plan
- 重构 `Executor`：合并多领域 Plan，构建全局 DAG，按拓扑顺序执行所有工具
- 新增 `Replanner`：检查执行结果完整性，决定是否需要重规划
- 工具分层：Discovery 工具（只读查询）由 Planner 调用补全参数，Action 工具由 Executor 执行

## Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                      Orchestrator Planner                         │
│                                                                  │
│   职责: 意图分析 + 领域识别 + 任务分发（并行）                       │
│   输入: 用户请求                                                   │
│   输出: [Domain1, Domain2, ...]                                   │
└────────────────────────────┬─────────────────────────────────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│ Infra Planner   │ │ Service Planner │ │ Monitor Planner │
│                 │ │                 │ │                 │
│ 只规划不执行     │ │ 只规划不执行     │ │ 只规划不执行     │
│                 │ │                 │ │                 │
│ 输出: DomainPlan│ │ 输出: DomainPlan│ │ 输出: DomainPlan│
│ (steps + deps)  │ │ (steps + deps)  │ │ (steps + deps)  │
└────────┬────────┘ └────────┬────────┘ └────────┬────────┘
         │                   │                   │
         └───────────────────┼───────────────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │    Executor     │
                    │                 │
                    │  合并所有 Plan   │
                    │  构建全局 DAG    │
                    │  按依赖执行工具  │
                    │  处理跨领域数据  │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │   Replanner     │
                    │                 │
                    │  检查结果完整性  │
                    │  决定是否重规划  │
                    └─────────────────┘
```

## Domain Boundaries

| 领域 | Planner | 核心工具 | 说明 |
|------|---------|----------|------|
| Infrastructure | InfraPlanner | `host_*`, `k8s_*`, `os_*` | 主机、K8s、容器运行时 |
| Service | ServicePlanner | `service_*`, `deployment_*`, `credential_*` | 服务、部署目标、凭证 |
| CICD | CICDPlanner | `cicd_*`, `job_*` | 流水线、任务 |
| Monitor | MonitorPlanner | `monitor_*`, `topology_*` | 监控、告警、拓扑 |
| Config | ConfigPlanner | `config_*` | 配置管理 |
| User | UserPlanner | `user_*`, `role_*`, `permission_*` | 用户、角色、权限 |

## Tool Layering

| 层级 | 工具类型 | 示例 | 调用者 |
|------|----------|------|--------|
| Discovery | 只读查询，获取 ID/枚举 | `host_list_inventory`, `service_list_inventory` | Planner（规划阶段调用） |
| Action | 真正的操作 | `service_deploy_apply`, `host_exec` | Executor（执行阶段调用） |

## Plan Format

每个 Domain Planner 输出标准化的 DomainPlan：

```go
type DomainPlan struct {
    Domain string
    Steps  []PlanStep
}

type PlanStep struct {
    ID        string            // 步骤唯一标识
    Tool      string            // 工具名称
    Params    map[string]any    // 静态参数 或 {$ref: "step_id.field"}
    DependsOn []string          // 领域内依赖
    Produces  []string          // 输出的变量名
    Requires  []string          // 需要从外部获取的变量
}
```

## Capabilities

### New Capabilities

- `multi-domain-agent-orchestration`: 定义多领域 Agent 编排能力，包括 Orchestrator、Domain Planner、Executor、Replanner 的职责边界与交互协议
- `domain-plan-dsl`: 定义 DomainPlan DSL 规范，包括步骤定义、依赖声明、变量引用语法

### Modified Capabilities

- `ai-assistant-command-bridge`: 适配新的多领域架构，调整命令路由与执行编排逻辑
- `ai-tool-registry`: 按领域重新组织工具注册，支持 Discovery/Action 分层

## Impact

### Backend

- `internal/ai/` 目录重构
  - 新增 `orchestrator/` - Orchestrator Planner 实现
  - 新增 `planner/` - Domain Planner 注册与实现
  - 新增 `executor/` - 合并 Plan、构建 DAG、执行工具
  - 新增 `replanner/` - 结果检查与重规划决策
  - 重构 `tools/` - 按领域分组，支持 Discovery/Action 标记
- `internal/ai/modes/agentic.go` - 适配新架构入口
- `internal/ai/hybrid.go` - 调整 HybridAgent 调用链

### API

- 无新增 API，内部架构重构

### Frontend

- 无直接影响，SSE 事件类型保持兼容

### Data

- 无数据库变更

## Risks

1. **迁移风险**：现有单领域 Plan-Execute 逻辑需要平滑过渡
   - 缓解：保留旧架构作为 fallback，通过配置开关切换

2. **性能开销**：多层 Planner 增加延迟
   - 缓解：Planner 并行执行，领域内 Discovery 工具缓存结果

3. **跨领域依赖复杂性**：变量引用语法可能导致执行失败
   - 缓解：Executor 在执行前校验 DAG 完整性，缺失变量时提前报错

## Success Criteria

- [ ] 单领域任务：与现有架构效果一致
- [ ] 跨领域并行任务：多个领域 Planner 并行输出 Plan，Executor 正确合并执行
- [ ] 跨领域依赖任务：Executor 正确处理 `{$ref}` 变量引用，按依赖顺序执行
- [ ] 工具分层：Discovery 工具被 Planner 调用，Action 工具被 Executor 调用
