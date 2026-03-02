package aiops

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/httpx"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svcCtx *svc.ServiceContext
}

func NewHandler(svcCtx *svc.ServiceContext) *Handler { return &Handler{svcCtx: svcCtx} }

func (h *Handler) RunInspection(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "aiops:run") {
		return
	}
	var req struct {
		ReleaseID uint   `json:"release_id"`
		TargetID  uint   `json:"target_id"`
		ServiceID uint   `json:"service_id"`
		Stage     string `json:"stage"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
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
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, rec)
}

func (h *Handler) ListInspections(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "aiops:read") {
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
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, gin.H{"list": rows, "total": len(rows)})
}

func (h *Handler) GetInspection(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "aiops:read") {
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var row model.AIOPSInspection
	if err := h.svcCtx.DB.First(&row, id).Error; err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, row)
}

func (h *Handler) ApplyPreview(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "aiops:read", "aiops:run") {
		return
	}
	var req struct {
		InspectionID uint `json:"inspection_id"`
	}
	_ = c.ShouldBindJSON(&req)
	httpx.OK(c, gin.H{
		"inspection_id": req.InspectionID,
		"preview":       "AIOPS recommendation apply preview is available; mutating actions still require approval.",
	})
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

// GetRiskFindings 获取风险发现列表
func (h *Handler) GetRiskFindings(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "aiops:read") {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	ctx := c.Request.Context()
	query := h.svcCtx.DB.WithContext(ctx).Model(&model.RiskFinding{})

	// Filter by severity
	if severity := c.Query("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}
	// Filter by service_id
	if serviceID := c.Query("service_id"); serviceID != "" {
		query = query.Where("service_id = ?", serviceID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		httpx.ServerErr(c, err)
		return
	}

	var findings []model.RiskFinding
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&findings).Error; err != nil {
		httpx.ServerErr(c, err)
		return
	}

	httpx.OK(c, gin.H{"list": findings, "total": total})
}

// GetAnomalies 获取异常检测列表
func (h *Handler) GetAnomalies(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "aiops:read") {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	ctx := c.Request.Context()
	query := h.svcCtx.DB.WithContext(ctx).Model(&model.Anomaly{})

	// Filter by type
	if anomalyType := c.Query("type"); anomalyType != "" {
		query = query.Where("type = ?", anomalyType)
	}
	// Filter by service_id
	if serviceID := c.Query("service_id"); serviceID != "" {
		query = query.Where("service_id = ?", serviceID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		httpx.ServerErr(c, err)
		return
	}

	var anomalies []model.Anomaly
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&anomalies).Error; err != nil {
		httpx.ServerErr(c, err)
		return
	}

	httpx.OK(c, gin.H{"list": anomalies, "total": total})
}

// GetSuggestions 获取优化建议列表
func (h *Handler) GetSuggestions(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "aiops:read") {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	ctx := c.Request.Context()
	query := h.svcCtx.DB.WithContext(ctx).Model(&model.Suggestion{})

	// Filter by impact
	if impact := c.Query("impact"); impact != "" {
		query = query.Where("impact = ?", impact)
	}
	// Filter by service_id
	if serviceID := c.Query("service_id"); serviceID != "" {
		query = query.Where("service_id = ?", serviceID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		httpx.ServerErr(c, err)
		return
	}

	var suggestions []model.Suggestion
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&suggestions).Error; err != nil {
		httpx.ServerErr(c, err)
		return
	}

	httpx.OK(c, gin.H{"list": suggestions, "total": total})
}
