# Design: Prometheus 监控系统集成

## Overview

本文档详细描述将监控系统从 MySQL 自建方案迁移到 Prometheus 生态的技术设计。

## Goals

1. 时序数据存储迁移到 Prometheus TSDB
2. Dashboard API 改用 Prometheus HTTP API
3. 告警系统对接 Alertmanager
4. 通知系统支持多渠道扩展

## Non-Goals

- Grafana 部署（后续独立规划）
- 长期存储方案（VictoriaMetrics/Thanos）
- K8s 集群监控部署
- Checkpoint/中断恢复

## Architecture

### 系统架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              监控系统架构                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────────────────── 数据采集层 ───────────────────────────┐  │
│  │                                                                   │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │  │
│  │  │node_exporter│  │kube-state-  │  │自定义业务   │              │  │
│  │  │ :9100       │  │metrics      │  │Exporter     │              │  │
│  │  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘              │  │
│  │         │                │                │                      │  │
│  └─────────┼────────────────┼────────────────┼──────────────────────┘  │
│            │                │                │                          │
│            └────────────────┼────────────────┘                          │
│                             │ pull (15s)                                │
│                             ▼                                           │
│  ┌─────────────────────────── 存储与计算层 ──────────────────────────┐  │
│  │                                                                   │  │
│  │  ┌─────────────────────────────────────────────────────────────┐ │  │
│  │  │                    Prometheus                               │ │  │
│  │  │  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐   │ │  │
│  │  │  │TSDB       │ │PromQL     │ │Recording  │ │Alerting   │   │ │  │
│  │  │  │(时序存储)  │ │Engine     │ │Rules      │ │Rules      │   │ │  │
│  │  │  └───────────┘ └───────────┘ └───────────┘ └─────┬─────┘   │ │  │
│  │  │                                                  │         │ │  │
│  │  │  HTTP API: /api/v1/query, /api/v1/query_range   │         │ │  │
│  │  └──────────────────────────────────────────────────┼─────────┘ │  │
│  │                                                     │           │  │
│  └─────────────────────────────────────────────────────┼───────────┘  │
│                                                        │              │
│                         ┌──────────────────────────────┘              │
│                         │                                             │
│                         ▼                                             │
│  ┌─────────────────────────── 告警管理层 ───────────────────────────┐  │
│  │                                                                   │  │
│  │  ┌─────────────────────────────────────────────────────────────┐ │  │
│  │  │                   Alertmanager                              │ │  │
│  │  │  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐   │ │  │
│  │  │  │Grouping   │ │Inhibition │ │Silencing  │ │Routing    │   │ │  │
│  │  │  │(告警分组)  │ │(告警抑制)  │ │(告警静默)  │ │(路由分发)  │   │ │  │
│  │  │  └───────────┘ └───────────┘ └───────────┘ └─────┬─────┘   │ │  │
│  │  │                                                     │       │ │  │
│  │  │  Webhook: /api/v1/alerts/receiver                   │       │ │  │
│  │  └─────────────────────────────────────────────────────┼───────┘ │  │
│  │                                                        │         │  │
│  └────────────────────────────────────────────────────────┼─────────┘  │
│                                                           │            │
│                         ┌─────────────────────────────────┘            │
│                         │                                              │
│                         ▼                                              │
│  ┌─────────────────────────── 应用服务层 ────────────────────────────┐  │
│  │                                                                   │  │
│  │  ┌─────────────────────────────────────────────────────────────┐ │  │
│  │  │                  Go Backend Service                         │ │  │
│  │  │                                                             │ │  │
│  │  │  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐   │ │  │
│  │  │  │Notification   │  │Rule Sync      │  │Dashboard API  │   │ │  │
│  │  │  │Gateway        │  │Service        │  │               │   │ │  │
│  │  │  │               │  │               │  │               │   │ │  │
│  │  │  │- 接收webhook  │  │- DB→Prometheus│  │- PromQL查询   │   │ │  │
│  │  │  │- 持久化事件   │  │- 规则同步     │  │- 聚合计算     │   │ │  │
│  │  │  │- 多渠道分发   │  │- 变更通知     │  │               │   │ │  │
│  │  │  └───────┬───────┘  └───────┬───────┘  └───────┬───────┘   │ │  │
│  │  │          │                  │                  │           │ │  │
│  │  │          ▼                  ▼                  ▼           │ │  │
│  │  │  ┌─────────────────────────────────────────────────────┐   │ │  │
│  │  │  │                    MySQL                           │   │ │  │
│  │  │  │  ┌───────────┐ ┌───────────┐ ┌───────────┐        │   │ │  │
│  │  │  │  │alert_rules│ │alerts     │ │notification│        │   │ │  │
│  │  │  │  │(规则配置)  │ │(事件历史)  │ │_channels  │        │   │ │  │
│  │  │  │  └───────────┘ └───────────┘ └───────────┘        │   │ │  │
│  │  │  └─────────────────────────────────────────────────────┘   │ │  │
│  │  └─────────────────────────────────────────────────────────────┘ │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 数据流

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              数据流向                                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  指标采集流:                                                            │
│  ┌─────────┐     ┌─────────────┐     ┌─────────────┐                   │
│  │Exporter │────▶│ Prometheus  │────▶│ Dashboard   │                   │
│  │         │pull │ (存储+计算)  │query│ (展示)      │                   │
│  └─────────┘     └─────────────┘     └─────────────┘                   │
│                                                                         │
│  告警触发流:                                                            │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐               │
│  │ Prometheus  │────▶│Alertmanager │────▶│Notification │               │
│  │ Alerting    │fire │ (分组/路由)  │webhook│Gateway    │               │
│  └─────────────┘     └─────────────┘     └──────┬──────┘               │
│                                                 │                       │
│                      ┌──────────────────────────┼───────────────────┐   │
│                      ▼                          ▼                   ▼   │
│                 ┌─────────┐              ┌─────────┐          ┌─────────┐
│                 │ MySQL   │              │钉钉/企微│          │邮件/短信│
│                 │(历史记录)│              │(即时通知)│          │(异步)   │
│                 └─────────┘              └─────────┘          └─────────┘
│                                                                         │
│  规则管理流:                                                            │
│  ┌─────────┐     ┌─────────────┐     ┌─────────────┐                   │
│  │ 前端UI  │────▶│ MySQL       │────▶│Rule Sync    │                   │
│  │(规则管理)│CRUD │ alert_rules │sync │Service      │                   │
│  └─────────┘     └─────────────┘     └──────┬──────┘                   │
│                                              │                          │
│                                              ▼                          │
│                                        ┌─────────────┐                 │
│                                        │ Prometheus  │                 │
│                                        │规则文件重载 │                 │
│                                        └─────────────┘                 │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

## Component Design

### 1. Prometheus Client

**路径**: `internal/infra/prometheus/`

```
internal/infra/prometheus/
├── client.go         # HTTP 客户端封装
├── query.go          # PromQL 构建器
├── response.go       # 响应解析
└── config.go         # 配置管理
```

**核心接口**:

```go
// client.go
type Client interface {
    // 即时查询
    Query(ctx context.Context, query string, ts time.Time) (*QueryResult, error)
    // 范围查询
    QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (*QueryResult, error)
    // 元数据查询
    Metadata(ctx context.Context, metric string) ([]MetricMetadata, error)
}

// query.go
type QueryBuilder struct {
    metric    string
    labels    map[string]string
    rangeExpr string
    aggFunc   string
}

// 使用示例
qb := NewQueryBuilder("cpu_usage").
    WithLabel("source", "host").
    WithLabel("host_id", "123").
    WithRange("5m").
    WithAggregation("avg")

query := qb.Build()  // avg(cpu_usage{source="host",host_id="123"}[5m])
```

**配置**:

```go
// config.go
type Config struct {
    Address         string        `yaml:"address" json:"address"`
    Timeout         time.Duration `yaml:"timeout" json:"timeout"`
    MaxConcurrent   int           `yaml:"max_concurrent" json:"max_concurrent"`
    RetryCount      int           `yaml:"retry_count" json:"retry_count"`
}
```

### 2. Rule Sync Service

**路径**: `internal/service/monitoring/rule_sync.go`

**职责**:
- 监听 alert_rules 表变更
- 生成 Prometheus alerting_rules.yml
- 触发 Prometheus 规则重载

**同步策略**:

```
┌─────────────────────────────────────────────────────────────────┐
│                    规则同步流程                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  MySQL alert_rules                                              │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ id | name       | metric      | threshold | operator    │   │
│  │ 1  | CPU高使用   | cpu_usage   | 85        | gt          │   │
│  │ 2  | 内存高使用  | memory_usage| 90        | gt          │   │
│  └─────────────────────────────────────────────────────────┘   │
│                            │                                    │
│                            ▼                                    │
│  ┌─────────────────────────────────────────────────────────┐   │
│  RuleSyncService.Sync()                                       │
│  1. 查询所有 enabled=true 的规则                              │
│  2. 转换为 Prometheus 规则格式                                │
│  3. 写入 alerting_rules.yml                                   │
│  4. 调用 POST /-/reload                                       │
│  └─────────────────────────────────────────────────────────┘   │
│                            │                                    │
│                            ▼                                    │
│  alerting_rules.yml                                             │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ groups:                                                  │   │
│  │   - name: host-alerts                                    │   │
│  │     rules:                                               │   │
│  │       - alert: CPU高使用                                  │   │
│  │         expr: cpu_usage > 85                             │   │
│  │         for: 5m                                          │   │
│  │         labels:                                          │   │
│  │           severity: warning                              │   │
│  │           rule_id: "1"                                   │   │
│  │         annotations:                                     │   │
│  │           summary: "CPU 使用率过高"                       │   │
│  │                                                           │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**触发时机**:
1. 服务启动时同步一次
2. 规则 CRUD 操作后触发同步
3. 定时校验（每 5 分钟）

### 3. Notification Gateway

**路径**: `internal/service/monitoring/notification_gateway.go`

**职责**:
- 接收 Alertmanager webhook
- 持久化告警事件
- 分发到通知渠道

**Webhook 接收**:

```go
// Alertmanager webhook payload
type AlertmanagerWebhook struct {
    Receiver          string  `json:"receiver"`
    Status            string  `json:"status"`  // firing, resolved
    Alerts            []Alert `json:"alerts"`
    GroupLabels       map[string]string `json:"groupLabels"`
    CommonLabels      map[string]string `json:"commonLabels"`
    CommonAnnotations map[string]string `json:"commonAnnotations"`
    ExternalURL       string  `json:"externalURL"`
}

type Alert struct {
    Status       string            `json:"status"`
    Labels       map[string]string `json:"labels"`
    Annotations  map[string]string `json:"annotations"`
    StartsAt     time.Time         `json:"startsAt"`
    EndsAt       time.Time         `json:"endsAt"`
    GeneratorURL string            `json:"generatorURL"`
    Fingerprint  string            `json:"fingerprint"`
}
```

**处理流程**:

```
┌─────────────────────────────────────────────────────────────────┐
│                   告警处理流程                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Alertmanager Webhook                                           │
│         │                                                       │
│         ▼                                                       │
│  ┌─────────────────┐                                           │
│  │ 1. 解析 Payload │                                           │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ 2. 幂等性检查   │  根据 fingerprint 判断是否已处理           │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ 3. 持久化事件   │  写入 alerts 表                           │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ 4. 查询通知渠道 │  从 notification_channels 表查询          │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────────────────────────────────────────┐       │
│  │ 5. 分发通知                                          │       │
│  │                                                      │       │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌────────┐ │       │
│  │  │ 钉钉    │  │ 企微    │  │ 邮件    │  │ 短信   │ │       │
│  │  │ DingTalk│  │ WeCom   │  │ Email   │  │ SMS    │ │       │
│  │  └─────────┘  └─────────┘  └─────────┘  └────────┘ │       │
│  │                                                      │       │
│  │  Provider 接口设计，支持扩展                          │       │
│  └─────────────────────────────────────────────────────┘       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 4. Notification Provider 接口

**路径**: `internal/service/notification/provider.go`

```go
// Provider 通知提供者接口
type Provider interface {
    // 名称
    Name() string
    // 发送通知
    Send(ctx context.Context, alert *AlertEvent, config *ChannelConfig) error
    // 验证配置
    ValidateConfig(config map[string]any) error
}

// ProviderRegistry 提供者注册表
type ProviderRegistry struct {
    providers map[string]Provider
}

func (r *ProviderRegistry) Register(p Provider) {
    r.providers[p.Name()] = p
}

// 内置提供者
type DingTalkProvider struct { ... }  // 钉钉机器人
type WeComProvider struct { ... }     // 企业微信
type EmailProvider struct { ... }     // 邮件
type SMSProvider struct { ... }       // 短信

// 扩展示例：自定义 Webhook
type CustomWebhookProvider struct {
    endpoint string
    headers  map[string]string
}
```

**渠道配置示例**:

```json
// notification_channels.config_json
{
  "dingtalk": {
    "webhook": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
    "secret": "SECxxx"
  }
}

{
  "wecom": {
    "webhook": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"
  }
}

{
  "email": {
    "smtp_host": "smtp.example.com",
    "smtp_port": 465,
    "username": "alert@example.com",
    "password": "xxx",
    "recipients": ["ops@example.com"]
  }
}
```

### 5. Dashboard API 改造

**路径**: `internal/service/monitoring/logic.go`

**当前实现**:

```go
func (l *Logic) GetMetrics(ctx context.Context, query MetricQuery) (*MetricQueryResult, error) {
    q := l.svcCtx.DB.WithContext(ctx).
        Where("metric = ? AND collected_at >= ? AND collected_at <= ?",
              query.Metric, query.Start, query.End)
    // ... 从 MySQL 查询
}
```

**改造后**:

```go
func (l *Logic) GetMetrics(ctx context.Context, query MetricQuery) (*MetricQueryResult, error) {
    // 构建 PromQL
    qb := prometheus.NewQueryBuilder(query.Metric)
    if query.Source != "" {
        qb.WithLabel("source", query.Source)
    }
    if query.HostID > 0 {
        qb.WithLabel("host_id", strconv.FormatUint(query.HostID, 10))
    }

    // 根据查询类型选择 API
    var result *prometheus.QueryResult
    var err error
    if query.Range {
        // 范围查询
        result, err = l.promClient.QueryRange(ctx, qb.Build(),
            query.Start, query.End, time.Duration(query.StepSec)*time.Second)
    } else {
        // 即时查询
        result, err = l.promClient.Query(ctx, qb.Build(), time.Now())
    }

    // 转换为业务格式
    return convertToMetricResult(result), nil
}
```

**新增查询能力**:

```go
// 聚合查询
func (l *Logic) GetMetricAggregation(ctx context.Context, query AggregationQuery) (*AggregationResult, error) {
    // 支持的聚合函数: avg, max, min, sum, count, stddev, stdvar
    // 示例: avg(cpu_usage{source="host"})
    qb := prometheus.NewQueryBuilder(query.Metric).
        WithAggregation(query.Func).
        WithRange(fmt.Sprintf("%dm", query.WindowMin))

    // ...
}

// 多指标查询
func (l *Logic) GetMultipleMetrics(ctx context.Context, metrics []string, start, end time.Time) (map[string]*MetricQueryResult, error) {
    // 并行查询多个指标
    // ...
}
```

## Data Model

### 数据库变更

```sql
-- +migrate Up

-- 1. 扩展告警规则表
ALTER TABLE alert_rules
  ADD COLUMN promql_expr VARCHAR(512) DEFAULT '' COMMENT 'PromQL 表达式（可选，默认自动生成）',
  ADD COLUMN labels_json LONGTEXT COMMENT 'Prometheus 标签 JSON',
  ADD COLUMN annotations_json LONGTEXT COMMENT '告警注解 JSON';

-- 2. 扩展通知渠道表
ALTER TABLE alert_notification_channels
  ADD COLUMN provider VARCHAR(32) DEFAULT 'builtin' COMMENT '提供者: builtin, dingtalk, wecom, email, sms',
  ADD COLUMN config_json LONGTEXT COMMENT '渠道配置 JSON';

-- 3. 新增：告警静默规则表（映射 Alertmanager Silence）
CREATE TABLE IF NOT EXISTS alert_silences (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  silence_id VARCHAR(64) NOT NULL COMMENT 'Alertmanager silence ID',
  matchers_json LONGTEXT NOT NULL COMMENT '匹配规则 JSON',
  starts_at TIMESTAMP NOT NULL,
  ends_at TIMESTAMP NOT NULL,
  created_by BIGINT UNSIGNED NOT NULL,
  comment VARCHAR(512) DEFAULT '',
  status VARCHAR(16) DEFAULT 'active' COMMENT 'active, expired',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_silence_time (starts_at, ends_at),
  INDEX idx_silence_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警静默规则';

-- 4. 删除不再需要的表
DROP TABLE IF EXISTS metric_points;
DROP TABLE IF EXISTS alert_rule_evaluations;

-- +migrate Down

-- 恢复表结构（略）
```

### Prometheus 配置

**prometheus.yml**:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

# 告警规则文件
rule_files:
  - /etc/prometheus/alerting_rules.yml

# Alertmanager 配置
alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  # node_exporter
  - job_name: 'node'
    static_configs:
      - targets: []
    # 动态发现后填充

  # kube-state-metrics
  - job_name: 'kubernetes'
    static_configs:
      - targets: []
```

**alertmanager.yml**:

```yaml
global:
  resolve_timeout: 5m

route:
  group_by: ['alertname', 'severity']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  receiver: 'webhook'

  routes:
    - match:
        severity: critical
      receiver: 'webhook-critical'
    - match:
        severity: warning
      receiver: 'webhook-warning'

receivers:
  - name: 'webhook'
    webhook_configs:
      - url: 'http://k8s-manage:8888/api/v1/alerts/receiver'
        send_resolved: true

  - name: 'webhook-critical'
    webhook_configs:
      - url: 'http://k8s-manage:8888/api/v1/alerts/receiver'
        send_resolved: true

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'host_id']
```

## API Design

### 新增 API

```
POST /api/v1/alerts/receiver
  - 接收 Alertmanager webhook
  - 无需认证（内部调用）

GET /api/v1/metrics/query
  - 查询指标（即时/范围）
  - 支持 PromQL

GET /api/v1/metrics/metadata
  - 获取指标元数据

POST /api/v1/alerts/silences
  - 创建告警静默

DELETE /api/v1/alerts/silences/:id
  - 删除告警静默

POST /api/v1/alerts/rules/sync
  - 手动触发规则同步
```

### 改造 API

```
GET /api/v1/metrics
  - 原: 从 MySQL 查询
  - 改: 从 Prometheus 查询
  - 请求参数不变，响应格式微调

GET /api/v1/alerts/rules
  - 不变，仍从 MySQL 查询

POST /api/v1/alerts/rules
  - 创建后触发规则同步

PUT /api/v1/alerts/rules/:id
  - 更新后触发规则同步

DELETE /api/v1/alerts/rules/:id
  - 删除后触发规则同步
```

## Migration Plan

### 阶段 1: 基础设施 (并行)

```
□ 部署 Alertmanager
□ 更新 Prometheus 配置（添加 alerting）
□ 配置 Alertmanager webhook 指向 k8s-manage
□ 验证告警链路连通性
```

### 阶段 2: 后端改造

```
□ 实现 Prometheus Client
□ 改造 Dashboard API（GetMetrics）
□ 实现 Rule Sync Service
□ 实现 Notification Gateway
□ 实现 Provider 接口和内置实现
□ 数据库迁移
```

### 阶段 3: 清理与验证

```
□ 删除 metric_points 相关代码
□ 删除 alert_rule_evaluations 相关代码
□ 删除 collectSnapshot 方法
□ 前端适配新响应格式
□ 端到端测试
```

## Risks & Mitigations

| 风险 | 影响 | 缓解措施 |
|-----|------|---------|
| 迁移期间监控中断 | 高 | 并行运行，逐步切换 |
| PromQL 复杂度 | 中 | 提供查询模板，封装 QueryBuilder |
| 规则同步延迟 | 中 | 变更后主动同步 + 定时校验 |
| Alertmanager 单点 | 高 | 部署多副本（后续） |
| 通知渠道故障 | 中 | 异步发送 + 重试 + 降级 |

## Rollback Plan

1. **保留 MySQL 表结构** - 迁移完成后暂不删除，观察 1 周后删除
2. **配置开关** - 通过配置切换 Prometheus/MySQL 数据源
3. **规则回退** - 保留原 alert_rules 表结构，可快速回退

## Success Criteria

- [ ] Dashboard 查询响应时间 < 50ms（P95）
- [ ] 告警从触发到通知延迟 < 30s
- [ ] 规则变更同步延迟 < 5s
- [ ] 支持 3+ 通知渠道
- [ ] 数据保留 30 天
