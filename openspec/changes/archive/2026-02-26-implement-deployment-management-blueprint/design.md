## Context

部署管理蓝图已经在前序 change 中明确了统一生命周期、审批门禁、审计时间线与 AI 入口复用要求，但当前代码仍存在分域实现差异：`deployment` 与 `cicd` 接口字段和状态词汇不完全一致，前端页面与命令中心入口也未共享同一审批追踪链路。本设计面向一次可落地实现，覆盖后端接口、数据结构、前端交互与回归验证。

## Goals / Non-Goals

**Goals:**
- 在 `/api/v1` 下统一 release 生命周期状态和响应字段，覆盖 preview/apply/rollback/query。
- 强制 apply 依赖有效 preview，并在生产环境走审批票据闭环。
- 建立 release timeline + diagnostics 的统一查询契约并接入 RBAC。
- 将 Deployment 页面与 AI 命令入口收敛到同一审批与审计链路。
- 输出可执行的分步任务，支持按后端、前端、验证三阶段推进。

**Non-Goals:**
- 不引入新的外部部署引擎或替换现有 runtime 执行器。
- 不重做全站权限系统，仅复用并补充 Casbin 权限点。
- 不一次性重写所有历史发布数据，仅提供兼容映射与增量迁移。

## Decisions

1. 采用“统一 release 应用服务 + runtime 适配执行器”结构。
- Why: 保持生命周期语义统一，同时隔离 K8s/Compose 执行差异。
- Alternative: 按 runtime 继续分散实现，代价是状态和审计长期分叉。

2. 对 apply 增加 preview 令牌校验（上下文哈希 + TTL + 参数一致性）。
- Why: 直接保障“先预览再执行”，并可追踪审批依据。
- Alternative: 仅前端约束 preview，后端无法防止绕过调用。

3. 审批采用全局收件箱模型，票据绑定 release_id 与 command_source。
- Why: UI 与 AI 请求统一进入审批队列，便于审计和接管。
- Alternative: 各入口独立审批，造成漏审与双轨治理。

4. 时间线事件采用统一事件模型（state_changed/approval/diagnostics/rollback）。
- Why: 前端详情页和 AI 回答可复用同一数据源。
- Alternative: 维持多表或多格式日志，查询拼装复杂且易丢字段。

5. 采用“兼容读取 + 渐进迁移”而非一次性数据重写。
- Why: 降低上线风险，允许旧记录继续查询。
- Alternative: 全量回填，风险和回滚成本高。

## Risks / Trade-offs

- [Risk] 生命周期状态收敛会影响旧接口依赖方。
  - Mitigation: 在响应中保留兼容字段并发布映射文档，逐步切换调用方。
- [Risk] 强制 preview 可能增加发布时延。
  - Mitigation: 引入 preview 缓存和快速重预览路径，控制二次确认开销。
- [Risk] 审批收件箱集中后可能出现待办堆积。
  - Mitigation: 提供按项目/环境/风险过滤和超时自动失效。
- [Risk] 增加审计与时间线写入可能影响高频发布性能。
  - Mitigation: 事件写入异步化，查询端按 release_id 建索引。

## Migration Plan

1. 数据层先补充 release/approval/timeline 相关字段与索引（含 Up/Down migration）。
2. 后端实现统一 lifecycle DTO 与 preview 校验，再接审批和审计写入。
3. 前端升级 Deployment 页面与 AI 命令入口，切换到统一接口和状态展示。
4. 发布后开启兼容映射观测，校验旧记录查询与新流程审批路径。
5. 若出现严重回归，回滚到旧路由处理并禁用新流程开关。

## Open Questions

- preview TTL 默认值是否统一为 30 分钟，或按环境分层配置。
- 生产审批是否需要二级审批（业务 owner + SRE）作为可选策略。
- AI 命令入口的审批提示是否需要提供“批量同类操作”能力。
