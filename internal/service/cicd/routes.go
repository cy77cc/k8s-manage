package cicd

import (
	"github.com/cy77cc/k8s-manage/internal/middleware"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterCICDHandlers(v1 *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := NewHandler(svcCtx)
	g := v1.Group("/cicd", middleware.JWTAuth())
	{
		g.GET("/services/:service_id/ci-config", h.GetServiceCIConfig)
		g.PUT("/services/:service_id/ci-config", h.PutServiceCIConfig)
		g.DELETE("/services/:service_id/ci-config", h.DeleteServiceCIConfig)
		g.POST("/services/:service_id/ci-runs/trigger", h.TriggerCIRun)
		g.GET("/services/:service_id/ci-runs", h.ListCIRuns)

		g.GET("/deployments/:deployment_id/cd-config", h.GetDeploymentCDConfig)
		g.PUT("/deployments/:deployment_id/cd-config", h.PutDeploymentCDConfig)

		g.POST("/releases", h.TriggerRelease)
		g.GET("/releases", h.ListReleases)
		g.POST("/releases/:id/approve", h.ApproveRelease)
		g.POST("/releases/:id/reject", h.RejectRelease)
		g.POST("/releases/:id/rollback", h.RollbackRelease)
		g.GET("/releases/:id/approvals", h.ListApprovals)

		g.GET("/services/:service_id/timeline", h.ServiceTimeline)
	}
}
