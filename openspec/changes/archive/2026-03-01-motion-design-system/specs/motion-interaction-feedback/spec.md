# Motion Interaction Feedback

交互反馈动效，包括按钮、卡片、输入框等元素的微交互。

## ADDED Requirements

### Requirement: Button hover feedback

按钮 hover 时 SHALL 有视觉反馈。

反馈效果：
- `scale(1.02)` - 轻微放大
- 阴影增强
- 时长 150ms
- 缓动 `--motion-ease-out`

#### Scenario: Button hover animation

- **WHEN** 用户鼠标悬停在按钮上
- **THEN** 按钮轻微放大并增强阴影
- **AND** 动画流畅自然

#### Scenario: Button hover respects reduced motion

- **WHEN** 用户设置 `prefers-reduced-motion: reduce`
- **THEN** 按钮直接显示 hover 状态，无动画过渡

---

### Requirement: Button active feedback

按钮点击时 SHALL 有即时反馈。

反馈效果：
- `scale(0.98)` - 轻微缩小
- 时长 100ms

#### Scenario: Button active animation

- **WHEN** 用户按下按钮
- **THEN** 按钮轻微缩小
- **AND** 释放后恢复原状

---

### Requirement: Card hover lift

卡片 hover 时 SHALL 有抬起效果。

反馈效果：
- `translateY(-2px)` - 向上移动 2px
- 阴影增强
- 时长 200ms

#### Scenario: Card hover animation

- **WHEN** 用户鼠标悬停在卡片上
- **THEN** 卡片轻微抬起并增强阴影

---

### Requirement: Table row hover transition

表格行 hover 时 SHALL 有背景色过渡。

#### Scenario: Table row hover animation

- **WHEN** 用户鼠标悬停在表格行上
- **THEN** 背景色平滑过渡
- **AND** 时长 150ms

---

### Requirement: Input focus feedback

输入框 focus 时 SHALL 有边框和阴影变化。

反馈效果：
- 边框颜色过渡
- 外发光阴影 (box-shadow)
- 时长 150ms

#### Scenario: Input focus animation

- **WHEN** 用户聚焦输入框
- **THEN** 边框颜色和阴影平滑变化

---

### Requirement: Tag status transition

Tag 状态变化时 SHALL 有颜色过渡动画。

#### Scenario: Tag status changes

- **WHEN** Tag 从 `online` 变为 `offline`
- **THEN** 颜色平滑过渡，时长 200ms

---

### Requirement: CSS classes for interaction

系统 SHALL 提供以下 CSS 辅助类：
- `.motion-hover-scale` - 按钮 hover 缩放
- `.motion-hover-lift` - 卡片 hover 抬起

#### Scenario: Developer applies interaction class

- **WHEN** 开发者给元素添加 `.motion-hover-lift` 类
- **THEN** 元素自动获得 hover 抬起效果
