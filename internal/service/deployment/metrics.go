package deployment

import (
	"context"
	"time"

	"github.com/cy77cc/OpsPilot/internal/httpx"
	"github.com/cy77cc/OpsPilot/internal/model"
	"github.com/cy77cc/OpsPilot/internal/svc"
	"github.com/gin-gonic/gin"
)

type MetricsHandler struct {
	svcCtx *svc.ServiceContext
}

func NewMetricsHandler(svcCtx *svc.ServiceContext) *MetricsHandler {
	return &MetricsHandler{svcCtx: svcCtx}
}

// MetricsSummary 指标汇总
type MetricsSummary struct {
	TotalReleases   int64                 `json:"total_releases"`
	SuccessRate     float64               `json:"success_rate"`
	FailureRate     float64               `json:"failure_rate"`
	AvgDurationSecs float64               `json:"avg_duration_seconds"`
	ByEnvironment   map[string]EnvMetrics `json:"by_environment"`
	ByStatus        map[string]int64      `json:"by_status"`
	RecentFailures  int64                 `json:"recent_failures"`
	RecentReleases  int64                 `json:"recent_releases"`
}

// EnvMetrics 环境指标
type EnvMetrics struct {
	Total       int64   `json:"total"`
	SuccessRate float64 `json:"success_rate"`
}

// MetricsTrend 指标趋势
type MetricsTrend struct {
	Date            string  `json:"date"`
	DeploymentCount int     `json:"deployment_count"`
	SuccessCount    int     `json:"success_count"`
	FailureCount    int     `json:"failure_count"`
	SuccessRate     float64 `json:"success_rate"`
}

// GetMetricsSummary 获取指标汇总
func (h *MetricsHandler) GetMetricsSummary(c *gin.Context) {
	ctx := c.Request.Context()

	summary, err := h.getMetricsSummary(ctx)
	if err != nil {
		httpx.BindErr(c, err)
		return
	}

	httpx.OK(c, summary)
}

// GetMetricsTrends 获取指标趋势
func (h *MetricsHandler) GetMetricsTrends(c *gin.Context) {
	ctx := c.Request.Context()
	timeRange := c.DefaultQuery("range", "daily")

	trends, err := h.getMetricsTrends(ctx, timeRange)
	if err != nil {
		httpx.BindErr(c, err)
		return
	}

	httpx.OK(c, trends)
}

func (h *MetricsHandler) getMetricsSummary(ctx context.Context) (*MetricsSummary, error) {
	summary := &MetricsSummary{
		ByEnvironment: make(map[string]EnvMetrics),
		ByStatus:      make(map[string]int64),
	}

	// 总发布数
	if err := h.svcCtx.DB.WithContext(ctx).Model(&model.DeploymentRelease{}).Count(&summary.TotalReleases).Error; err != nil {
		return nil, err
	}

	// 按状态统计
	var statusCounts []struct {
		State string
		Count int64
	}
	if err := h.svcCtx.DB.WithContext(ctx).Model(&model.DeploymentRelease{}).
		Select("state, count(*) as count").
		Group("state").
		Scan(&statusCounts).Error; err != nil {
		return nil, err
	}

	var successCount, failureCount int64
	for _, sc := range statusCounts {
		summary.ByStatus[sc.State] = sc.Count
		if sc.State == "applied" || sc.State == "success" {
			successCount = sc.Count
		}
		if sc.State == "failed" {
			failureCount = sc.Count
		}
	}

	// 成功率
	if summary.TotalReleases > 0 {
		summary.SuccessRate = float64(successCount) / float64(summary.TotalReleases) * 100
		summary.FailureRate = float64(failureCount) / float64(summary.TotalReleases) * 100
	}

	// 按环境统计 (从 targets 关联)
	var envCounts []struct {
		Env   string
		Total int64
	}
	if err := h.svcCtx.DB.WithContext(ctx).
		Table("deploy_releases r").
		Select("t.env, count(*) as total").
		Joins("JOIN deploy_targets t ON r.target_id = t.id").
		Group("t.env").
		Scan(&envCounts).Error; err != nil {
		return nil, err
	}

	for _, ec := range envCounts {
		summary.ByEnvironment[ec.Env] = EnvMetrics{
			Total:       ec.Total,
			SuccessRate: 0, // 可以后续优化
		}
	}

	// 最近7天的统计
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if err := h.svcCtx.DB.WithContext(ctx).
		Model(&model.DeploymentRelease{}).
		Where("created_at > ?", sevenDaysAgo).
		Count(&summary.RecentReleases).Error; err != nil {
		return nil, err
	}

	if err := h.svcCtx.DB.WithContext(ctx).
		Model(&model.DeploymentRelease{}).
		Where("state = ? AND created_at > ?", "failed", sevenDaysAgo).
		Count(&summary.RecentFailures).Error; err != nil {
		return nil, err
	}

	return summary, nil
}

func (h *MetricsHandler) getMetricsTrends(ctx context.Context, timeRange string) ([]MetricsTrend, error) {
	var trends []MetricsTrend

	// 根据时间范围确定分组格式
	var dateFormat string
	var startDate time.Time
	switch timeRange {
	case "weekly":
		dateFormat = "%Y-%u"
		startDate = time.Now().AddDate(0, 0, -28) // 最近4周
	case "monthly":
		dateFormat = "%Y-%m"
		startDate = time.Now().AddDate(0, -6, 0) // 最近6个月
	default: // daily
		dateFormat = "%Y-%m-%d"
		startDate = time.Now().AddDate(0, 0, -7) // 最近7天
	}

	var results []struct {
		Date   string
		Total  int64
		Failed int64
	}

	// 按日期分组统计
	if err := h.svcCtx.DB.WithContext(ctx).
		Model(&model.DeploymentRelease{}).
		Select("DATE_FORMAT(created_at, '"+dateFormat+"') as date, count(*) as total, sum(case when state = 'failed' then 1 else 0 end) as failed").
		Where("created_at > ?", startDate).
		Group("date").
		Order("date").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	for _, r := range results {
		successCount := r.Total - r.Failed
		var successRate float64
		if r.Total > 0 {
			successRate = float64(successCount) / float64(r.Total) * 100
		}
		trends = append(trends, MetricsTrend{
			Date:            r.Date,
			DeploymentCount: int(r.Total),
			SuccessCount:    int(successCount),
			FailureCount:    int(r.Failed),
			SuccessRate:     successRate,
		})
	}

	return trends, nil
}
