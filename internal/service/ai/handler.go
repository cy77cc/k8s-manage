package ai

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

type chatRequest struct {
	SessionID string         `json:"sessionId"`
	Message   string         `json:"message" binding:"required"`
	Context   map[string]any `json:"context"`
}

type aiSession struct {
	ID        string           `json:"id"`
	Title     string           `json:"title"`
	Messages  []map[string]any `json:"messages"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

var (
	sessionMu sync.Mutex
	sessions  = map[string]*aiSession{}
)

func RegisterAIHandlers(v1 *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	g := v1.Group("/ai")
	g.POST("/chat", chatHandler(svcCtx))
	g.GET("/sessions", listSessionsHandler())
	g.GET("/sessions/:id", getSessionHandler())
	g.DELETE("/sessions/:id", deleteSessionHandler())
	g.POST("/analyze", analyzeHandler())
	g.POST("/recommendations", recommendationsHandler())
	g.POST("/k8s/analyze", k8sAnalyzeHandler())
	g.POST("/k8s/actions/preview", actionPreviewHandler())
	g.POST("/k8s/actions/execute", actionExecuteHandler())
	g.POST("/actions/preview", actionPreviewHandler())
	g.POST("/actions/execute", actionExecuteHandler())
}

func chatHandler(_ *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req chatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
			return
		}
		respText := "已收到你的问题：" + req.Message
		sessionMu.Lock()
		defer sessionMu.Unlock()
		sid := req.SessionID
		if sid == "" {
			sid = fmt.Sprintf("sess-%d", time.Now().UnixNano())
		}
		s, ok := sessions[sid]
		if !ok {
			s = &aiSession{ID: sid, Title: "AI Session", CreatedAt: time.Now()}
			sessions[sid] = s
		}
		now := time.Now()
		s.Messages = append(s.Messages, map[string]any{"id": fmt.Sprintf("u-%d", now.UnixNano()), "role": "user", "content": req.Message, "timestamp": now}, map[string]any{"id": fmt.Sprintf("a-%d", now.UnixNano()+1), "role": "assistant", "content": respText, "timestamp": now})
		s.UpdatedAt = now
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"session": s, "response": gin.H{"id": fmt.Sprintf("a-%d", now.UnixNano()+1), "role": "assistant", "content": respText, "timestamp": now}}})
	}
}

func listSessionsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionMu.Lock()
		defer sessionMu.Unlock()
		out := make([]*aiSession, 0, len(sessions))
		for _, s := range sessions {
			out = append(out, s)
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": out})
	}
}

func getSessionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionMu.Lock()
		defer sessionMu.Unlock()
		s, ok := sessions[c.Param("id")]
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "session not found"}})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": s})
	}
}

func deleteSessionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionMu.Lock()
		defer sessionMu.Unlock()
		delete(sessions, c.Param("id"))
		c.JSON(http.StatusOK, gin.H{"success": true, "data": nil})
	}
}

func analyzeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": fmt.Sprintf("ana-%d", time.Now().UnixNano()), "type": "generic", "title": "AI 分析结果", "summary": "MVP阶段分析能力已启用", "details": gin.H{}, "createdAt": time.Now()}})
	}
}

func recommendationsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": []gin.H{{"id": fmt.Sprintf("rec-%d", time.Now().UnixNano()), "type": "suggestion", "title": "建议 #1", "content": "建议先观察资源使用，再执行变更。", "relevance": 0.8, "action": "k8s.scale.deployment", "params": gin.H{"replicas": 2}}}})
	}
}

func k8sAnalyzeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"insights": []string{"建议优先检查异常 Pod 的重启次数和事件。", "建议确认 Deployment 副本数与实际运行数是否一致。"}, "risks": []string{"高峰时段直接变更副本可能引发抖动。"}, "recommended_actions": []gin.H{{"action": "k8s.scale.deployment", "params": gin.H{"replicas": 2}, "reason": "优先验证扩缩容链路"}}}})
	}
}

func actionPreviewHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Action string         `json:"action"`
			Params map[string]any `json:"params"`
		}
		_ = c.ShouldBindJSON(&req)
		token := fmt.Sprintf("approve-%d", time.Now().UnixNano())
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"approval_token": token, "intent": req.Action, "risk": "medium", "params": req.Params, "previewDiff": "MVP preview"}})
	}
}

func actionExecuteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ApprovalToken string `json:"approval_token"`
		}
		_ = c.ShouldBindJSON(&req)
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"approval_token": req.ApprovalToken, "status": "executed"}})
	}
}
