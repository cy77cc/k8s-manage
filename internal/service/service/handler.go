package service

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

func (h *Handler) Preview(c *gin.Context) {
	if !h.authorize(c, "service", "write") {
		return
	}
	var req RenderPreviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	resp, err := h.logic.Preview(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) Transform(c *gin.Context) {
	if !h.authorize(c, "service", "write") {
		return
	}
	var req TransformReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	resp, err := h.logic.Transform(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) Create(c *gin.Context) {
	if !h.authorize(c, "service", "write") {
		return
	}
	var req ServiceCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	if !h.checkOwnershipHeaders(c, req.ProjectID, req.TeamID) {
		return
	}
	uid, _ := c.Get("uid")
	resp, err := h.logic.Create(c.Request.Context(), toUint(uid), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) Update(c *gin.Context) {
	if !h.authorize(c, "service", "write") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "invalid id"})
		return
	}
	var req ServiceCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	if !h.checkOwnershipHeaders(c, req.ProjectID, req.TeamID) {
		return
	}
	resp, err := h.logic.Update(c.Request.Context(), uint(id), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) List(c *gin.Context) {
	if !h.authorize(c, "service", "read") {
		return
	}
	filters := map[string]string{
		"project_id":     c.Query("project_id"),
		"team_id":        c.Query("team_id"),
		"runtime_type":   c.Query("runtime_type"),
		"env":            c.Query("env"),
		"label_selector": c.Query("label_selector"),
		"q":              c.Query("q"),
	}
	if !h.isAdmin(c) {
		if hp := strings.TrimSpace(c.GetHeader("X-Project-ID")); hp != "" {
			filters["project_id"] = hp
		}
		if ht := strings.TrimSpace(c.GetHeader("X-Team-ID")); ht != "" {
			filters["team_id"] = ht
		}
	}
	list, total, err := h.logic.List(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": list, "total": total}})
}

func (h *Handler) Get(c *gin.Context) {
	if !h.authorize(c, "service", "read") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "invalid id"})
		return
	}
	resp, err := h.logic.Get(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) Delete(c *gin.Context) {
	if !h.authorize(c, "service", "write") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "invalid id"})
		return
	}
	if err := h.logic.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) Deploy(c *gin.Context) {
	if !h.authorize(c, "service", "deploy") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "invalid id"})
		return
	}
	var req DeployReq
	_ = c.ShouldBindJSON(&req)
	item, err := h.logic.Get(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	env := defaultIfEmpty(req.Env, item.Env)
	if strings.EqualFold(env, "production") && !h.hasPermission(c, "service:approve") {
		c.JSON(http.StatusForbidden, gin.H{"code": 2004, "msg": "production deploy requires service:approve"})
		return
	}
	uid, _ := c.Get("uid")
	recordID, err := h.logic.Deploy(c.Request.Context(), uint(id), toUint(uid), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": DeployResp{ReleaseRecordID: recordID}})
}

func (h *Handler) DeployPreview(c *gin.Context) {
	if !h.authorize(c, "service", "deploy") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "invalid id"})
		return
	}
	var req DeployReq
	_ = c.ShouldBindJSON(&req)
	resp, err := h.logic.DeployPreview(c.Request.Context(), uint(id), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) HelmImport(c *gin.Context) {
	if !h.authorize(c, "service", "write") {
		return
	}
	var req HelmImportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid, _ := c.Get("uid")
	resp, err := h.logic.HelmImport(c.Request.Context(), toUint(uid), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) HelmRender(c *gin.Context) {
	if !h.authorize(c, "service", "write") {
		return
	}
	var req HelmRenderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	rendered, diags, err := h.logic.HelmRender(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error(), "data": gin.H{"rendered_yaml": rendered, "diagnostics": diags}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"rendered_yaml": rendered, "diagnostics": diags}})
}

func (h *Handler) DeployHelm(c *gin.Context) {
	if !h.authorize(c, "service", "deploy") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "invalid id"})
		return
	}
	if err := h.logic.deployHelm(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) Rollback(c *gin.Context) {
	if !h.authorize(c, "service", "deploy") {
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "rollback is not implemented in MVP", "data": nil})
}

func (h *Handler) Events(c *gin.Context) {
	if !h.authorize(c, "service", "read") {
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": []gin.H{
		{"id": 1, "service_id": c.Param("id"), "type": "deploy", "level": "info", "message": "service event", "created_at": time.Now()},
	}})
}

func (h *Handler) Quota(c *gin.Context) {
	if !h.authorize(c, "service", "read") {
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{
		"cpuLimit":    8000,
		"memoryLimit": 16384,
		"cpuUsed":     1200,
		"memoryUsed":  2048,
	}})
}

func (h *Handler) ExtractVariables(c *gin.Context) {
	if !h.authorize(c, "service", "write") {
		return
	}
	var req VariableExtractReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	resp, err := h.logic.ExtractVariables(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) GetVariableSchema(c *gin.Context) {
	if !h.authorize(c, "service", "read") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "invalid id"})
		return
	}
	resp, err := h.logic.GetVariableSchema(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"vars": resp}})
}

func (h *Handler) GetVariableValues(c *gin.Context) {
	if !h.authorize(c, "service", "read") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "invalid id"})
		return
	}
	resp, err := h.logic.GetVariableValues(c.Request.Context(), uint(id), c.Query("env"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) UpsertVariableValues(c *gin.Context) {
	if !h.authorize(c, "service", "write") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "invalid id"})
		return
	}
	var req VariableValuesUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid, _ := c.Get("uid")
	resp, err := h.logic.UpsertVariableValues(c.Request.Context(), uint(id), toUint(uid), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) ListRevisions(c *gin.Context) {
	if !h.authorize(c, "service", "read") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "invalid id"})
		return
	}
	resp, err := h.logic.ListRevisions(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": resp, "total": len(resp)}})
}

func (h *Handler) CreateRevision(c *gin.Context) {
	if !h.authorize(c, "service", "write") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "invalid id"})
		return
	}
	var req RevisionCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid, _ := c.Get("uid")
	resp, err := h.logic.CreateRevision(c.Request.Context(), uint(id), toUint(uid), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) UpsertDeployTarget(c *gin.Context) {
	if !h.authorize(c, "service", "write") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "invalid id"})
		return
	}
	var req DeployTargetUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid, _ := c.Get("uid")
	resp, err := h.logic.UpsertDeployTarget(c.Request.Context(), uint(id), toUint(uid), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) ListReleaseRecords(c *gin.Context) {
	if !h.authorize(c, "service", "read") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "invalid id"})
		return
	}
	resp, err := h.logic.ListReleaseRecords(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": resp, "total": len(resp)}})
}

func (h *Handler) authorize(c *gin.Context, resource, action string) bool {
	if h.hasPermission(c, resource+":"+action) || h.hasPermission(c, resource+":*") || h.hasPermission(c, "*:*") {
		return true
	}
	c.JSON(http.StatusForbidden, gin.H{"code": 2004, "msg": "forbidden"})
	return false
}

func (h *Handler) hasPermission(c *gin.Context, code string) bool {
	if h.isAdmin(c) {
		return true
	}
	uid, ok := c.Get("uid")
	if !ok {
		return false
	}
	userID := toUint(uid)
	type row struct {
		Code string `gorm:"column:code"`
	}
	var rows []row
	err := h.svcCtx.DB.Table("permissions").
		Select("permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Scan(&rows).Error
	if err != nil {
		return false
	}
	for _, r := range rows {
		if strings.TrimSpace(r.Code) == code {
			return true
		}
	}
	return false
}

func (h *Handler) isAdmin(c *gin.Context) bool {
	uid, ok := c.Get("uid")
	if !ok {
		return false
	}
	id := toUint(uid)
	if id == 0 {
		return false
	}
	var u model.User
	if err := h.svcCtx.DB.Select("id", "username").Where("id = ?", id).First(&u).Error; err == nil {
		if strings.EqualFold(strings.TrimSpace(u.Username), "admin") {
			return true
		}
	}
	var rows []struct {
		Code string `gorm:"column:code"`
	}
	if err := h.svcCtx.DB.Table("roles").Select("roles.code").Joins("JOIN user_roles ON user_roles.role_id = roles.id").Where("user_roles.user_id = ?", id).Scan(&rows).Error; err != nil {
		return false
	}
	for _, row := range rows {
		if strings.EqualFold(strings.TrimSpace(row.Code), "admin") {
			return true
		}
	}
	return false
}

func (h *Handler) checkOwnershipHeaders(c *gin.Context, projectID, teamID uint) bool {
	if h.isAdmin(c) {
		return true
	}
	if hp := strings.TrimSpace(c.GetHeader("X-Project-ID")); hp != "" && projectID > 0 {
		if strconv.FormatUint(uint64(projectID), 10) != hp {
			c.JSON(http.StatusForbidden, gin.H{"code": 2004, "msg": "project ownership mismatch"})
			return false
		}
	}
	if ht := strings.TrimSpace(c.GetHeader("X-Team-ID")); ht != "" && teamID > 0 {
		if strconv.FormatUint(uint64(teamID), 10) != ht {
			c.JSON(http.StatusForbidden, gin.H{"code": 2004, "msg": "team ownership mismatch"})
			return false
		}
	}
	return true
}

func toUint(v any) uint64 {
	switch x := v.(type) {
	case uint:
		return uint64(x)
	case uint64:
		return x
	case int:
		return uint64(x)
	case int64:
		return uint64(x)
	case float64:
		return uint64(x)
	default:
		return 0
	}
}
