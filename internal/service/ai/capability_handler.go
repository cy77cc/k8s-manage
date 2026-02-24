package ai

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	ai2 "github.com/cy77cc/k8s-manage/internal/ai"
	"github.com/gin-gonic/gin"
)

func (h *handler) capabilities(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized"})
		return
	}
	if h.svcCtx.AI == nil {
		c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": []any{}})
		return
	}
	all := h.svcCtx.AI.ToolMetas()
	out := make([]ai2.ToolMeta, 0, len(all))
	for _, item := range all {
		if h.hasPermission(uid, item.Permission) {
			out = append(out, item)
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": out})
}

func (h *handler) previewTool(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized"})
		return
	}
	var req struct {
		Tool   string         `json:"tool" binding:"required"`
		Params map[string]any `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	meta, ok := h.findMeta(req.Tool)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "tool not found"})
		return
	}
	if h.svcCtx.AI == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "ai agent not initialized"})
		return
	}
	if !h.hasPermission(uid, meta.Permission) {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "permission denied"})
		return
	}
	data := gin.H{
		"tool":              meta.Name,
		"mode":              meta.Mode,
		"risk":              meta.Risk,
		"params":            req.Params,
		"approval_required": meta.Mode == ai2.ToolModeMutating,
	}
	if meta.Mode == ai2.ToolModeMutating {
		t := h.store.newApproval(uid, approvalTicket{
			Tool:   meta.Name,
			Params: req.Params,
			Risk:   meta.Risk,
			Mode:   meta.Mode,
			Meta:   meta,
		})
		data["approval_token"] = t.ID
		data["expiresAt"] = t.ExpiresAt
		data["previewDiff"] = "Mutating operation requires approval."
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": data})
}

func (h *handler) executeTool(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized"})
		return
	}
	var req struct {
		Tool          string         `json:"tool" binding:"required"`
		Params        map[string]any `json:"params"`
		ApprovalToken string         `json:"approval_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	meta, ok := h.findMeta(req.Tool)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "tool not found"})
		return
	}
	rec := &executionRecord{
		ID:         "exe-" + strconvFormatInt(time.Now().UnixNano()),
		Tool:       req.Tool,
		Params:     req.Params,
		Mode:       meta.Mode,
		Status:     "running",
		RequestUID: uid,
		CreatedAt:  time.Now(),
	}
	start := time.Now()
	ctx := ai2.WithToolUser(c.Request.Context(), uid, strings.TrimSpace(req.ApprovalToken))
	ctx = ai2.WithToolPolicyChecker(ctx, h.toolPolicy)
	result, err := h.svcCtx.AI.RunTool(ctx, req.Tool, req.Params)
	finished := time.Now()
	rec.FinishedAt = &finished
	rec.Result = &result
	if err != nil {
		rec.Status = "failed"
		rec.Error = err.Error()
	} else if result.OK {
		rec.Status = "succeeded"
	} else {
		rec.Status = "failed"
		rec.Error = result.Error
	}
	if apErr, ok := ai2.IsApprovalRequired(err); ok {
		rec.Status = "failed"
		rec.Error = apErr.Error()
	}
	if result.LatencyMS == 0 {
		rec.Result.LatencyMS = time.Since(start).Milliseconds()
	}
	h.store.saveExecution(rec)
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rec})
}

func (h *handler) getExecution(c *gin.Context) {
	if _, ok := uidFromContext(c); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized"})
		return
	}
	rec, ok := h.store.getExecution(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "execution not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rec})
}

func (h *handler) createApproval(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized"})
		return
	}
	var req struct {
		Tool   string         `json:"tool" binding:"required"`
		Params map[string]any `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	meta, ok := h.findMeta(req.Tool)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "tool not found"})
		return
	}
	if meta.Mode == ai2.ToolModeReadonly {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "readonly tool does not require approval"})
		return
	}
	t := h.store.newApproval(uid, approvalTicket{
		Tool:   meta.Name,
		Params: req.Params,
		Risk:   meta.Risk,
		Mode:   meta.Mode,
		Meta:   meta,
	})
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": t})
}

func (h *handler) confirmApproval(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized"})
		return
	}
	if !h.hasPermission(uid, "ai:approval:review") {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "permission denied"})
		return
	}
	var req struct {
		Approve bool `json:"approve"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	id := c.Param("id")
	t, ok := h.store.getApproval(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "approval not found"})
		return
	}
	if time.Now().After(t.ExpiresAt) {
		_, _ = h.store.setApprovalStatus(id, "expired", uid)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "approval expired"})
		return
	}
	status := "rejected"
	if req.Approve {
		status = "approved"
	}
	out, _ := h.store.setApprovalStatus(id, status, uid)
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": out})
}

func (h *handler) findMeta(name string) (ai2.ToolMeta, bool) {
	if h.svcCtx.AI == nil {
		return ai2.ToolMeta{}, false
	}
	for _, item := range h.svcCtx.AI.ToolMetas() {
		if item.Name == name {
			return item, true
		}
	}
	return ai2.ToolMeta{}, false
}

func strconvFormatInt(v int64) string {
	return fmt.Sprintf("%d", v)
}
