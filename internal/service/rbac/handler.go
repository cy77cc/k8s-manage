package rbac

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svcCtx *svc.ServiceContext }

func NewHandler(svcCtx *svc.ServiceContext) *Handler { return &Handler{svcCtx: svcCtx} }

func (h *Handler) MyPermissions(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": gin.H{"message": "unauthorized"}})
		return
	}
	userID := toUint64(uid)
	perms, _ := h.fetchPermissionsByUserID(userID)
	if h.isAdminUser(userID) {
		perms = mergePermissions(perms, adminPermissionSet()...)
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": perms})
}

func (h *Handler) Check(c *gin.Context) {
	var req struct{ Resource, Action string }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	code := req.Resource + ":" + req.Action
	uid, _ := c.Get("uid")
	userID := toUint64(uid)
	perms, _ := h.fetchPermissionsByUserID(userID)
	if h.isAdminUser(userID) {
		perms = mergePermissions(perms, adminPermissionSet()...)
	}
	has := hasPermission(perms, code, req.Resource)
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"hasPermission": has}})
}

func (h *Handler) ListUsers(c *gin.Context) {
	var users []model.User
	if err := h.svcCtx.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	list := make([]gin.H, 0, len(users))
	for _, u := range users {
		list = append(list, gin.H{"id": u.ID, "username": u.Username, "name": u.Username, "email": u.Email, "roles": []string{}, "status": "active", "createdAt": time.Unix(u.CreateTime, 0), "updatedAt": time.Unix(u.UpdateTime, 0)})
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": list, "total": len(list)})
}

func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid id"}})
		return
	}
	var u model.User
	if err := h.svcCtx.DB.First(&u, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "user not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"id": u.ID, "username": u.Username, "name": u.Username, "email": u.Email, "roles": []string{}, "status": "active", "createdAt": time.Unix(u.CreateTime, 0), "updatedAt": time.Unix(u.UpdateTime, 0)}})
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req struct {
		Username, Name, Email, Password string
		Roles                           []string
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	now := time.Now().Unix()
	u := model.User{Username: req.Username, PasswordHash: req.Password, Email: req.Email, CreateTime: now, UpdateTime: now, Status: 1}
	if err := h.svcCtx.DB.Create(&u).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"id": u.ID, "username": u.Username, "name": u.Username, "email": u.Email, "roles": req.Roles, "status": "active", "createdAt": time.Unix(u.CreateTime, 0), "updatedAt": time.Unix(u.UpdateTime, 0)}})
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid id"}})
		return
	}
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	delete(req, "roles")
	req["update_time"] = time.Now().Unix()
	if err := h.svcCtx.DB.Model(&model.User{}).Where("id = ?", id).Updates(req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	h.GetUser(c)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid id"}})
		return
	}
	if err := h.svcCtx.DB.Delete(&model.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) ListRoles(c *gin.Context) {
	var roles []model.Role
	if err := h.svcCtx.DB.Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	list := make([]gin.H, 0, len(roles))
	for _, r := range roles {
		list = append(list, gin.H{"id": r.ID, "name": r.Name, "description": r.Description, "permissions": []string{}, "createdAt": time.Unix(r.CreateTime, 0), "updatedAt": time.Unix(r.UpdateTime, 0)})
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": list, "total": len(list)})
}

func (h *Handler) GetRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid id"}})
		return
	}
	var r model.Role
	if err := h.svcCtx.DB.First(&r, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "role not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"id": r.ID, "name": r.Name, "description": r.Description, "permissions": []string{}, "createdAt": time.Unix(r.CreateTime, 0), "updatedAt": time.Unix(r.UpdateTime, 0)}})
}

func (h *Handler) CreateRole(c *gin.Context) {
	var req struct {
		Name, Description string
		Permissions       []string
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	now := time.Now().Unix()
	r := model.Role{Name: req.Name, Code: req.Name, Description: req.Description, Status: 1, CreateTime: now, UpdateTime: now}
	if err := h.svcCtx.DB.Create(&r).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"id": r.ID, "name": r.Name, "description": r.Description, "permissions": req.Permissions, "createdAt": time.Unix(r.CreateTime, 0), "updatedAt": time.Unix(r.UpdateTime, 0)}})
}

func (h *Handler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid id"}})
		return
	}
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	delete(req, "permissions")
	req["update_time"] = time.Now().Unix()
	if err := h.svcCtx.DB.Model(&model.Role{}).Where("id = ?", id).Updates(req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	h.GetRole(c)
}

func (h *Handler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid id"}})
		return
	}
	if err := h.svcCtx.DB.Delete(&model.Role{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) ListPermissions(c *gin.Context) {
	var permissions []model.Permission
	if err := h.svcCtx.DB.Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	list := make([]gin.H, 0, len(permissions))
	for _, p := range permissions {
		list = append(list, gin.H{"id": p.ID, "name": p.Name, "code": p.Code, "description": p.Description, "category": p.Resource, "createdAt": time.Unix(p.CreateTime, 0)})
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": list, "total": len(list)})
}

func (h *Handler) GetPermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid id"}})
		return
	}
	var p model.Permission
	if err := h.svcCtx.DB.First(&p, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "permission not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"id": p.ID, "name": p.Name, "code": p.Code, "description": p.Description, "category": p.Resource, "createdAt": time.Unix(p.CreateTime, 0)}})
}

func (h *Handler) fetchPermissionsByUserID(userID uint64) ([]string, error) {
	type row struct {
		Code string `gorm:"column:code"`
	}
	var rows []row
	err := h.svcCtx.DB.Table("permissions").Select("permissions.code").Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").Where("user_roles.user_id = ?", userID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Code)
	}
	return out, nil
}

func toUint64(v any) uint64 {
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

func (h *Handler) isAdminUser(userID uint64) bool {
	if userID == 0 {
		return false
	}

	var u model.User
	if err := h.svcCtx.DB.Select("id", "username").Where("id = ?", userID).First(&u).Error; err == nil {
		if strings.EqualFold(strings.TrimSpace(u.Username), "admin") {
			return true
		}
	}

	type roleRow struct {
		Code string `gorm:"column:code"`
	}
	var rows []roleRow
	err := h.svcCtx.DB.Table("roles").
		Select("roles.code").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Scan(&rows).Error
	if err != nil {
		return false
	}
	for _, row := range rows {
		if strings.EqualFold(strings.TrimSpace(row.Code), "admin") {
			return true
		}
	}
	return false
}

func hasPermission(perms []string, code string, resource string) bool {
	resourceWildcard := resource + ":*"
	for _, p := range perms {
		if p == code || p == resourceWildcard || p == "*:*" {
			return true
		}
	}
	return false
}

func mergePermissions(base []string, extras ...string) []string {
	seen := make(map[string]struct{}, len(base)+len(extras))
	merged := make([]string, 0, len(base)+len(extras))
	for _, p := range base {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		merged = append(merged, p)
	}
	for _, p := range extras {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		merged = append(merged, p)
	}
	return merged
}

func adminPermissionSet() []string {
	return []string{
		"*:*",
		"host:read", "host:write", "host:*",
		"task:read", "task:write", "task:*",
		"kubernetes:read", "kubernetes:write", "kubernetes:*",
		"monitoring:read", "monitoring:write", "monitoring:*",
		"config:read", "config:write", "config:*",
		"rbac:read", "rbac:write", "rbac:*",
		"automation:*",
		"cicd:*",
		"cmdb:*",
	}
}
