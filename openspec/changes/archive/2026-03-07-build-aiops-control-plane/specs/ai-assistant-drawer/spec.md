## MODIFIED Requirements

### Requirement: 消息渲染

系统 SHALL 使用独立的消息块渲染机制渲染消息，并在富渲染失败时安全降级。随着后端演进到 AIOps control-plane，抽屉 MUST 能消费更丰富的 application-card 事件，而不再仅依赖纯文本聊天输出。

#### Scenario: 用户消息渲染
- **GIVEN** 用户发送消息
- **WHEN** 消息添加到列表
- **THEN** 用户消息靠右显示
- **AND** 显示用户头像

#### Scenario: AI 消息渲染
- **GIVEN** AI 返回消息
- **WHEN** 消息添加到列表
- **THEN** AI 消息靠左显示
- **AND** 消息先被规范化为可渲染的消息块
- **AND** 支持 Markdown、代码、思考过程、推荐内容等独立块渲染

#### Scenario: 富渲染块失败时安全降级
- **GIVEN** AI 消息包含富渲染块
- **WHEN** 某个块的渲染器执行失败
- **THEN** 仅失败块降级为安全文本或回退视图
- **AND** 同一条消息中的其他块继续渲染
- **AND** AI 助手抽屉保持可用

#### Scenario: 流式输出
- **GIVEN** AI 正在生成回复
- **WHEN** 收到 SSE delta 事件
- **THEN** 实时更新消息内容
- **AND** 显示打字效果

#### Scenario: application-card 事件渲染
- **GIVEN** 后端产生计划、步骤、证据、审批或下一步动作等平台事件
- **WHEN** 抽屉接收到这些事件
- **THEN** 抽屉 MUST 能将其渲染为对应的应用卡片或消息块
- **AND** 抽屉在兼容期内仍可继续消费现有聊天流事件

### Requirement: 工具执行卡片

系统 SHALL 显示工具执行卡片，并逐步扩展为更完整的操作步骤与证据卡片体系。

#### Scenario: 工具开始执行
- **GIVEN** AI 调用工具
- **WHEN** 收到 tool_call 事件
- **THEN** 显示工具卡片
- **AND** 状态显示为 "执行中"
- **AND** 显示加载动画

#### Scenario: 工具执行成功
- **GIVEN** 工具正在执行
- **WHEN** 收到 tool_result 事件且成功
- **THEN** 状态更新为 "成功"
- **AND** 显示执行耗时
- **AND** 图标变为绿色勾号

#### Scenario: 工具执行失败
- **GIVEN** 工具正在执行
- **WHEN** 收到 tool_result 事件且失败
- **THEN** 状态更新为 "失败"
- **AND** 图标变为红色叉号

#### Scenario: control-plane steps render as richer cards
- **GIVEN** 后端为某次 AIOps 任务生成计划步骤、证据、审批或下一步动作
- **WHEN** 抽屉收到对应的平台事件
- **THEN** UI MUST 能把这些状态渲染为操作步骤卡片、证据卡片、审批卡片或下一步动作卡片
