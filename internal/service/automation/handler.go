package automation

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

func (h *Handler) ListInventories(c *gin.Context) {
	if !h.authorize(c, "automation:read", "automation:*") {
		return
	}
	rows, err := h.logic.listInventories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rows, "total": len(rows)})
}

func (h *Handler) CreateInventory(c *gin.Context) {
	if !h.authorize(c, "automation:write", "automation:*") {
		return
	}
	var req createInventoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	row, err := h.logic.createInventory(c.Request.Context(), h.uidFromContext(c), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) ListPlaybooks(c *gin.Context) {
	if !h.authorize(c, "automation:read", "automation:*") {
		return
	}
	rows, err := h.logic.listPlaybooks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rows, "total": len(rows)})
}

func (h *Handler) CreatePlaybook(c *gin.Context) {
	if !h.authorize(c, "automation:write", "automation:*") {
		return
	}
	var req createPlaybookReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	row, err := h.logic.createPlaybook(c.Request.Context(), h.uidFromContext(c), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) PreviewRun(c *gin.Context) {
	if !h.authorize(c, "automation:read", "automation:*") {
		return
	}
	var req previewRunReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	out, err := h.logic.previewRun(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": out})
}

func (h *Handler) ExecuteRun(c *gin.Context) {
	// Mutating action gate.
	if !h.authorize(c, "automation:execute", "automation:write", "automation:*") {
		return
	}
	var req executeRunReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	row, err := h.logic.executeRun(c.Request.Context(), h.uidFromContext(c), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) GetRun(c *gin.Context) {
	if !h.authorize(c, "automation:read", "automation:*") {
		return
	}
	row, err := h.logic.getRun(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 3004, "msg": "run not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) GetRunLogs(c *gin.Context) {
	if !h.authorize(c, "automation:read", "automation:*") {
		return
	}
	rows, err := h.logic.listRunLogs(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rows, "total": len(rows)})
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
			if r.Code == code || r.Code == "*:*" || r.Code == "automation:*" {
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
