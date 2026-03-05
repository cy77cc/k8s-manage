# Spec: Prometheus Integration

## Overview

本规格定义 Prometheus 时序数据库与 k8s-manage 平台的集成要求。

## Requirements

### REQ-PROM-001: Prometheus 客户端

系统 SHALL 提供 Prometheus HTTP API 客户端，支持：

- 即时查询 (Query)
- 范围查询 (QueryRange)
- 元数据查询 (Metadata)

#### Scenario: 即时查询指标

```gherkin
GIVEN Prometheus 服务可用
WHEN 调用 Query("up", now)
THEN 返回所有 up 指标的当前值
AND 响应时间 < 100ms
```

#### Scenario: 范围查询指标

```gherkin
GIVEN Prometheus 服务可用
WHEN 调用 QueryRange("cpu_usage", start, end, 1m)
THEN 返回指定时间范围内的 cpu_usage 数据点
AND 数据点按时间排序
```

### REQ-PROM-002: PromQL 构建器

系统 SHALL 提供 PromQL 查询构建器，支持：

- 指标名称
- 标签过滤
- 时间范围
- 聚合函数 (avg, sum, max, min, count)

#### Scenario: 构建带标签的查询

```gherkin
GIVEN 查询 cpu_usage 指标
WHEN 使用 QueryBuilder 添加标签 source=host, host_id=123
THEN 生成 PromQL: cpu_usage{source="host",host_id="123"}
```

#### Scenario: 构建聚合查询

```gherkin
GIVEN 查询 cpu_usage 指标
WHEN 使用 QueryBuilder 添加 avg 聚合和 5m 范围
THEN 生成 PromQL: avg(cpu_usage[5m])
```

### REQ-PROM-003: Dashboard API 集成

系统 SHALL 将 Dashboard 指标查询改为使用 Prometheus 数据源。

#### Scenario: 查询主机 CPU 使用率

```gherkin
GIVEN 用户请求主机 CPU 使用率
WHEN 调用 GET /api/v1/metrics?metric=cpu_usage&source=host
THEN 从 Prometheus 查询数据
AND 返回格式保持兼容
AND 响应时间 < 50ms (P95)
```

#### Scenario: 查询时间范围指标

```gherkin
GIVEN 用户请求过去 1 小时的指标数据
WHEN 调用 GET /api/v1/metrics?metric=cpu_usage&start=<1h_ago>&end=<now>
THEN 使用 QueryRange API 查询
AND 返回聚合后的时间序列数据
```

### REQ-PROM-004: 配置管理

系统 SHALL 支持 Prometheus 连接配置：

```yaml
prometheus:
  address: "http://prometheus:9090"
  timeout: 10s
  max_concurrent: 10
  retry_count: 3
```

#### Scenario: 配置缺失降级

```gherkin
GIVEN Prometheus 配置未设置
WHEN 系统启动
THEN 记录警告日志
AND 指标查询返回空数据或错误
```

### REQ-PROM-005: 错误处理

系统 SHALL 处理 Prometheus 服务不可用情况：

#### Scenario: Prometheus 不可用

```gherkin
GIVEN Prometheus 服务不可达
WHEN 查询指标
THEN 返回明确的错误信息
AND 记录错误日志
AND 不影响其他系统功能
```

## API Contract

### GET /api/v1/metrics

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| metric | string | 是 | 指标名称 |
| source | string | 否 | 数据源过滤 |
| host_id | uint64 | 否 | 主机 ID 过滤 |
| start | timestamp | 否 | 开始时间 |
| end | timestamp | 否 | 结束时间 |
| step_sec | int | 否 | 步长（秒） |

**响应**:

```json
{
  "window": {
    "start": "2026-03-05T10:00:00Z",
    "end": "2026-03-05T11:00:00Z",
    "granularity_sec": 60
  },
  "dimensions": {
    "metric": "cpu_usage",
    "source": "host"
  },
  "series": [
    {
      "timestamp": "2026-03-05T10:00:00Z",
      "value": 45.5,
      "labels": {"host_id": "123"}
    }
  ]
}
```

## Dependencies

- Prometheus v2.40+ 已部署
- 网络连通性（k8s-manage → Prometheus:9090）

## Migration Notes

- 原 MySQL `metric_points` 表将在迁移完成后删除
- 原 `collectSnapshot` 定时任务将被移除
- 数据采集改由 Exporter + Prometheus pull 模式
