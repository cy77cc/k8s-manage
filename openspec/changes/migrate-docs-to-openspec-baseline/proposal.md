## Why

当前项目的能力说明、架构设计、产品流程、QA 计划和开发进度分散在 `docs/` 多层目录，缺少统一的规范化变更入口。后续转向 OpenSpec 开发时，需要先把既有文档与代码现状收敛到可追踪的 specs/tasks 体系，避免“文档与实现脱节”。

## What Changes

- 将 `docs/` 现有文档迁移为 OpenSpec 变更工件，建立统一的“文档域 -> capability spec”映射。
- 基于当前代码路由与关键模块（`internal/service/*`、`internal/ai/*`、`web/src/api/modules/*`、`storage/migrations/*`）同步能力基线到 OpenSpec specs。
- 在 OpenSpec 中定义进度同步机制：以代码事实为准，记录能力状态（Done/In Progress）与后续补齐项。
- 保留 `docs/` 作为历史参考输入，不在本次变更中删除旧文档。

## Capabilities

### New Capabilities
- `docs-to-openspec-mapping`: 定义 `docs/` 到 OpenSpec 工件的映射、覆盖范围与迁移完成标准。
- `platform-capability-baseline`: 定义平台核心域（Auth/Host/Cluster/Service/Deploy/RBAC/Monitoring）的当前能力基线与接口覆盖要求。
- `ai-control-plane-baseline`: 定义 AI 聊天、工具调用、审批与执行查询的控制面基线能力。
- `progress-sync-governance`: 定义“进度如何从代码同步到 OpenSpec”的规则、频率和验收约束。

### Modified Capabilities
- None.

## Impact

- Affected docs: `docs/` 下架构、产品、全栈、QA、AI、roadmap/progress 文档作为迁移输入。
- Affected spec artifacts: `openspec/changes/migrate-docs-to-openspec-baseline/*`。
- Affected systems for事实校验: 后端路由与 handlers、前端 API 模块、DB migration。
- No runtime API behavior change in this change; this is a documentation/spec governance baseline change.
