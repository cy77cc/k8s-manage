package host

import (
	"github.com/cy77cc/k8s-manage/internal/middleware"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterHostHandlers(v1 *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := NewHandler(svcCtx)
	g := v1.Group("/hosts", middleware.JWTAuth())
	{
		g.GET("", h.List)
		g.POST("", h.Create)
		g.POST("/batch", h.Batch)
		g.POST("/batch/exec", h.BatchExec)
		g.GET("/:id", h.Get)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
		g.POST("/:id/actions", h.Action)
		g.POST("/:id/ssh/check", h.SSHCheck)
		g.POST("/:id/ssh/exec", h.SSHExec)
		g.GET("/:id/facts", h.Facts)
		g.GET("/:id/tags", h.Tags)
		g.POST("/:id/tags", h.AddTag)
		g.DELETE("/:id/tags/:tag", h.RemoveTag)
		g.GET("/:id/metrics", h.Metrics)
		g.GET("/:id/audits", h.Audits)
	}
}
