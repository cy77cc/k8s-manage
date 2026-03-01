# Motion AI Assistant

AI 助手专属动效，增强工具调用、状态指示、思考过程的视觉体验。

## ADDED Requirements

### Requirement: First token waiting animation

AI 等待首包时 SHALL 显示状态指示动画。

已有：三点波浪动画（ai-dot-wave）
增强：根据状态显示不同提示文字

状态类型：
- `thinking` - "正在思考..."
- `tool_calling` - "调用工具: {tool_name}..."
- `approval_pending` - "等待审批..."

#### Scenario: Thinking state animation

- **WHEN** AI 模型正在生成
- **THEN** 显示思考状态文字和动画

#### Scenario: Tool calling state animation

- **WHEN** AI 正在调用工具
- **THEN** 显示工具名称和执行状态

---

### Requirement: Tool trace timeline animation

工具调用轨迹 SHALL 以时间线样式展示，每项有状态图标和摘要。

状态类型：
- `running` - 脉冲动画 + 进度指示
- `success` - 绿色勾号弹入 + 摘要淡入
- `error` - 红色叉号 + 边框闪红
- `approval` - 警告图标 + 边框脉冲

#### Scenario: Tool success animation

- **WHEN** 工具执行成功
- **THEN** 显示绿色勾号
- **AND** 摘要信息淡入

#### Scenario: Tool error animation

- **WHEN** 工具执行失败
- **THEN** 显示红色叉号
- **AND** 错误信息高亮

#### Scenario: Tool approval animation

- **WHEN** 工具需要审批
- **THEN** 显示警告图标和审批按钮
- **AND** 边框持续脉冲

---

### Requirement: Thinking process expand animation

思考过程展开时 SHALL 有平滑动画。

动画效果：
- Collapse 展开动画
- 背景渐变
- 内容淡入

#### Scenario: Thinking block expand

- **WHEN** 用户点击"查看思考过程"
- **THEN** 内容平滑展开
- **AND** 动画时长 200ms

---

### Requirement: Recommendation chip animation

建议卡片 SHALL 依次淡入，比现有动画更流畅。

增强效果：
- 从上到下依次出现
- 每项延迟 80ms
- 动画时长 180ms

#### Scenario: Recommendations appear

- **WHEN** AI 返回建议卡片
- **THEN** 卡片依次淡入
- **AND** 有轻微的上滑效果

---

### Requirement: Streaming cursor enhancement

流式输出光标 SHALL 闪烁流畅。

已有：ai-cursor-blink 动画
增强：考虑添加轻微的颜色脉冲

#### Scenario: Typewriter cursor

- **WHEN** AI 正在输出内容
- **THEN** 光标在最后一个字符后闪烁

---

### Requirement: Session list animation

会话列表切换时 SHALL 有过渡动画。

#### Scenario: Session switch animation

- **WHEN** 用户切换会话
- **THEN** 消息列表平滑过渡

---

### Requirement: User message bubble animation

用户消息气泡 SHALL 有轻微的进入动画。

#### Scenario: User message appear

- **WHEN** 用户发送消息
- **THEN** 消息气泡从右侧淡入
