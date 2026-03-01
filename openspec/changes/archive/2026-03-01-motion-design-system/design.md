# Motion Design System - Technical Design

## Context

OpsPilot 前端使用 React 19 + Ant Design + Tailwind CSS，当前动效实现分散在各组件中，缺少统一的设计系统和代码复用机制。

**现状分析:**
- `web/src/index.css` 中有零散的 `@keyframes` 定义（pulse-glow, fade-in）
- `web/src/components/AI/ai-assistant.css` 中有 AI 助手专属动效（ai-dot-wave, ai-cursor-blink, ai-chip-in）
- `web/tailwind.config.js` 中有 status-pulse 动画
- 无页面切换过渡，路由变化时组件瞬间替换
- 列表页无 stagger 动效，数据加载后立即显示

**约束:**
- 纯前端改动，不涉及后端 API
- 需与 Ant Design 组件风格协调
- 动效强度: 中等（类似 Linear/Vercel）
- 目标设备: 桌面端（1024px+）
- 性能要求: 使用 GPU 加速属性（opacity, transform）

## Goals / Non-Goals

**Goals:**
- 建立统一的动效设计 Token（时长、缓动曲线、距离）
- 实现页面切换过渡动画
- 实现列表交错进入动效
- 增强交互反馈（hover, active, focus）
- 增强操作结果反馈（成功/失败/高风险）
- 优化 AI 助手动效体验

**Non-Goals:**
- 不做移动端响应式适配（仅桌面端）
- 不重写现有 Ant Design 组件
- 不改变业务逻辑
- 不引入额外的 UI 组件库

## Decisions

### Decision 1: 动效技术栈选型

**选择: CSS Variables + Framer Motion 混合方案**

| 方案 | 优点 | 缺点 |
|------|------|------|
| 纯 CSS | 零依赖、性能最佳 | 复杂动效（stagger）需要 JS 配合 |
| Framer Motion | 声明式、强大能力、社区活跃 | ~25KB gzip |
| Ant Design 内置 | 最小改动 | 能力有限，难实现现代流畅感 |

**选择 Framer Motion 的理由:**
- 页面切换需要 `AnimatePresence` 处理退出动画
- stagger 动效声明式实现更简洁
- spring 物理动画效果更自然
- 与 React 19 兼容良好

**混合策略:**
- 基础交互反馈（hover/active）: 纯 CSS
- 页面切换/列表交错: Framer Motion
- AI 助手复杂动效: Framer Motion

### Decision 2: 动效 Token 设计

**选择: CSS Custom Properties 集中管理**

```css
:root {
  /* Duration */
  --motion-duration-instant: 100ms;
  --motion-duration-fast: 150ms;
  --motion-duration-normal: 200ms;
  --motion-duration-slow: 300ms;

  /* Easing */
  --motion-ease-out: cubic-bezier(0.16, 1, 0.3, 1);
  --motion-ease-in-out: cubic-bezier(0.65, 0, 0.35, 1);
  --motion-ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1);

  /* Distance */
  --motion-distance-xs: 4px;
  --motion-distance-sm: 8px;
  --motion-distance-md: 12px;
}
```

**替代方案:** Tailwind 配置扩展
- 缺点: 无法在 CSS animation 中引用 Tailwind 变量
- 选择 CSS Variables 可同时用于 Tailwind 类和原生 CSS

### Decision 3: 页面切换动画方案

**选择: 缩放淡入 (Scale + Fade)**

```
退出: opacity 1 → 0, scale 1 → 0.98, 200ms, ease-out
进入: opacity 0 → 1, scale 0.98 → 1, 200ms, ease-out
```

**替代方案:**
- 滑动切换 (y: 12px): 方向感强，但运动幅度大
- 纯淡入淡出: 简单，但缺少层次感

**选择理由:**
- 缩放微妙，传达"层级关系"
- 类似 Linear 风格，克制专业
- 性能好，只使用 opacity + scale

### Decision 4: 列表 Stagger 实现方案

**选择: 根据场景使用不同实现**

| 场景 | 实现方案 | 理由 |
|------|---------|------|
| 卡片网格 | Framer Motion StaggerList | 元素数量可控，组件化更简洁 |
| Ant Design Table | CSS 变量 + rowClassName | 不破坏 Table 组件封装 |

**Table Stagger CSS 方案:**
```tsx
// TSX
onRow={(record, index) => ({
  style: { '--stagger-delay': `${Math.min(index, 20) * 30}ms` }
})}
rowClassName={() => 'stagger-row'}
```

```css
/* CSS */
.stagger-row {
  animation: stagger-fade-in 150ms var(--motion-ease-out) forwards;
  animation-delay: var(--stagger-delay, 0ms);
  opacity: 0;
}
```

**限制:** 行数 > 20 时只动画前 20 行（`Math.min(index, 20)`）

### Decision 5: 文件组织结构

**选择: 集中式 Motion 模块**

```
web/src/
├── styles/
│   └── motion.css          # Token + 基础动画类
├── components/
│   └── Motion/
│       ├── index.ts         # 导出
│       ├── PageTransition.tsx
│       ├── StaggerList.tsx
│       └── OperationFeedback.tsx
```

**替代方案:** 分散到各业务组件
- 缺点: 代码重复，难以维护一致性

## Risks / Trade-offs

### Risk 1: Framer Motion 包体积

**风险:** framer-motion 约 25KB gzip，增加首屏加载时间

**缓解措施:**
- 使用动态导入: `const { motion } = await import('framer-motion')`
- 仅在需要复杂动效的页面使用
- 基础交互反馈使用纯 CSS，不依赖 Framer Motion

### Risk 2: Table Stagger 性能

**风险:** 大数据量表格（100+ 行）动画可能卡顿

**缓解措施:**
- 限制动画行数（最多 20 行）
- 使用 CSS animation（GPU 加速）
- 添加 `will-change: transform` 优化

### Risk 3: 动画过度影响工作效率

**风险:** 运维平台用户追求效率，过多动画可能影响操作速度

**缓解措施:**
- 时长控制在 150-300ms
- 高频操作按钮无动画延迟
- 尊重 `prefers-reduced-motion` 媒体查询

### Trade-off: 动效强度 vs 专业感

**选择:** 中等强度，类似 Linear
- 太弱: 感觉生硬，缺少现代感
- 太强: 影响效率，不符合运维平台定位

## Migration Plan

### Phase 1: 基础设施 (无风险)
1. 添加 `framer-motion` 依赖
2. 创建 `motion.css` Token 文件
3. 在 `index.css` 中引入
4. 创建 `Motion/` 组件目录

### Phase 2: 全局组件 (低风险)
1. 实现 `PageTransition` 组件
2. 修改 `App.tsx` 集成页面切换
3. 测试所有页面切换效果

### Phase 3: 列表页应用 (中风险)
1. 实现 `StaggerList` 组件
2. 修改 `Dashboard.tsx` 卡片网格
3. 修改 `HostListPage.tsx` 表格行
4. 修改 `ServiceListPage.tsx` 表格行
5. 验证性能

### Phase 4: 细节打磨 (低风险)
1. 添加按钮/卡片 hover 效果
2. 增强操作反馈动效
3. 优化 AI 助手动效

### 回滚策略
- 所有改动都是增量添加，不破坏现有功能
- 如需回滚，删除 `Motion/` 目录和 `motion.css`，移除依赖即可
- 页面切换可通过移除 `PageTransition` 包裹立即禁用

## Open Questions

1. **骨架屏动效是否纳入此变更?**
   - 当前未包含，可作为后续迭代
   - 需要与 Suspense/Loading 状态统一设计

2. **是否需要动效预览/调试工具?**
   - Storybook 集成?
   - 开发环境动效慢放?

3. **暗色模式适配?**
   - 当前动效 Token 不涉及颜色
   - 部分效果（如阴影）可能需要暗色模式调整
