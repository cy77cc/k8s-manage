package handler

import (
	"net/http"

	hostlogic "github.com/cy77cc/k8s-manage/internal/service/host/logic"
	"github.com/gin-gonic/gin"
)

func (h *Handler) KVMPreview(c *gin.Context) {
	hostID, ok := parseID(c)
	if !ok {
		return
	}
	var req hostlogic.KVMPreviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	result, err := h.hostService.KVMPreview(c.Request.Context(), hostID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": result})
}

func (h *Handler) KVMProvision(c *gin.Context) {
	hostID, ok := parseID(c)
	if !ok {
		return
	}
	var req hostlogic.KVMProvisionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	task, node, err := h.hostService.KVMProvision(c.Request.Context(), getUID(c), hostID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"task": task, "node": node}})
}

func (h *Handler) GetVirtualizationTask(c *gin.Context) {
	task, err := h.hostService.GetVirtualizationTask(c.Request.Context(), c.Param("task_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "task not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": task})
}
