package monitoring

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	logic  *Logic
	svcCtx *svc.ServiceContext
}

func NewHandler(svcCtx *svc.ServiceContext) *Handler {
	return &Handler{logic: NewLogic(svcCtx), svcCtx: svcCtx}
}

func (h *Handler) StartCollector() {
	h.logic.StartCollector()
}

func (h *Handler) ListAlerts(c *gin.Context) {
	if !h.authorize(c, "monitoring:read") {
		return
	}
	alerts, total, err := h.logic.ListAlerts(
		c.Request.Context(),
		strings.TrimSpace(c.Query("severity")),
		strings.TrimSpace(c.Query("status")),
		intFromQuery(c, "page", 1),
		intFromQuery(c, "page_size", 20),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": alerts, "total": total}})
}

func (h *Handler) ListRules(c *gin.Context) {
	if !h.authorize(c, "monitoring:read") {
		return
	}
	rules, total, err := h.logic.ListRules(c.Request.Context(), intFromQuery(c, "page", 1), intFromQuery(c, "page_size", 50))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": rules, "total": total}})
}

func (h *Handler) CreateRule(c *gin.Context) {
	if !h.authorize(c, "monitoring:write") {
		return
	}
	var req struct {
		Name           string  `json:"name" binding:"required"`
		Metric         string  `json:"metric" binding:"required"`
		Operator       string  `json:"operator"`
		Threshold      float64 `json:"threshold"`
		Severity       string  `json:"severity"`
		Enabled        *bool   `json:"enabled"`
		WindowSec      int     `json:"window_sec"`
		GranularitySec int     `json:"granularity_sec"`
		DimensionsJSON string  `json:"dimensions_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	rule, err := h.logic.CreateRule(c.Request.Context(), model.AlertRule{
		Name:           req.Name,
		Metric:         req.Metric,
		Operator:       defaultIfEmpty(req.Operator, "gt"),
		Threshold:      req.Threshold,
		Severity:       defaultIfEmpty(req.Severity, "warning"),
		Enabled:        req.Enabled == nil || *req.Enabled,
		WindowSec:      positiveOr(req.WindowSec, 3600),
		GranularitySec: positiveOr(req.GranularitySec, 60),
		DimensionsJSON: strings.TrimSpace(req.DimensionsJSON),
		Source:         "custom",
		Scope:          "global",
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rule})
}

func (h *Handler) UpdateRule(c *gin.Context) {
	if !h.authorize(c, "monitoring:write") {
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Name           string  `json:"name"`
		Operator       string  `json:"operator"`
		Threshold      float64 `json:"threshold"`
		Severity       string  `json:"severity"`
		Enabled        *bool   `json:"enabled"`
		WindowSec      int     `json:"window_sec"`
		GranularitySec int     `json:"granularity_sec"`
		DimensionsJSON *string `json:"dimensions_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	payload := map[string]any{}
	if strings.TrimSpace(req.Name) != "" {
		payload["name"] = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.Operator) != "" {
		payload["operator"] = strings.TrimSpace(req.Operator)
	}
	if req.Threshold > 0 {
		payload["threshold"] = req.Threshold
	}
	if strings.TrimSpace(req.Severity) != "" {
		payload["severity"] = strings.TrimSpace(req.Severity)
	}
	if req.Enabled != nil {
		payload["enabled"] = *req.Enabled
	}
	if req.WindowSec > 0 {
		payload["window_sec"] = req.WindowSec
	}
	if req.GranularitySec > 0 {
		payload["granularity_sec"] = req.GranularitySec
	}
	if req.DimensionsJSON != nil {
		payload["dimensions_json"] = strings.TrimSpace(*req.DimensionsJSON)
	}
	rule, err := h.logic.UpdateRule(c.Request.Context(), uint(id), payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rule})
}

func (h *Handler) EnableRule(c *gin.Context) {
	h.setRuleEnabled(c, true)
}

func (h *Handler) DisableRule(c *gin.Context) {
	h.setRuleEnabled(c, false)
}

func (h *Handler) setRuleEnabled(c *gin.Context, enabled bool) {
	if !h.authorize(c, "monitoring:write") {
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	rule, err := h.logic.SetRuleEnabled(c.Request.Context(), uint(id), enabled)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rule})
}

func (h *Handler) ListRuleEvaluations(c *gin.Context) {
	if !h.authorize(c, "monitoring:read") {
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	items, total, err := h.logic.ListRuleEvaluations(c.Request.Context(), uint(id), intFromQuery(c, "page", 1), intFromQuery(c, "page_size", 50))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": items, "total": total}})
}

func (h *Handler) GetMetrics(c *gin.Context) {
	if !h.authorize(c, "monitoring:read") {
		return
	}
	metric := strings.TrimSpace(c.Query("metric"))
	if metric == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "metric is required"})
		return
	}
	start, err := parseTime(defaultIfEmpty(c.Query("start_time"), time.Now().Add(-24*time.Hour).Format(time.RFC3339)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "invalid start_time"})
		return
	}
	end, err := parseTime(defaultIfEmpty(c.Query("end_time"), time.Now().Format(time.RFC3339)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "invalid end_time"})
		return
	}
	out, err := h.logic.GetMetrics(c.Request.Context(), MetricQuery{
		Metric:         metric,
		Start:          start,
		End:            end,
		GranularitySec: intFromQuery(c, "granularity_sec", 60),
		Source:         strings.TrimSpace(c.Query("source")),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": out})
}

func (h *Handler) ListChannels(c *gin.Context) {
	if !h.authorize(c, "monitoring:read") {
		return
	}
	items, err := h.logic.ListChannels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": items, "total": len(items)}})
}

func (h *Handler) CreateChannel(c *gin.Context) {
	if !h.authorize(c, "monitoring:write") {
		return
	}
	var req struct {
		Name       string `json:"name" binding:"required"`
		Type       string `json:"type"`
		Provider   string `json:"provider"`
		Target     string `json:"target"`
		ConfigJSON string `json:"config_json"`
		Enabled    *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	item, err := h.logic.CreateChannel(c.Request.Context(), model.AlertNotificationChannel{
		Name:       strings.TrimSpace(req.Name),
		Type:       strings.TrimSpace(req.Type),
		Provider:   strings.TrimSpace(req.Provider),
		Target:     strings.TrimSpace(req.Target),
		ConfigJSON: strings.TrimSpace(req.ConfigJSON),
		Enabled:    req.Enabled == nil || *req.Enabled,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": item})
}

func (h *Handler) UpdateChannel(c *gin.Context) {
	if !h.authorize(c, "monitoring:write") {
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Name       string  `json:"name"`
		Type       string  `json:"type"`
		Provider   string  `json:"provider"`
		Target     string  `json:"target"`
		ConfigJSON *string `json:"config_json"`
		Enabled    *bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	payload := map[string]any{}
	if strings.TrimSpace(req.Name) != "" {
		payload["name"] = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.Type) != "" {
		payload["type"] = strings.TrimSpace(req.Type)
	}
	if strings.TrimSpace(req.Provider) != "" {
		payload["provider"] = strings.TrimSpace(req.Provider)
	}
	if strings.TrimSpace(req.Target) != "" {
		payload["target"] = strings.TrimSpace(req.Target)
	}
	if req.ConfigJSON != nil {
		payload["config_json"] = strings.TrimSpace(*req.ConfigJSON)
	}
	if req.Enabled != nil {
		payload["enabled"] = *req.Enabled
	}
	item, err := h.logic.UpdateChannel(c.Request.Context(), uint(id), payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": item})
}

func (h *Handler) ListDeliveries(c *gin.Context) {
	if !h.authorize(c, "monitoring:read") {
		return
	}
	alertID := uint(intFromQuery(c, "alert_id", 0))
	items, total, err := h.logic.ListDeliveries(
		c.Request.Context(),
		alertID,
		strings.TrimSpace(c.Query("channel_type")),
		strings.TrimSpace(c.Query("status")),
		intFromQuery(c, "page", 1),
		intFromQuery(c, "page_size", 20),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": items, "total": total}})
}

func (h *Handler) authorize(c *gin.Context, code string) bool {
	uid, ok := c.Get("uid")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized"})
		return false
	}
	var user model.User
	if err := h.svcCtx.DB.Select("id,username").Where("id = ?", toUint(uid)).First(&user).Error; err == nil && strings.EqualFold(user.Username, "admin") {
		return true
	}
	var rows []struct {
		Code string `gorm:"column:code"`
	}
	if err := h.svcCtx.DB.Table("permissions").
		Select("permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", toUint(uid)).
		Scan(&rows).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden"})
		return false
	}
	for _, r := range rows {
		if r.Code == code || r.Code == "monitoring:*" || r.Code == "*:*" {
			return true
		}
	}
	c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden"})
	return false
}

func toUint(v any) uint64 {
	switch x := v.(type) {
	case uint64:
		return x
	case uint:
		return uint64(x)
	case int:
		if x < 0 {
			return 0
		}
		return uint64(x)
	case int64:
		if x < 0 {
			return 0
		}
		return uint64(x)
	default:
		return 0
	}
}

func parseTime(raw string) (time.Time, error) {
	return time.Parse(time.RFC3339, raw)
}

func intFromQuery(c *gin.Context, key string, def int) int {
	v, err := strconv.Atoi(strings.TrimSpace(c.Query(key)))
	if err != nil || v < 0 {
		return def
	}
	if v == 0 {
		return def
	}
	return v
}

func defaultIfEmpty(v, d string) string {
	if strings.TrimSpace(v) == "" {
		return d
	}
	return strings.TrimSpace(v)
}

func positiveOr(v, d int) int {
	if v > 0 {
		return v
	}
	return d
}
