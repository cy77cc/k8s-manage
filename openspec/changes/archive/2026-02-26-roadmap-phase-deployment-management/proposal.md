## Why

`docs/roadmap.md` 已将 Deployment Management 列为持续演进方向，但缺少独立 OpenSpec change 承载后续需求与实现任务，导致推进路径不清晰。

## What Changes

- 建立 Deployment Management 独立 change，沉淀目标能力与实施任务。
- 将 deploy target/release/governance/aiops 相关后续增强纳入统一规格。
- 为后续代码实现提供可跟踪的任务清单。

## Capabilities

### New Capabilities
- `deployment-management-phase`: 定义部署管理阶段能力（目标、发布、回滚、治理与审计）的增量要求。

### Modified Capabilities
- None.

## Impact

- Affected modules: `internal/service/deployment/*`, `internal/service/service/*`, `internal/service/aiops/*`, `web/src/api/modules/deployment.ts`.
- Affected docs: `docs/roadmap.md`, `docs/progress.md`, `docs/fullstack/service-studio-api-contract.md`.
