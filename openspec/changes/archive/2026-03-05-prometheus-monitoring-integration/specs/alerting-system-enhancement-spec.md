# Spec: Alerting System Enhancement

## Overview

本规格定义告警系统与 Prometheus Alertmanager 的集成要求，包括规则同步、告警接收和事件管理。

## Requirements

### REQ-ALERT-001: 规则同步服务

系统 SHALL 将数据库中的告警规则同步到 Prometheus。

#### Scenario: 规则自动同步

```gherkin
GIVEN alert_rules 表中有 enabled=true 的规则
WHEN Rule Sync Service 执行同步
THEN 生成 alerting_rules.yml 文件
AND Prometheus 重载规则成功
AND 规则生效延迟 < 5s
```

#### Scenario: 规则格式转换

```gherkin
GIVEN 数据库中有一条告警规则:
  - metric: cpu_usage
  - operator: gt
  - threshold: 85
  - duration_sec: 300
  - severity: warning
WHEN 转换为 Prometheus 格式
THEN 生成的规则为:
  - expr: cpu_usage > 85
  - for: 5m
  - labels.severity: warning
```

### REQ-ALERT-002: 规则 CRUD 触发同步

系统 SHALL 在规则变更后自动触发同步。

#### Scenario: 创建规则后同步

```gherkin
GIVEN 用户创建新的告警规则
WHEN 规则写入数据库成功
THEN 触发 Rule Sync Service 同步
AND 新规则在 Prometheus 中生效
```

#### Scenario: 更新规则后同步

```gherkin
GIVEN 用户修改告警规则
WHEN 规则更新到数据库成功
THEN 触发 Rule Sync Service 同步
AND 更新后的规则在 Prometheus 中生效
```

#### Scenario: 删除规则后同步

```gherkin
GIVEN 用户删除告警规则
WHEN 规则从数据库删除成功
THEN 触发 Rule Sync Service 同步
AND 规则从 Prometheus 中移除
```

### REQ-ALERT-003: Alertmanager Webhook 接收

系统 SHALL 接收 Alertmanager 发送的告警 webhook。

#### Scenario: 接收 firing 告警

```gherkin
GIVEN Alertmanager 触发告警
WHEN 发送 webhook 到 POST /api/v1/alerts/receiver
THEN 系统解析告警 payload
AND 创建 alert_events 记录
AND 触发通知分发
```

#### Scenario: 接收 resolved 告警

```gherkin
GIVEN Alertmanager 发送告警恢复
WHEN webhook status 为 "resolved"
THEN 更新对应的 alert_events 状态
AND 发送恢复通知
```

### REQ-ALERT-004: 告警事件持久化

系统 SHALL 持久化所有告警事件。

#### Scenario: 告警事件记录

```gherkin
GIVEN 收到 Alertmanager webhook
WHEN 处理告警事件
THEN 写入 alerts 表:
  - rule_id: 从 labels.rule_id 提取
  - title: alertname
  - status: firing/resolved
  - triggered_at: startsAt
  - labels: 保存所有 labels
```

#### Scenario: 幂等性处理

```gherkin
GIVEN 已收到相同 fingerprint 的告警
WHEN 再次收到相同告警
THEN 不重复创建事件记录
AND 更新现有记录
```

### REQ-ALERT-005: 告警静默管理

系统 SHALL 支持告警静默规则管理（映射 Alertmanager Silence）。

#### Scenario: 创建静默规则

```gherkin
GIVEN 管理员创建静默规则
WHEN 指定匹配条件和时间范围
THEN 调用 Alertmanager API 创建 silence
AND 记录到 alert_silences 表
```

#### Scenario: 静默规则过期

```gherkin
GIVEN 静默规则到达 ends_at 时间
WHEN 系统检测到过期
THEN 更新 alert_silences.status 为 "expired"
```

### REQ-ALERT-006: 规则扩展字段

alert_rules 表 SHALL 支持扩展字段：

| 字段 | 类型 | 说明 |
|-----|------|------|
| promql_expr | VARCHAR(512) | 自定义 PromQL（可选） |
| labels_json | LONGTEXT | Prometheus 标签 JSON |
| annotations_json | LONGTEXT | 告警注解 JSON |

#### Scenario: 使用自定义 PromQL

```gherkin
GIVEN 告警规则设置了 promql_expr
WHEN 同步规则到 Prometheus
THEN 使用 promql_expr 而非自动生成
AND 添加 labels_json 中的标签
```

## API Contract

### POST /api/v1/alerts/receiver

**请求体** (Alertmanager webhook):

```json
{
  "receiver": "webhook",
  "status": "firing",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "CPU高使用",
        "severity": "warning",
        "rule_id": "1",
        "host_id": "123"
      },
      "annotations": {
        "summary": "CPU 使用率过高",
        "value": "92.5"
      },
      "startsAt": "2026-03-05T10:00:00Z",
      "fingerprint": "abc123"
    }
  ]
}
```

**响应**:

```json
{
  "status": "success"
}
```

### POST /api/v1/alerts/rules/sync

**响应**:

```json
{
  "status": "success",
  "synced_count": 6,
  "synced_at": "2026-03-05T10:00:00Z"
}
```

## Dependencies

- Alertmanager 已部署
- Prometheus 已配置 alerting
- k8s-manage 可被 Alertmanager 访问

## Migration Notes

- 原 `evaluateRules` 方法将被移除
- 原 `alert_rule_evaluations` 表将被删除
- 告警评估完全由 Prometheus 处理
