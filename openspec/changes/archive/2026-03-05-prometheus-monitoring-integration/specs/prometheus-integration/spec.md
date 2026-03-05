# Spec: Prometheus Integration

## ADDED Requirements

### Requirement: Prometheus client queries

系统 SHALL 提供 Prometheus HTTP API 查询能力，支持即时查询、范围查询、元数据查询。

#### Scenario: Query instant metric
- **WHEN** 调用 Prometheus `Query` 查询即时指标
- **THEN** 返回 vector 结果
- **AND** 查询失败时返回明确错误

#### Scenario: Query range metric
- **WHEN** 调用 Prometheus `QueryRange` 查询时间序列
- **THEN** 返回 matrix 结果
- **AND** 数据可转换为业务 `series` 格式

### Requirement: Dashboard metrics use Prometheus

系统 SHALL 将监控指标查询切换为 Prometheus 数据源并保持响应兼容。

#### Scenario: Get metrics from Prometheus
- **WHEN** 调用 `GET /api/v1/metrics`
- **THEN** 后端优先通过 Prometheus 查询
- **AND** 返回 `window/dimensions/series` 结构

#### Scenario: Fallback when Prometheus unavailable
- **WHEN** Prometheus 查询失败
- **THEN** 系统执行降级查询
- **AND** 不影响接口可用性

### Requirement: Aggregation and metadata APIs

系统 SHALL 支持聚合查询与指标元数据查询能力。

#### Scenario: Aggregation query
- **WHEN** 请求指标聚合（如 avg）
- **THEN** 使用 PromQL 聚合表达式查询
- **AND** 返回聚合值与时间戳

#### Scenario: Metadata query
- **WHEN** 请求指标 metadata
- **THEN** 返回类型、帮助信息、单位等元数据
