package ai

import (
	"github.com/cy77cc/k8s-manage/internal/middleware"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterAIHandlers(v1 *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := newHandler(svcCtx)
	g := v1.Group("/ai", middleware.JWTAuth())
	{
		g.POST("/chat", h.chat)
		g.GET("/capabilities", h.capabilities)
		g.POST("/tools/preview", h.previewTool)
		g.POST("/tools/execute", h.executeTool)
		g.GET("/executions/:id", h.getExecution)
		g.POST("/approvals", h.createApproval)
		g.POST("/approvals/:id/confirm", h.confirmApproval)
		g.GET("/sessions", h.listSessions)
		g.GET("/sessions/current", h.currentSession)
		g.GET("/sessions/:id", h.getSession)
		g.DELETE("/sessions/:id", h.deleteSession)
		g.POST("/analyze", h.analyze)
		g.POST("/recommendations", h.recommendations)
		g.POST("/k8s/analyze", h.k8sAnalyze)
		g.POST("/k8s/actions/preview", h.actionPreview)
		g.POST("/k8s/actions/execute", h.actionExecute)
		g.POST("/actions/preview", h.actionPreview)
		g.POST("/actions/execute", h.actionExecute)
		g.GET("/commands/suggestions", h.commandSuggestions)
		g.POST("/commands/preview", h.previewCommand)
		g.POST("/commands/execute", h.executeCommand)
		g.GET("/commands/history", h.listCommandHistory)
		g.GET("/commands/history/:id", h.getCommandHistory)
	}
}
