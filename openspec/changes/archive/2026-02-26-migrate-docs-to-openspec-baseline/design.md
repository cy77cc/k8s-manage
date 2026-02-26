## Context

项目已形成较多高价值文档（架构、产品、QA、合同、进度），但这些文档按目录分散在 `docs/`，缺乏统一的“规范源”。

同时，代码侧已实现大量能力（`/api/v1` 多域路由、AI SSE + tool calling、迁移框架等），且 `docs/progress.md` 的迭代记录与代码需要周期性校准。若继续以非结构化文档推进，将导致以下问题：
- 需求变化难以追踪到规范与任务。
- 代码实际能力与文档状态容易漂移。
- 后续变更难以复用一致的验收模板。

## Goals / Non-Goals

**Goals:**
- 建立 `docs/` 到 OpenSpec capability 的稳定映射。
- 输出可执行的基线 specs，覆盖平台核心域与 AI 控制面。
- 在 OpenSpec 中固化“进度同步规则”，明确数据来源优先级：代码 > 进度文档 > 规划文档。
- 让本次 change 达到 `tasks` 可追踪状态，为后续 `/opsx:apply` 持续维护打基础。

**Non-Goals:**
- 不在本次变更中重构后端/前端业务代码。
- 不在本次变更中删除旧 `docs/` 文件。
- 不在本次变更中归档 change（由后续实现完成后执行 archive）。

## Decisions

### 1) 按“能力域”而非“原文档一比一”迁移
- Decision: 采用 4 个 capability（映射、平台基线、AI 基线、进度治理）聚合 `docs/` 内容。
- Why: 原文档数量多且跨域重叠，一比一迁移会造成 specs 颗粒度过碎、维护成本高。
- Alternative: 每个 `docs/*.md` 生成独立 spec；放弃，因会制造大量重复 requirement。

### 2) 以代码事实作为进度同步的主来源
- Decision: 进度状态同步必须先核对路由注册、handler 能力、前端 API 模块与迁移脚本，再写入 OpenSpec。
- Why: `docs/progress.md` 反映历史记录，可能滞后于代码；OpenSpec 需保持可验证。
- Alternative: 仅迁移 `docs/progress.md` 文本；放弃，因无法保证真实性。

### 3) 本次输出“基线规范 + 后续更新任务”，而非一次性终局文档
- Decision: 在 tasks 中标注完成项与后续增量维护项，支持持续同步。
- Why: 项目正在高频迭代，基线需要可持续更新流程。
- Alternative: 尝试一次性补齐所有未来需求；放弃，因不可维护。

## Risks / Trade-offs

- [Risk] 能力聚合后，某些细节从“文档级”下降到“能力级”。
  - Mitigation: 在 specs 中保留来源目录与关键约束，并在后续 change 中按需细分。
- [Risk] 当前代码包含进行中能力，状态判定可能受时间点影响。
  - Mitigation: 在进度规范中写明快照日期与证据来源（路由/模块/迁移文件）。
- [Risk] OpenSpec CLI 版本差异可能影响命令行为。
  - Mitigation: 工件按 `spec-driven` 基本结构直接落盘，并用 `openspec validate` 校验。

## Migration Plan

1. 建立 change：`migrate-docs-to-openspec-baseline`。
2. 生成 proposal，定义迁移范围与 capability 清单。
3. 读取 `docs/` 与代码关键路径，形成能力基线快照。
4. 生成 specs（映射、平台、AI、进度治理）。
5. 生成 tasks，标记已完成迁移步骤和后续维护步骤。
6. 执行 `openspec validate` 与 `openspec status`，确保 change 进入可实施状态。

Rollback strategy:
- 本次仅新增 OpenSpec 文档工件，不改业务代码。若需回滚，仅删除该 change 目录即可。

## Open Questions

- 是否在下一轮 change 中将 `docs/roadmap.md` 的 Phase 目标拆成独立能力 specs（例如 deployment-governance、monitoring-alerting）？
- 是否建立 CI 检查，要求变更 PR 同步更新对应 OpenSpec tasks/spec delta？
