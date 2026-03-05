# Proposal: Prometheus 监控系统集成

## Summary

将现有自建监控系统迁移到 Prometheus 生态，实现：
1. **时序数据存储** - Prometheus TSDB 替代 MySQL metric_points
2. **告警规则管理** - 数据库管理规则，自动同步到 Prometheus
3. **告警通知网关** - 接收 Alertmanager webhook，支持多渠道扩展

## Motivation

### 当前问题

**1. MySQL 存储时序数据性能差**
```
当前状态:
- 10万条数据查询 > 200ms
- 无降采样机制
- 无内置聚合函数
- 数据保留策略缺失（代码写 7 天，规划 30 天）
```

**2. 告警系统功能有限**
```
当前实现:
- 简单阈值告警
- 无告警分组/抑制/静默
- 无 PromQL 级复杂查询
- 通知中心无实际渠道
```

**3. 运维可观测性不足**
```
缺失能力:
- 无标准 Exporter 支持
- 无服务发现机制
- 无法对接 Grafana 等生态工具
```

### Prometheus 优势

| 对比项 | 当前 MySQL | Prometheus |
|--------|-----------|------------|
| 写入性能 | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| 查询性能 | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| 数据压缩 | ⭐ | ⭐⭐⭐⭐⭐ |
| 聚合能力 | ⭐ | ⭐⭐⭐⭐⭐ (PromQL) |
| 告警管理 | ⭐⭐ | ⭐⭐⭐⭐⭐ (Alertmanager) |
| 生态集成 | ⭐ | ⭐⭐⭐⭐⭐ |

## Goals

1. **存储迁移** - 时序指标存储从 MySQL 迁移到 Prometheus
2. **查询改造** - Dashboard API 改用 Prometheus HTTP API
3. **告警集成** - Alertmanager 对接，保留数据库规则管理
4. **通知扩展** - Notification Gateway 支持多渠道，预留扩展

## Non-Goals

- Grafana 可视化部署（后续独立规划）
- 长期存储方案（VictoriaMetrics/Thanos，后续按需引入）
- K8s 集群内监控（node_exporter/kube-state-metrics 部署，另行处理）
- 前端 UI 大改（仅适配 API 返回格式变化）

## Proposed Changes

### 架构变更

```
改造前:
┌─────────┐    ┌─────────┐    ┌─────────┐
│Collector│───▶│ MySQL   │───▶│Alert    │
│(自建)   │    │metric_  │    │Engine   │
│         │    │points   │    │(自建)   │
└─────────┘    └─────────┘    └─────────┘

改造后:
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│Exporter │───▶│Prometheus│──▶│Alertmgr │───▶│Notify   │
│(标准)   │    │(TSDB)   │    │(成熟)   │    │Gateway  │
└─────────┘    └─────────┘    └─────────┘    └─────────┘
                    │
                    │ PromQL
                    ▼
              ┌─────────┐
              │Dashboard│
              │ API     │
              └─────────┘
```

### 新增组件

1. **Rule Sync Service** (`internal/service/monitoring/rule_sync.go`)
   - 监听 alert_rules 表变更
   - 生成 Prometheus alerting_rules.yml
   - 通过 API 重载规则

2. **Prometheus Client** (`internal/infra/prometheus/`)
   - HTTP API 封装
   - PromQL 构建器
   - 响应解析

3. **Notification Gateway** (`internal/service/monitoring/notification_gateway.go`)
   - Alertmanager webhook 接收
   - 告警事件持久化
   - 多渠道通知分发

### 数据库变更

```sql
-- 扩展告警规则表
ALTER TABLE alert_rules ADD COLUMN promql_expr VARCHAR(512);
ALTER TABLE alert_rules ADD COLUMN labels_json LONGTEXT;
ALTER TABLE alert_rules ADD COLUMN annotations_json LONGTEXT;

-- 扩展通知渠道表
ALTER TABLE alert_notification_channels ADD COLUMN config_json LONGTEXT;
ALTER TABLE alert_notification_channels ADD COLUMN provider VARCHAR(32) DEFAULT 'builtin';

-- 删除不再需要的表
-- DROP TABLE metric_points;
-- DROP TABLE alert_rule_evaluations;
```

### 保留内容

- `alert_rules` 表 - UI 管理告警规则
- `alerts` 表 - 告警事件历史记录
- `alert_notification_channels` 表 - 通知渠道配置
- 现有前端页面结构

## Capabilities

- `prometheus-integration` - Prometheus 时序存储集成
- `alerting-system-enhancement` - 告警系统增强（Alertmanager 对接）
- `notification-gateway` - 通知网关（多渠道支持）

## Impact

### Backend

| 路径 | 变更类型 | 说明 |
|------|---------|------|
| `internal/service/monitoring/logic.go` | 重构 | 删除 Collector，改用 Prometheus API |
| `internal/service/monitoring/rule_sync.go` | 新增 | 规则同步服务 |
| `internal/infra/prometheus/` | 新增 | Prometheus 客户端 |
| `internal/service/monitoring/notification_gateway.go` | 新增 | 通知网关 |
| `internal/model/monitoring.go` | 修改 | 表结构扩展 |

### Frontend

| 路径 | 变更类型 | 说明 |
|------|---------|------|
| `web/src/api/modules/monitoring.ts` | 修改 | 适配新 API 响应格式 |

### Infrastructure

| 路径 | 变更类型 | 说明 |
|------|---------|------|
| `deploy/compose/prometheus/` | 修改 | 更新配置，添加 alerting |
| `deploy/compose/alertmanager/` | 新增 | Alertmanager 部署配置 |

### Storage

| 文件 | 变更类型 | 说明 |
|------|---------|------|
| `storage/migrations/202603XX_prometheus_migration.sql` | 新增 | 表结构变更 |

## Risks

1. **迁移期间监控中断**
   - 缓解：并行运行两套系统，确认 Prometheus 正常后再切换

2. **PromQL 学习曲线**
   - 缓解：提供常用查询模板，逐步迁移

3. **规则同步延迟**
   - 缓解：规则变更时主动触发重载，添加版本校验

## Timeline

- Phase 1: Prometheus 客户端 + Dashboard 查询改造 (2-3 天)
- Phase 2: Alertmanager 部署 + 规则同步服务 (2-3 天)
- Phase 3: Notification Gateway + 通知渠道集成 (2-3 天)
- Phase 4: 清理旧代码 + 文档更新 (1-2 天)
