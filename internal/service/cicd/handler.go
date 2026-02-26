package cicd

import (
	"net/http"
	"strconv"
	"strings"

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

func (h *Handler) GetServiceCIConfig(c *gin.Context) {
	if !h.authorize(c, "cicd:ci:read", "cicd:*") {
		return
	}
	row, err := h.logic.GetServiceCIConfig(c.Request.Context(), uintFromParam(c, "service_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 3004, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) PutServiceCIConfig(c *gin.Context) {
	if !h.authorize(c, "cicd:ci:write", "cicd:*") {
		return
	}
	var req UpsertServiceCIConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid, _ := c.Get("uid")
	row, err := h.logic.UpsertServiceCIConfig(c.Request.Context(), uint(toUint(uid)), uintFromParam(c, "service_id"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) DeleteServiceCIConfig(c *gin.Context) {
	if !h.authorize(c, "cicd:ci:write", "cicd:*") {
		return
	}
	uid, _ := c.Get("uid")
	if err := h.logic.DeleteServiceCIConfig(c.Request.Context(), uint(toUint(uid)), uintFromParam(c, "service_id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) TriggerCIRun(c *gin.Context) {
	if !h.authorize(c, "cicd:ci:run", "cicd:*") {
		return
	}
	var req TriggerCIRunReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid, _ := c.Get("uid")
	row, err := h.logic.TriggerCIRun(c.Request.Context(), uint(toUint(uid)), uintFromParam(c, "service_id"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) ListCIRuns(c *gin.Context) {
	if !h.authorize(c, "cicd:ci:read", "cicd:*") {
		return
	}
	rows, err := h.logic.ListCIRuns(c.Request.Context(), uintFromParam(c, "service_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": rows, "total": len(rows)}})
}

func (h *Handler) GetDeploymentCDConfig(c *gin.Context) {
	if !h.authorize(c, "cicd:cd:read", "cicd:*") {
		return
	}
	row, err := h.logic.GetDeploymentCDConfig(c.Request.Context(), uintFromParam(c, "deployment_id"), strings.TrimSpace(c.Query("env")), strings.TrimSpace(c.Query("runtime_type")))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 3004, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) PutDeploymentCDConfig(c *gin.Context) {
	if !h.authorize(c, "cicd:cd:write", "cicd:*") {
		return
	}
	var req UpsertDeploymentCDConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid, _ := c.Get("uid")
	row, err := h.logic.UpsertDeploymentCDConfig(c.Request.Context(), uint(toUint(uid)), uintFromParam(c, "deployment_id"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) TriggerRelease(c *gin.Context) {
	if !h.authorize(c, "cicd:release:run", "cicd:*") {
		return
	}
	var req TriggerReleaseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid, _ := c.Get("uid")
	row, err := h.logic.TriggerRelease(c.Request.Context(), uint(toUint(uid)), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) ListReleases(c *gin.Context) {
	if !h.authorize(c, "cicd:cd:read", "cicd:audit:read", "cicd:*") {
		return
	}
	rows, err := h.logic.ListReleases(c.Request.Context(), uintFromQuery(c, "service_id"), uintFromQuery(c, "deployment_id"), strings.TrimSpace(c.Query("runtime_type")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": rows, "total": len(rows)}})
}

func (h *Handler) ApproveRelease(c *gin.Context) {
	if !h.authorize(c, "cicd:release:approve", "cicd:*") {
		return
	}
	var req ReleaseDecisionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid, _ := c.Get("uid")
	row, err := h.logic.ApproveRelease(c.Request.Context(), uint(toUint(uid)), uintFromParam(c, "id"), req.Comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) RejectRelease(c *gin.Context) {
	if !h.authorize(c, "cicd:release:approve", "cicd:*") {
		return
	}
	var req ReleaseDecisionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid, _ := c.Get("uid")
	row, err := h.logic.RejectRelease(c.Request.Context(), uint(toUint(uid)), uintFromParam(c, "id"), req.Comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) RollbackRelease(c *gin.Context) {
	if !h.authorize(c, "cicd:release:rollback", "cicd:*") {
		return
	}
	var req RollbackReleaseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid, _ := c.Get("uid")
	row, err := h.logic.RollbackRelease(c.Request.Context(), uint(toUint(uid)), uintFromParam(c, "id"), req.TargetVersion, req.Comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) ListApprovals(c *gin.Context) {
	if !h.authorize(c, "cicd:audit:read", "cicd:release:approve", "cicd:*") {
		return
	}
	rows, err := h.logic.ListApprovals(c.Request.Context(), uintFromParam(c, "id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": rows, "total": len(rows)}})
}

func (h *Handler) ServiceTimeline(c *gin.Context) {
	if !h.authorize(c, "cicd:audit:read", "cicd:*") {
		return
	}
	rows, err := h.logic.ServiceTimeline(c.Request.Context(), uintFromParam(c, "service_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": rows, "total": len(rows)}})
}

func (h *Handler) ListAuditEvents(c *gin.Context) {
	if !h.authorize(c, "cicd:audit:read", "cicd:*") {
		return
	}
	limit := 100
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	rows, err := h.logic.ListAuditEvents(
		c.Request.Context(),
		uintFromQuery(c, "service_id"),
		strings.TrimSpace(c.Query("trace_id")),
		strings.TrimSpace(c.Query("command_id")),
		limit,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": rows, "total": len(rows)}})
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
	for _, code := range codes {
		if h.hasPermission(toUint(uid), code) {
			return true
		}
	}
	c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden"})
	return false
}

func (h *Handler) hasPermission(uid uint64, code string) bool {
	var rows []struct {
		Code string `gorm:"column:code"`
	}
	if err := h.svcCtx.DB.Table("permissions").
		Select("permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", uid).
		Scan(&rows).Error; err != nil {
		return false
	}
	for _, r := range rows {
		if r.Code == code || r.Code == "*:*" || r.Code == "cicd:*" {
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
	var user model.User
	if err := h.svcCtx.DB.Select("id,username").Where("id = ?", toUint(uid)).First(&user).Error; err == nil && strings.EqualFold(user.Username, "admin") {
		return true
	}
	return false
}

func uintFromParam(c *gin.Context, key string) uint {
	v := strings.TrimSpace(c.Param(key))
	n, _ := strconv.ParseUint(v, 10, 64)
	return uint(n)
}

func uintFromQuery(c *gin.Context, key string) uint {
	v := strings.TrimSpace(c.Query(key))
	n, _ := strconv.ParseUint(v, 10, 64)
	return uint(n)
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
	case string:
		n, _ := strconv.ParseUint(strings.TrimSpace(x), 10, 64)
		return n
	default:
		return 0
	}
}
