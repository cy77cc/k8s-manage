package handler

import (
	"net/http"
	"strconv"

	hostlogic "github.com/cy77cc/k8s-manage/internal/service/host/logic"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListSSHKeys(c *gin.Context) {
	list, err := h.hostService.ListSSHKeys(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": list, "total": len(list)})
}

func (h *Handler) CreateSSHKey(c *gin.Context) {
	var req hostlogic.SSHKeyCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	item, err := h.hostService.CreateSSHKey(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": item})
}

func (h *Handler) DeleteSSHKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid id"}})
		return
	}
	if err := h.hostService.DeleteSSHKey(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) VerifySSHKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid id"}})
		return
	}
	var req hostlogic.SSHKeyVerifyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	result, err := h.hostService.VerifySSHKey(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": result})
}
