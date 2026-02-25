## Context

自动化执行与拓扑洞察属于高阶运营能力，需要先规范作业模型、依赖关系、执行审计和图查询行为。

## Goals / Non-Goals

**Goals:**
- 定义 automation job 与 execution 的领域模型。
- 定义 topology 节点/边模型与查询接口行为。

**Non-Goals:**
- 不在本 change 中完成全量工作流引擎实现。

## Decisions

- 先以任务执行可追踪为核心，后续再扩展编排 DSL。
- 拓扑查询先覆盖关键资源关系，再演进可视化增强。

## Risks / Trade-offs

- [Risk] 领域边界与现有 service/deployment/aiops 交叉。
  - Mitigation: 通过 capability spec 明确 owner 与跨域接口。
