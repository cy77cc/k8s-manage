# Spec: Alerting System Enhancement

## ADDED Requirements

### Requirement: Rule sync to Prometheus

系统 SHALL 将数据库告警规则同步为 Prometheus 规则文件并触发重载。

#### Scenario: Manual sync API
- **WHEN** 调用 `POST /api/v1/alerts/rules/sync`
- **THEN** 读取启用规则并写入 `alerting_rules.yml`
- **AND** 调用 Prometheus `/-/reload`

#### Scenario: Auto sync on CRUD
- **WHEN** 规则被创建、更新或启停
- **THEN** 自动触发规则同步
- **AND** 同步结果可记录

### Requirement: Rule conversion

系统 SHALL 支持规则字段到 Prometheus 规则格式转换。

#### Scenario: Build expression from threshold rule
- **WHEN** 规则包含 `metric/operator/threshold`
- **THEN** 生成 Prometheus `expr`
- **AND** 生成 `for/labels/annotations`

#### Scenario: Use custom PromQL and extra labels/annotations
- **WHEN** 规则设置 `promql_expr`、`labels_json`、`annotations_json`
- **THEN** 优先使用自定义表达式
- **AND** 合并扩展标签与注解

### Requirement: Alertmanager webhook intake

系统 SHALL 接收 Alertmanager webhook 并持久化告警事件。

#### Scenario: Receive firing webhook
- **WHEN** Alertmanager 发送 `firing` 告警到 `/api/v1/alerts/receiver`
- **THEN** 创建或更新告警事件
- **AND** 触发通知分发流程

#### Scenario: Receive resolved webhook
- **WHEN** Alertmanager 发送 `resolved` 告警
- **THEN** 更新事件为 resolved
- **AND** 记录恢复时间
