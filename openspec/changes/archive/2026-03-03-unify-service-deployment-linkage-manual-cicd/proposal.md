## Why

当前服务发布存在三条并行链路（`/services/:id/deploy`、`/deploy/releases/*`、`/cicd/releases`），导致手动部署与 CI/CD 无法共享同一发布状态机、审批门禁和审计视图。现在需要统一“服务到部署”的联动模型，避免状态分叉和运维认知成本持续扩大。

## What Changes

- 引入统一的 Release 编排入口，手动部署与 CI/CD 触发都进入同一发布流程。
- 统一发布状态流为 `previewed -> pending_approval/approved -> applying -> applied/failed -> rollback`，并要求两类触发一致执行。
- 为发布记录补充触发来源（manual/ci）、CI 运行上下文、版本元数据，形成端到端追踪。
- 规定服务模块的“手动部署”改为发布请求入口，不再维护独立执行状态机。
- 保留旧接口兼容窗口，将旧链路映射到统一编排路径并产出兼容审计事件。

## Capabilities

### New Capabilities
- `service-deployment-linkage`: 定义服务与部署的统一联动契约，包括触发入口、统一状态机、来源追踪与兼容映射规则。

### Modified Capabilities
- `service-ci-management`: 增加 CI 触发到统一 Release 的联动要求与上下文传递要求。
- `deployment-cd-management`: 调整为统一编排前提下的 CD 策略、审批门禁和执行约束。
- `deployment-release-management`: 将发布生命周期与回滚行为收敛到统一入口并约束状态一致性。
- `service-configuration-management`: 明确服务侧手动部署入口变为“创建发布请求”而非独立部署流程。

## Impact

- Backend
  - `internal/service/service/*`: 手动部署相关 handler/logic 需要改为调用统一发布编排。
  - `internal/service/deployment/*`: 作为统一发布编排核心，承接手动与 CI/CD 两类触发。
  - `internal/service/cicd/*`: CI 触发与 release 触发改为统一 release 编排路径。
  - `internal/model/*` 与 `storage/migrations/*`: 补充或对齐 release 元数据字段与关联关系。
- API
  - `/api/v1/services/*`、`/api/v1/deploy/*`、`/api/v1/cicd/*` 的发布相关接口语义将统一，并定义兼容策略。
- Frontend
  - `web/src/api/modules/{deployment,services,cicd}.ts` 与部署/服务页面需要对齐统一发布状态与来源展示。
- Governance/Audit
  - 统一审批、审计和时间线查询口径，减少跨表/跨模块拼接。
