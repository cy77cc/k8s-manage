package topology

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

func (h *Handler) ServiceTopology(c *gin.Context) {
	if !h.authorize(c, "topology:read", "cmdb:read", "cmdb:*") {
		return
	}
	serviceID := uint(atoiDefault(c.Param("id"), 0))
	out, err := h.logic.getServiceTopology(c.Request.Context(), serviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 3004, "msg": "service not found"})
		return
	}
	h.logic.writeAccessAudit(c.Request.Context(), h.uidFromContext(c), "topology.service", "service", map[string]any{"service_id": serviceID})
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": out})
}

func (h *Handler) HostServices(c *gin.Context) {
	if !h.authorize(c, "topology:read", "cmdb:read", "cmdb:*") {
		return
	}
	hostID := uint(atoiDefault(c.Param("id"), 0))
	out, err := h.logic.getHostServices(c.Request.Context(), hostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	h.logic.writeAccessAudit(c.Request.Context(), h.uidFromContext(c), "topology.host.services", "host", map[string]any{"host_id": hostID})
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": out, "total": len(out)})
}

func (h *Handler) ClusterServices(c *gin.Context) {
	if !h.authorize(c, "topology:read", "cmdb:read", "cmdb:*") {
		return
	}
	clusterID := uint(atoiDefault(c.Param("id"), 0))
	out, err := h.logic.getClusterServices(c.Request.Context(), clusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	h.logic.writeAccessAudit(c.Request.Context(), h.uidFromContext(c), "topology.cluster.services", "cluster", map[string]any{"cluster_id": clusterID})
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": out, "total": len(out)})
}

func (h *Handler) Graph(c *gin.Context) {
	if !h.authorize(c, "topology:read", "cmdb:read", "cmdb:*") {
		return
	}
	filter := QueryFilter{
		ProjectID:    uint(atoiDefault(c.Query("project_id"), 0)),
		ClusterID:    uint(atoiDefault(c.Query("cluster_id"), 0)),
		ResourceType: strings.TrimSpace(c.Query("resource_type")),
		Keyword:      strings.TrimSpace(c.Query("keyword")),
	}
	out, err := h.logic.queryGraph(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	h.logic.writeAccessAudit(c.Request.Context(), h.uidFromContext(c), "topology.graph.query", "graph", filter)
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": out})
}

func (h *Handler) authorize(c *gin.Context, codes ...string) bool {
	if h.isAdmin(c) {
		return true
	}
	uid := h.uidFromContext(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized"})
		return false
	}
	var rows []struct {
		Code string `gorm:"column:code"`
	}
	if err := h.svcCtx.DB.Table("permissions").
		Select("permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", uid).
		Scan(&rows).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden"})
		return false
	}
	for _, r := range rows {
		for _, code := range codes {
			if r.Code == code || r.Code == "*:*" || r.Code == "topology:*" || r.Code == "cmdb:*" {
				return true
			}
		}
	}
	c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden"})
	return false
}

func (h *Handler) isAdmin(c *gin.Context) bool {
	uid := h.uidFromContext(c)
	if uid == 0 {
		return false
	}
	var user model.User
	if err := h.svcCtx.DB.Select("id,username").Where("id = ?", uid).First(&user).Error; err != nil {
		return false
	}
	return strings.EqualFold(user.Username, "admin")
}

func (h *Handler) uidFromContext(c *gin.Context) uint {
	v, ok := c.Get("uid")
	if !ok {
		return 0
	}
	switch x := v.(type) {
	case uint:
		return x
	case uint64:
		return uint(x)
	case int:
		if x < 0 {
			return 0
		}
		return uint(x)
	case int64:
		if x < 0 {
			return 0
		}
		return uint(x)
	case string:
		n, _ := strconv.ParseUint(strings.TrimSpace(x), 10, 64)
		return uint(n)
	default:
		return 0
	}
}

func atoiDefault(s string, d int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return d
	}
	return n
}
