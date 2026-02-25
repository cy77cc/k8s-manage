## Context

当前部署管理虽已具备目标管理与基础发布流程，但 Compose 与 K8s 运行时在参数、执行链路、校验机制和可观测数据上不一致，导致发布流程不可复用、排障成本高。该变更需要在现有 `internal/service/deployment` 与 `internal/service/cicd` 基础上，建立统一部署域模型并保留运行时差异化执行。

## Goals / Non-Goals

**Goals:**
- 提供 K8s 与 Compose 两套可落地的部署管理能力，并共享统一的发布状态与审计模型。
- 在部署配置、发布执行、回滚、验收、诊断查询层面形成统一 API 约定。
- 明确运行时维度的 RBAC 授权与审批策略，避免越权发布。
- 统一前端部署视图，支持运行时筛选、聚合状态和故障诊断。

**Non-Goals:**
- 不实现新的底层编排器，仅复用已有 K8s 客户端与 Compose 执行路径。
- 不在本次变更中引入多集群联邦调度与复杂流量治理。
- 不覆盖非 Compose/K8s 的新运行时。

## Decisions

### Decision 1: 引入统一部署运行时抽象层
- Choice: 在部署逻辑层定义 `runtime` 抽象（k8s/compose），统一输入输出结构，底层分支执行。
- Rationale: 使发布、审批、回滚、审计可复用，减少分散实现。
- Alternative considered:
  - 为每个运行时独立服务。缺点是状态与审计模型重复，前端聚合复杂。

### Decision 2: 统一发布状态机并增加运行时上下文
- Choice: 保持统一状态机（pending_approval/approved/rejected/executing/succeeded/failed/rolled_back），每次状态变更附带 runtime 与目标上下文。
- Rationale: 便于跨运行时查询和告警；兼容既有 CD/审计能力。
- Alternative considered:
  - 运行时各自定义状态机。缺点是无法统一查询与审批规则。

### Decision 3: 将部署诊断数据结构化入库并暴露聚合查询
- Choice: 将执行日志摘要、校验结果、失败原因以结构化 JSON 持久化，并在 API 层提供按 service/target/runtime 维度查询。
- Rationale: 支撑故障定位与历史追踪，避免仅依赖临时日志。
- Alternative considered:
  - 仅保留文本日志。缺点是不可索引、不可聚合分析。

### Decision 4: 权限模型细化到运行时与动作
- Choice: 为 deploy/read, deploy/apply, deploy/rollback, deploy/approve 在运行时维度做授权检查（k8s/compose）。
- Rationale: 降低越权风险，满足最小权限原则。
- Alternative considered:
  - 仅保留 deploy:* 粗粒度权限。缺点是无法按运行时分责。

## Risks / Trade-offs

- [Risk] Compose 节点执行环境差异导致发布失败率上升。 → Mitigation: 发布前增加节点能力检查与预检失败快速返回。
- [Risk] 统一抽象后掩盖运行时特性。 → Mitigation: API 中保留 runtime-specific 扩展字段。
- [Risk] 结构化日志落库导致存储增长。 → Mitigation: 增加 retention 与归档策略，热数据分页查询。
- [Risk] 权限规则细化导致初期配置复杂。 → Mitigation: 提供默认角色权限模板并在管理界面提示缺失权限。

## Migration Plan

1. 新增 migration：扩展部署记录与诊断字段，补充运行时维度索引。
2. 后端先上线只读查询与兼容写入（双字段策略），确保旧数据可读。
3. 上线 K8s/Compose 新发布流程与回滚接口，启用统一状态机。
4. 上线前端统一部署视图与运行时筛选。
5. 完成回滚演练并切换默认入口。

Rollback strategy:
- 通过开关恢复到旧发布入口，保留新表/新字段以便追溯。
- 如需数据库回退，执行 migration Down 并降级 API 字段返回。

## Open Questions

- Compose 发布执行是否需要引入批次并发控制上限？
- K8s 部署后验收指标（探针/日志/指标）首期采用哪些必选项？
- 前端是否需要在一个页面中同时展示运行时差异配置，还是分步向导？
