# Spec: Prometheus 指标管道

## Overview

定义主机和集群指标的数据格式、采集频率、Prometheus 查询接口。

## Metrics Specification

### 主机指标 (Host Metrics)

所有主机指标包含标签：`host_id`, `host_name`, `host_ip`

| 指标名称 | 类型 | 单位 | 描述 |
|----------|------|------|------|
| `host_cpu_load` | Gauge | 数值 | 系统 1 分钟负载 |
| `host_memory_used_mb` | Gauge | MB | 已用内存 |
| `host_memory_total_mb` | Gauge | MB | 总内存 |
| `host_memory_usage_percent` | Gauge | 百分比 | 内存使用率 (0-100) |
| `host_disk_usage_percent` | Gauge | 百分比 | 磁盘使用率 (0-100) |
| `host_inode_usage_percent` | Gauge | 百分比 | Inode 使用率 (0-100) |
| `host_health_state` | Gauge | 状态码 | 健康状态: 0=healthy, 1=degraded, 2=critical, 3=unknown |
| `host_connectivity_status` | Gauge | 状态码 | 连接状态: 0=healthy, 1=degraded, 2=critical, 3=unknown |

### 集群指标 (Cluster Metrics)

所有集群指标包含标签：`cluster_id`, `cluster_name`

| 指标名称 | 类型 | 单位 | 描述 |
|----------|------|------|------|
| `k8s_cluster_node_count` | Gauge | 数量 | 集群节点总数 |
| `k8s_cluster_node_ready` | Gauge | 数量 | Ready 状态节点数 |
| `k8s_cluster_node_not_ready` | Gauge | 数量 | NotReady 状态节点数 |
| `k8s_cluster_pod_count` | Gauge | 数量 | Pod 总数 |
| `k8s_cluster_pod_running` | Gauge | 数量 | Running 状态 Pod 数 |
| `k8s_cluster_pod_pending` | Gauge | 数量 | Pending 状态 Pod 数 |
| `k8s_cluster_pod_failed` | Gauge | 数量 | Failed 状态 Pod 数 |
| `k8s_node_status` | Gauge | 状态码 | 节点状态: 0=Ready, 1=NotReady |
| `k8s_node_cpu_capacity` | Gauge | 核数 | 节点 CPU 容量 |
| `k8s_node_memory_capacity_mb` | Gauge | MB | 节点内存容量 |

标签：`k8s_node_status` 额外包含 `node_name`

## Collection Frequency

| 采集类型 | 频率 | 超时 |
|----------|------|------|
| 主机健康检查 | 2 分钟 | 18 秒/主机 |
| 集群状态采集 | 1 分钟 | 30 秒/集群 |

## Prometheus Query Examples

### 主机指标查询

```promql
# 查询所有主机 CPU 负载
host_cpu_load

# 查询特定主机内存使用率
host_memory_usage_percent{host_id="123"}

# 查询过去 1 小时的内存使用趋势
host_memory_usage_percent[1h]

# 平均内存使用率
avg(host_memory_usage_percent)

# 不健康主机数量
count(host_health_state > 0)
```

### 集群指标查询

```promql
# 查询集群 NotReady 节点数
k8s_cluster_node_not_ready{cluster_id="1"}

# 查询各命名空间 Pod 数量（需要额外采集）
# 暂不支持，后续扩展

# 集群健康状态
k8s_cluster_node_not_ready == 0
```

## API Changes

### Dashboard API

`GET /api/v1/dashboard/overview?time_range=1h`

响应变化：
- 移除从 MySQL `metric_points` 查询
- 改用 Prometheus `query_range` API

```json
{
  "hosts": { "total": 10, "healthy": 8, "degraded": 2 },
  "clusters": { "total": 3, "healthy": 3 },
  "metrics": {
    "cpu_usage": [
      { "host_id": 1, "host_name": "node-01", "data": [...] }
    ],
    "memory_usage": [...]
  }
}
```

### AI Tools

`monitor_metric` 工具：
- 移除 MySQL 查询
- 改用 Prometheus 查询

## Data Retention

- Prometheus: 30 天（已有配置）
- MySQL `host_health_snapshots`: 保留用于历史审计
- MySQL `metric_points`: 已删除

## Dependencies

- Prometheus Pushgateway
- Prometheus Client (已实现)
- K8s Clientset (已实现)
