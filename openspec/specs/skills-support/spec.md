# skills-support Specification

## Purpose

为 AI 助手提供技能（Skills）支持，允许管理员通过 YAML 配置预定义多步骤操作流程，简化常见运维操作的执行。

## Requirements

### Requirement: 技能配置管理

系统 SHALL 支持通过 YAML 文件定义技能。

#### Scenario: 技能配置加载
- **GIVEN** configs/skills.yaml 文件存在
- **WHEN** 系统启动或配置重载
- **THEN** 系统 MUST 解析 YAML 并加载所有技能定义
- **AND** 系统 MUST 验证配置格式正确性

#### Scenario: 技能配置验证
- **GIVEN** 技能配置文件被修改
- **WHEN** 加载配置时
- **THEN** 系统 MUST 验证必需字段（name, trigger_patterns, steps）
- **AND** 配置错误时 MUST 返回明确的错误信息

#### Scenario: 技能热重载
- **GIVEN** 运行时修改技能配置
- **WHEN** 管理员触发重载
- **THEN** 系统 MUST 重新加载配置
- **AND** 新配置 MUST 立即生效

### Requirement: 技能路由匹配

系统 SHALL 根据用户消息自动匹配技能。

#### Scenario: 关键词匹配
- **GIVEN** 用户消息包含触发关键词
- **WHEN** 系统进行技能匹配
- **THEN** 系统 MUST 返回匹配度最高的技能
- **AND** 匹配度 MUST 考虑关键词完整度

#### Scenario: 多技能冲突处理
- **GIVEN** 用户消息匹配多个技能
- **WHEN** 系统选择执行哪个技能
- **THEN** 系统 MUST 选择匹配度最高的技能
- **OR** 系统 MUST 询问用户选择

#### Scenario: 无匹配技能
- **GIVEN** 用户消息不匹配任何技能
- **WHEN** 技能匹配阶段
- **THEN** 系统 MUST 回退到 Expert 路由流程
- **AND** 正常执行 AI 对话

### Requirement: 技能参数提取

系统 SHALL 从用户消息中提取技能所需的参数。

#### Scenario: 必需参数提取
- **GIVEN** 技能定义了必需参数
- **WHEN** 用户消息触发该技能
- **THEN** 系统 MUST 尝试从消息中提取参数值
- **AND** 缺少必需参数时 MUST 提示用户补充

#### Scenario: 枚举参数解析
- **GIVEN** 参数定义了 enum_source
- **WHEN** 系统提取参数值
- **THEN** 系统 MUST 从对应数据源查询可选项
- **AND** 系统 MUST 支持模糊匹配资源名称

#### Scenario: 参数默认值
- **GIVEN** 参数定义了默认值
- **WHEN** 用户未提供该参数
- **THEN** 系统 MUST 使用默认值
- **AND** 默认值 MUST 在执行确认时展示

### Requirement: 技能步骤执行

系统 SHALL 按定义的顺序执行技能步骤。

#### Scenario: 工具步骤执行
- **GIVEN** 步骤类型为 tool
- **WHEN** 执行该步骤
- **THEN** 系统 MUST 调用指定的工具
- **AND** 步骤结果 MUST 存储供后续步骤使用

#### Scenario: 审批步骤执行
- **GIVEN** 步骤类型为 approval
- **WHEN** 执行该步骤
- **THEN** 系统 MUST 创建审批工单
- **AND** 系统 MUST 等待审批结果

#### Scenario: 步骤参数模板渲染
- **GIVEN** 步骤定义了 params_template
- **WHEN** 执行该步骤
- **THEN** 系统 MUST 渲染模板中的变量引用
- **AND** 系统 MUST 支持引用前序步骤的结果

### Requirement: 技能与现有系统集成

系统 SHALL 将技能集成到现有的 AI 助手流程中。

#### Scenario: 技能优先级
- **GIVEN** 用户发送消息
- **WHEN** 系统处理请求
- **THEN** 系统 MUST 先尝试匹配技能
- **AND** 无匹配时再走 Expert 路由

#### Scenario: 工具复用
- **GIVEN** 技能步骤指定工具名称
- **WHEN** 执行步骤
- **THEN** 系统 MUST 使用现有的 Tool Registry
- **AND** 工具权限检查 MUST 正常执行

#### Scenario: 流式输出支持
- **GIVEN** 技能执行产生输出
- **WHEN** 向用户返回结果
- **THEN** 系统 MUST 支持流式输出
- **AND** 长时间操作 MUST 显示进度

## Constraints

- 技能配置 MUST 仅由管理员管理
- 系统 MUST 不存储技能执行历史
- 技能执行 MUST 不支持回放功能
- 技能步骤数量 SHOULD 不超过 10 个

## Dependencies

- 现有 Tool Registry
- 现有 Approval Service
- YAML 解析库
