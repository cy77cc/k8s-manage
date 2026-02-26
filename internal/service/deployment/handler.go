package deployment

import (
	"fmt"
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

func (h *Handler) ListTargets(c *gin.Context) {
	if !h.authorize(c, "deploy:target:read") {
		return
	}
	list, err := h.logic.ListTargets(c.Request.Context(), uintFromQuery(c, "project_id"), uintFromQuery(c, "team_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": list, "total": len(list)}})
}

func (h *Handler) CreateTarget(c *gin.Context) {
	var req TargetUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	if !h.authorize(c, "deploy:target:write") || !h.authorizeRuntime(c, req.TargetType, "apply") {
		return
	}
	uid, _ := c.Get("uid")
	resp, err := h.logic.CreateTarget(c.Request.Context(), toUint(uid), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) GetTarget(c *gin.Context) {
	if !h.authorize(c, "deploy:target:read") {
		return
	}
	id := uintFromParam(c, "id")
	resp, err := h.logic.GetTarget(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) UpdateTarget(c *gin.Context) {
	var req TargetUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	if !h.authorize(c, "deploy:target:write") || !h.authorizeRuntime(c, req.TargetType, "apply") {
		return
	}
	resp, err := h.logic.UpdateTarget(c.Request.Context(), uintFromParam(c, "id"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) DeleteTarget(c *gin.Context) {
	if !h.authorize(c, "deploy:target:write") {
		return
	}
	if err := h.logic.DeleteTarget(c.Request.Context(), uintFromParam(c, "id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) PutTargetNodes(c *gin.Context) {
	if !h.authorize(c, "deploy:target:write") {
		return
	}
	var req struct {
		Nodes []TargetNodeReq `json:"nodes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	if err := h.logic.ReplaceTargetNodes(c.Request.Context(), uintFromParam(c, "id"), req.Nodes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	resp, _ := h.logic.GetTarget(c.Request.Context(), uintFromParam(c, "id"))
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) PreviewRelease(c *gin.Context) {
	var req ReleasePreviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	target, terr := h.logic.GetTarget(c.Request.Context(), req.TargetID)
	if terr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": terr.Error()})
		return
	}
	if !h.authorize(c, "deploy:release:apply") || !h.authorizeRuntime(c, target.RuntimeType, "apply") {
		return
	}
	resp, err := h.logic.PreviewRelease(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) ApplyRelease(c *gin.Context) {
	var req ReleasePreviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	target, terr := h.logic.GetTarget(c.Request.Context(), req.TargetID)
	if terr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": terr.Error()})
		return
	}
	if !h.authorize(c, "deploy:release:apply") || !h.authorizeRuntime(c, target.RuntimeType, "apply") {
		return
	}
	uid, _ := c.Get("uid")
	resp, err := h.logic.ApplyRelease(c.Request.Context(), toUint(uid), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) RollbackRelease(c *gin.Context) {
	row, err := h.logic.GetRelease(c.Request.Context(), uintFromParam(c, "id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	if !h.authorize(c, "deploy:release:rollback") || !h.authorizeRuntime(c, row.RuntimeType, "rollback") {
		return
	}
	uid, _ := c.Get("uid")
	resp, err := h.logic.RollbackRelease(c.Request.Context(), uintFromParam(c, "id"), toUint(uid))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) ApproveRelease(c *gin.Context) {
	row, err := h.logic.GetRelease(c.Request.Context(), uintFromParam(c, "id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	if !h.authorize(c, "deploy:release:approve", "deploy:release:apply") || !h.authorizeRuntime(c, row.RuntimeType, "apply") {
		return
	}
	var req ReleaseDecisionReq
	_ = c.ShouldBindJSON(&req)
	uid, _ := c.Get("uid")
	resp, err := h.logic.ApproveRelease(c.Request.Context(), row.ID, toUint(uid), req.Comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) RejectRelease(c *gin.Context) {
	row, err := h.logic.GetRelease(c.Request.Context(), uintFromParam(c, "id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	if !h.authorize(c, "deploy:release:approve", "deploy:release:apply") || !h.authorizeRuntime(c, row.RuntimeType, "apply") {
		return
	}
	var req ReleaseDecisionReq
	_ = c.ShouldBindJSON(&req)
	uid, _ := c.Get("uid")
	resp, err := h.logic.RejectRelease(c.Request.Context(), row.ID, toUint(uid), req.Comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) ListReleaseTimeline(c *gin.Context) {
	row, err := h.logic.GetRelease(c.Request.Context(), uintFromParam(c, "id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	if !h.authorize(c, "deploy:release:read") || !h.authorizeRuntime(c, row.RuntimeType, "read") {
		return
	}
	list, err := h.logic.ListReleaseTimeline(c.Request.Context(), row.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": list, "total": len(list)}})
}

func (h *Handler) ListReleases(c *gin.Context) {
	runtime := strings.TrimSpace(c.Query("runtime_type"))
	if !h.authorize(c, "deploy:release:read") {
		return
	}
	if runtime != "" && !h.authorizeRuntime(c, runtime, "read") {
		return
	}
	rows, err := h.logic.ListReleases(c.Request.Context(), uintFromQuery(c, "service_id"), uintFromQuery(c, "target_id"), runtime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": rows, "total": len(rows)}})
}

func (h *Handler) GetRelease(c *gin.Context) {
	row, err := h.logic.GetRelease(c.Request.Context(), uintFromParam(c, "id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	if !h.authorize(c, "deploy:release:read") || !h.authorizeRuntime(c, row.RuntimeType, "read") {
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) GetGovernance(c *gin.Context) {
	if !h.authorize(c, "service:governance:read", "service:read") {
		return
	}
	row, err := h.logic.GetGovernance(c.Request.Context(), uintFromParam(c, "id"), strings.TrimSpace(c.Query("env")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) PutGovernance(c *gin.Context) {
	if !h.authorize(c, "service:governance:write", "service:write") {
		return
	}
	var req GovernanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid, _ := c.Get("uid")
	row, err := h.logic.UpsertGovernance(c.Request.Context(), toUint(uid), uintFromParam(c, "id"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) PreviewClusterBootstrap(c *gin.Context) {
	if !h.authorize(c, "deploy:target:write") {
		return
	}
	var req ClusterBootstrapPreviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	resp, err := h.logic.PreviewClusterBootstrap(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) ApplyClusterBootstrap(c *gin.Context) {
	if !h.authorize(c, "deploy:target:write") {
		return
	}
	var req ClusterBootstrapPreviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid, _ := c.Get("uid")
	resp, err := h.logic.ApplyClusterBootstrap(c.Request.Context(), toUint(uid), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) GetClusterBootstrapTask(c *gin.Context) {
	if !h.authorize(c, "deploy:target:read") {
		return
	}
	task, err := h.logic.GetClusterBootstrapTask(c.Request.Context(), strings.TrimSpace(c.Param("task_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": task})
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
		if r.Code == code || r.Code == "*:*" || r.Code == "deploy:*" || (strings.HasSuffix(code, ":read") && r.Code == "deploy:*") {
			return true
		}
	}
	return false
}

func (h *Handler) authorizeRuntime(c *gin.Context, runtime, action string) bool {
	r := strings.TrimSpace(runtime)
	if r == "" {
		return true
	}
	code := fmt.Sprintf("deploy:%s:%s", r, action)
	return h.authorize(c, code)
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

func uintFromParam(c *gin.Context, key string) uint {
	v, _ := strconv.ParseUint(c.Param(key), 10, 64)
	return uint(v)
}

func uintFromQuery(c *gin.Context, key string) uint {
	v, _ := strconv.ParseUint(strings.TrimSpace(c.Query(key)), 10, 64)
	return uint(v)
}
