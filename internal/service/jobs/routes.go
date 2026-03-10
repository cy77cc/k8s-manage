package jobs

import (
	"github.com/cy77cc/OpsPilot/internal/middleware"
	"github.com/cy77cc/OpsPilot/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterJobsHandlers(v1 *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := NewHandler(svcCtx)
	g := v1.Group("/jobs", middleware.JWTAuth())
	{
		g.GET("", h.ListJobs)
		g.POST("", h.CreateJob)
		g.GET("/:id", h.GetJob)
		g.PUT("/:id", h.UpdateJob)
		g.DELETE("/:id", h.DeleteJob)
		g.POST("/:id/start", h.StartJob)
		g.POST("/:id/stop", h.StopJob)
		g.GET("/:id/executions", h.GetJobExecutions)
		g.GET("/:id/logs", h.GetJobLogs)
	}
}
