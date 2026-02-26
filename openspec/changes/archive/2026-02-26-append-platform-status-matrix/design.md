## Context

基线规范已经可用，但状态信息仍偏叙述化。团队需要标准矩阵视图来做迭代决策和回顾。

## Goals / Non-Goals

**Goals:**
- 定义统一状态矩阵字段与更新频率。
- 保持状态矩阵可被代码证据验证。

**Non-Goals:**
- 不替代各能力 spec 的行为需求定义。

## Decisions

- 采用固定三态：`Done | In Progress | Risk`，并强制证据字段。
- 状态矩阵以 change 附录形式维护，避免污染行为 spec 主体。

## Risks / Trade-offs

- [Risk] 状态矩阵维护额外成本。
  - Mitigation: 将更新动作纳入 PR checklist 与 tasks。
