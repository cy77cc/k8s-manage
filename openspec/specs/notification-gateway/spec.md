# Spec: Notification Gateway

## Overview

本规格定义通知网关的设计，支持多渠道通知分发，预留扩展机制。

## Requirements

### REQ-NOTIFY-001: Provider 接口

系统 SHALL 定义 Provider 接口，支持多通知渠道实现：

```go
type Provider interface {
    Name() string
    Send(ctx context.Context, alert *AlertEvent, config *ChannelConfig) error
    ValidateConfig(config map[string]any) error
}
```

#### Scenario: 注册新 Provider

```gherkin
GIVEN 实现了 Provider 接口的新渠道
WHEN 调用 ProviderRegistry.Register(provider)
THEN 该渠道可被使用
AND 可通过 provider name 查找
```

### REQ-NOTIFY-002: 内置 Provider

系统 SHALL 提供以下内置 Provider：

| Provider | 名称 | 功能 |
|----------|-----|------|
| Log | log | 记录日志（默认） |
| DingTalk | dingtalk | 钉钉机器人 |
| WeCom | wecom | 企业微信 |
| Email | email | 邮件通知 |
| SMS | sms | 短信通知（接口预留） |

#### Scenario: 钉钉通知发送

```gherkin
GIVEN 配置了钉钉机器人 webhook
WHEN 有告警需要通知
THEN 发送 Markdown 格式消息到钉钉
AND 消息包含告警名称、级别、时间、详情
```

#### Scenario: 企业微信通知发送

```gherkin
GIVEN 配置了企业微信 webhook
WHEN 有告警需要通知
THEN 发送 Markdown 格式消息到企业微信
AND 消息包含告警名称、级别、时间、详情
```

#### Scenario: 邮件通知发送

```gherkin
GIVEN 配置了 SMTP 服务器
WHEN 有告警需要通知
THEN 发送邮件到指定收件人
AND 邮件标题包含告警级别和名称
AND 邮件正文包含告警详情
```

### REQ-NOTIFY-003: 渠道配置

notification_channels 表 SHALL 支持扩展配置：

| 字段 | 类型 | 说明 |
|-----|------|------|
| provider | VARCHAR(32) | 提供者名称 |
| config_json | LONGTEXT | 渠道配置 JSON |

#### Scenario: 钉钉渠道配置

```json
{
  "webhook": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
  "secret": "SECxxx"
}
```

#### Scenario: 邮件渠道配置

```json
{
  "smtp_host": "smtp.example.com",
  "smtp_port": 465,
  "username": "alert@example.com",
  "password": "xxx",
  "recipients": ["ops@example.com", "dev@example.com"]
}
```

### REQ-NOTIFY-004: 通知分发

系统 SHALL 将告警分发给所有启用的通知渠道。

#### Scenario: 多渠道分发

```gherkin
GIVEN 配置了钉钉和邮件两个渠道
WHEN 告警触发通知
THEN 同时发送钉钉消息和邮件
AND 各渠道独立处理，互不影响
AND 任一渠道失败不影响其他渠道
```

#### Scenario: 渠道配置错误

```gherkin
GIVEN 某渠道配置有误
WHEN 尝试发送通知
THEN 记录错误日志
AND 不阻塞其他渠道
AND 返回部分失败状态
```

### REQ-NOTIFY-005: 异步发送

系统 SHALL 支持异步发送通知，避免阻塞主流程。

#### Scenario: 异步发送

```gherkin
GIVEN 收到告警 webhook
WHEN 处理通知
THEN 持久化事件后立即返回
AND 异步执行通知发送
AND 发送结果记录到日志
```

### REQ-NOTIFY-006: 重试机制

系统 SHALL 对失败的通知进行重试。

#### Scenario: 发送失败重试

```gherkin
GIVEN 通知发送失败（网络超时等）
WHEN 重试条件满足
THEN 进行最多 3 次重试
AND 每次重试间隔递增（1s, 2s, 4s）
AND 最终失败记录到日志
```

### REQ-NOTIFY-007: 告警格式

系统 SHALL 统一告警通知格式。

#### Scenario: Firing 告警格式

```markdown
🚨 告警: CPU高使用

**状态**: 🔴 触发中
**级别**: warning
**时间**: 2026-03-05 10:00:00
**主机**: node-01 (ID: 123)

**详情**:
- 指标: cpu_usage
- 当前值: 92.5%
- 阈值: 85%

[查看详情](http://k8s-manage/alerts/1)
```

#### Scenario: Resolved 告警格式

```markdown
✅ 告警恢复: CPU高使用

**状态**: 🟢 已恢复
**级别**: warning
**时间**: 2026-03-05 10:15:00
**主机**: node-01 (ID: 123)

**详情**:
- 指标: cpu_usage
- 恢复值: 72.3%
- 持续时间: 15分钟

[查看详情](http://k8s-manage/alerts/1)
```

## API Contract

### POST /api/v1/alerts/receiver

（见 Alerting System Enhancement 规格）

### GET /api/v1/notification/channels

**响应**:

```json
{
  "channels": [
    {
      "id": 1,
      "name": "钉钉告警群",
      "type": "dingtalk",
      "provider": "dingtalk",
      "enabled": true
    },
    {
      "id": 2,
      "name": "运维邮件组",
      "type": "email",
      "provider": "email",
      "enabled": true
    }
  ]
}
```

### POST /api/v1/notification/channels

**请求**:

```json
{
  "name": "钉钉告警群",
  "type": "dingtalk",
  "provider": "dingtalk",
  "config_json": {
    "webhook": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
    "secret": "SECxxx"
  },
  "enabled": true
}
```

## Extension Points

### 添加新通知渠道

1. 实现 `Provider` 接口
2. 在服务初始化时注册到 `ProviderRegistry`
3. 前端添加对应配置表单（可选）

```go
// 示例：添加 Slack 支持
type SlackProvider struct{}

func (p *SlackProvider) Name() string { return "slack" }

func (p *SlackProvider) Send(ctx context.Context, alert *AlertEvent, config *ChannelConfig) error {
    // 实现发送逻辑
}

// 注册
registry.Register(&SlackProvider{})
```

## Dependencies

- Alertmanager webhook 配置正确
- 各渠道外部服务可访问（钉钉/企微/SMTP）

## Security Considerations

- Webhook secret 验证（可选）
- 敏感配置加密存储（secret 字段）
- 访问日志记录
