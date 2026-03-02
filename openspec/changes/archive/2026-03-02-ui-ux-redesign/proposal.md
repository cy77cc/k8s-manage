# Proposal: UI/UX 重设计

**Status**: Draft
**Created**: 2026-03-02
**Author**: Design Team

---

## Why

当前 OpsPilot PaaS 平台的用户界面存在多个痛点，严重影响用户体验和工作效率：

### 视觉设计问题
- 深色渐变侧边栏过于厚重，缺乏现代感
- 蓝色主题 (#3a7afe) 不够优雅
- 组件样式陈旧，使用标准 Ant Design 未定制
- 缺少流畅的动画和微交互

### 信息架构问题
- Dashboard 信息过载，缺乏重点
- 表格行高过小（默认），难以快速扫视
- 卡片内容拥挤，缺乏呼吸空间
- 视觉层次不清晰，主次不分

### 交互体验问题
- 常用操作需要多次点击，缺少快捷方式
- 加载状态不明显，操作反馈延迟
- 没有命令面板和键盘快捷键
- 缺少批量操作和拖拽功能

### 性能问题
- 首屏加载时间过长
- 页面切换有明显延迟
- 长列表渲染卡顿（无虚拟滚动）
- 打包体积过大，无代码分割

### 响应式问题
- 移动端布局不适配，操作按钮过小
- 平板体验不佳，空间利用率低

这些问题导致：
- 用户学习成本高
- 操作效率低
- 用户满意度低
- 与现代 PaaS 平台（Vercel、Railway、Render）相比缺乏竞争力

---

## What Changes

### 1. 建立现代化设计系统

**设计风格**: 极简主义（Minimalism）
- 灵感来源: Vercel, Linear, Stripe
- 大量留白，降低视觉噪音
- 细线条，微妙阴影
- 优雅动画，流畅交互

**色彩系统**:
- 主色: Indigo 500 (#6366f1) 替代当前的 #3a7afe
- 中性色: Gray 系列 (#fafbfc ~ #212529)
- 语义色: Success (#10b981), Warning (#f59e0b), Error (#ef4444)

**排版系统**:
- 字体: 系统字体栈（-apple-system, PingFang SC 等）
- 字号: 12px ~ 32px（8 个层级）
- 行高: 1.25 (tight) ~ 1.75 (relaxed)

**间距系统**:
- 基于 8px 基准: 4px, 8px, 16px, 24px, 32px, 48px, 64px

**其他系统**:
- 圆角: 4px ~ 16px
- 阴影: 5 个层级（sm ~ 2xl）
- 动画: 150ms ~ 500ms，4 种缓动函数

### 2. 重构组件库

基于 Ant Design 6.3.0 深度定制：

**核心组件**:
- Button: 更大圆角（8px）、更好的 hover 效果
- Input: 高度 40px、更明显的 focus 状态
- Card: 圆角 12px、微妙阴影、hover 效果
- Table: 行高 56px、更清晰的列分隔、内联操作
- Modal: 圆角 12px、背景模糊效果
- Form: 更大的 label、更清晰的验证反馈
- Tag: 重新设计色彩方案
- Notification: 左侧色条、更好的动画

### 3. 重设计核心页面

**布局**:
- 侧边栏: 240px 宽，白色背景，浅色主题
- 顶部导航: 64px 高，面包屑 + 搜索 + 快捷操作
- 内容区: padding 32px，更多留白
- 移动端: 底部导航栏

**Dashboard**:
- 4 个统计卡片（服务总数、运行中、部署次数、告警数）
- 服务健康状态列表（实时指标）
- 最近部署和活跃告警（快速访问）
- 资源使用趋势图表（24小时）

**服务列表**:
- 卡片式布局（替代表格）
- 内联操作按钮（hover 显示）
- 快速筛选和搜索
- 批量操作

**服务详情**:
- 清晰的页面头部（状态、版本、快捷操作）
- 6 个实时指标卡片（CPU、内存、网络、请求数、错误率、响应时间）
- 实例列表和部署历史
- Tab 导航（概览、配置、日志、监控）

**部署流程**:
- 步骤式部署界面（4 步）
- 清晰的进度展示
- 实时日志输出
- 一键回滚

### 4. 添加高级交互

**命令面板**:
- 快捷键: Cmd+K / Ctrl+K
- 导航命令（跳转到任意页面）
- 搜索命令（搜索服务、主机等）
- 操作命令（创建服务、部署等）

**键盘快捷键**:
- `j/k`: 列表上下移动
- `/`: 聚焦搜索
- `g+h`: 回到首页
- `g+s`: 服务列表
- `g+d`: 部署管理
- `Esc`: 关闭模态框/面板

**动画**:
- 页面切换动画（淡入淡出）
- 组件动画（卡片 hover、按钮点击）
- 微交互（表单验证、操作反馈）

### 5. 性能优化

**代码优化**:
- 虚拟滚动（react-window）用于长列表
- 路由级别代码分割
- 组件懒加载
- 图片懒加载和压缩

**数据优化**:
- 数据预加载
- 乐观更新
- 请求缓存和去抖
- 骨架屏替代 loading spinner

**打包优化**:
- Tree Shaking
- 代码压缩
- 依赖优化

### 6. 响应式适配

**断点**:
- xs: 0px (手机竖屏)
- sm: 640px (手机横屏)
- md: 768px (平板)
- lg: 1024px (笔记本)
- xl: 1280px (桌面)

**适配策略**:
- 桌面: 完整侧边栏 + 多列布局
- 平板: 可折叠侧边栏 + 两列布局
- 手机: 底部导航 + 单列布局

---

## Capabilities

本次重设计涉及以下能力：

- `ui-design-system` - 设计系统（色彩、排版、间距等）
- `ui-component-library` - 组件库（Button、Input、Card 等）
- `ui-layout` - 布局系统（侧边栏、顶部导航、响应式）
- `ui-dashboard` - Dashboard 页面
- `ui-service-management` - 服务管理页面（列表、详情）
- `ui-deployment` - 部署管理页面
- `ui-command-palette` - 命令面板
- `ui-keyboard-shortcuts` - 键盘快捷键
- `ui-animations` - 动画系统
- `ui-performance` - 性能优化

---

## Impact

### 受影响的文件和模块

**新增文件**:
- `web/src/design-system/tokens.ts` - 设计 Token
- `web/src/theme/antd-theme.ts` - Ant Design 主题配置
- `web/src/components/CommandPalette/` - 命令面板组件
- `web/src/hooks/useKeyboardShortcuts.ts` - 键盘快捷键 Hook
- `web/src/styles/animations.css` - 动画样式

**修改文件**:
- `web/tailwind.config.js` - Tailwind 配置
- `web/src/styles/global.css` - 全局样式
- `web/src/main.tsx` - 应用入口（添加命令面板）
- `web/src/pages/Dashboard/Dashboard.tsx` - Dashboard 重设计
- `web/src/pages/Services/*.tsx` - 服务管理页面重设计
- `web/src/pages/Deployment/*.tsx` - 部署管理页面重设计
- `web/src/components/Layout/AppLayout.tsx` - 布局重构
- 所有使用 Ant Design 组件的页面（样式调整）

**新增依赖**:
- `cmdk`: ^1.0.0 (命令面板)
- `react-window`: ^1.8.10 (虚拟滚动)

**配置文件**:
- `web/package.json` - 添加新依赖
- `web/vite.config.ts` - 打包优化配置

### 影响范围

**前端**:
- 所有页面的视觉样式
- 所有组件的交互行为
- 整体布局结构
- 性能和加载速度

**后端**:
- 无影响（纯前端变更）

**用户**:
- 需要适应新的界面和交互方式
- 学习新的快捷键
- 体验更快的加载速度和更流畅的交互

### 风险

**技术风险**:
- Ant Design 深度定制可能导致升级困难
  - 缓解: 使用主题系统，避免直接修改组件源码
- 性能优化可能引入新的 bug
  - 缓解: 充分测试，渐进式优化

**项目风险**:
- 时间估算不准确，项目延期
  - 缓解: 预留 buffer 时间，定期评估进度
- 设计和开发不一致
  - 缓解: 建立设计系统，定期 review

**用户风险**:
- 用户不适应新界面
  - 缓解: 提供用户指南，收集反馈，快速迭代

---

## Timeline

### 阶段 1: 设计系统建立 (2周)
- Week 1: 设计 Token 定义
- Week 2: Tailwind 配置和 Ant Design 主题

### 阶段 2: 组件库重构 (3周)
- Week 1: 基础组件（Button、Input、Card、Tag、Modal）
- Week 2: 数据展示组件（Table、List、Descriptions、Empty、Skeleton）
- Week 3: 反馈组件和布局组件（Notification、Message、Progress、Spin、Layout）

### 阶段 3: 核心页面重设计 (4周)
- Week 1: 布局和 Dashboard
- Week 2: 服务管理（列表、详情）
- Week 3: 部署管理
- Week 4: 其他核心页面（主机、监控、配置）

### 阶段 4: 交互优化 (2周)
- Week 1: 命令面板和键盘快捷键
- Week 2: 动画和微交互

### 阶段 5: 性能优化 (2周)
- Week 1: 代码优化（虚拟滚动、代码分割、资源优化）
- Week 2: 数据优化（预加载、乐观更新、骨架屏）

### 阶段 6: 测试和修复 (2周)
- Week 1: 功能测试（浏览器兼容性、响应式、功能、性能）
- Week 2: Bug 修复和优化

**总计**: 15周（约 3.5 个月）

---

## Success Metrics

### 用户体验指标
- 用户满意度提升 30%
- 任务完成时间减少 20%
- 错误率降低 50%

### 性能指标
- 首屏加载时间 < 2秒
- 页面切换时间 < 300ms
- Lighthouse 性能评分 > 90
- 打包体积减小 30%

### 代码质量指标
- 代码覆盖率 > 80%
- 无严重 bug
- 技术债务减少 50%

---

## References

- [UI_UX_REDESIGN_PROPOSAL.md](../../UI_UX_REDESIGN_PROPOSAL.md) - 完整设计提案
- [design-system-spec.md](../../design-system-spec.md) - 设计系统规范
- [component-library-spec.md](../../component-library-spec.md) - 组件库规范
- [page-redesign-spec.md](../../page-redesign-spec.md) - 页面设计规范
- [implementation-plan.md](../../implementation-plan.md) - 实施计划

