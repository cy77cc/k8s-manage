## Context

当前监控模块已注册路由，但 roadmap 目标中的指标治理与告警联动仍需规范化拆解。

## Goals / Non-Goals

**Goals:**
- 定义指标查询、告警规则生命周期、通知投递最小闭环。
- 定义可审计告警事件查询与状态跟踪。

**Non-Goals:**
- 不在本 change 中绑定特定监控后端实现细节。

## Decisions

- 先定义通用 alert rule 模型与状态机，再映射到具体后端。
- 通知渠道采用可扩展 provider 接口，不在初期绑定单厂商。

## Risks / Trade-offs

- [Risk] 不同监控后端语义不一致。
  - Mitigation: 在规范层保留统一抽象字段并定义适配层。
