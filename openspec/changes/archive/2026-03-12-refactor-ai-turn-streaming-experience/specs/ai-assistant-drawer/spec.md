## MODIFIED Requirements

### Requirement: 消息渲染

系统 SHALL 使用独立的消息块渲染机制渲染消息，并在富渲染失败时安全降级。随着后端演进到 AIOps control-plane，抽屉 MUST 以 turn-and-block 状态模型消费聊天与平台事件，而不再仅依赖单条纯文本消息输出。

#### Scenario: 用户消息渲染
- **GIVEN** 用户发送消息
- **WHEN** 消息添加到列表
- **THEN** 用户消息靠右显示
- **AND** 显示用户头像

#### Scenario: AI 消息渲染
- **GIVEN** AI 返回消息
- **WHEN** assistant turn 添加到列表
- **THEN** AI turn 靠左显示
- **AND** turn MUST 先被规范化为可渲染的消息块集合
- **AND** 支持 Markdown、状态、计划、工具、审批、证据、思考过程、推荐内容等独立块渲染

#### Scenario: 富渲染块失败时安全降级
- **GIVEN** AI turn 包含富渲染块
- **WHEN** 某个块的渲染器执行失败
- **THEN** 仅失败块降级为安全文本或回退视图
- **AND** 同一 turn 中的其他块继续渲染
- **AND** AI 助手抽屉保持可用

#### Scenario: 流式输出
- **GIVEN** AI 正在生成回复
- **WHEN** 收到 turn 或 block 增量事件
- **THEN** 抽屉 MUST 实时更新对应 turn 中的目标块
- **AND** 最终回答文本块 MUST 显示打字效果
- **AND** 状态、工具和审批类块 MUST 使用状态更新而非逐字打印

#### Scenario: 默认展示策略限制内部噪音
- **GIVEN** assistant turn 同时包含最终回答、thinking、证据细节和原始工具载荷
- **WHEN** 抽屉以普通模式渲染该 turn
- **THEN** 抽屉 MUST 默认展示最终回答、关键状态、审批信息和证据摘要
- **AND** thinking 与原始工具 JSON MUST 默认折叠或隐藏
- **AND** 用户无需阅读内部 agent 原始输出即可理解结论与下一步动作

#### Scenario: 展示模式由显式用户偏好控制
- **GIVEN** 用户首次打开 AI 抽屉
- **WHEN** 前端初始化 turn/block 渲染设置
- **THEN** 抽屉 MUST 默认使用普通模式
- **AND** 抽屉 MUST 提供显式的展示模式切换入口
- **AND** 用户切换到调试模式后，前端 MUST 持久化该偏好供后续打开复用
- **AND** 展示模式切换 MUST 独立于后端 rollout flag

#### Scenario: application-card 事件渲染
- **GIVEN** 后端产生计划、步骤、证据、审批或下一步动作等平台事件
- **WHEN** 抽屉接收到这些事件
- **THEN** 抽屉 MUST 能将其渲染为对应的应用卡片或消息块
- **AND** 抽屉在兼容期内仍可继续消费现有聊天流事件

## MODIFIED Requirements

### Requirement: 工具执行卡片

系统 SHALL 显示工具执行卡片，并逐步扩展为更完整的操作步骤与证据卡片体系；工具卡片 MUST 作为 assistant turn 内的独立块进行创建、更新和完成。

#### Scenario: 工具开始执行
- **GIVEN** AI 调用工具
- **WHEN** 收到工具开始事件
- **THEN** 显示工具卡片
- **AND** 状态显示为 "执行中"
- **AND** 显示加载动画

#### Scenario: 工具执行成功
- **GIVEN** 工具正在执行
- **WHEN** 收到工具完成事件且成功
- **THEN** 状态更新为 "成功"
- **AND** 显示执行耗时
- **AND** 图标变为绿色勾号

#### Scenario: 工具执行失败
- **GIVEN** 工具正在执行
- **WHEN** 收到工具完成事件且失败
- **THEN** 状态更新为 "失败"
- **AND** 图标变为红色叉号

#### Scenario: control-plane steps render as richer cards
- **GIVEN** 后端为某次 AIOps 任务生成计划步骤、证据、审批或下一步动作
- **WHEN** 抽屉收到对应的平台事件
- **THEN** UI MUST 能把这些状态渲染为操作步骤卡片、证据卡片、审批卡片或下一步动作卡片
- **AND** 同一 assistant turn 中的多个卡片 MUST 保持稳定顺序和可回放状态

#### Scenario: 调试模式允许查看深度细节
- **GIVEN** 用户处于专家模式或调试模式
- **WHEN** 抽屉渲染 assistant turn
- **THEN** thinking、原始工具参数、原始工具结果和扩展证据细节 MAY 被显式展开
- **AND** 这些调试细节 MUST 仍然附着于原有 turn/block 结构而不是单独生成另一条消息

#### Scenario: 动效遵守 reduced-motion 偏好
- **GIVEN** 用户系统启用了 reduced-motion 偏好
- **WHEN** 抽屉渲染流式 text、status 或卡片更新
- **THEN** 最终回答块 MUST 关闭逐字动画并按 chunk 直接追加
- **AND** 抽屉 MUST 禁用非必要的平滑滚动和卡片过渡动画
- **AND** 状态变化与内容更新仍 MUST 保持可见

#### Scenario: 交互入口支持键盘与触控
- **GIVEN** 抽屉中存在审批按钮、展开按钮、工具详情入口或跳转到最新操作
- **WHEN** 用户通过键盘或触控设备操作这些入口
- **THEN** 所有入口 MUST 可聚焦并可通过键盘触发
- **AND** 触控目标 MUST 满足最小可点击尺寸
- **AND** 焦点顺序 MUST 与视觉顺序一致

#### Scenario: 条件自动跟随不会抢走用户滚动位置
- **GIVEN** assistant turn 正在持续流式更新
- **WHEN** 用户仍停留在消息列表底部附近
- **THEN** 抽屉 MAY 自动跟随到最新 block

#### Scenario: 用户离开底部后显示回到底部入口
- **GIVEN** assistant turn 正在持续流式更新
- **WHEN** 用户主动上滑离开列表底部
- **THEN** 后续增量更新 MUST NOT 强制把滚动位置拉回底部
- **AND** 抽屉 MUST 显示明确的“跳转到最新”入口
- **AND** 用户点击该入口后，抽屉 MUST 恢复对活跃 turn 的自动跟随
