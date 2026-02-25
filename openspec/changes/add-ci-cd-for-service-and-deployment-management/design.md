## Context

k8s-manage 当前在服务管理与部署管理之间缺少统一的 CI/CD 编排层，导致构建触发、环境发布、审批和回滚由多个人工步骤拼接完成。现有平台已具备 Go + Gin + GORM 后端、React 前端、Casbin 权限体系及 Kubernetes 运行基础，适合在现有 `/api/v1` 体系中增加 CI/CD 领域能力并与服务管理、部署管理联动。

约束包括：
- 受保护操作必须接入 JWT + Casbin。
- 涉及发布变更的动作需要可审计、可回滚。
- 需要保持对既有服务与部署数据模型的兼容，避免一次性重构。

## Goals / Non-Goals

**Goals:**
- 在服务管理中引入可配置的 CI 流水线定义与触发策略。
- 在部署管理中引入可配置的 CD 发布策略、审批门禁和回滚策略。
- 建立发布全链路审计，关联服务、构建、部署、审批与操作者。
- 提供面向前端的统一查询接口以展示流水线与发布状态。

**Non-Goals:**
- 不在本次变更中实现新的 CI 执行引擎，仅定义与集成执行接口。
- 不替换现有 Kubernetes 部署机制，仅在其上增加编排与治理能力。
- 不覆盖多租户跨集群联邦调度场景。

## Decisions

### Decision 1: 采用独立 CI/CD 领域模块并与服务/部署模块通过引用关联
- Choice: 在 `internal/service/cicd` 下建立 routes/handler/logic/repo（或等价分层），通过 service_id、deployment_id 建立关联。
- Rationale: 避免将 CI/CD 逻辑耦合进既有 service/deployment 代码，降低改动面并便于后续扩展。
- Alternative considered:
  - 将 CI 逻辑放入 service 模块、CD 逻辑放入 deployment 模块。缺点是职责分散、审批与审计难以统一。

### Decision 2: 流水线与发布配置持久化到关系数据库，运行状态缓存到 Redis
- Choice: 配置与审计主数据落库（GORM + migration），运行中状态与短期查询加速使用 Redis。
- Rationale: 落库满足审计和追溯，Redis 满足列表与看板实时性。
- Alternative considered:
  - 仅落库不缓存。缺点是高频状态查询性能不足。
  - 仅缓存不落库。缺点是不满足审计与合规要求。

### Decision 3: 所有变更类 API 强制 RBAC，发布执行引入审批状态机
- Choice: CI/CD 配置变更、发布触发、回滚等接口全部接入 Casbin；CD 发布引入 `pending_approval -> approved/rejected -> executing -> succeeded/failed/rolled_back` 状态机。
- Rationale: 将安全治理与流程治理前置，减少越权发布和不可追踪操作。
- Alternative considered:
  - 仅在 UI 层控制审批。缺点是可被绕过，后端无强约束。

### Decision 4: 通过统一 API 聚合返回服务、构建、部署、审批链路
- Choice: 提供聚合查询接口（例如发布详情和时间线），由后端拼装多表数据返回前端。
- Rationale: 减少前端多次请求与拼接复杂度，统一口径。
- Alternative considered:
  - 前端自行拼装多接口数据。缺点是一致性与性能不可控。

## Risks / Trade-offs

- [Risk] 外部构建执行器或代码仓库回调不稳定导致流水线状态延迟。 → Mitigation: 增加重试、超时与死信补偿任务，并在 UI 明确“状态同步中”。
- [Risk] 审批流过严影响发布效率。 → Mitigation: 支持按环境配置审批策略（例如 dev 可免审批，prod 强制审批）。
- [Risk] 新增数据模型与既有部署记录关联不完整。 → Mitigation: 增加外键约束与发布前校验，灰度期间启用双写核对。
- [Risk] 聚合接口复杂度上升影响响应时间。 → Mitigation: 对关键列表引入 Redis 缓存与分页索引优化。

## Migration Plan

1. 新增数据库迁移：创建 CI 配置、CD 配置、发布记录、审批记录、审计事件等表（含索引与回滚脚本）。
2. 上线只读查询接口与前端展示，验证数据链路与权限模型。
3. 上线配置管理接口（CI/CD 配置 CRUD），默认不触发真实发布。
4. 灰度开启发布触发与审批流（先 dev/staging，后 prod）。
5. 全量启用回滚能力与审计看板。

Rollback strategy:
- 功能开关关闭发布触发入口，保留查询能力。
- 通过 migration Down 回退新增表结构（若未承载关键生产数据）；若已承载，保留表并禁用入口，执行数据归档方案。

## Open Questions

- 构建执行器优先对接哪一种（Jenkins/GitLab CI/GitHub Actions）以及统一抽象字段集？
- 金丝雀与蓝绿策略是否都需要首期支持，还是先支持滚动+蓝绿？
- 审批人来源是否直接复用现有组织/角色模型，是否需要值班轮转策略？
