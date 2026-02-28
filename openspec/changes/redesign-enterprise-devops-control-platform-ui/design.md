## Context

平台需要面向运维工程师与 SRE 的长时使用场景重构为企业级控制台体验：在暗黑界面下持续提供高可读监控信息、低认知负担的高频操作路径，以及强约束的安全操作反馈。本次变更是跨壳层布局、主题系统、监控面板、服务列表和治理交互的横切改造，主要落地在 `web/src/components/Layout`、`web/src/pages/Dashboard`、`web/src/pages/Monitor`、`web/src/pages/Services`、`web/src/styles` 与 `web/src/theme`。

约束：
- 保持现有路由与权限边界，不破坏 `/api/v1` 业务契约。
- 默认暗黑模式，但支持保留可回滚开关。
- 优先保证监控与告警处理路径效率，而非一次性重写所有页面。

## Goals / Non-Goals

**Goals:**
- 建立现代科技感企业级暗黑视觉体系（深蓝 + 冷灰，低饱和，蓝色强调，10px 圆角，轻阴影）。
- 建立左侧可折叠导航 + 顶部健康状态栏 + 12 栅格主区的统一框架。
- 强化监控优先布局：首屏优先展示系统状态图表与告警处理入口，下方提供可操作服务表格。
- 统一 UX 基线：实时刷新、Skeleton、toast、搜索筛选排序、危险操作二次确认、响应式。
- 形成可扩展组件化结构（React + Tailwind + Ant Design 主题整合）。

**Non-Goals:**
- 不引入新的后端领域模型，不替换核心 API。
- 不在本次完成全站每个页面的视觉重绘，优先覆盖壳层和核心运维页面。
- 不改变 RBAC 判定逻辑，只重构其在 UI 层的展示方式。

## Decisions

### Decision 1: 采用“暗黑 token 驱动 + AntD/Tailwind 双轨消费”
- Rationale: 通过同一组 design tokens 同时驱动 AntD 组件与 Tailwind 原子类，减少风格漂移。
- Alternatives considered:
  - 仅靠页面 CSS 覆盖：维护成本高且难统一。
  - 更换组件库：迁移成本高、收益低。

### Decision 2: 信息架构按“状态感知优先”组织
- Rationale: SRE 首要任务是快速发现异常并决策，导航与首页信息结构应围绕监控、告警、服务健康展开。
- Alternatives considered:
  - 按系统模块平铺菜单：学习成本更高，关键告警路径不突出。

### Decision 3: 交互策略统一为“显式入口 + 安全确认 + 即时反馈”
- Rationale: 降低高压场景误操作，保持操作可审计和可恢复。
- Alternatives considered:
  - 保留现有各页不同交互：一致性不足。

### Decision 4: 实时刷新采用前端定时轮询与可控节奏
- Rationale: 不引入新后端推送协议前提下，轮询更易落地并可按页面负载控制频率。
- Alternatives considered:
  - WebSocket 全量改造：超出本次范围。

## Risks / Trade-offs

- [Risk] 暗黑模式下对比度不足影响可读性 → Mitigation: 制定对比度阈值并加入视觉验收清单。
- [Risk] 高频轮询造成前端性能抖动 → Mitigation: 分级刷新频率与页面可见性暂停策略。
- [Risk] 导航重组导致用户短期找不到入口 → Mitigation: 保持旧路径映射与过渡提示。
- [Risk] 治理操作按钮增多造成视觉拥挤 → Mitigation: 主次操作分层和表格列优先级控制。

## Migration Plan

1. Phase 1：上线暗黑 token、壳层布局、顶部健康栏和导航分组。
2. Phase 2：上线监控图表优先主内容区和服务列表操作增强。
3. Phase 3：治理页面交互一致化、响应式细节和 Skeleton/toast 全面补齐。
4. Rollback：通过 UI 主题开关回切旧主题与旧壳层样式（不改动 API 与权限后端）。

## Open Questions

- 实时刷新默认周期是否按页面场景区分（如告警页 15s、列表页 30s）？
- 顶部健康度是否需要统一 SLI 口径（可用性、告警总数、变更风险）？
- 移动端优先级是否限定为“只读监控视图”还是支持完整操作？
