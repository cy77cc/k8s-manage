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
	alerts, total, err := h.logic.ListAlerts(c.Request.Context(), strings.TrimSpace(c.Query("severity")), strings.TrimSpace(c.Query("status")), intFromQuery(c, "page", 1), intFromQuery(c, "page_size", 20))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": alerts, "total": total})
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
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rules, "total": total})
}

func (h *Handler) CreateRule(c *gin.Context) {
	if !h.authorize(c, "monitoring:write") {
		return
	}
	var req struct {
		Name      string  `json:"name" binding:"required"`
		Metric    string  `json:"metric" binding:"required"`
		Operator  string  `json:"operator"`
		Threshold float64 `json:"threshold"`
		Severity  string  `json:"severity"`
		Enabled   *bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	rule, err := h.logic.CreateRule(c.Request.Context(), model.AlertRule{
		Name:      req.Name,
		Metric:    req.Metric,
		Operator:  defaultIfEmpty(req.Operator, "gt"),
		Threshold: req.Threshold,
		Severity:  defaultIfEmpty(req.Severity, "warning"),
		Enabled:   req.Enabled == nil || *req.Enabled,
		Source:    "custom",
		Scope:     "global",
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
		Name      string  `json:"name"`
		Operator  string  `json:"operator"`
		Threshold float64 `json:"threshold"`
		Severity  string  `json:"severity"`
		Enabled   *bool   `json:"enabled"`
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
	rule, err := h.logic.UpdateRule(c.Request.Context(), uint(id), payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rule})
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
	rows, err := h.logic.GetMetrics(c.Request.Context(), metric, start, end)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rows})
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
	if err != nil || v <= 0 {
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
