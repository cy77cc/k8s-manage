package deployment

import (
	"github.com/cy77cc/k8s-manage/internal/middleware"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterDeploymentHandlers(v1 *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := NewHandler(svcCtx)
	g := v1.Group("/deploy", middleware.JWTAuth())
	{
		g.GET("/targets", h.ListTargets)
		g.POST("/targets", h.CreateTarget)
		g.GET("/targets/:id", h.GetTarget)
		g.PUT("/targets/:id", h.UpdateTarget)
		g.DELETE("/targets/:id", h.DeleteTarget)
		g.PUT("/targets/:id/nodes", h.PutTargetNodes)

		g.POST("/releases/preview", h.PreviewRelease)
		g.POST("/releases/apply", h.ApplyRelease)
		g.POST("/releases/:id/approve", h.ApproveRelease)
		g.POST("/releases/:id/reject", h.RejectRelease)
		g.POST("/releases/:id/rollback", h.RollbackRelease)
		g.GET("/releases", h.ListReleases)
		g.GET("/releases/:id", h.GetRelease)
		g.GET("/releases/:id/timeline", h.ListReleaseTimeline)

		g.POST("/clusters/bootstrap/preview", h.PreviewClusterBootstrap)
		g.POST("/clusters/bootstrap/apply", h.ApplyClusterBootstrap)
		g.GET("/clusters/bootstrap/:task_id", h.GetClusterBootstrapTask)
	}

	sg := v1.Group("/services", middleware.JWTAuth())
	{
		sg.GET("/:id/governance", h.GetGovernance)
		sg.PUT("/:id/governance", h.PutGovernance)
	}
}
