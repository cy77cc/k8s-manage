package monitor

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/cy77cc/k8s-manage/internal/ai/tools/core"
	"github.com/cy77cc/k8s-manage/internal/model"
)

func MonitorAlertRuleList(ctx context.Context, deps PlatformDeps, input MonitorAlertRuleListInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "monitor_alert_rule_list",
			Description: "查询告警规则列表。可选参数 status/keyword/limit。示例: {\"status\":\"enabled\"}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"limit": 50},
			SceneScope:  []string{"monitor", "deployment:metrics"},
		},
		input,
		func(in MonitorAlertRuleListInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			limit := in.Limit
			if limit <= 0 {
				limit = 50
			}
			if limit > 200 {
				limit = 200
			}
			query := deps.DB.Model(&model.AlertRule{})
			if status := strings.TrimSpace(in.Status); status != "" {
				query = query.Where("state = ? OR status = ?", status, status)
			}
			if kw := strings.TrimSpace(in.Keyword); kw != "" {
				pattern := "%" + kw + "%"
				query = query.Where("name LIKE ? OR metric LIKE ?", pattern, pattern)
			}
			var rules []model.AlertRule
			if err := query.Order("id desc").Limit(limit).Find(&rules).Error; err != nil {
				return nil, "db", err
			}
			return map[string]any{"total": len(rules), "list": rules}, "db", nil
		},
	)
}

func MonitorAlert(ctx context.Context, deps PlatformDeps, input MonitorAlertInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "monitor_alert",
			Description: "查询活跃告警列表。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"limit": 50},
			SceneScope:  []string{"monitor", "deployment:metrics"},
		},
		input,
		func(in MonitorAlertInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			limit := in.Limit
			if limit <= 0 {
				limit = 50
			}
			if limit > 200 {
				limit = 200
			}
			query := deps.DB.Model(&model.AlertEvent{}).Where("status = ?", "firing")
			if severity := strings.TrimSpace(in.Severity); severity != "" {
				query = query.Where("severity = ?", severity)
			}
			if in.ServiceID > 0 {
				query = query.Where("source LIKE ?", fmt.Sprintf("%%service:%d%%", in.ServiceID))
			}
			var alerts []model.AlertEvent
			if err := query.Order("triggered_at desc").Limit(limit).Find(&alerts).Error; err != nil {
				return nil, "db", err
			}
			return map[string]any{"total": len(alerts), "list": alerts}, "db", nil
		},
	)
}

func MonitorAlertActive(ctx context.Context, deps PlatformDeps, input MonitorAlertActiveInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "monitor_alert_active",
			Description: "查询活跃告警。可选 severity/service_id/limit。示例: {\"severity\":\"critical\"}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"limit": 50},
			SceneScope:  []string{"monitor", "deployment:metrics"},
		},
		input,
		func(in MonitorAlertActiveInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			limit := in.Limit
			if limit <= 0 {
				limit = 50
			}
			if limit > 200 {
				limit = 200
			}
			query := deps.DB.Model(&model.AlertEvent{}).Where("status = ?", "firing")
			if severity := strings.TrimSpace(in.Severity); severity != "" {
				query = query.Where("severity = ?", severity)
			}
			if in.ServiceID > 0 {
				query = query.Where("source LIKE ?", fmt.Sprintf("%%service:%d%%", in.ServiceID))
			}
			var alerts []model.AlertEvent
			if err := query.Order("triggered_at desc").Limit(limit).Find(&alerts).Error; err != nil {
				return nil, "db", err
			}
			return map[string]any{"total": len(alerts), "list": alerts}, "db", nil
		},
	)
}

func MonitorMetric(ctx context.Context, deps PlatformDeps, input MonitorMetricInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "monitor_metric",
			Description: "查询监控指标数据。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"query"},
			DefaultHint: map[string]any{"time_range": "1h", "step": 60},
			SceneScope:  []string{"monitor", "deployment:metrics"},
		},
		input,
		func(in MonitorMetricInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			queryName := strings.TrimSpace(in.Query)
			if queryName == "" {
				return nil, "validation", NewMissingParam("query", "query is required")
			}
			rangeDuration := parseTimeRange(strings.TrimSpace(in.TimeRange), time.Hour)
			step := in.Step
			if step <= 0 {
				step = 60
			}
			since := time.Now().Add(-rangeDuration)
			var points []model.MetricPoint
			if err := deps.DB.Where("metric = ? AND collected_at >= ?", queryName, since).
				Order("collected_at asc").
				Limit(2000).
				Find(&points).Error; err != nil {
				return nil, "db", err
			}
			return map[string]any{
				"query":      queryName,
				"time_range": rangeDuration.String(),
				"step":       step,
				"points":     points,
				"count":      len(points),
			}, "db", nil
		},
	)
}

func MonitorMetricQuery(ctx context.Context, deps PlatformDeps, input MonitorMetricQueryInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "monitor_metric_query",
			Description: "查询指标数据。query 必填，可选 time_range/step。示例: {\"query\":\"cpu_usage\",\"time_range\":\"1h\"}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"query"},
			DefaultHint: map[string]any{"time_range": "1h", "step": 60},
			SceneScope:  []string{"monitor", "deployment:metrics"},
		},
		input,
		func(in MonitorMetricQueryInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			queryName := strings.TrimSpace(in.Query)
			if queryName == "" {
				return nil, "validation", NewMissingParam("query", "query is required")
			}
			rangeDuration := parseTimeRange(strings.TrimSpace(in.TimeRange), time.Hour)
			step := in.Step
			if step <= 0 {
				step = 60
			}
			since := time.Now().Add(-rangeDuration)
			var points []model.MetricPoint
			if err := deps.DB.Where("metric = ? AND collected_at >= ?", queryName, since).
				Order("collected_at asc").
				Limit(2000).
				Find(&points).Error; err != nil {
				return nil, "db", err
			}
			return map[string]any{
				"query":      queryName,
				"time_range": rangeDuration.String(),
				"step":       step,
				"points":     points,
				"count":      len(points),
			}, "db", nil
		},
	)
}

func parseTimeRange(raw string, fallback time.Duration) time.Duration {
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}
