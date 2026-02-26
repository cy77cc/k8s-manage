## Why

当前部署管理能力已具备基础发布与审批机制，但缺少统一的功能蓝图，导致目标模型、发布生命周期、治理与可观测能力在产品与研发侧缺乏一致边界，后续迭代容易出现接口与交互分裂。需要先完成蓝图重设计，形成跨后端、前端与 AI 助手命令入口的一致能力地图与落地路径。

## What Changes

- 重新定义部署管理功能蓝图，统一“目标建模、发布编排、审批治理、运行观测、回滚恢复、AI 命令代理”六大能力域。
- 明确 Compose 与 K8s 双运行时在同一发布域中的共性抽象与差异化扩展点。
- 规范发布生命周期阶段与状态词汇，统一 API 返回、前端状态展示与审计事件语义。
- 增补命令中心/AI 助手作为部署操作入口的能力边界、审批门禁与可追溯要求。
- 输出按阶段实施的里程碑和验收口径，作为后续实现 change 的基线。

## Capabilities

### New Capabilities
- `deployment-management-blueprint`: 部署管理整体功能蓝图，包括能力域划分、运行时抽象、生命周期、交互入口与实施阶段。

### Modified Capabilities
- `deployment-management-phase`: 将现有阶段能力从“任务集合”升级为“蓝图驱动”的能力分层与验收要求。
- `deployment-cd-management`: 对齐发布编排、审批策略和环境门禁语义到统一蓝图。
- `unified-deployment-observability`: 对齐发布时间线、诊断事件、审计追踪与跨运行时可观测约束。

## Impact

- Affected specs: `openspec/specs/deployment-management-phase/spec.md`, `openspec/specs/deployment-cd-management/spec.md`, `openspec/specs/unified-deployment-observability/spec.md`.
- Affected backend domains: `internal/service/deployment/*`, `internal/service/ai/*`, `internal/service/cicd/*`.
- Affected frontend domains: `web/src/pages/Deployment/*`, `web/src/pages/AI/*`, `web/src/api/modules/deployment.ts`.
- Affected docs: `docs/roadmap.md`, `docs/progress.md`, `docs/fullstack/service-studio-api-contract.md`.
