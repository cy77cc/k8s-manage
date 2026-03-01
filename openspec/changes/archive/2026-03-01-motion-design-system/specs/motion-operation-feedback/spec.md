# Motion Operation Feedback

操作结果反馈动效，包括成功、失败、高风险操作的视觉反馈。

## ADDED Requirements

### Requirement: Success feedback animation

操作成功时 SHALL 提供视觉反馈。

反馈效果：
- Message 图标 bounce 弹入（scale 0→1.2→1）
- 时长 300ms
- 缓动 `--motion-ease-spring`
- 可选：目标元素绿色边框闪烁

#### Scenario: Success message animation

- **WHEN** 操作成功（如重启主机成功）
- **THEN** Message 图标有弹跳动画
- **AND** 用户感知到"成功"

#### Scenario: Success element flash

- **WHEN** 操作按钮对应的操作成功
- **THEN** 按钮短暂显示绿色边框脉冲
- **AND** 时长 200ms

---

### Requirement: Failure feedback animation

操作失败时 SHALL 提供明确的警告反馈。

反馈效果：
- Message shake 抖动
- 时长 300ms
- 红色边框闪烁

#### Scenario: Failure message animation

- **WHEN** 操作失败（如网络超时）
- **THEN** Message 水平抖动 3 次
- **AND** 用户感知到"失败"

#### Scenario: Failure element flash

- **WHEN** 操作失败
- **THEN** 相关元素显示红色边框脉冲

---

### Requirement: High-risk operation warning animation

高风险操作确认时 SHALL 有持续的视觉警告。

反馈效果：
- Alert 边框脉冲（红色/橙色）
- 确认按钮呼吸效果
- 持续动画直到用户确认或取消

#### Scenario: High-risk command alert

- **WHEN** AI 助手检测到高风险命令
- **THEN** Alert 组件边框持续脉冲
- **AND** 确认按钮有呼吸效果

#### Scenario: High-risk batch operation

- **WHEN** 用户执行批量重启等高风险操作
- **THEN** 确认对话框有警告动画

---

### Requirement: Batch operation progress animation

批量操作时 SHALL 提供逐项进度反馈。

反馈效果：
- 每项结果依次出现
- 成功/失败用颜色区分
- 进度条平滑更新

#### Scenario: Batch operation progress

- **WHEN** 批量重启 50 台主机
- **THEN** 每台主机结果依次显示
- **AND** 成功显示绿色，失败显示红色
- **AND** 进度条平滑填充

---

### Requirement: CSS classes for feedback

系统 SHALL 提供以下 CSS 辅助类：
- `.motion-success-flash` - 成功边框闪烁
- `.motion-error-shake` - 错误抖动
- `.motion-high-risk-pulse` - 高风险脉冲
- `.motion-progress-bar` - 进度条动画

#### Scenario: Developer applies feedback class

- **WHEN** 开发者给元素添加 `.motion-success-flash` 类
- **THEN** 元素执行成功反馈动画
