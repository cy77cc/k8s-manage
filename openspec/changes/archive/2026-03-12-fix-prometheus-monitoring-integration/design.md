# Design: Prometheus 监控集成修复

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         监控数据流架构                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   数据采集层                                                                 │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                                                                     │   │
│   │   ┌────────────────────────┐     ┌────────────────────────┐        │   │
│   │   │ HostService            │     │ ClusterCollector       │        │   │
│   │   │ (SSH 采集)              │     │ (K8s API 采集)          │        │   │
│   │   │                        │     │                        │        │   │
│   │   │ - CPU Load             │     │ - Node Status          │        │   │
│   │   │ - Memory Used/Total    │     │ - Pod Count            │        │   │
│   │   │ - Disk Usage           │     │ - Node Capacity        │        │   │
│   │   │ - Inode Usage          │     │                        │        │   │
│   │   └───────────┬────────────┘     └───────────┬────────────┘        │   │
│   │               │                              │                     │   │
│   │               │ 每 2 分钟                     │ 每 1 分钟            │   │
│   │               ▼                              ▼                     │   │
│   │   ┌─────────────────────────────────────────────────────────────┐  │   │
│   │   │                    MetricsPusher                            │  │   │
│   │   │  (internal/infra/prometheus/pusher.go)                      │  │   │
│   │   │                                                             │  │   │
│   │   │  - PushHostMetrics(hostID, hostName, snapshot)              │  │   │
│   │   │  - PushClusterMetrics(clusterID, clusterName, status)       │  │   │
│   │   └─────────────────────────┬───────────────────────────────────┘  │   │
│   │                             │                                       │   │
│   └─────────────────────────────┼───────────────────────────────────────┘   │
│                                 │                                           │
│                                 ▼                                           │
│   数据存储层                                                               │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                                                                     │   │
│   │   ┌───────────────────┐     ┌───────────────────┐                  │   │
│   │   │ Pushgateway       │     │ Prometheus        │                  │   │
│   │   │ :9091             │────▶│ :9090             │                  │   │
│   │   │                   │pull │                   │                  │   │
│   │   └───────────────────┘     │ - TSDB 存储        │                  │   │
│   │                             │ - PromQL 引擎      │                  │   │
│   │                             │ - 告警规则         │                  │   │
│   │                             └─────────┬─────────┘                  │   │
│   │                                       │                             │   │
│   └───────────────────────────────────────┼─────────────────────────────┘   │
│                                           │                                 │
│                                           ▼                                 │
│   数据查询层                                                               │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                                                                     │   │
│   │   ┌────────────────────────┐     ┌────────────────────────┐        │   │
│   │   │ dashboard/logic.go     │     │ ai/tools/monitor/      │        │   │
│   │   │                        │     │                        │        │   │
│   │   │ getMetricsSeries()     │     │ MonitorMetric()        │        │   │
│   │   │   └─ Prometheus API    │     │   └─ Prometheus API    │        │   │
│   │   └────────────────────────┘     └────────────────────────┘        │   │
│   │                                                                     │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Component Design

### 1. MetricsPusher

**路径**: `internal/infra/prometheus/pusher.go`

```go
package prometheus

import (
    "context"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/push"
)

// MetricsPusher 推送指标到 Pushgateway
type MetricsPusher struct {
    gatewayURL string
    registry   *prometheus.Registry

    // 主机指标
    hostCPULoad           *prometheus.GaugeVec
    hostMemoryUsedMB      *prometheus.GaugeVec
    hostMemoryTotalMB     *prometheus.GaugeVec
    hostMemoryUsage       *prometheus.GaugeVec
    hostDiskUsage         *prometheus.GaugeVec
    hostInodeUsage        *prometheus.GaugeVec
    hostHealthState       *prometheus.GaugeVec
    hostConnectivity      *prometheus.GaugeVec

    // 集群指标
    clusterNodeCount      *prometheus.GaugeVec
    clusterNodeReady      *prometheus.GaugeVec
    clusterNodeNotReady   *prometheus.GaugeVec
    clusterPodCount       *prometheus.GaugeVec
    clusterPodRunning     *prometheus.GaugeVec
    clusterPodPending     *prometheus.GaugeVec
    clusterPodFailed      *prometheus.GaugeVec
}

// HostMetricSnapshot 主机指标快照
type HostMetricSnapshot struct {
    HostID           uint64
    HostName         string
    HostIP           string
    CPULoad          float64
    MemoryUsedMB     int
    MemoryTotalMB    int
    DiskUsagePercent float64
    InodeUsagePercent float64
    HealthState      string  // healthy, degraded, critical, unknown
    ConnectivityStatus string
}

// ClusterMetricSnapshot 集群指标快照
type ClusterMetricSnapshot struct {
    ClusterID    uint64
    ClusterName  string
    NodeTotal    int
    NodeReady    int
    NodeNotReady int
    PodTotal     int
    PodRunning   int
    PodPending   int
    PodFailed    int
}

// PushHostMetrics 推送主机指标
func (p *MetricsPusher) PushHostMetrics(ctx context.Context, snapshot HostMetricSnapshot) error

// PushClusterMetrics 推送集群指标
func (p *MetricsPusher) PushClusterMetrics(ctx context.Context, snapshot ClusterMetricSnapshot) error
```

### 2. HostService 改造

**路径**: `internal/service/host/logic/host_service.go`

```go
// persistHealthSnapshot 改造
func (s *HostService) persistHealthSnapshot(ctx context.Context, snapshot *model.HostHealthSnapshot, node *model.Node) error {
    // 1. 保存到数据库 (保留用于历史审计)
    if err := s.svcCtx.DB.WithContext(ctx).Create(snapshot).Error; err != nil {
        return err
    }

    // 2. 更新节点状态
    updates := map[string]any{
        "health_state":  snapshot.State,
        "last_check_at": snapshot.CheckedAt,
    }
    if err := s.svcCtx.DB.WithContext(ctx).Model(&model.Node{}).Where("id = ?", node.ID).Updates(updates).Error; err != nil {
        return err
    }

    // 3. 推送到 Prometheus (新增)
    if s.svcCtx.MetricsPusher != nil {
        metricSnapshot := prominfra.HostMetricSnapshot{
            HostID:            uint64(node.ID),
            HostName:          node.Name,
            HostIP:            node.IP,
            CPULoad:           snapshot.CpuLoad,
            MemoryUsedMB:      snapshot.MemoryUsedMB,
            MemoryTotalMB:     snapshot.MemoryTotalMB,
            DiskUsagePercent:  snapshot.DiskUsagePct,
            InodeUsagePercent: snapshot.InodeUsedPct,
            HealthState:       snapshot.State,
            ConnectivityStatus: snapshot.ConnectivityStatus,
        }
        if err := s.svcCtx.MetricsPusher.PushHostMetrics(ctx, metricSnapshot); err != nil {
            // 推送失败不影响主流程，记录日志
            logger.L().Warn("failed to push host metrics", logger.Error(err), logger.Uint64("host_id", uint64(node.ID)))
        }
    }

    return nil
}
```

### 3. ClusterCollector (新增)

**路径**: `internal/service/cluster/collector.go`

```go
package cluster

import (
    "context"
    "sync"
    "time"

    prominfra "github.com/cy77cc/OpsPilot/internal/infra/prometheus"
    "github.com/cy77cc/OpsPilot/internal/logger"
    "github.com/cy77cc/OpsPilot/internal/model"
    "github.com/cy77cc/OpsPilot/internal/svc"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterCollector 集群状态采集器
type ClusterCollector struct {
    svcCtx        *svc.ServiceContext
    pusher        *prominfra.MetricsPusher
    collectorOnce sync.Once
}

// NewClusterCollector 创建采集器
func NewClusterCollector(svcCtx *svc.ServiceContext, pusher *prominfra.MetricsPusher) *ClusterCollector

// Start 启动定时采集
func (c *ClusterCollector) Start() {
    c.collectorOnce.Do(func() {
        go func() {
            ticker := time.NewTicker(1 * time.Minute)
            defer ticker.Stop()
            for {
                ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
                c.Collect(ctx)
                cancel()
                <-ticker.C
            }
        }()
    })
}

// Collect 执行一轮采集
func (c *ClusterCollector) Collect(ctx context.Context) {
    var clusters []model.Cluster
    if err := c.svcCtx.DB.WithContext(ctx).Find(&clusters).Error; err != nil {
        return
    }

    for _, cluster := range clusters {
        c.collectClusterMetrics(ctx, &cluster)
    }
}

// collectClusterMetrics 采集单个集群指标
func (c *ClusterCollector) collectClusterMetrics(ctx context.Context, cluster *model.Cluster) {
    cli, err := c.getK8sClient(cluster)
    if err != nil {
        return
    }

    // 采集节点状态
    nodes, err := cli.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        return
    }

    var readyCount, notReadyCount int
    for _, node := range nodes.Items {
        if isNodeReady(&node) {
            readyCount++
        } else {
            notReadyCount++
        }
    }

    // 采集 Pod 状态
    pods, err := cli.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
    if err != nil {
        return
    }

    var runningCount, pendingCount, failedCount int
    for _, pod := range pods.Items {
        switch pod.Status.Phase {
        case corev1.PodRunning:
            runningCount++
        case corev1.PodPending:
            pendingCount++
        case corev1.PodFailed:
            failedCount++
        }
    }

    // 推送指标
    snapshot := prominfra.ClusterMetricSnapshot{
        ClusterID:    uint64(cluster.ID),
        ClusterName:  cluster.Name,
        NodeTotal:    len(nodes.Items),
        NodeReady:    readyCount,
        NodeNotReady: notReadyCount,
        PodTotal:     len(pods.Items),
        PodRunning:   runningCount,
        PodPending:   pendingCount,
        PodFailed:    failedCount,
    }

    if err := c.pusher.PushClusterMetrics(ctx, snapshot); err != nil {
        logger.L().Warn("failed to push cluster metrics", logger.Error(err), logger.Uint64("cluster_id", uint64(cluster.ID)))
    }
}
```

### 4. Dashboard Logic 改造

**路径**: `internal/service/dashboard/logic.go`

```go
// getMetricsSeries 改造
func (l *Logic) getMetricsSeries(ctx context.Context, since, now time.Time) (dashboardv1.MetricsSeries, error) {
    if l.svcCtx.Prometheus == nil {
        // Prometheus 不可用时返回空数据
        return dashboardv1.MetricsSeries{}, nil
    }

    step := time.Duration(calculateStep(since, now)) * time.Second

    // 查询 CPU 使用率 (使用 CPU 负载)
    cpuSeries, err := l.queryHostMetrics(ctx, "host_cpu_load", since, now, step)
    if err != nil {
        return dashboardv1.MetricsSeries{}, err
    }

    // 查询内存使用率
    memorySeries, err := l.queryHostMetrics(ctx, "host_memory_usage_percent", since, now, step)
    if err != nil {
        return dashboardv1.MetricsSeries{}, err
    }

    return dashboardv1.MetricsSeries{
        CPUUsage:    cpuSeries,
        MemoryUsage: memorySeries,
    }, nil
}

// queryHostMetrics 从 Prometheus 查询主机指标
func (l *Logic) queryHostMetrics(ctx context.Context, metric string, start, end time.Time, step time.Duration) ([]dashboardv1.MetricSeries, error) {
    query := metric
    result, err := l.svcCtx.Prometheus.QueryRange(ctx, query, start, end, step)
    if err != nil {
        return nil, err
    }

    // 按 host_id 分组
    groups := make(map[uint64]*dashboardv1.MetricSeries)

    for _, series := range result.Matrix {
        hostID := parseHostID(series.Metric["host_id"])
        hostName := series.Metric["host_name"]

        if hostID == 0 {
            continue
        }

        if groups[hostID] == nil {
            groups[hostID] = &dashboardv1.MetricSeries{
                HostID:   hostID,
                HostName: hostName,
                Data:     []dashboardv1.MetricPoint{},
            }
        }

        for _, pair := range series.Values {
            if len(pair) >= 2 {
                groups[hostID].Data = append(groups[hostID].Data, dashboardv1.MetricPoint{
                    Timestamp: time.Unix(int64(pair[0].(float64)), 0),
                    Value:     pair[1].(float64),
                })
            }
        }
    }

    // 转换为切片
    out := make([]dashboardv1.MetricSeries, 0, len(groups))
    for _, series := range groups {
        out = append(out, *series)
    }

    // 按主机名排序
    sort.Slice(out, func(i, j int) bool {
        return out[i].HostName < out[j].HostName
    })

    return out, nil
}

// 删除 listMetricPointsGrouped 方法
```

### 5. AI Monitor Tools 改造

**路径**: `internal/ai/tools/monitor/tools.go`

```go
// MonitorMetric 改造
func MonitorMetric(ctx context.Context, deps common.PlatformDeps) tool.InvokableTool {
    t, err := einoutils.InferOptionableTool(
        "monitor_metric",
        "Query time-series metric data from Prometheus. ...",
        func(ctx context.Context, input *MonitorMetricInput, opts ...tool.Option) (*MonitorMetricOutput, error) {
            if deps.Prometheus == nil {
                return nil, fmt.Errorf("prometheus client unavailable")
            }

            queryName := strings.TrimSpace(input.Query)
            if queryName == "" {
                return nil, fmt.Errorf("query is required")
            }

            rangeDuration := parseTimeRange(strings.TrimSpace(input.TimeRange), time.Hour)
            step := input.Step
            if step <= 0 {
                step = 60
            }

            start := time.Now().Add(-rangeDuration)
            end := time.Now()

            result, err := deps.Prometheus.QueryRange(ctx, queryName, start, end, time.Duration(step)*time.Second)
            if err != nil {
                return nil, err
            }

            // 转换为输出格式
            points := make([]MetricPoint, 0, 2000)
            for _, series := range result.Matrix {
                for _, pair := range series.Values {
                    if len(pair) >= 2 {
                        points = append(points, MetricPoint{
                            Timestamp: time.Unix(int64(pair[0].(float64)), 0),
                            Value:     pair[1].(float64),
                            Labels:    series.Metric,
                        })
                    }
                }
            }

            return &MonitorMetricOutput{
                Query:     queryName,
                TimeRange: rangeDuration.String(),
                Step:      step,
                Points:    points,
                Count:     len(points),
            }, nil
        },
    )
    // ...
}

// MetricPoint 输出结构 (不再依赖 model.MetricPoint)
type MetricPoint struct {
    Timestamp time.Time         `json:"timestamp"`
    Value     float64           `json:"value"`
    Labels    map[string]string `json:"labels,omitempty"`
}
```

## Infrastructure Changes

### docker-compose.yml

```yaml
services:
  prometheus:
    image: prom/prometheus:v2.55.1
    # ... existing config ...

  pushgateway:
    image: prom/pushgateway:v1.9.0
    container_name: pushgateway
    restart: unless-stopped
    ports:
      - "9091:9091"
    networks:
      - monitoring
```

### prometheus.yml

```yaml
scrape_configs:
  # ... existing configs ...

  - job_name: 'pushgateway'
    honor_labels: true  # 保留推送时的标签
    static_configs:
      - targets: ['pushgateway:9091']
```

## Configuration

### config.yaml 新增

```yaml
prometheus:
  enable: true
  address: ${PROMETHEUS_ADDRESS}
  host: ${PROMETHEUS_HOST}
  port: ${PROMETHEUS_PORT}
  pushgateway_url: ${PROMETHEUS_PUSHGATEWAY_URL}  # 新增
  timeout: 10s
  max_concurrent: 10
  retry_count: 3
```

## Migration

无需数据库迁移，仅代码改造。
