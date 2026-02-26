package automation

import (
	"github.com/cy77cc/k8s-manage/internal/middleware"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterAutomationHandlers(v1 *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := NewHandler(svcCtx)
	g := v1.Group("/automation", middleware.JWTAuth())
	{
		g.GET("/inventories", h.ListInventories)
		g.POST("/inventories", h.CreateInventory)
		g.GET("/playbooks", h.ListPlaybooks)
		g.POST("/playbooks", h.CreatePlaybook)
		g.POST("/runs/preview", h.PreviewRun)
		g.POST("/runs/execute", h.ExecuteRun)
		g.GET("/runs/:id", h.GetRun)
		g.GET("/runs/:id/logs", h.GetRunLogs)
	}
}
