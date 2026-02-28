## Why

现有运维控制台在长时间高频使用场景下，信息密度、视觉一致性与操作效率不足，难以支撑 SRE/运维工程师快速识别系统状态并处理告警。现在进行企业级 UI/UX 重设计，可显著提升可读性、响应速度感知与关键操作安全性。

## What Changes

- 重构平台整体视觉体系为现代科技感暗黑风（深蓝 + 冷灰低饱和，蓝色强调色，10px 圆角，轻阴影）。
- 重构主界面信息架构：左侧可折叠导航、顶部健康状态栏、主区 12 栅格监控优先布局。
- 新增实时刷新体验基线（自动刷新节奏、loading skeleton、操作反馈 toast）。
- 强化数据操作体验：表格统一搜索/筛选/排序，危险操作统一二次确认。
- 建立可扩展组件化规范（React + Tailwind），支持响应式布局与后续模块扩展。

## Capabilities

### New Capabilities
- `enterprise-devops-console-ui-redesign`: 定义企业级 DevOps 运维控制平台的新视觉语言、信息布局和高频操作体验标准。

### Modified Capabilities
- `platform-ui-ux-redesign-foundation`: 将全局 UI 基线更新为默认暗黑模式、低饱和科技风与 12 栅格监控优先布局规范。
- `role-aware-navigation-visibility`: 在任务导向导航重构后，继续保证基于角色/权限的入口可见性与受限操作隐藏。
- `role-permission-management-ux`: 治理页面在新设计体系下保持显式入口、反馈一致性与可访问性要求。

## Impact

- Frontend: `web/src/components/Layout`、`web/src/pages/*`、`web/src/styles`、`web/src/theme`、共享表格/反馈组件会有系统级改动。
- API/Backend: 无需新增核心业务 API；可能增强前端对现有监控/告警接口的轮询与状态聚合呈现。
- UX/Interaction: 影响导航、监控看板、服务列表、治理操作流、危险操作确认与全局反馈模式。
- OpenSpec: 新增 1 个 capability spec，并更新 3 个现有 capability 的 requirement delta。
