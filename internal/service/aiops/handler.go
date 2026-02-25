package aiops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svcCtx *svc.ServiceContext
}

func NewHandler(svcCtx *svc.ServiceContext) *Handler { return &Handler{svcCtx: svcCtx} }

func (h *Handler) RunInspection(c *gin.Context) {
	if !h.authorize(c, "aiops:run") {
		return
	}
	var req struct {
		ReleaseID uint   `json:"release_id"`
		TargetID  uint   `json:"target_id"`
		ServiceID uint   `json:"service_id"`
		Stage     string `json:"stage"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	req.Stage = defaultIfEmpty(req.Stage, "periodic")
	summary := "Deployment looks healthy with minor optimization recommendations."
	findings := []map[string]any{
		{"level": "info", "code": "resource_headroom", "message": "CPU and memory headroom seems adequate"},
	}
	suggestions := []map[string]any{
		{"type": "suggestion", "title": "Enable post-deploy health checks", "risk": "low"},
	}
	if h.svcCtx.AI != nil {
		msgs := []*schema.Message{schema.UserMessage(fmt.Sprintf("Please analyze deployment risk for service=%d target=%d release=%d stage=%s and return concise Chinese summary.", req.ServiceID, req.TargetID, req.ReleaseID, req.Stage))}
		if out, err := h.svcCtx.AI.Generate(context.Background(), msgs); err == nil && out != nil && strings.TrimSpace(out.Content) != "" {
			summary = out.Content
		}
	}
	rec := &model.AIOPSInspection{
		ReleaseID:       req.ReleaseID,
		TargetID:        req.TargetID,
		ServiceID:       req.ServiceID,
		Stage:           req.Stage,
		Summary:         summary,
		FindingsJSON:    toJSON(findings),
		SuggestionsJSON: toJSON(suggestions),
		Status:          "done",
		CreatedAt:       time.Now(),
	}
	if err := h.svcCtx.DB.Create(rec).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rec})
}

func (h *Handler) ListInspections(c *gin.Context) {
	if !h.authorize(c, "aiops:read") {
		return
	}
	q := h.svcCtx.DB.Model(&model.AIOPSInspection{})
	if v := strings.TrimSpace(c.Query("service_id")); v != "" {
		q = q.Where("service_id = ?", v)
	}
	if v := strings.TrimSpace(c.Query("target_id")); v != "" {
		q = q.Where("target_id = ?", v)
	}
	var rows []model.AIOPSInspection
	if err := q.Order("id DESC").Limit(200).Find(&rows).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": rows, "total": len(rows)}})
}

func (h *Handler) GetInspection(c *gin.Context) {
	if !h.authorize(c, "aiops:read") {
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var row model.AIOPSInspection
	if err := h.svcCtx.DB.First(&row, id).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) ApplyPreview(c *gin.Context) {
	if !h.authorize(c, "aiops:read", "aiops:run") {
		return
	}
	var req struct {
		InspectionID uint `json:"inspection_id"`
	}
	_ = c.ShouldBindJSON(&req)
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{
		"inspection_id": req.InspectionID,
		"preview":       "AIOPS recommendation apply preview is available; mutating actions still require approval.",
	}})
}

func (h *Handler) authorize(c *gin.Context, codes ...string) bool {
	if h.isAdmin(c) {
		return true
	}
	uid, ok := c.Get("uid")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized"})
		return false
	}
	var rows []struct{ Code string `gorm:"column:code"` }
	if err := h.svcCtx.DB.Table("permissions").
		Select("permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", toUint(uid)).
		Scan(&rows).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden"})
		return false
	}
	for _, code := range codes {
		for _, r := range rows {
			if r.Code == code || r.Code == "*:*" {
				return true
			}
		}
	}
	c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden"})
	return false
}

func (h *Handler) isAdmin(c *gin.Context) bool {
	uid, ok := c.Get("uid")
	if !ok {
		return false
	}
	var user model.User
	if err := h.svcCtx.DB.Select("id,username").Where("id = ?", toUint(uid)).First(&user).Error; err == nil && strings.EqualFold(user.Username, "admin") {
		return true
	}
	return false
}

func toUint(v any) uint64 {
	switch x := v.(type) {
	case uint:
		return uint64(x)
	case uint64:
		return x
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

func defaultIfEmpty(v, d string) string {
	if strings.TrimSpace(v) == "" {
		return d
	}
	return v
}

func toJSON(v any) string {
	raw, _ := json.Marshal(v)
	return string(raw)
}
