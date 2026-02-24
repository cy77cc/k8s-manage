package rbac

import (
	"github.com/cy77cc/k8s-manage/internal/middleware"
	rbachandler "github.com/cy77cc/k8s-manage/internal/service/rbac/handler"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterRBACHandlers(v1 *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := rbachandler.NewHandler(svcCtx)
	g := v1.Group("/rbac", middleware.JWTAuth())
	{
		g.GET("/me/permissions", h.MyPermissions)
		g.POST("/check", h.Check)
		g.GET("/users", h.ListUsers)
		g.GET("/users/:id", h.GetUser)
		g.POST("/users", h.CreateUser)
		g.PUT("/users/:id", h.UpdateUser)
		g.DELETE("/users/:id", h.DeleteUser)
		g.GET("/roles", h.ListRoles)
		g.GET("/roles/:id", h.GetRole)
		g.POST("/roles", h.CreateRole)
		g.PUT("/roles/:id", h.UpdateRole)
		g.DELETE("/roles/:id", h.DeleteRole)
		g.GET("/permissions", h.ListPermissions)
		g.GET("/permissions/:id", h.GetPermission)
	}
}
