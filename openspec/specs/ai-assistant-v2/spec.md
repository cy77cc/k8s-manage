# ai-assistant-v2 Specification

## Purpose

定义 AI 助手 V2 的功能规格，实现高质量的运维对话体验。

## Requirements

### Requirement: 混合模式 Agent 架构

系统 SHALL 支持混合模式 Agent，根据用户意图自动选择简单问答或 Agent 执行模式。

#### Scenario: 简单问答直接响应

- **GIVEN** 用户发送简单问候或知识问答
- **WHEN** 意图分类器判断为 "simple" 意图
- **THEN** 系统直接生成文本响应，不调用任何工具
- **AND** 响应时间应小于 1 秒

#### Scenario: 复杂任务走 Agent 流程

- **GIVEN** 用户请求需要查询或操作资源
- **WHEN** 意图分类器判断为 "agentic" 意图
- **THEN** 系统启动 Agent 模式，规划并执行工具调用
- **AND** 返回结构化的执行结果

#### Scenario: 意图分类降级

- **GIVEN** 意图分类器不可用或出错
- **WHEN** 分类失败
- **THEN** 系统自动降级为简单问答模式
- **AND** 记录降级日志

---

### Requirement: 交互式确认机制

系统 SHALL 对高风险操作提供交互式确认机制，用户可选择确认执行或取消。

#### Scenario: 高风险操作触发确认

- **GIVEN** Agent 执行高风险工具（如批量删除、服务部署）
- **WHEN** 工具标记为 "high" 风险等级
- **THEN** 系统中断执行，发送 ask_user 事件
- **AND** 事件包含操作描述、风险等级、详情

#### Scenario: 用户确认后继续执行

- **GIVEN** 系统发送确认请求等待响应
- **WHEN** 用户通过 POST /chat/respond 确认执行
- **THEN** 系统恢复 Agent 执行，完成工具调用
- **AND** 返回执行结果

#### Scenario: 用户取消操作

- **GIVEN** 系统发送确认请求等待响应
- **WHEN** 用户选择取消
- **THEN** 系统返回取消消息，不执行操作
- **AND** 会话状态正常，用户可继续对话

---

### Requirement: K8s 资源查询能力

系统 SHALL 提供统一的 K8s 资源查询工具，支持多种资源类型。

#### Scenario: 查询 Pod 列表

- **GIVEN** 用户请求查看某个命名空间的 Pod
- **WHEN** Agent 调用 k8s_query 工具，resource="pods"
- **THEN** 系统返回 Pod 列表，包含名称、命名空间、状态
- **AND** 支持分页和标签过滤

#### Scenario: 查询 Pod 日志

- **GIVEN** 用户请求查看某个 Pod 的日志
- **WHEN** Agent 调用 k8s_logs 工具
- **THEN** 系统返回 Pod 日志内容
- **AND** 支持指定容器、行数限制

#### Scenario: 查询 K8s 事件

- **GIVEN** 用户请求查看 K8s 事件
- **WHEN** Agent 调用 k8s_events 工具
- **THEN** 系统返回事件列表，包含类型、原因、消息
- **AND** 支持按资源类型和名称过滤

---

### Requirement: 主机运维能力

系统 SHALL 提供主机命令执行能力，支持单机和批量操作。

#### Scenario: 单机命令执行

- **GIVEN** 用户请求在指定主机执行只读命令
- **WHEN** Agent 调用 host_exec 工具
- **THEN** 系统通过 SSH 执行命令并返回结果
- **AND** 执行过程有日志记录

#### Scenario: 批量主机操作

- **GIVEN** 用户请求在多台主机执行命令
- **WHEN** Agent 调用 host_batch 工具
- **THEN** 系统触发交互式确认
- **AND** 用户确认后批量执行

---

### Requirement: 服务部署能力

系统 SHALL 提供服务部署能力，支持预览和执行。

#### Scenario: 部署预览

- **GIVEN** 用户请求部署某个服务
- **WHEN** Agent 调用 service_deploy 工具，preview=true
- **THEN** 系统返回部署预览，包含资源清单、变更影响
- **AND** 不实际执行部署

#### Scenario: 执行部署

- **GIVEN** 用户确认部署预览
- **WHEN** Agent 调用 service_deploy 工具，apply=true
- **THEN** 系统执行部署并返回状态
- **AND** 部署过程可追踪

---

### Requirement: 监控告警能力

系统 SHALL 提供监控指标查询和告警管理能力。

#### Scenario: 查询活跃告警

- **GIVEN** 用户请求查看告警
- **WHEN** Agent 调用 monitor_alert 工具
- **THEN** 系统返回活跃告警列表
- **AND** 支持按严重程度过滤

#### Scenario: 查询监控指标

- **GIVEN** 用户请求查看某个资源的监控指标
- **WHEN** Agent 调用 monitor_metric 工具
- **THEN** 系统返回指标数据
- **AND** 支持时间范围查询

---

### Requirement: 会话持久化

系统 SHALL 持久化存储会话和消息，支持历史查询。

#### Scenario: 创建会话

- **GIVEN** 用户开始新对话
- **WHEN** 系统创建会话
- **THEN** 会话元数据存入数据库
- **AND** 返回会话 ID

#### Scenario: 加载会话历史

- **GIVEN** 用户选择某个历史会话
- **WHEN** 前端请求会话详情
- **THEN** 系统返回会话元数据和消息列表
- **AND** 消息按时间排序

#### Scenario: 消息持久化

- **GIVEN** 用户发送消息或 AI 响应
- **WHEN** 消息生成完成
- **THEN** 消息存入数据库
- **AND** 更新会话的 updated_at 时间

---

### Requirement: MCP 工具扩展

系统 SHALL 支持 MCP (Model Context Protocol) 工具扩展。

#### Scenario: 加载 MCP 工具

- **GIVEN** 配置了 MCP 服务端点
- **WHEN** 系统初始化
- **THEN** 从 MCP 服务获取工具列表
- **AND** 注册到工具注册表

#### Scenario: 调用 MCP 工具

- **GIVEN** Agent 需要调用 MCP 工具
- **WHEN** 工具名称匹配 MCP 工具前缀
- **THEN** 通过 MCP 协议调用外部工具
- **AND** 返回结果给 Agent

---

### Requirement: SSE 流式响应

系统 SHALL 通过 SSE 提供流式响应，支持实时交互。

#### Scenario: 流式文本输出

- **GIVEN** Agent 生成文本响应
- **WHEN** 模型流式输出内容
- **THEN** 系统实时发送 text_delta 事件
- **AND** 前端实时渲染

#### Scenario: 工具执行事件

- **GIVEN** Agent 执行工具
- **WHEN** 工具开始执行和执行完成
- **THEN** 系统发送 tool_start 和 tool_result 事件
- **AND** 前端展示工具执行状态

#### Scenario: 错误处理

- **GIVEN** 执行过程中发生错误
- **WHEN** 错误可恢复
- **THEN** 系统发送 error 事件，payload.recoverable=true
- **AND** 用户可选择重试

---

### Requirement: 前端组件化

前端 SHALL 使用 Ant Design X 组件库实现模块化架构。

#### Scenario: 消息气泡渲染

- **GIVEN** 接收到消息数据
- **WHEN** 消息类型为文本
- **THEN** 使用 Bubble 组件渲染
- **AND** 支持 Markdown 渲染

#### Scenario: 确认面板渲染

- **GIVEN** 接收到 ask_user 事件
- **WHEN** 前端渲染确认面板
- **THEN** 显示操作描述、风险等级、确认按钮
- **AND** 用户交互后发送响应

#### Scenario: 工具执行卡片

- **GIVEN** 接收到 tool_result 事件
- **WHEN** 工具执行完成
- **THEN** 显示工具名称、执行时间、结果摘要
- **AND** 支持展开查看详情
