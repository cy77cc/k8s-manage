package monitoring

import (
	"github.com/cy77cc/k8s-manage/internal/middleware"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterMonitoringHandlers(v1 *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := NewHandler(svcCtx)
	h.StartCollector()
	g := v1.Group("", middleware.JWTAuth())
	{
		g.GET("/alerts", h.ListAlerts)
		g.GET("/alert-rules", h.ListRules)
		g.POST("/alert-rules", h.CreateRule)
		g.PUT("/alert-rules/:id", h.UpdateRule)
		g.POST("/alert-rules/:id/enable", h.EnableRule)
		g.POST("/alert-rules/:id/disable", h.DisableRule)
		g.GET("/alert-rules/:id/evaluations", h.ListRuleEvaluations)
		g.GET("/metrics", h.GetMetrics)
		g.GET("/alert-channels", h.ListChannels)
		g.POST("/alert-channels", h.CreateChannel)
		g.PUT("/alert-channels/:id", h.UpdateChannel)
		g.GET("/alert-deliveries", h.ListDeliveries)
	}
}
