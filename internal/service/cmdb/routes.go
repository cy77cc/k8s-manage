package cmdb

import (
	"github.com/cy77cc/k8s-manage/internal/middleware"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterCMDBHandlers(v1 *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := NewHandler(svcCtx)
	g := v1.Group("/cmdb", middleware.JWTAuth())
	{
		g.GET("/assets", h.ListAssets)
		g.POST("/assets", h.CreateAsset)
		g.GET("/assets/:id", h.GetAsset)
		g.PUT("/assets/:id", h.UpdateAsset)
		g.DELETE("/assets/:id", h.DeleteAsset)

		g.GET("/relations", h.ListRelations)
		g.POST("/relations", h.CreateRelation)
		g.DELETE("/relations/:id", h.DeleteRelation)

		g.GET("/topology", h.Topology)

		g.POST("/sync/jobs", h.TriggerSync)
		g.GET("/sync/jobs/:id", h.GetSyncJob)
		g.POST("/sync/jobs/:id/retry", h.RetrySyncJob)

		g.GET("/changes", h.ListChanges)
		g.GET("/audits", h.ListAudits)
	}
}
