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
		g.GET("/metrics", h.GetMetrics)
	}
}
