# Motion Design System - Implementation Tasks

## 1. 基础设施搭建

- [x] 1.1 添加 framer-motion 依赖到 web/package.json
- [x] 1.2 创建 web/src/styles/motion.css 动效 Token 文件
- [x] 1.3 在 web/src/index.css 中引入 motion.css
- [x] 1.4 创建 web/src/components/Motion/ 目录结构
- [x] 1.5 创建 web/src/components/Motion/index.ts 导出文件

## 2. 动效 Token 实现

- [x] 2.1 定义 duration Token（instant/fast/normal/slow）
- [x] 2.2 定义 easing Token（ease-out/ease-in-out/ease-spring）
- [x] 2.3 定义 distance Token（xs/sm/md）
- [x] 2.4 添加 prefers-reduced-motion 媒体查询支持

## 3. 页面切换过渡

- [x] 3.1 创建 PageTransition.tsx 组件
- [x] 3.2 实现 AnimatePresence + motion.div 包裹
- [x] 3.3 定义进入动画（opacity + scale 0.98→1, 200ms）
- [x] 3.4 定义退出动画（opacity + scale 1→0.98, 200ms）
- [x] 3.5 使用 location.pathname 作为 key 触发动画
- [x] 3.6 添加滚动重置逻辑
- [x] 3.7 在 App.tsx 中集成 PageTransition 包裹 Routes

## 4. 列表交错动效组件

- [x] 4.1 创建 StaggerList.tsx 组件
- [x] 4.2 实现 staggerChildren 动画配置（delay 50ms）
- [x] 4.3 定义子项动画（opacity 0→1, y 8px→0, 200ms）
- [x] 4.4 添加 reduced-motion 降级处理
- [x] 4.5 创建表格行 stagger CSS 类（.stagger-row）
- [x] 4.6 定义 stagger-fade-in keyframes（x -8px→0）

## 5. 交互反馈动效

- [x] 5.1 创建 .motion-hover-scale CSS 类（按钮 hover）
- [x] 5.2 创建 .motion-hover-lift CSS 类（卡片 hover）
- [x] 5.3 添加按钮 active 缩放效果（scale 0.98, 100ms）
- [x] 5.4 添加表格行 hover 背景过渡（150ms）
- [x] 5.5 添加输入框 focus 边框/阴影过渡（150ms）
- [x] 5.6 添加 Tag 状态变化颜色过渡（200ms）

## 6. 操作反馈动效

- [x] 6.1 创建 .motion-success-flash CSS 类
- [x] 6.2 创建 .motion-error-shake CSS 类
- [x] 6.3 创建 .motion-high-risk-pulse CSS 类
- [ ] 6.4 创建 OperationFeedback.tsx 组件（可选）

## 7. 页面集成 - Dashboard

- [x] 7.1 在 Dashboard.tsx 中使用 StaggerList 包裹卡片网格
- [x] 7.2 为统计卡片添加交错进入动画
- [x] 7.3 为 widget 卡片添加交错进入动画

## 8. 页面集成 - 主机列表

- [x] 8.1 在 HostListPage.tsx 表格添加 stagger-row 类
- [x] 8.2 使用 onRow 传递 --stagger-delay CSS 变量
- [x] 8.3 限制最多 20 行动画（Math.min(index, 20)）

## 9. 页面集成 - 服务列表

- [x] 9.1 在 ServiceListPage.tsx 表格添加 stagger-row 类
- [x] 9.2 使用 onRow 传递 --stagger-delay CSS 变量

## 10. AI 助手动效增强

- [x] 10.1 增强首包等待状态指示（thinking/tool_calling/approval_pending）
- [x] 10.2 优化工具调用轨迹时间线样式
- [x] 10.3 添加工具状态图标动画（running/success/error/approval）
- [x] 10.4 增强思考过程展开动画
- [x] 10.5 优化建议卡片交错动画节奏
- [x] 10.6 增强高风险操作确认动画（CommandPanel.tsx）

## 11. 测试与验证

- [x] 11.1 验证页面切换动画在所有路由正常工作
- [x] 11.2 验证列表 stagger 动画性能（大数据量场景）
- [x] 11.3 验证 prefers-reduced-motion 正确响应
- [x] 11.4 验证动效不影响业务功能
- [x] 11.5 检查 framer-motion 包体积影响
