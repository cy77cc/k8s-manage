# Motion Stagger List

列表交错进入动效，支持卡片网格和表格行两种场景。

## ADDED Requirements

### Requirement: StaggerList component for card grid

系统 SHALL 提供 `StaggerList` 组件，用于卡片网格的交错进入动画。

组件 SHALL 使用 Framer Motion 的 `staggerChildren` 实现。

#### Scenario: Cards enter with stagger animation

- **WHEN** 卡片网格加载完成
- **THEN** 每张卡片依次进入，动画为 opacity 0→1 和 y: 8px→0
- **AND** 每张卡片动画时长为 200ms
- **AND** 卡片间延迟为 50ms

#### Scenario: Stagger respects reduced motion

- **WHEN** 用户设置 `prefers-reduced-motion: reduce`
- **THEN** 所有卡片同时显示，无交错延迟

---

### Requirement: Table row stagger via CSS

Ant Design Table 的行 SHALL 使用 CSS 方案实现 stagger 动画，不破坏组件封装。

实现方式：
- 使用 `rowClassName` 添加 `stagger-row` 类
- 使用 `onRow` 传递 `--stagger-delay` CSS 变量
- CSS 动画使用 `animation-delay: var(--stagger-delay)`

#### Scenario: Table rows enter with stagger animation

- **WHEN** 表格数据加载完成
- **THEN** 每行依次进入，动画为 opacity 0→1 和 x: -8px→0
- **AND** 每行动画时长为 150ms
- **AND** 行间延迟为 30ms

#### Scenario: Large table limits stagger count

- **WHEN** 表格行数超过 20 行
- **THEN** 只对前 20 行应用交错动画
- **AND** 后续行直接显示

#### Scenario: Table filter or pagination resets animation

- **WHEN** 用户切换分页或应用筛选
- **THEN** 新显示的行重新执行交错动画

---

### Requirement: Stagger animation CSS class

系统 SHALL 提供 `.stagger-row` CSS 类，用于表格行动画。

CSS 规范：
```css
.stagger-row {
  animation: stagger-fade-in 150ms var(--motion-ease-out) forwards;
  animation-delay: var(--stagger-delay, 0ms);
  opacity: 0;
}

@keyframes stagger-fade-in {
  from { opacity: 0; transform: translateX(-8px); }
  to { opacity: 1; transform: translateX(0); }
}
```

#### Scenario: CSS class is reusable

- **WHEN** 开发者给任意元素添加 `.stagger-row` 类和 `--stagger-delay` 变量
- **THEN** 元素执行交错进入动画

---

### Requirement: Dashboard card grid stagger

`Dashboard.tsx` 的卡片网格 SHALL 使用 `StaggerList` 组件包裹。

#### Scenario: Dashboard cards animate on load

- **WHEN** 用户打开 Dashboard 页面
- **THEN** 统计卡片和 widget 卡片依次进入

---

### Requirement: HostListPage table stagger

`HostListPage.tsx` 的表格 SHALL 使用 CSS stagger 方案。

#### Scenario: Host list rows animate on load

- **WHEN** 用户打开主机列表页面
- **THEN** 主机行依次进入
