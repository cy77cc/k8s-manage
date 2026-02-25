package aiops

import (
	"github.com/cy77cc/k8s-manage/internal/middleware"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterAIOPSHandlers(v1 *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := NewHandler(svcCtx)
	g := v1.Group("/aiops", middleware.JWTAuth())
	{
		g.POST("/inspections/run", h.RunInspection)
		g.GET("/inspections", h.ListInspections)
		g.GET("/inspections/:id", h.GetInspection)
		g.POST("/recommendations/apply-preview", h.ApplyPreview)
	}
}

