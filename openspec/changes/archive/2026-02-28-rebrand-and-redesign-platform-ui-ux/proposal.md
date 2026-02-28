## Why

当前平台在品牌识别与视觉一致性上较弱，名称辨识度不高、界面风格分散，影响用户第一印象与日常使用效率。现在进行统一的品牌与体验升级，可以同时提升专业形象、可用性与操作流畅度。

## What Changes

- 为平台定义新的产品名称与品牌叙事，建立统一命名规范。
- 设计并落地新 Logo（主标、简化标、单色版）及基础品牌应用规则。
- 重构系统整体 UI 视觉语言：色彩、字体、间距、组件样式、状态反馈。
- 重构核心交互流程与信息架构：导航、页面布局、关键任务路径与反馈机制。
- 建立可复用的设计基线（Design Tokens + 组件规范），确保后续页面一致演进。

## Capabilities

### New Capabilities
- `platform-brand-identity-system`: 定义平台新名称、Logo 方案与品牌基础规范，并规定在系统内的展示和替换策略。
- `platform-ui-ux-redesign-foundation`: 定义全局视觉与交互重设计标准，包括设计令牌、核心布局模式和关键任务流程体验要求。

### Modified Capabilities
- `role-aware-navigation-visibility`: 在保留角色可见性控制前提下，调整导航结构与信息层级，以适配新的全局 IA 与交互框架。
- `role-permission-management-ux`: 调整权限管理相关界面的交互与可读性要求，使其符合新的 UI/UX 设计基线。

## Impact

- Frontend: `web/src` 下的布局、导航、页面容器、主题样式、共享组件与交互逻辑将大范围调整。
- API/Backend: 预计不引入新的核心业务 API；如需品牌配置动态化，可能新增只读配置接口（`/api/v1`）。
- Assets: 新增/替换品牌资源（Logo SVG/PNG、应用图标、登录页品牌素材）。
- Documentation/OpenSpec: 新增 2 个 capability spec，并更新 2 个现有 capability 的需求增量。
