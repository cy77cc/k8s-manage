# Motion Design System

## Why

OpsPilot 当前缺少统一的动效系统，交互反馈零散且缺乏一致性。页面切换瞬间完成、列表无交错进入、操作结果无视觉反馈，用户体验生硬。为提升运维平台的专业感和流畅度，需要建立一套现代化的动效设计系统。

动效系统目标：中等强度（类似 Linear/Vercel），用户能清晰感知动效但不夸张，传达专业、可靠的运维平台形象。

## What Changes

### 新增能力

- **动效设计 Token 系统**: 统一的时长、缓动曲线、距离等 CSS 变量
- **页面切换过渡组件**: 基于 Framer Motion 的 AnimatePresence 实现页面进入/退出动画
- **列表交错进入动效**: 卡片网格和表格行的 stagger 动画
- **交互反馈动效**: 按钮 hover/active、卡片悬浮、输入框 focus 等微交互
- **操作结果反馈**: 成功/失败/高风险操作的视觉反馈动画
- **AI 助手动效增强**: 工具调用时间线、状态指示、思考过程展开动画

### 技术改动

- 新增 `framer-motion` 依赖
- 新增 `web/src/styles/motion.css` 动效 Token 和基础类
- 新增 `web/src/components/Motion/` 动效组件目录
- 修改 `AppLayout.tsx` 集成页面切换过渡
- 修改各列表页集成 stagger 动效
- 增强 AI 助手组件的动效体验

## Capabilities

### New Capabilities

- `motion-tokens`: 动效设计 Token（时长、缓动曲线、距离等 CSS 变量）
- `motion-page-transition`: 页面切换过渡组件（PageTransition）
- `motion-stagger-list`: 列表交错进入动效（StaggerList 组件 + CSS 方案）
- `motion-interaction-feedback`: 交互反馈动效（按钮、卡片、输入框微交互）
- `motion-operation-feedback`: 操作结果反馈（成功/失败/高风险动画）
- `motion-ai-assistant`: AI 助手专属动效（工具调用时间线、状态指示）

### Modified Capabilities

无现有能力的需求变更。此变更为纯新增能力。

## Impact

### 受影响文件

**新增文件:**
- `web/src/styles/motion.css` - 动效 Token 和基础动画类
- `web/src/components/Motion/PageTransition.tsx` - 页面切换组件
- `web/src/components/Motion/StaggerList.tsx` - 列表交错组件
- `web/src/components/Motion/OperationFeedback.tsx` - 操作反馈组件
- `web/src/components/Motion/index.ts` - 导出

**修改文件:**
- `web/package.json` - 添加 framer-motion 依赖
- `web/src/index.css` - 引入 motion.css
- `web/src/App.tsx` - 集成 PageTransition
- `web/src/components/Layout/AppLayout.tsx` - 页面切换过渡
- `web/src/pages/Dashboard/Dashboard.tsx` - 卡片 stagger 动效
- `web/src/pages/Hosts/HostListPage.tsx` - 表格行动效 + 响应式
- `web/src/pages/Services/ServiceListPage.tsx` - 表格行动效
- `web/src/pages/Deployment/DeploymentPage.tsx` - 表单动效
- `web/src/components/AI/ChatInterface.tsx` - AI 助手动效增强
- `web/src/components/AI/CommandPanel.tsx` - 高风险操作反馈增强

### 依赖变更

- 新增 `framer-motion` (~25KB gzip)

### 风险评估

- **低风险**: 纯前端视觉增强，不影响业务逻辑
- **性能**: 动效使用 opacity/transform，GPU 加速，对性能影响小
- **兼容性**: Framer Motion 兼容 React 19，无兼容性问题
