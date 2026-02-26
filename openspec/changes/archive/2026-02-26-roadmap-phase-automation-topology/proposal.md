## Why

roadmap Phase 3 中的 Automation/Topology 能力尚无独立 OpenSpec change，后续实现缺少明确的规格入口和任务拆解。

## What Changes

- 建立 Automation/Topology 独立 change，定义作业编排、执行记录、资源拓扑关系与查询能力。
- 形成后续实现任务清单，避免无规范先开发。

## Capabilities

### New Capabilities
- `automation-topology-phase`: 定义自动化作业与资源拓扑建模的阶段能力。

### Modified Capabilities
- None.

## Impact

- Affected modules: future `internal/service/automation/*`, `internal/service/topology/*`, existing `web/src/api/modules/automation.ts`, `web/src/api/modules/topology.ts`.
- Affected docs: `docs/roadmap.md`, `docs/fullstack/refactor-task-breakdown.md`.
