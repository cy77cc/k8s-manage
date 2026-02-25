package handler

import (
	"fmt"
	"net/http"
	"strings"

	sshclient "github.com/cy77cc/k8s-manage/internal/client/ssh"
	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/utils"
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
	privateKey, passphrase, err := h.loadNodePrivateKey(c, node)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"reachable": false, "message": err.Error()}})
		return
	}
	password := strings.TrimSpace(node.SSHPassword)
	if strings.TrimSpace(privateKey) != "" {
		password = ""
	}
	cli, err := sshclient.NewSSHClient(node.SSHUser, password, node.IP, node.Port, privateKey, passphrase)
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
	privateKey, passphrase, err := h.loadNodePrivateKey(c, node)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"stdout": "", "stderr": err.Error(), "exit_code": 1}})
		return
	}
	password := strings.TrimSpace(node.SSHPassword)
	if strings.TrimSpace(privateKey) != "" {
		password = ""
	}
	cli, err := sshclient.NewSSHClient(node.SSHUser, password, node.IP, node.Port, privateKey, passphrase)
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
		privateKey, passphrase, err := h.loadNodePrivateKey(c, node)
		if err != nil {
			results[fmt.Sprintf("%d", id)] = gin.H{"stdout": "", "stderr": err.Error(), "exit_code": 1}
			continue
		}
		password := strings.TrimSpace(node.SSHPassword)
		if strings.TrimSpace(privateKey) != "" {
			password = ""
		}
		cli, err := sshclient.NewSSHClient(node.SSHUser, password, node.IP, node.Port, privateKey, passphrase)
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

func (h *Handler) loadNodePrivateKey(c *gin.Context, node *model.Node) (string, string, error) {
	if node == nil || node.SSHKeyID == nil {
		return "", "", nil
	}
	var key model.SSHKey
	if err := h.svcCtx.DB.WithContext(c.Request.Context()).
		Select("id", "private_key", "passphrase", "encrypted").
		Where("id = ?", uint64(*node.SSHKeyID)).
		First(&key).Error; err != nil {
		return "", "", err
	}
	passphrase := strings.TrimSpace(key.Passphrase)
	if !key.Encrypted {
		return strings.TrimSpace(key.PrivateKey), passphrase, nil
	}
	privateKey, err := utils.DecryptText(strings.TrimSpace(key.PrivateKey), config.CFG.Security.EncryptionKey)
	if err != nil {
		return "", "", err
	}
	return privateKey, passphrase, nil
}
