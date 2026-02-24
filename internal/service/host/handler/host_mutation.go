package handler

import (
	"net/http"
	"strings"

	hostlogic "github.com/cy77cc/k8s-manage/internal/service/host/logic"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Probe(c *gin.Context) {
	var req hostlogic.ProbeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}, "error_code": "validation_error"})
		return
	}
	resp, err := h.hostService.Probe(c.Request.Context(), getUID(c), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) Create(c *gin.Context) {
	var req hostlogic.CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	uid := getUID(c)
	node, err := h.hostService.CreateWithProbe(c.Request.Context(), uid, isAdminByUserID(h.svcCtx, uid), req)
	if err != nil {
		errorCode := "validation_error"
		if strings.Contains(err.Error(), "probe_expired") {
			errorCode = "probe_expired"
		} else if strings.Contains(err.Error(), "probe_not_found") {
			errorCode = "probe_not_found"
		}
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}, "error_code": errorCode})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": node})
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	node, err := h.hostService.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": node})
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.hostService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) Action(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req struct {
		Action string `json:"action"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.hostService.UpdateStatus(c.Request.Context(), id, req.Action); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"id": id, "action": req.Action}})
}

func (h *Handler) Batch(c *gin.Context) {
	var req struct {
		HostIDs []uint64 `json:"host_ids"`
		Action  string   `json:"action"`
		Tags    []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	if err := h.hostService.BatchUpdateStatus(c.Request.Context(), req.HostIDs, req.Action); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) AddTag(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req struct {
		Tag string `json:"tag" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	node, err := h.hostService.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "host not found"}})
		return
	}
	labels := hostlogic.ParseLabels(node.Labels)
	labels = append(labels, req.Tag)
	_, err = h.hostService.Update(c.Request.Context(), id, map[string]any{"labels": strings.Join(labels, ",")})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) RemoveTag(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	tag := c.Param("tag")
	node, err := h.hostService.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "host not found"}})
		return
	}
	labels := hostlogic.ParseLabels(node.Labels)
	filtered := make([]string, 0, len(labels))
	for _, item := range labels {
		if item != tag {
			filtered = append(filtered, item)
		}
	}
	_, err = h.hostService.Update(c.Request.Context(), id, map[string]any{"labels": strings.Join(filtered, ",")})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) UpdateCredentials(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req hostlogic.UpdateCredentialsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}, "error_code": "validation_error"})
		return
	}
	node, probeResp, err := h.hostService.UpdateCredentials(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}, "error_code": "auth_error", "data": gin.H{"probe": probeResp, "node": node}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"node": node, "probe": probeResp}})
}
