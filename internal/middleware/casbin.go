package middleware

import (
	"fmt"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/cy77cc/k8s-manage/internal/response"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

// CasbinAuth returns a middleware that enforces Casbin authorization.
func CasbinAuth(enforcer *casbin.Enforcer) gin.HandlerFunc {
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
		// obj: request path
		// act: request method
		sub := fmt.Sprintf("%v", uid)
		obj := c.Request.URL.Path
		act := c.Request.Method

		// 4. Enforce
		ok, err := enforcer.Enforce(sub, obj, act)
		if err != nil {
			resp := response.NewResp(xcode.ServerError, "权限验证错误", nil)
			c.JSON(http.StatusInternalServerError, resp)
			c.Abort()
			return
		}

		if !ok {
			resp := response.NewResp(xcode.Forbidden, "无权限访问该资源", nil)
			c.JSON(http.StatusForbidden, resp)
			c.Abort()
			return
		}

		c.Next()
	}
}
