package catalog

import (
	"context"
	"log"

	"github.com/cy77cc/k8s-manage/internal/middleware"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterCatalogHandlers(v1 *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := NewHandler(svcCtx)
	if err := h.logic.InitPresetCategories(context.Background()); err != nil {
		log.Printf("catalog preset categories init failed: %v", err)
	}
	g := v1.Group("/catalog", middleware.JWTAuth())
	{
		g.GET("/categories", h.ListCategories)
		g.POST("/categories", h.CreateCategory)
		g.PUT("/categories/:id", h.UpdateCategory)
		g.DELETE("/categories/:id", h.DeleteCategory)

		g.GET("/templates", h.ListTemplates)
		g.GET("/templates/:id", h.GetTemplate)
		g.POST("/templates", h.CreateTemplate)
		g.PUT("/templates/:id", h.UpdateTemplate)
		g.DELETE("/templates/:id", h.DeleteTemplate)
		g.POST("/templates/:id/submit", h.SubmitTemplate)
		g.POST("/templates/:id/publish", h.PublishTemplate)
		g.POST("/templates/:id/reject", h.RejectTemplate)

		g.POST("/preview", h.Preview)
		g.POST("/deploy", h.Deploy)
	}
}
