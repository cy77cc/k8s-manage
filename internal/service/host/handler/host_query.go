package handler

import (
	"net/http"
	"time"

	hostlogic "github.com/cy77cc/k8s-manage/internal/service/host/logic"
	"github.com/gin-gonic/gin"
)

func (h *Handler) List(c *gin.Context) {
	list, err := h.hostService.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": list, "total": len(list)})
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	node, err := h.hostService.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "host not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": node})
}

func (h *Handler) Facts(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	node, err := h.hostService.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "host not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"os": node.OS, "arch": node.Arch, "kernel": node.Kernel, "cpu_cores": node.CpuCores, "memory_mb": node.MemoryMB, "disk_gb": node.DiskGB, "source": "node"}})
}

func (h *Handler) Tags(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	node, err := h.hostService.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "host not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": hostlogic.ParseLabels(node.Labels)})
}

func (h *Handler) Metrics(c *gin.Context) {
	now := time.Now()
	rows := []gin.H{{"id": 1, "cpu": 10, "memory": 256, "disk": 20, "network": 5, "created_at": now.Add(-2 * time.Minute)}, {"id": 2, "cpu": 12, "memory": 260, "disk": 20, "network": 6, "created_at": now.Add(-time.Minute)}, {"id": 3, "cpu": 14, "memory": 262, "disk": 20, "network": 8, "created_at": now}}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rows})
}

func (h *Handler) Audits(c *gin.Context) {
	rows := []gin.H{{"id": 1, "action": "query", "operator": "system", "detail": "host detail viewed", "created_at": time.Now()}}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rows})
}
