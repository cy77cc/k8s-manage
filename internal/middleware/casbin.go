package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/cy77cc/k8s-manage/internal/response"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

// CasbinAuth returns a middleware that enforces Casbin authorization using a specific permission code.
func CasbinAuth(enforcer *casbin.Enforcer, permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Check if enforcer is initialized
		if enforcer == nil {
			resp := response.NewResp(xcode.ServerError, "权限服务不可用", nil)
			c.JSON(http.StatusInternalServerError, resp)
			c.Abort()
			return
		}

		// 2. Get User ID from Context (set by JWT)
		uid, exists := c.Get("uid")
		if !exists {
			resp := response.NewResp(xcode.Unauthorized, "未登录或Token无效", nil)
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
			return
		}

		// 3. Prepare Enforce parameters
		// sub: user id (string)
		// obj: permission code (e.g., "user:add")
		sub := fmt.Sprintf("%v", uid)
		obj := permissionCode

		if isPrivilegedSubject(enforcer, sub) {
			c.Next()
			return
		}

		// 4. Enforce
		ok, err := enforcer.Enforce(sub, obj)
		if err != nil {
			resp := response.NewResp(xcode.ServerError, "权限验证错误", nil)
			c.JSON(http.StatusInternalServerError, resp)
			c.Abort()
			return
		}

		if !ok {
			auditAccessDenied(c, sub, obj)
			resp := response.NewResp(xcode.Forbidden, "无权限访问该资源", gin.H{
				"code": "RBAC_FORBIDDEN",
			})
			c.JSON(http.StatusForbidden, resp)
			c.Abort()
			return
		}

		c.Next()
	}
}

func auditAccessDenied(c *gin.Context, actor, action string) {
	resource := c.FullPath()
	if resource == "" {
		resource = c.Request.URL.Path
	}
	c.Set("rbac_deny_audit", gin.H{
		"actor":     actor,
		"resource":  resource,
		"action":    action,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	log.Printf("rbac deny actor=%s resource=%s action=%s timestamp=%s", actor, resource, action, time.Now().UTC().Format(time.RFC3339))
}

func isPrivilegedSubject(enforcer *casbin.Enforcer, subject string) bool {
	roles, err := enforcer.GetRolesForUser(subject)
	if err != nil {
		return false
	}
	for _, role := range roles {
		normalized := strings.ToLower(strings.TrimSpace(role))
		if normalized == "admin" || normalized == "super-admin" || normalized == "super_admin" || normalized == "root" || normalized == "超级管理员" {
			return true
		}
	}
	return false
}
