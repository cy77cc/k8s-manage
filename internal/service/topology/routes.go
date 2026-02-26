package topology

import (
	"github.com/cy77cc/k8s-manage/internal/middleware"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterTopologyHandlers(v1 *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := NewHandler(svcCtx)
	g := v1.Group("/topology", middleware.JWTAuth())
	{
		g.GET("/services/:id", h.ServiceTopology)
		g.GET("/hosts/:id/services", h.HostServices)
		g.GET("/clusters/:id/services", h.ClusterServices)
		g.GET("/graph", h.Graph)
	}
}
