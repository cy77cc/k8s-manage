package handler

import (
	"fmt"
	"net/http"

	sshclient "github.com/cy77cc/k8s-manage/internal/client/ssh"
	"github.com/gin-gonic/gin"
)

func (h *Handler) SSHCheck(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	node, err := h.hostService.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "host not found"}})
		return
	}
	cli, err := sshclient.NewSSHClient(node.SSHUser, node.SSHPassword, node.IP, node.Port, "")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"reachable": false, "message": err.Error()}})
		return
	}
	_ = cli.Close()
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"reachable": true}})
}

func (h *Handler) SSHExec(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req struct {
		Command string `json:"command" binding:"required"`
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
	cli, err := sshclient.NewSSHClient(node.SSHUser, node.SSHPassword, node.IP, node.Port, "")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"stdout": "", "stderr": err.Error(), "exit_code": 1}})
		return
	}
	defer cli.Close()
	out, err := sshclient.RunCommand(cli, req.Command)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"stdout": out, "stderr": err.Error(), "exit_code": 1}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"stdout": out, "stderr": "", "exit_code": 0}})
}

func (h *Handler) BatchExec(c *gin.Context) {
	var req struct {
		HostIDs []uint64 `json:"host_ids"`
		Command string   `json:"command" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	results := map[string]any{}
	for _, id := range req.HostIDs {
		node, err := h.hostService.Get(c.Request.Context(), id)
		if err != nil {
			results[fmt.Sprintf("%d", id)] = gin.H{"stdout": "", "stderr": "host not found", "exit_code": 1}
			continue
		}
		cli, err := sshclient.NewSSHClient(node.SSHUser, node.SSHPassword, node.IP, node.Port, "")
		if err != nil {
			results[fmt.Sprintf("%d", id)] = gin.H{"stdout": "", "stderr": err.Error(), "exit_code": 1}
			continue
		}
		out, err := sshclient.RunCommand(cli, req.Command)
		_ = cli.Close()
		if err != nil {
			results[fmt.Sprintf("%d", id)] = gin.H{"stdout": out, "stderr": err.Error(), "exit_code": 1}
			continue
		}
		results[fmt.Sprintf("%d", id)] = gin.H{"stdout": out, "stderr": "", "exit_code": 0}
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": results})
}
