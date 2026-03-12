# Tasks: Prometheus 监控集成修复

## Phase 1: 基础设施 (P0)

- [x] **TASK-001**: 部署 Prometheus Pushgateway
  - 更新 `deploy/compose/prometheus/docker-compose.yml`
  - 添加 pushgateway 服务
  - 更新 `deploy/compose/prometheus/prometheus.yml`
  - 添加 pushgateway scrape 配置
  - 验证 Pushgateway 可访问

- [x] **TASK-002**: 更新配置结构
  - 更新 `internal/config/config.go`
  - 添加 `PushgatewayURL` 字段
  - 更新 `configs/config.yaml` 示例

## Phase 2: 数据写入改造 (P0)

- [x] **TASK-003**: 实现 MetricsPusher
  - 创建 `internal/infra/prometheus/pusher.go`
  - 定义主机指标 Gauge
  - 定义集群指标 Gauge
  - 实现 PushHostMetrics 方法
  - 实现 PushClusterMetrics 方法
  - 编写单元测试

- [x] **TASK-004**: 更新 ServiceContext
  - 更新 `internal/svc/svc.go`
  - 初始化 MetricsPusher
  - 添加到 ServiceContext

- [x] **TASK-005**: 改造 HostService
  - 更新 `internal/service/host/logic/host_service.go`
  - 修改 persistHealthSnapshot 方法
  - 添加 Prometheus 推送逻辑
  - 确保推送失败不影响主流程

- [x] **TASK-006**: 实现 ClusterCollector
  - 创建 `internal/service/cluster/collector.go`
  - 实现定时采集逻辑
  - 采集 K8s 节点状态
  - 采集 Pod 状态
  - 推送指标到 Prometheus
  - 在服务启动时初始化

## Phase 3: 数据读取改造 (P0)

- [x] **TASK-007**: 改造 Dashboard Logic
  - 更新 `internal/service/dashboard/logic.go`
  - 重写 getMetricsSeries 方法
  - 删除 listMetricPointsGrouped 方法
  - 使用 Prometheus QueryRange API
  - 处理 Prometheus 不可用的情况
  - 更新单元测试

- [x] **TASK-008**: 改造 AI Monitor Tools
  - 更新 `internal/ai/tools/monitor/tools.go`
  - 重写 MonitorMetric 方法
  - 重写 MonitorMetricQuery 方法
  - 删除 model.MetricPoint 依赖
  - 定义本地 MetricPoint 结构
  - 使用 deps.Prometheus 查询
  - 更新 common.PlatformDeps 添加 Prometheus 字段

- [x] **TASK-009**: 更新 PlatformDeps
  - 更新 `internal/ai/tools/common/common.go`
  - 添加 Prometheus 字段到 PlatformDeps
  - 确保初始化时注入 Prometheus 客户端

## Phase 4: 代码清理 (P1)

- [x] **TASK-010**: 清理废弃代码
  - 删除 `internal/model/monitoring.go` 中的 MetricPoint 结构体
  - 删除 `internal/service/monitoring/logic.go` 中的 queryMetricsFromDB 方法
  - 删除 StartCollector 方法（已标记废弃）
  - 检查并删除其他对 MetricPoint 的引用

- [x] **TASK-011**: 更新测试
  - 更新 dashboard/logic_test.go
  - 更新 monitoring/logic_test.go
  - 更新 ai/tools/monitor 相关测试
  - Mock Prometheus 客户端

## Phase 5: 验证与文档 (P2)

- [x] **TASK-012**: 端到端验证
  - 启动完整服务栈
  - 验证主机指标推送到 Prometheus
  - 验证集群指标推送到 Prometheus
  - 验证 Dashboard 正常加载
  - 验证 AI 工具查询正常

- [x] **TASK-013**: 更新告警规则
  - 更新 `deploy/compose/prometheus/alerting_rules.yml`
  - 添加主机指标告警规则
  - 添加集群指标告警规则

## Dependencies

```
TASK-001 ──┬── TASK-002 ──┬── TASK-003 ──┬── TASK-004
           │              │              │
           │              └── TASK-005 ──┤
           │                             │
           └── TASK-006 ─────────────────┤
                                         │
TASK-007 ────────────────────────────────┤
TASK-008 ─── TASK-009 ───────────────────┤
                                         │
TASK-010 ────────────────────────────────┤
TASK-011 ────────────────────────────────┤
                                         │
TASK-012 ────────────────────────────────┘
TASK-013
```

## Estimated Effort

| Phase | 任务数 | 预计时间 |
|-------|--------|----------|
| Phase 1 | 2 | 0.5 天 |
| Phase 2 | 4 | 0.5 天 |
| Phase 3 | 3 | 0.5 天 |
| Phase 4 | 2 | 0.25 天 |
| Phase 5 | 2 | 0.25 天 |
| **总计** | **13** | **2 天** |
