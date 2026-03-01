# Motion Page Transition

页面切换过渡组件，基于 Framer Motion 实现页面进入/退出动画。

## ADDED Requirements

### Requirement: Page transition component

系统 SHALL 提供 `PageTransition` 组件，用于包裹页面内容实现切换动画。

组件 SHALL 使用 Framer Motion 的 `AnimatePresence` 和 `motion.div` 实现。

#### Scenario: Page enters with animation

- **WHEN** 用户导航到新页面
- **THEN** 页面以 opacity 0→1 和 scale 0.98→1 进入
- **AND** 动画时长为 200ms
- **AND** 使用 `--motion-ease-out` 缓动曲线

#### Scenario: Page exits with animation

- **WHEN** 用户离开当前页面
- **THEN** 页面以 opacity 1→0 和 scale 1→0.98 退出
- **AND** 动画时长为 200ms

#### Scenario: Exit animation completes before enter

- **WHEN** 页面切换发生
- **THEN** 旧页面退出动画完成后，新页面进入动画才开始
- **AND** 使用 `AnimatePresence` 的 `mode="wait"` 模式

---

### Requirement: Route-based key for transition

`PageTransition` 组件 SHALL 使用路由路径作为 `key`，确保每次路由变化触发新动画。

#### Scenario: Navigation triggers transition

- **WHEN** 用户从 `/hosts` 导航到 `/services`
- **THEN** key 从 `/hosts` 变为 `/services`
- **AND** 触发退出和进入动画

#### Scenario: Same route with different params

- **WHEN** 用户从 `/hosts/detail/1` 导航到 `/hosts/detail/2`
- **THEN** 视为不同页面，触发过渡动画

---

### Requirement: Scroll reset on page change

页面切换时 SHALL 重置滚动位置到顶部。

#### Scenario: Scroll position resets

- **WHEN** 用户从长页面（已滚动到中间位置）导航到新页面
- **THEN** 新页面从顶部开始显示

---

### Requirement: Integration with AppLayout

`PageTransition` 组件 SHALL 在 `App.tsx` 中包裹 `<Routes>` 内容。

#### Scenario: All pages use transition

- **WHEN** 用户在任意页面间导航
- **THEN** 都有过渡动画效果

---

### Requirement: Performance optimization

页面过渡 SHALL 只使用 GPU 加速属性（opacity, transform），避免触发重排。

#### Scenario: No layout thrashing

- **WHEN** 页面过渡动画执行
- **THEN** 不使用 width/height/margin/left/top 等触发重排的属性
