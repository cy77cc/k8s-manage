## Why

`redesign-deployment-management-blueprint` 已经完成蓝图与规范统一，但当前仓库中的后端 API、审批流、审计时间线与前端交互尚未完整落地这些约束。需要一个实现型 change 将蓝图从文档状态推进到可交付能力，避免规范与系统行为长期偏离。

## What Changes

- 在后端 `deployment`/`cicd`/`ai` 域实现统一发布生命周期状态与标准响应字段。
- 落地“先 preview 再 apply”的强约束，并在生产环境执行审批门禁。
- 补齐发布审计事件与时间线查询能力，统一 UI 与 AI 命令入口的追踪链路。
- 升级前端部署页与命令中心入口，统一展示 lifecycle、approval、diagnostics、timeline。
- 增加回归验证（API、审批路径、UI 状态）并同步 roadmap/progress 文档。

## Capabilities

### New Capabilities
- `deployment-blueprint-implementation`: 将部署管理蓝图要求落地为可运行的后端接口、治理流程、前端交互与验证清单。

### Modified Capabilities
- 无（本 change 以新增实现能力编排为主，不直接改写既有 capability 文本）。

## Impact

- Backend: `internal/service/deployment/*`, `internal/service/cicd/*`, `internal/service/ai/*`, `api/*/v1`。
- Frontend: `web/src/pages/Deployment/*`, `web/src/pages/AI/*`, `web/src/api/modules/deployment.ts`。
- Persistence: `storage/migrations/*`（若新增 release/approval/audit 字段或索引）。
- Docs and governance: `docs/roadmap.md`, `docs/progress.md`, OpenSpec artifacts for this change.
