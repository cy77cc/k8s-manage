# Spec: Notification Gateway

## ADDED Requirements

### Requirement: Provider abstraction

系统 SHALL 提供可扩展通知 Provider 接口与注册机制。

#### Scenario: Provider registry lookup
- **WHEN** Provider 注册后按名称查询
- **THEN** 可获取对应 Provider 实现

#### Scenario: Builtin providers available
- **WHEN** 系统启动通知网关
- **THEN** 内置 `log/dingtalk/wecom/email/sms` Provider 可用

### Requirement: Async fan-out with retries

系统 SHALL 对启用通知渠道执行异步分发与重试。

#### Scenario: Async send to multiple channels
- **WHEN** 同一告警匹配多个渠道
- **THEN** 各渠道并发分发
- **AND** 单渠道失败不阻塞其他渠道

#### Scenario: Retry on failure
- **WHEN** 渠道发送失败
- **THEN** 按 1s/2s/4s 进行最多 3 次重试
- **AND** 最终结果写入 delivery 记录

### Requirement: Delivery audit

系统 SHALL 持久化通知投递结果。

#### Scenario: Record success or failure
- **WHEN** 通知发送流程结束
- **THEN** 写入 `alert_notification_deliveries`
- **AND** 包含状态、目标、错误信息与时间
