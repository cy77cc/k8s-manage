## Why

当前部署管理仅覆盖基础发布动作，Compose 与 K8s 两种运行时在配置模型、发布校验、回滚与可观测信息上不一致，导致运维操作复杂且高风险。需要建立统一但可区分运行时差异的部署管理能力，提升发布稳定性和可运维性。

## What Changes

- 新增 K8s 部署管理能力，覆盖目标绑定、发布参数、发布状态、回滚与运行时检查。
- 新增 Compose 部署管理能力，覆盖主机/节点编排、发布参数、发布状态、回滚与运行时检查。
- 新增统一部署管理视图与 API，支持按运行时聚合查询发布记录、执行结果与异常信息。
- 增强部署前校验与部署后验收流程，统一失败处理与重试策略。
- 修改现有 CD 需求：发布策略与审批流程需要明确区分 K8s/Compose 运行时，并保证状态机行为一致。

## Capabilities

### New Capabilities
- `k8s-runtime-deployment-management`: 管理 Kubernetes 运行时下的部署目标、发布执行、健康校验与回滚。
- `compose-runtime-deployment-management`: 管理 Compose 运行时下的主机目标、发布执行、健康校验与回滚。
- `unified-deployment-observability`: 提供跨运行时的统一发布记录、执行日志、状态与诊断信息查询能力。

### Modified Capabilities
- `deployment-cd-management`: 调整发布策略与审批门禁需求，使其在 K8s/Compose 两种运行时下均具备一致状态机和可追溯行为。

## Impact

- Backend:
  - `internal/service/deployment` 扩展运行时分支逻辑与统一发布状态模型。
  - 新增/扩展 `/api/v1/deploy/*` 与 `/api/v1/cicd/*` 中与运行时相关的接口字段。
  - 强化 Casbin 权限点，细分运行时相关发布与回滚操作权限。
- Frontend:
  - `web/src/api/modules/deployment.ts` 与 `web/src/api/modules/cicd.ts` 增加运行时区分字段与查询接口。
  - `web/src/pages/Deployment` 增加 Compose/K8s 差异化配置表单与统一发布观测面板。
- Data:
  - `storage/migrations` 增加或调整部署记录、执行日志、诊断信息与运行时维度字段（含 Up/Down）。
- Ops/Quality:
  - 需要补齐 Compose 与 K8s 两条发布链路的测试与回滚演练流程。
