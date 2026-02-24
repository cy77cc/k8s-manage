package ai

import (
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

func (h *handler) listSessions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": h.store.listSessions()})
}

func (h *handler) getSession(c *gin.Context) {
	session, ok := h.store.getSession(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "session not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": session})
}

func (h *handler) deleteSession(c *gin.Context) {
	h.store.deleteSession(c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"success": true, "data": nil})
}

func (h *handler) analyze(c *gin.Context) {
	var req map[string]any
	_ = c.ShouldBindJSON(&req)
	summary := "MVP阶段分析能力已启用"
	if h.svcCtx.AI != nil && h.svcCtx.AI.Model != nil {
		msg, err := h.svcCtx.AI.Model.Generate(c.Request.Context(), []*schema.Message{
			schema.UserMessage("请根据输入生成简短运维分析摘要，最多120字：" + toString(req)),
		})
		if err == nil && msg != nil && strings.TrimSpace(msg.Content) != "" {
			summary = msg.Content
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"id":        "ana-" + strconvFormatInt(time.Now().UnixNano()),
		"type":      "generic",
		"title":     "AI 分析结果",
		"summary":   summary,
		"details":   req,
		"createdAt": time.Now(),
	}})
}

func (h *handler) recommendations(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": []gin.H{{
		"id":        "rec-" + strconvFormatInt(time.Now().UnixNano()),
		"type":      "suggestion",
		"title":     "建议 #1",
		"content":   "建议先观察资源使用，再执行变更。",
		"relevance": 0.8,
	}}})
}

func (h *handler) k8sAnalyze(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"insights": []string{"建议优先检查异常 Pod 的重启次数和事件。"},
		"risks":    []string{"高峰时段直接变更副本可能引发抖动。"},
	}})
}

func (h *handler) actionPreview(c *gin.Context) {
	var req struct {
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
	}
	_ = c.ShouldBindJSON(&req)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"approval_token": "approve-" + strconvFormatInt(time.Now().UnixNano()),
		"intent":         req.Action,
		"risk":           "medium",
		"params":         req.Params,
		"previewDiff":    "MVP preview",
	}})
}

func (h *handler) actionExecute(c *gin.Context) {
	var req struct {
		ApprovalToken string `json:"approval_token"`
	}
	_ = c.ShouldBindJSON(&req)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"approval_token": req.ApprovalToken,
		"status":         "executed",
	}})
}
