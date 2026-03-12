# Proposal: 修复 Prometheus 监控集成

## Summary

修复半完成的 Prometheus 迁移，实现主机和集群状态数据采集推送到 Prometheus，修复 Dashboard 读取时报错 `metric_points` 表不存在的问题。

## Motivation

### 问题背景

之前的 Prometheus 迁移只完成了一半：
- ✅ Prometheus Client 已实现
- ✅ monitoring/logic.go 已改造
- ✅ 数据库迁移删除了 `metric_points` 表
- ❌ dashboard/logic.go 仍查询已删除的 `metric_points` 表
- ❌ ai/tools/monitor/tools.go 仍查询已删除的 `metric_points` 表
- ❌ 主机健康采集数据未写入 Prometheus
- ❌ 集群状态数据未写入 Prometheus

### 影响范围

- 主控台 Dashboard 加载时报错 `Error 1146: Table 'devops.metric_points' doesn't exist`
- AI 工具无法查询监控指标
- 实时监控图表无数据

## Goals

1. 部署 Prometheus Pushgateway 作为指标推送入口
2. 主机健康采集完成后推送指标到 Prometheus
3. 集群状态定时采集并推送指标到 Prometheus
4. Dashboard 和 AI Tools 改用 Prometheus API 查询
5. 清理过时的 MetricPoint 相关代码

## Non-Goals

- 部署 node_exporter agent 方案（后续完善）
- 长期存储方案（VictoriaMetrics/Thanos）
- Alertmanager 集成（已有基础实现）

## Scope

### In Scope

| 模块 | 改造内容 |
|------|----------|
| 基础设施 | 部署 Pushgateway，更新 Prometheus 配置 |
| 数据写入 | HostService 采集后推送主机指标到 Prometheus |
| 数据写入 | 定时采集 K8s 节点/Pod 状态并推送集群指标 |
| 数据读取 | dashboard/logic.go 改用 Prometheus API |
| 数据读取 | ai/tools/monitor/tools.go 改用 Prometheus API |
| 代码清理 | 删除 MetricPoint 结构体和 queryMetricsFromDB 方法 |

### Out of Scope

- 前端 UI 改造
- 新增告警规则
- 其他监控数据源集成

## Success Criteria

- [ ] 主控台 Dashboard 正常加载，无数据库表不存在错误
- [ ] 主机 CPU/内存/磁盘指标可在 Prometheus 中查询
- [ ] 集群节点/Pod 状态指标可在 Prometheus 中查询
- [ ] AI 工具可正常查询监控指标
- [ ] 所有相关测试通过

## Risks

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Pushgateway 单点 | 中 | 后续可部署多副本 |
| 指标格式变更 | 低 | 保持与现有告警规则兼容 |
| 采集频率过高 | 中 | 合理设置采集间隔 |

## Timeline

预计 1-2 个工作日完成。
