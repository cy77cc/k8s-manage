package ai

import (
	"strings"

	askills "github.com/cy77cc/k8s-manage/internal/ai/skills"
	"github.com/cy77cc/k8s-manage/internal/httpx"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

func (h *handler) listSkills(c *gin.Context) {
	if _, ok := uidFromContext(c); !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	if h.skillRegistry == nil {
		httpx.OK(c, []askills.Skill{})
		return
	}
	httpx.OK(c, h.skillRegistry.List())
}

func (h *handler) executeSkill(c *gin.Context) {
	if _, ok := uidFromContext(c); !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	if h.skillRegistry == nil || h.skillExecutor == nil {
		httpx.Fail(c, xcode.ServerError, "skill executor not initialized")
		return
	}
	name := strings.TrimSpace(c.Param("name"))
	skill, ok := h.skillRegistry.Get(name)
	if !ok {
		httpx.Fail(c, xcode.NotFound, "skill not found")
		return
	}
	var req struct {
		Message string         `json:"message"`
		Params  map[string]any `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}
	var (
		result *askills.ExecutionResult
		err    error
	)
	if strings.TrimSpace(req.Message) != "" {
		result, err = h.skillExecutor.ExecuteFromMessage(c.Request.Context(), skill, req.Message, nil)
	} else {
		result, err = h.skillExecutor.Execute(c.Request.Context(), skill, req.Params)
	}
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, result)
}

func (h *handler) reloadSkills(c *gin.Context) {
	if _, ok := uidFromContext(c); !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	if h.skillRegistry == nil {
		httpx.Fail(c, xcode.ServerError, "skill registry not initialized")
		return
	}
	if err := h.skillRegistry.Reload(); err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, gin.H{"reloaded": true})
}
