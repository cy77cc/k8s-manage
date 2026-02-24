package cluster

import (
	"github.com/cy77cc/k8s-manage/internal/middleware"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterClusterHandlers(v1 *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := NewHandler(svcCtx)
	g := v1.Group("/clusters", middleware.JWTAuth())
	{
		g.GET("", h.List)
		g.POST("", h.Create)
		g.GET("/:id", h.Get)
		g.GET("/:id/nodes", h.Nodes)
		g.GET("/:id/deployments", h.Deployments)
		g.GET("/:id/pods", h.Pods)
		g.GET("/:id/services", h.Services)
		g.GET("/:id/ingresses", h.Ingresses)
		g.GET("/:id/events", h.Events)
		g.GET("/:id/logs", h.Logs)
		g.POST("/:id/connect/test", h.ConnectTest)
		g.POST("/:id/deploy/preview", h.DeployPreview)
		g.POST("/:id/deploy/apply", h.DeployApply)
	}
}
