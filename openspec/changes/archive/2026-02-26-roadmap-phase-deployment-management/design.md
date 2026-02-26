## Context

当前已具备 deploy target/release/governance 的基础能力，但缺少跨环境策略、审批链与审计一致性规范。

## Goals / Non-Goals

**Goals:**
- 统一部署目标模型与发布流程约束。
- 明确审批、回滚、审计、查询链路。

**Non-Goals:**
- 本 change 不直接实现所有代码，仅定义后续实现边界。

## Decisions

- 以 release workflow 为主线，覆盖 preview/apply/rollback/list/get。
- production 变更保持审批门禁能力，且与 AI 建议/诊断可联动。

## Risks / Trade-offs

- [Risk] 多环境策略复杂度上升。
  - Mitigation: 先定义最小可执行策略集，再逐步扩展。
