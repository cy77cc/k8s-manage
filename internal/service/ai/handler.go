package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/logger"
	"github.com/cy77cc/k8s-manage/internal/middleware"
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
	ensureControlPlane(svcCtx)
	maybeAddAIPermissions(svcCtx)

	g := v1.Group("/ai", middleware.JWTAuth())
	g.POST("/chat", chatHandler(svcCtx))
	g.GET("/capabilities", capabilitiesHandler(svcCtx))
	g.POST("/tools/preview", previewToolHandler(svcCtx))
	g.POST("/tools/execute", executeToolHandler(svcCtx))
	g.GET("/executions/:id", executionHandler(svcCtx))
	g.POST("/approvals", createApprovalHandler(svcCtx))
	g.POST("/approvals/:id/confirm", confirmApprovalHandler(svcCtx))
	g.GET("/sessions", listSessionsHandler())
	g.GET("/sessions/:id", getSessionHandler())
	g.DELETE("/sessions/:id", deleteSessionHandler())
	g.POST("/analyze", analyzeHandler(svcCtx))
	g.POST("/recommendations", recommendationsHandler(svcCtx))
	g.POST("/k8s/analyze", k8sAnalyzeHandler(svcCtx))
	g.POST("/k8s/actions/preview", actionPreviewHandler())
	g.POST("/k8s/actions/execute", actionExecuteHandler())
	g.POST("/actions/preview", actionPreviewHandler())
	g.POST("/actions/execute", actionExecuteHandler())
}

func chatHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req chatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
			return
		}

		msg := strings.TrimSpace(req.Message)
		if msg == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "message is required"}})
			return
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")
		c.Status(http.StatusOK)

		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "streaming not supported"}})
			return
		}

		sid := req.SessionID
		if sid == "" {
			sid = fmt.Sprintf("sess-%d", time.Now().UnixNano())
		}

		userTime := time.Now()
		session := appendMessage(sid, map[string]any{
			"id":        fmt.Sprintf("u-%d", userTime.UnixNano()),
			"role":      "user",
			"content":   msg,
			"timestamp": userTime,
		})
		if !writeSSE(c, flusher, "meta", gin.H{
			"sessionId": session.ID,
			"createdAt": session.CreatedAt,
		}) {
			return
		}

		uid, _ := getUIDFromContext(c)
		if toolName, params, approvalToken, ok := extractToolRequest(msg, req.Context); ok {
			_ = writeSSE(c, flusher, "tool_call", gin.H{"tool": toolName, "params": params})

			cp := ensureControlPlane(svcCtx)
			previewData, err := cp.previewTool(uid, toolName, params)
			if err != nil {
				_ = writeSSE(c, flusher, "error", gin.H{"message": err.Error()})
				return
			}

			needApproval, _ := previewData["approval_required"].(bool)
			if needApproval && strings.TrimSpace(approvalToken) == "" {
				_ = writeSSE(c, flusher, "approval_required", previewData)
				assistantTime := time.Now()
				session = appendMessage(sid, map[string]any{
					"id":        fmt.Sprintf("a-%d", assistantTime.UnixNano()),
					"role":      "assistant",
					"content":   "该操作需要审批，请先确认 approval token 后重试执行。",
					"timestamp": assistantTime,
				})
				_ = writeSSE(c, flusher, "done", gin.H{"session": session})
				return
			}

			execRec, err := cp.executeTool(c.Request.Context(), uid, toolName, params, approvalToken)
			if err != nil {
				_ = writeSSE(c, flusher, "tool_result", gin.H{"execution": execRec})
				_ = writeSSE(c, flusher, "error", gin.H{"message": err.Error()})
				return
			}
			_ = writeSSE(c, flusher, "tool_result", gin.H{"execution": execRec})

			toolText := "工具调用已完成。"
			if execRec.Result != nil {
				toolText = fmt.Sprintf("工具 %s 执行完成，状态：%s。", execRec.Tool, execRec.Status)
			}
			assistantTime := time.Now()
			session = appendMessage(sid, map[string]any{
				"id":        fmt.Sprintf("a-%d", assistantTime.UnixNano()),
				"role":      "assistant",
				"content":   toolText,
				"timestamp": assistantTime,
			})
			_ = writeSSE(c, flusher, "done", gin.H{"session": session})
			return
		}

		var assistantContent strings.Builder
		if err := streamAssistant(c.Request.Context(), svcCtx, msg, req.Context, func(chunk string) bool {
			if chunk == "" {
				return true
			}
			assistantContent.WriteString(chunk)
			return writeSSE(c, flusher, "delta", gin.H{"contentChunk": chunk})
		}); err != nil {
			logger.L().Warn("ai stream failed", logger.Error(err))
			_ = writeSSE(c, flusher, "error", gin.H{"message": err.Error()})
			return
		}

		content := strings.TrimSpace(assistantContent.String())
		if content == "" {
			content = "当前没有可返回的结果，请稍后重试。"
		}
		assistantTime := time.Now()
		session = appendMessage(sid, map[string]any{
			"id":        fmt.Sprintf("a-%d", assistantTime.UnixNano()),
			"role":      "assistant",
			"content":   content,
			"timestamp": assistantTime,
		})

		_ = writeSSE(c, flusher, "done", gin.H{"session": session})
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

func analyzeHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]any
		_ = c.ShouldBindJSON(&req)

		summary := "MVP阶段分析能力已启用"
		dataSource := "fallback"
		if out, err := generateByLLM(c.Request.Context(), svcCtx, "请根据输入生成简短运维分析摘要，最多120字："+mustJSON(req)); err == nil && strings.TrimSpace(out) != "" {
			summary = out
			dataSource = "llm"
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"id":          fmt.Sprintf("ana-%d", time.Now().UnixNano()),
				"type":        "generic",
				"title":       "AI 分析结果",
				"summary":     summary,
				"details":     req,
				"createdAt":   time.Now(),
				"data_source": dataSource,
			},
		})
	}
}

func recommendationsHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]any
		_ = c.ShouldBindJSON(&req)

		content := "建议先观察资源使用，再执行变更。"
		dataSource := "fallback"
		if out, err := generateByLLM(c.Request.Context(), svcCtx, "请基于运维上下文给出一条可执行建议："+mustJSON(req)); err == nil && strings.TrimSpace(out) != "" {
			content = out
			dataSource = "llm"
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": []gin.H{{
				"id":          fmt.Sprintf("rec-%d", time.Now().UnixNano()),
				"type":        "suggestion",
				"title":       "建议 #1",
				"content":     content,
				"relevance":   0.8,
				"action":      "k8s.scale.deployment",
				"params":      gin.H{"replicas": 2},
				"data_source": dataSource,
			}},
		})
	}
}

func k8sAnalyzeHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]any
		_ = c.ShouldBindJSON(&req)

		insights := []string{
			"建议优先检查异常 Pod 的重启次数和事件。",
			"建议确认 Deployment 副本数与实际运行数是否一致。",
		}
		dataSource := "fallback"
		if out, err := generateByLLM(c.Request.Context(), svcCtx, "你是K8s运维助手。根据如下输入给出2条诊断建议（每条一句）:"+mustJSON(req)); err == nil && strings.TrimSpace(out) != "" {
			insights = strings.Split(out, "\n")
			dataSource = "llm"
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
			"insights": insights,
			"risks": []string{
				"高峰时段直接变更副本可能引发抖动。",
			},
			"recommended_actions": []gin.H{{
				"action": "k8s.scale.deployment",
				"params": gin.H{"replicas": 2},
				"reason": "优先验证扩缩容链路",
			}},
			"data_source": dataSource,
		}})
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

func streamAssistant(ctx context.Context, svcCtx *svc.ServiceContext, message string, msgContext map[string]any, onDelta func(chunk string) bool) error {
	if svcCtx == nil || svcCtx.AI == nil || svcCtx.AI.Runnable == nil {
		if !onDelta("AI 功能未初始化，请检查 LLM 配置与服务连接。") {
			return errors.New("client closed")
		}
		return nil
	}

	prompt := message
	if len(msgContext) > 0 {
		prompt = message + "\n\n上下文:\n" + mustJSON(msgContext)
	}
	stream, err := svcCtx.AI.Runnable.Stream(ctx, []*schema.Message{schema.UserMessage(prompt)})
	if err != nil {
		return err
	}
	defer stream.Close()

	for {
		msg, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			return nil
		}
		if recvErr != nil {
			return recvErr
		}
		if msg == nil || msg.Role != schema.Assistant || msg.Content == "" {
			continue
		}
		if !onDelta(msg.Content) {
			return errors.New("client closed")
		}
	}
}

func generateByLLM(ctx context.Context, svcCtx *svc.ServiceContext, prompt string) (string, error) {
	if svcCtx == nil || svcCtx.AI == nil || svcCtx.AI.Runnable == nil {
		return "", errors.New("ai not initialized")
	}
	msg, err := svcCtx.AI.Runnable.Generate(ctx, []*schema.Message{schema.UserMessage(prompt)})
	if err != nil {
		return "", err
	}
	if msg == nil {
		return "", errors.New("empty ai response")
	}
	return strings.TrimSpace(msg.Content), nil
}

func appendMessage(sessionID string, message map[string]any) *aiSession {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	now := time.Now()
	s, ok := sessions[sessionID]
	if !ok {
		s = &aiSession{
			ID:        sessionID,
			Title:     "AI Session",
			CreatedAt: now,
		}
		sessions[sessionID] = s
	}
	s.Messages = append(s.Messages, message)
	s.UpdatedAt = now
	return cloneSession(s)
}

func cloneSession(in *aiSession) *aiSession {
	out := *in
	out.Messages = make([]map[string]any, 0, len(in.Messages))
	for _, m := range in.Messages {
		cloned := make(map[string]any, len(m))
		for k, v := range m {
			cloned[k] = v
		}
		out.Messages = append(out.Messages, cloned)
	}
	return &out
}

func writeSSE(c *gin.Context, flusher http.Flusher, event string, payload any) bool {
	raw, err := json.Marshal(payload)
	if err != nil {
		return false
	}
	if _, err = fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, raw); err != nil {
		return false
	}
	flusher.Flush()
	return true
}

func mustJSON(v any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func extractToolRequest(message string, msgContext map[string]any) (string, map[string]any, string, bool) {
	toolName := strings.TrimSpace(toString(msgContext["tool_name"]))
	params, _ := msgContext["tool_params"].(map[string]any)
	approvalToken := strings.TrimSpace(toString(msgContext["approval_token"]))
	if toolName != "" {
		if params == nil {
			params = map[string]any{}
		}
		return toolName, params, approvalToken, true
	}

	msg := strings.TrimSpace(message)
	if !strings.HasPrefix(msg, "/tool ") {
		return "", nil, "", false
	}
	parts := strings.SplitN(strings.TrimSpace(strings.TrimPrefix(msg, "/tool ")), " ", 2)
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return "", nil, "", false
	}
	toolName = strings.TrimSpace(parts[0])
	params = map[string]any{}
	if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
		_ = json.Unmarshal([]byte(parts[1]), &params)
	}
	return toolName, params, approvalToken, true
}
