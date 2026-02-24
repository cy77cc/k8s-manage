package handler

import (
	"net/http"

	hostlogic "github.com/cy77cc/k8s-manage/internal/service/host/logic"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListCloudAccounts(c *gin.Context) {
	provider := c.Query("provider")
	list, err := h.hostService.ListCloudAccounts(c.Request.Context(), provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": list, "total": len(list)})
}

func (h *Handler) CreateCloudAccount(c *gin.Context) {
	var req hostlogic.CloudAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	item, err := h.hostService.CreateCloudAccount(c.Request.Context(), getUID(c), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": item})
}

func (h *Handler) TestCloudAccount(c *gin.Context) {
	provider := c.Param("provider")
	var req hostlogic.CloudAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	req.Provider = provider
	result, err := h.hostService.TestCloudAccount(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": result})
}

func (h *Handler) QueryCloudInstances(c *gin.Context) {
	provider := c.Param("provider")
	var req hostlogic.CloudQueryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	req.Provider = provider
	list, err := h.hostService.QueryCloudInstances(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": list, "total": len(list)})
}

func (h *Handler) ImportCloudInstances(c *gin.Context) {
	provider := c.Param("provider")
	var req hostlogic.CloudImportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	req.Provider = provider
	task, nodes, err := h.hostService.ImportCloudInstances(c.Request.Context(), getUID(c), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"task": task, "created": nodes}})
}

func (h *Handler) GetCloudImportTask(c *gin.Context) {
	task, err := h.hostService.GetImportTask(c.Request.Context(), c.Param("task_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "task not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": task})
}
