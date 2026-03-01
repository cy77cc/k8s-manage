# Motion Tokens

统一的动效设计 Token 系统，定义时长、缓动曲线、距离等基础变量。

## ADDED Requirements

### Requirement: Duration tokens

系统 SHALL 提供标准化的动画时长 Token，以 CSS Custom Properties 形式定义。

支持的时长 Token：
- `--motion-duration-instant: 100ms` - 即时反馈（button active）
- `--motion-duration-fast: 150ms` - 快速过渡（hover, focus）
- `--motion-duration-normal: 200ms` - 标准动画（展开、切换）
- `--motion-duration-slow: 300ms` - 页面级动画（进入、退出）

#### Scenario: Developer uses duration token in CSS

- **WHEN** 开发者在 CSS 中使用 `transition-duration: var(--motion-duration-fast)`
- **THEN** 动画时长为 150ms

#### Scenario: Duration tokens are consistent across components

- **WHEN** 多个组件使用相同的 duration token
- **THEN** 动画节奏保持一致

---

### Requirement: Easing tokens

系统 SHALL 提供标准化的缓动曲线 Token。

支持的缓动 Token：
- `--motion-ease-out: cubic-bezier(0.16, 1, 0.3, 1)` - 进入动画，开始快结束慢
- `--motion-ease-in-out: cubic-bezier(0.65, 0, 0.35, 1)` - 双向过渡，如展开/折叠
- `--motion-ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1)` - 弹跳效果，如成功反馈

#### Scenario: Developer uses easing token for enter animation

- **WHEN** 开发者创建进入动画
- **THEN** 使用 `--motion-ease-out` 实现自然的减速效果

#### Scenario: Spring easing for success feedback

- **WHEN** 显示操作成功反馈
- **THEN** 使用 `--motion-ease-spring` 实现轻微回弹效果

---

### Requirement: Distance tokens

系统 SHALL 提供标准化的移动距离 Token。

支持的距离 Token：
- `--motion-distance-xs: 4px` - 微小移动
- `--motion-distance-sm: 8px` - 小幅度移动
- `--motion-distance-md: 12px` - 中等幅度移动

#### Scenario: Stagger animation uses distance token

- **WHEN** 列表项执行 stagger 进入动画
- **THEN** 使用 `--motion-distance-sm` (8px) 作为移动距离

---

### Requirement: Token file location

动效 Token 文件 SHALL 位于 `web/src/styles/motion.css`，并在 `web/src/index.css` 中通过 `@import` 引入。

#### Scenario: Tokens are available globally

- **WHEN** 任何组件的 CSS 文件加载
- **THEN** 所有 motion token 变量均可使用

---

### Requirement: Reduced motion support

系统 SHALL 尊重用户的 `prefers-reduced-motion` 设置，为需要减少动画的用户禁用或简化动画。

#### Scenario: User prefers reduced motion

- **WHEN** 用户系统设置为 `prefers-reduced-motion: reduce`
- **THEN** 所有动画时长变为 0 或显著减少
- **AND** 交错动画变为同时显示
