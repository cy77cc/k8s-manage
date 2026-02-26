## Why

当前基线规范已覆盖能力范围，但缺少按域细化的状态矩阵（Done/In Progress/Risk），难以快速定位推进优先级。

## What Changes

- 新增 per-domain 状态矩阵能力 change。
- 为核心域补齐状态维度、证据来源、风险与下一步动作字段。
- 为后续周期同步提供统一模板。

## Capabilities

### New Capabilities
- `platform-status-matrix`: 定义平台域状态矩阵的结构和同步要求。

### Modified Capabilities
- None.

## Impact

- Affected artifacts: baseline governance specs in `migrate-docs-to-openspec-baseline`.
- Evidence inputs: `internal/service/*/routes.go`, `docs/progress.md`, `docs/roadmap.md`.
