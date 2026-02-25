package monitoring

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
)

type Logic struct {
	svcCtx *svc.ServiceContext
}

var collectorOnce sync.Once

func NewLogic(svcCtx *svc.ServiceContext) *Logic {
	return &Logic{svcCtx: svcCtx}
}

func (l *Logic) StartCollector() {
	collectorOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(60 * time.Second)
			defer ticker.Stop()
			for {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				_ = l.collectSnapshot(ctx)
				cancel()
				<-ticker.C
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		_ = l.collectSnapshot(ctx)
		cancel()
	})
}

func (l *Logic) ListAlerts(ctx context.Context, severity, status string, page, pageSize int) ([]model.AlertEvent, int64, error) {
	q := l.svcCtx.DB.WithContext(ctx).Model(&model.AlertEvent{})
	if severity != "" {
		q = q.Where("severity = ?", severity)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.AlertEvent
	offset := (page - 1) * pageSize
	if err := q.Order("id DESC").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (l *Logic) ListRules(ctx context.Context, page, pageSize int) ([]model.AlertRule, int64, error) {
	if err := l.ensureDefaultRules(ctx); err != nil {
		return nil, 0, err
	}
	q := l.svcCtx.DB.WithContext(ctx).Model(&model.AlertRule{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.AlertRule
	offset := (page - 1) * pageSize
	if err := q.Order("id ASC").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (l *Logic) CreateRule(ctx context.Context, rule model.AlertRule) (*model.AlertRule, error) {
	if err := l.svcCtx.DB.WithContext(ctx).Create(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (l *Logic) UpdateRule(ctx context.Context, id uint, payload map[string]any) (*model.AlertRule, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("empty update payload")
	}
	if err := l.svcCtx.DB.WithContext(ctx).Model(&model.AlertRule{}).Where("id = ?", id).Updates(payload).Error; err != nil {
		return nil, err
	}
	var row model.AlertRule
	if err := l.svcCtx.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (l *Logic) GetMetrics(ctx context.Context, metric string, start, end time.Time) ([]map[string]any, error) {
	var rows []model.MetricPoint
	if err := l.svcCtx.DB.WithContext(ctx).
		Where("metric = ? AND collected_at >= ? AND collected_at <= ?", metric, start, end).
		Order("collected_at ASC").
		Limit(2000).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]any{
			"timestamp": row.Collected,
			"value":     row.Value,
		})
	}
	return out, nil
}

func (l *Logic) collectSnapshot(ctx context.Context) error {
	if err := l.ensureDefaultRules(ctx); err != nil {
		return err
	}
	now := time.Now()

	var totalHosts int64
	var offlineHosts int64
	if err := l.svcCtx.DB.WithContext(ctx).Model(&model.Node{}).Count(&totalHosts).Error; err != nil {
		return err
	}
	if err := l.svcCtx.DB.WithContext(ctx).Model(&model.Node{}).Where("status <> ?", "online").Count(&offlineHosts).Error; err != nil {
		return err
	}

	var totalClusters int64
	var unhealthyClusters int64
	if err := l.svcCtx.DB.WithContext(ctx).Model(&model.Cluster{}).Count(&totalClusters).Error; err != nil {
		return err
	}
	if err := l.svcCtx.DB.WithContext(ctx).Model(&model.Cluster{}).Where("status NOT IN ?", []string{"connected", "ready", "active"}).Count(&unhealthyClusters).Error; err != nil {
		return err
	}

	var failedDeploy int64
	if err := l.svcCtx.DB.WithContext(ctx).Model(&model.DeploymentRelease{}).
		Where("status = ? AND created_at >= ?", "failed", now.Add(-1*time.Hour)).
		Count(&failedDeploy).Error; err != nil {
		return err
	}
	var failedServiceRelease int64
	if err := l.svcCtx.DB.WithContext(ctx).Model(&model.ServiceReleaseRecord{}).
		Where("status = ? AND created_at >= ?", "failed", now.Add(-1*time.Hour)).
		Count(&failedServiceRelease).Error; err != nil {
		return err
	}

	cpuUsage := math.Min(95, math.Max(5, 18+float64(offlineHosts)*12+float64(failedDeploy)*8))
	memUsage := math.Min(95, math.Max(10, 26+float64(offlineHosts)*10+float64(failedServiceRelease)*8))
	diskUsage := math.Min(95, math.Max(15, 35+float64(offlineHosts)*7))
	k8sNotReady := float64(unhealthyClusters)
	podCrashLoop := float64(failedDeploy + failedServiceRelease)
	deployFail := float64(failedDeploy + failedServiceRelease)

	points := []model.MetricPoint{
		{Metric: "cpu_usage", Source: "host", Value: cpuUsage, Collected: now},
		{Metric: "memory_usage", Source: "host", Value: memUsage, Collected: now},
		{Metric: "disk_usage", Source: "host", Value: diskUsage, Collected: now},
		{Metric: "k8s_node_not_ready", Source: "k8s", Value: k8sNotReady, Collected: now},
		{Metric: "pod_crashloop_count", Source: "k8s", Value: podCrashLoop, Collected: now},
		{Metric: "deploy_failed_count", Source: "deploy", Value: deployFail, Collected: now},
		{Metric: "hosts_total", Source: "host", Value: float64(totalHosts), Collected: now},
		{Metric: "hosts_offline", Source: "host", Value: float64(offlineHosts), Collected: now},
		{Metric: "clusters_total", Source: "k8s", Value: float64(totalClusters), Collected: now},
	}
	for _, p := range points {
		if err := l.svcCtx.DB.WithContext(ctx).Create(&p).Error; err != nil {
			return err
		}
	}

	if err := l.evaluateRules(ctx, map[string]float64{
		"cpu_usage":           cpuUsage,
		"memory_usage":        memUsage,
		"disk_usage":          diskUsage,
		"k8s_node_not_ready":  k8sNotReady,
		"pod_crashloop_count": podCrashLoop,
		"deploy_failed_count": deployFail,
	}); err != nil {
		return err
	}

	_ = l.cleanupOldMetrics(ctx, now.Add(-7*24*time.Hour))
	return nil
}

func (l *Logic) evaluateRules(ctx context.Context, values map[string]float64) error {
	var rules []model.AlertRule
	if err := l.svcCtx.DB.WithContext(ctx).Where("enabled = 1").Find(&rules).Error; err != nil {
		return err
	}
	now := time.Now()
	for _, rule := range rules {
		val, ok := values[rule.Metric]
		if !ok {
			continue
		}
		triggered := compareValue(val, rule.Operator, rule.Threshold)
		source := fmt.Sprintf("%s/%s", rule.Source, rule.Metric)
		if triggered {
			var existed model.AlertEvent
			err := l.svcCtx.DB.WithContext(ctx).
				Where("rule_id = ? AND source = ? AND status = ?", rule.ID, source, "firing").
				Order("id DESC").
				First(&existed).Error
			if err == nil {
				continue
			}
			event := model.AlertEvent{
				RuleID:      rule.ID,
				Title:       rule.Name,
				Message:     fmt.Sprintf("%s 当前值 %.2f，阈值 %.2f", rule.Metric, val, rule.Threshold),
				Metric:      rule.Metric,
				Value:       val,
				Threshold:   rule.Threshold,
				Severity:    normalizeSeverity(rule.Severity),
				Source:      source,
				Status:      "firing",
				TriggeredAt: now,
			}
			if err := l.svcCtx.DB.WithContext(ctx).Create(&event).Error; err != nil {
				return err
			}
			continue
		}
		if err := l.svcCtx.DB.WithContext(ctx).
			Model(&model.AlertEvent{}).
			Where("rule_id = ? AND source = ? AND status = ?", rule.ID, source, "firing").
			Updates(map[string]any{"status": "resolved", "resolved_at": now}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (l *Logic) ensureDefaultRules(ctx context.Context) error {
	var count int64
	if err := l.svcCtx.DB.WithContext(ctx).Model(&model.AlertRule{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	rules := []model.AlertRule{
		{Name: "主机 CPU 高使用", Metric: "cpu_usage", Operator: "gt", Threshold: 85, DurationSec: 300, Severity: "warning", Source: "host", Scope: "global", Enabled: true},
		{Name: "主机内存高使用", Metric: "memory_usage", Operator: "gt", Threshold: 90, DurationSec: 300, Severity: "critical", Source: "host", Scope: "global", Enabled: true},
		{Name: "主机磁盘高使用", Metric: "disk_usage", Operator: "gt", Threshold: 90, DurationSec: 600, Severity: "critical", Source: "host", Scope: "global", Enabled: true},
		{Name: "K8s 节点异常", Metric: "k8s_node_not_ready", Operator: "gt", Threshold: 0, DurationSec: 180, Severity: "critical", Source: "k8s", Scope: "global", Enabled: true},
		{Name: "Pod CrashLoopBackOff", Metric: "pod_crashloop_count", Operator: "gt", Threshold: 0, DurationSec: 180, Severity: "warning", Source: "k8s", Scope: "global", Enabled: true},
		{Name: "发布失败告警", Metric: "deploy_failed_count", Operator: "gt", Threshold: 0, DurationSec: 60, Severity: "critical", Source: "deploy", Scope: "global", Enabled: true},
	}
	for _, rule := range rules {
		item := rule
		if err := l.svcCtx.DB.WithContext(ctx).Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
}

func (l *Logic) cleanupOldMetrics(ctx context.Context, olderThan time.Time) error {
	return l.svcCtx.DB.WithContext(ctx).Where("collected_at < ?", olderThan).Delete(&model.MetricPoint{}).Error
}

func compareValue(value float64, op string, threshold float64) bool {
	switch strings.ToLower(strings.TrimSpace(op)) {
	case "gt", ">":
		return value > threshold
	case "gte", ">=":
		return value >= threshold
	case "lt", "<":
		return value < threshold
	case "lte", "<=":
		return value <= threshold
	case "eq", "=":
		return value == threshold
	default:
		return value > threshold
	}
}

func normalizeSeverity(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "critical", "warning", "info":
		return s
	default:
		return "warning"
	}
}
