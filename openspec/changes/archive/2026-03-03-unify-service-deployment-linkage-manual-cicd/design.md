## Context

当前发布能力分散在 `internal/service/service`、`internal/service/deployment`、`internal/service/cicd` 三条链路，分别落库到 `service_release_records`、`deployment_releases`、`cicd_releases`。手动发布与 CI/CD 发布在状态定义、审批触发和审计口径上不一致，导致前端页面与运维排障都需要跨模块拼接信息。

本变更是跨域改造，涉及 API 语义收敛、状态机统一、数据模型兼容迁移和前端展示口径统一，需要先形成明确设计后再进入编码。

## Goals / Non-Goals

**Goals:**
- 建立统一的 Release 编排入口，手动与 CI/CD 两种触发共享同一状态机和审批门禁。
- 统一发布审计与时间线口径，保证按 `release_id` 可追踪从触发到回滚全过程。
- 为旧接口提供兼容映射，降低一次性迁移风险。
- 前后端统一展示发布来源（manual/ci）与版本上下文（CI run、artifact、revision）。

**Non-Goals:**
- 不在本变更引入新的发布策略类型（仅沿用 rolling/blue-green/canary/rollback）。
- 不在本变更重做部署可视化页面的完整交互，仅调整到统一数据源。
- 不在本变更引入外部 CI 执行器，仅定义本平台内的触发与编排契约。

## Decisions

### Decision 1: 以 deployment release 生命周期作为统一事实源
- 选择：以 `deployment_releases` 生命周期为统一发布事实源，手动发布与 CI/CD 均归并到该状态机。
- 原因：该链路已具备 preview token、审批、回滚、runtime 验证、timeline 审计等完整语义，扩展成本最低。
- 备选：
  - 扩展 `cicd_releases` 为唯一源：会丢失现有 deployment 的 preview 与 runtime 验证模型，需要大规模回填。
  - 保留三套并行并新增聚合层：短期快，但长期继续放大一致性负担。

### Decision 2: 保留现有 API 路径，但统一到单一编排逻辑
- 选择：`/services/:id/deploy` 与 `/cicd/releases` 保留入口语义，内部都转为统一 release create/approve/apply/rollback 流程。
- 原因：降低前端与外部调用改造量，支持渐进迁移。
- 备选：
  - 直接废弃旧 API：迁移窗口过陡，风险高。

### Decision 3: 增加触发来源与关联上下文
- 选择：统一 release 记录必须带 `trigger_source`（manual/ci）和 `trigger_context`（ci_run_id、artifact、operator、service_revision 等）。
- 原因：解决“同一服务不同入口触发后无法串联”的审计与排障问题。
- 备选：
  - 仅在审计表附加来源：检索和页面过滤成本高，难做一致状态统计。

### Decision 4: 服务侧手动部署转为“创建发布请求”语义
- 选择：服务模块不再维护独立执行状态，而是负责组装 release draft 并调用统一编排。
- 原因：避免重复执行器、重复审批判断和重复状态转换。
- 备选：
  - 继续 service 自执行，再异步同步 deployment：一致性与补偿逻辑复杂。

## Risks / Trade-offs

- [历史数据口径不一致] 旧 `service_release_records` 与新统一 release 难直接对齐  
  → Mitigation: 提供只读兼容查询映射与时间线合并视图，增量迁移不强制回写历史。

- [接口语义漂移] 旧接口继续存在可能导致“看起来没变，行为已变”  
  → Mitigation: 在 API 文档与返回字段中显式增加 `trigger_source`、`unified_release_id` 并发布迁移说明。

- [审批规则冲突] service 与 deployment 现有审批策略可能不一致  
  → Mitigation: 统一以 deployment CD config + env policy 为唯一审批判断来源。

- [前端短期复杂度上升] 页面需要兼容旧字段与新字段  
  → Mitigation: 先改 API adapter 层（`web/src/api/modules/*`）做字段归一，再改页面消费。

## Migration Plan

1. 数据层先补充统一 release 来源与上下文字段（含索引），保持旧表可读。
2. 后端新增统一 release 编排服务（内部），先接入 CI/CD 触发链路。
3. 将 service 手动部署入口改为调用统一编排，并产出兼容响应字段。
4. 前端 API 模块切换到统一 release 响应结构，页面逐步替换。
5. 发布窗口内保留旧查询接口；稳定后评估下线重复 release 记录写入。

回滚策略：如出现异常，可临时切回旧 service deploy 直发逻辑开关，并保持 deployment/cicd 原有流程不受影响。

## Open Questions

- 旧 `service_release_records` 是否保留写入（双写）多久，还是立刻降级为只读历史。
- `trigger_context` 字段最小必填集合是否包含 commit SHA 和 artifact digest。
- 统一 release API 是否需要显式区分“草稿创建”和“立即执行”两阶段。
