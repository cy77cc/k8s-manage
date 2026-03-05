# Tasks: Prometheus 监控系统集成

## Phase 1: 基础设施部署

### 1.1 Alertmanager 部署

- [x] 1.1.1 创建 `deploy/compose/alertmanager/` 目录
- [x] 1.1.2 编写 `docker-compose.yml`（Alertmanager 服务）
- [x] 1.1.3 编写 `alertmanager.yml`（基础配置 + webhook receiver）

### 1.2 Prometheus 配置更新

- [x] 1.2.1 更新 `prometheus.yml` 添加 alerting 配置
- [x] 1.2.2 创建 `alerting_rules.yml` 空规则文件（占位）
- [x] 1.2.3 配置数据保留时间为 30 天（`--storage.tsdb.retention.time=30d`）

### 1.3 网络连通性

- [x] 1.3.1 确保 Prometheus 可访问 Alertmanager
- [x] 1.3.2 确保 Alertmanager 可访问 k8s-manage webhook

## Phase 2: Prometheus 客户端实现

### 2.1 客户端基础

- [x] 2.1.1 创建 `internal/infra/prometheus/` 目录
- [x] 2.1.2 实现 `config.go` 配置结构
- [x] 2.1.3 实现 `client.go` HTTP 客户端
  - Query (即时查询)
  - QueryRange (范围查询)
  - Metadata (元数据查询)
- [x] 2.1.4 实现 `response.go` 响应解析

### 2.2 PromQL 构建器

- [x] 2.2.1 实现 `query.go` QueryBuilder
  - WithLabel (标签过滤)
  - WithRange (时间范围)
  - WithAggregation (聚合函数)
  - Build() 生成 PromQL

### 2.3 服务集成

- [x] 2.3.1 在 `svc.ServiceContext` 添加 PrometheusClient
- [x] 2.3.2 添加配置项到 `configs/config.yaml`
- [x] 2.3.3 编写单元测试

## Phase 3: Dashboard API 改造

### 3.1 查询接口改造

- [x] 3.1.1 改造 `GetMetrics` 方法使用 Prometheus 客户端
- [x] 3.1.2 实现指标格式转换（Prometheus → 业务格式）
- [x] 3.1.3 添加错误处理和降级逻辑

### 3.2 新增查询能力

- [x] 3.2.1 实现 `GetMetricAggregation` 聚合查询
- [x] 3.2.2 实现 `GetMetricMetadata` 元数据查询
- [x] 3.2.3 添加查询缓存（Redis）

### 3.3 清理旧代码

- [x] 3.3.1 删除 `collectSnapshot` 方法
- [x] 3.3.2 删除 `cleanupOldMetrics` 方法
- [x] 3.3.3 删除 `collectPerHostMetrics` 方法
- [x] 3.3.4 删除 `generateSimulatedHostMetrics` 方法

## Phase 4: 规则同步服务

### 4.1 同步服务实现

- [x] 4.1.1 创建 `internal/service/monitoring/rule_sync.go`
- [x] 4.1.2 实现 `SyncRules` 方法
  - 查询 enabled=true 的规则
  - 转换为 Prometheus 格式
  - 写入 alerting_rules.yml
  - 调用 reload API

### 4.2 规则格式转换

- [x] 4.2.1 实现 `convertRuleToPrometheus` 转换函数
  - metric + operator + threshold → expr
  - duration_sec → for
  - severity → labels
  - name → annotations

### 4.3 触发机制

- [x] 4.3.1 服务启动时触发同步
- [x] 4.3.2 规则 CRUD 后触发同步
- [x] 4.3.3 实现定时校验（每 5 分钟）

### 4.4 API 端点

- [x] 4.4.1 添加 `POST /api/v1/alerts/rules/sync` 手动同步接口
- [x] 4.4.2 在规则 CRUD handler 中调用同步

## Phase 5: Notification Gateway

### 5.1 Webhook 接收

- [x] 5.1.1 创建 `internal/service/monitoring/notification_gateway.go`
- [x] 5.1.2 定义 Alertmanager webhook 数据结构
- [x] 5.1.3 实现 `HandleWebhook` 方法
  - 解析 payload
  - 幂等性检查（fingerprint）
  - 持久化到 alerts 表

### 5.2 Provider 接口设计

- [x] 5.2.1 创建 `internal/service/notification/provider.go`
- [x] 5.2.2 定义 `Provider` 接口
- [x] 5.2.3 实现 `ProviderRegistry` 注册表

### 5.3 内置 Provider 实现

- [x] 5.3.1 实现 `LogProvider`（默认，记录日志）
- [x] 5.3.2 实现 `DingTalkProvider`（钉钉机器人）
- [x] 5.3.3 实现 `WeComProvider`（企业微信）
- [x] 5.3.4 实现 `EmailProvider`（邮件）
- [x] 5.3.5 实现 `SMSProvider`（短信，接口预留）

### 5.4 通知分发

- [x] 5.4.1 实现通知分发逻辑
- [x] 5.4.2 添加异步发送支持
- [x] 5.4.3 添加重试机制

### 5.5 API 端点

- [x] 5.5.1 添加 `POST /api/v1/alerts/receiver` webhook 接收接口
- [x] 5.5.2 更新路由配置

## Phase 6: 数据库迁移

### 6.1 迁移脚本

- [x] 6.1.1 创建 `storage/migrations/202603XX_prometheus_migration.sql`
- [x] 6.1.2 添加 alert_rules 表字段扩展
- [x] 6.1.3 添加 notification_channels 表字段扩展
- [x] 6.1.4 创建 alert_silences 表
- [x] 6.1.5 编写 Down 回滚脚本

### 6.2 模型更新

- [x] 6.2.1 更新 `internal/model/monitoring.go`
  - AlertRule 添加新字段
  - AlertNotificationChannel 添加新字段
  - 新增 AlertSilence 模型

## Phase 7: 前端适配

### 7.1 API 适配

- [x] 7.1.1 更新 `web/src/api/modules/monitoring.ts`
- [x] 7.1.2 适配新的响应格式

### 7.2 UI 调整

- [x] 7.2.1 规则管理页面添加 PromQL 预览
- [x] 7.2.2 通知渠道配置页面更新

## Phase 8: 测试与验证

### 8.1 单元测试

- [x] 8.1.1 Prometheus Client 测试
- [x] 8.1.2 QueryBuilder 测试
- [x] 8.1.3 Rule Sync 测试
- [x] 8.1.4 Notification Gateway 测试
- [x] 8.1.5 Provider 测试

### 8.2 集成测试

- [x] 8.2.1 指标查询端到端测试
- [x] 8.2.2 告警触发到通知测试
- [x] 8.2.3 规则同步测试

### 8.3 性能验证

- [x] 8.3.1 Dashboard 查询响应 < 50ms (P95)
- [x] 8.3.2 告警通知延迟 < 30s
- [x] 8.3.3 规则同步延迟 < 5s

## Phase 9: 清理与文档

### 9.1 代码清理

- [x] 9.1.1 删除 metric_points 表
- [x] 9.1.2 删除 alert_rule_evaluations 表
- [x] 9.1.3 清理无用代码

### 9.2 文档更新

- [x] 9.2.1 更新 README 部署说明
- [x] 9.2.2 更新 API 文档
- [x] 9.2.3 编写告警配置指南

## Phase 10: OpenSpec 同步

- [x] 10.1 创建 `openspec/specs/prometheus-integration/spec.md`
- [x] 10.2 创建 `openspec/specs/alerting-system-enhancement/spec.md`
- [x] 10.3 创建 `openspec/specs/notification-gateway/spec.md`
- [x] 10.4 运行 `openspec validate --json`

---

## 验证命令

```bash
# 后端测试
go test ./internal/infra/prometheus/... -v
go test ./internal/service/monitoring/... -v
go test ./internal/service/notification/... -v

# 前端构建
cd web && npm run build

# 集成测试
go test ./... -tags=integration -v

# OpenSpec 验证
openspec validate --json
```
