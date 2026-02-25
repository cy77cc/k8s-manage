## Why

当前平台缺少统一的 CI/CD 流程与发布治理能力，服务管理和部署管理仍需人工拼接流程，导致交付不稳定、回滚成本高、审计链路不完整。现在引入平台级 CI/CD 能力，可以将构建、发布、审批和回滚纳入同一条受控链路。

## What Changes

- 新增面向服务管理的 CI 流水线配置能力，支持仓库、分支策略、构建步骤、镜像产物与触发策略管理。
- 新增面向部署管理的 CD 发布配置能力，支持环境分层（如 dev/staging/prod）、发布策略（滚动/蓝绿/金丝雀）和版本晋级流程。
- 新增发布流程中的审批与审计要求，记录触发人、审批人、发布结果与回滚操作。
- 新增流水线与部署状态可视化接口，打通服务配置、构建结果与部署结果的关联查询。
- 对接现有 RBAC/Casbin 权限模型，确保服务与部署配置变更具备细粒度授权控制。

## Capabilities

### New Capabilities
- `service-ci-management`: 管理服务级 CI 配置、构建触发策略、构建步骤及产物元数据。
- `deployment-cd-management`: 管理部署级 CD 配置、环境发布策略、审批流程与回滚动作。
- `release-audit-tracking`: 提供发布全链路审计与状态追踪，关联服务、构建、部署与操作人信息。

### Modified Capabilities
- None.

## Impact

- Backend:
  - `internal/service/` 下新增 CI/CD 相关领域模块（routes/handler/logic）与 `/api/v1` 端点。
  - `api/<domain>/v1` 增加流水线、发布配置、审批与审计相关接口契约。
  - 接入 Casbin 权限点与操作审计记录逻辑。
- Frontend:
  - `web/src/api/modules` 增加 CI/CD 与发布管理 API 模块。
  - 新增服务管理与部署管理页面中的 CI/CD 配置、发布记录、审批记录视图。
- Data:
  - `storage/migrations` 增加 CI/CD 配置、发布记录、审批记录等表结构（含 Up/Down）。
- Integrations:
  - 对接 Git 仓库/构建执行器/镜像仓库及 Kubernetes 部署执行链路。
