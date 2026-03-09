package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	coreai "github.com/cy77cc/k8s-manage/internal/ai"
	airag "github.com/cy77cc/k8s-manage/internal/ai/rag"
	"github.com/cy77cc/k8s-manage/internal/ai/tools/core"
	"github.com/cy77cc/k8s-manage/internal/httpx"
	"github.com/cy77cc/k8s-manage/internal/service/ai/logic"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *AIHandler) capabilities(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	if h.ai == nil {
		httpx.OK(c, []any{})
		return
	}
	all := h.ai.ToolMetas()
	out := make([]core.ToolMeta, 0, len(all))
	for _, item := range all {
		if h.hasPermission(uid, item.Permission) {
			out = append(out, item)
		}
	}
	httpx.OK(c, out)
}

func (h *AIHandler) previewTool(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	var req struct {
		Tool   string         `json:"tool" binding:"required"`
		Params map[string]any `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}
	data, err := h.gateway.PreviewTool(uid, req.Tool, req.Params)
	if err != nil {
		switch {
		case errors.Is(err, coreai.ErrToolNotFound):
			httpx.Fail(c, xcode.NotFound, "tool not found")
		case errors.Is(err, coreai.ErrPermissionDenied):
			httpx.Fail(c, xcode.Forbidden, "permission denied")
		default:
			httpx.Fail(c, xcode.ServerError, err.Error())
		}
		return
	}
	httpx.OK(c, data)
}

func (h *AIHandler) executeTool(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	var req struct {
		Tool          string         `json:"tool" binding:"required"`
		Params        map[string]any `json:"params"`
		ApprovalToken string         `json:"approval_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}
	rec, err := h.gateway.ExecuteTool(c.Request.Context(), uid, req.Tool, req.Params, req.ApprovalToken)
	if err != nil {
		switch {
		case errors.Is(err, coreai.ErrToolNotFound):
			httpx.Fail(c, xcode.NotFound, "tool not found")
		case errors.Is(err, coreai.ErrPermissionDenied):
			httpx.Fail(c, xcode.Forbidden, "permission denied")
		default:
			httpx.Fail(c, xcode.ServerError, err.Error())
		}
		return
	}
	httpx.OK(c, rec)
}

func (h *AIHandler) getExecution(c *gin.Context) {
	if _, ok := uidFromContext(c); !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	rec, ok := h.gateway.GetExecution(c.Param("id"))
	if !ok {
		httpx.Fail(c, xcode.NotFound, "tool not found")
		return
	}
	httpx.OK(c, rec)
}

func (h *AIHandler) createApproval(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	var req struct {
		Tool   string         `json:"tool" binding:"required"`
		Params map[string]any `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}
	t, err := h.gateway.CreateApprovalTask(c.Request.Context(), uid, req.Tool, req.Params)
	if err != nil {
		switch {
		case errors.Is(err, coreai.ErrToolNotFound):
			httpx.Fail(c, xcode.NotFound, "tool not found")
		case errors.Is(err, coreai.ErrPermissionDenied):
			httpx.Fail(c, xcode.Forbidden, "permission denied")
		default:
			httpx.Fail(c, xcode.ParamError, err.Error())
		}
		return
	}
	httpx.OK(c, t)
}

func (h *AIHandler) listApprovals(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	items, err := h.gateway.ListApprovalTasks(c.Request.Context(), uid, strings.TrimSpace(c.Query("status")))
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, items)
}

func (h *AIHandler) getApproval(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	item, err := h.gateway.GetApprovalTask(c.Request.Context(), uid, c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, coreai.ErrPermissionDenied):
			httpx.Fail(c, xcode.Forbidden, "permission denied")
		case errors.Is(err, gorm.ErrRecordNotFound):
			httpx.Fail(c, xcode.NotFound, "approval not found")
		default:
			httpx.Fail(c, xcode.ServerError, err.Error())
		}
		return
	}
	httpx.OK(c, item)
}

func (h *AIHandler) confirmApproval(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	if !h.hasPermission(uid, "ai:approval:review") {
		httpx.Fail(c, xcode.Forbidden, "permission denied")
		return
	}
	var req struct {
		Approve bool `json:"approve"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}
	task, execution, err := h.gateway.ReviewApprovalTask(c.Request.Context(), uid, c.Param("id"), req.Approve, "")
	if err != nil {
		switch {
		case errors.Is(err, coreai.ErrPermissionDenied):
			httpx.Fail(c, xcode.Forbidden, "permission denied")
		case errors.Is(err, gorm.ErrRecordNotFound):
			httpx.Fail(c, xcode.NotFound, "approval not found")
		case errors.Is(err, coreai.ErrApprovalExpired):
			httpx.Fail(c, xcode.ParamError, "approval expired")
		default:
			httpx.Fail(c, xcode.ParamError, err.Error())
		}
		return
	}
	httpx.OK(c, gin.H{"task": task, "execution": execution})
}

func (h *AIHandler) approveApproval(c *gin.Context) {
	h.reviewApproval(c, true)
}

func (h *AIHandler) rejectApproval(c *gin.Context) {
	h.reviewApproval(c, false)
}

func (h *AIHandler) submitFeedback(c *gin.Context) {
	var req struct {
		SessionID   string `json:"session_id"`
		Namespace   string `json:"namespace"`
		IsEffective bool   `json:"is_effective"`
		Comment     string `json:"comment"`
		Question    string `json:"question"`
		Answer      string `json:"answer"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}
	manager, ok := any(h.ai).(interface {
		CollectFeedback(ctx context.Context, sessionID, namespace string, feedback airag.Feedback, question, answer string) (*airag.KnowledgeEntry, error)
	})
	if !ok {
		httpx.Fail(c, xcode.ServerError, "feedback collector not initialized")
		return
	}
	entry, err := manager.CollectFeedback(c.Request.Context(), req.SessionID, req.Namespace, airag.Feedback{
		IsEffective: req.IsEffective,
		Comment:     strings.TrimSpace(req.Comment),
	}, req.Question, req.Answer)
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, entry)
}

func (h *AIHandler) reviewApproval(c *gin.Context, approve bool) {
	uid, ok := uidFromContext(c)
	if !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.BindErr(c, err)
			return
		}
	}
	task, execution, err := h.gateway.ReviewApprovalTask(c.Request.Context(), uid, c.Param("id"), approve, strings.TrimSpace(req.Reason))
	if err != nil {
		switch {
		case errors.Is(err, coreai.ErrPermissionDenied):
			httpx.Fail(c, xcode.Forbidden, "permission denied")
		case errors.Is(err, gorm.ErrRecordNotFound):
			httpx.Fail(c, xcode.NotFound, "approval not found")
		case errors.Is(err, coreai.ErrApprovalExpired):
			httpx.Fail(c, xcode.ParamError, "approval expired")
		default:
			httpx.Fail(c, xcode.ServerError, err.Error())
		}
		return
	}
	httpx.OK(c, gin.H{"task": task, "execution": execution})
}

func (h *AIHandler) confirmConfirmation(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	var req struct {
		Approve bool `json:"approve"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		httpx.Fail(c, xcode.ParamError, "confirmation id is required")
		return
	}
	out, err := h.gateway.ConfirmConfirmation(c.Request.Context(), uid, id, req.Approve)
	if err != nil {
		switch {
		case errors.Is(err, logic.ErrConfirmationNotFound):
			httpx.Fail(c, xcode.NotFound, "confirmation not found")
		case errors.Is(err, coreai.ErrPermissionDenied):
			httpx.Fail(c, xcode.Forbidden, "permission denied")
		case errors.Is(err, coreai.ErrApprovalExpired):
			httpx.Fail(c, xcode.ParamError, "confirmation expired")
		default:
			httpx.Fail(c, xcode.ParamError, err.Error())
		}
		return
	}
	httpx.OK(c, out)
}

func (h *AIHandler) findMeta(name string) (core.ToolMeta, bool) {
	return h.control.FindMeta(name)
}

func strconvFormatInt(v int64) string {
	return fmt.Sprintf("%d", v)
}

func (h *AIHandler) toolParamHints(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	name := strings.TrimSpace(c.Param("name"))
	meta, ok := h.findMeta(name)
	if !ok {
		httpx.Fail(c, xcode.NotFound, "tool not found")
		return
	}
	if !h.hasPermission(uid, meta.Permission) {
		httpx.Fail(c, xcode.Forbidden, "permission denied")
		return
	}
	resp := core.ResolveToolParamHints(c.Request.Context(), core.PlatformDeps{DB: h.svcCtx.DB}, meta)
	httpx.OK(c, resp)
}
