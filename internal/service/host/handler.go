package host

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	sshclient "github.com/cy77cc/k8s-manage/internal/client/ssh"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svcCtx *svc.ServiceContext
}

func NewHandler(svcCtx *svc.ServiceContext) *Handler { return &Handler{svcCtx: svcCtx} }

type createHostReq struct {
	Name        string `json:"name" binding:"required"`
	IP          string `json:"ip" binding:"required"`
	Status      string `json:"status"`
	Username    string `json:"username"`
	Port        int    `json:"port"`
	Password    string `json:"password"`
	Description string `json:"description"`
}

func (h *Handler) List(c *gin.Context) {
	var list []model.Node
	if err := h.svcCtx.DB.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": list, "total": len(list)})
}

func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid id"}})
		return
	}
	var node model.Node
	if err := h.svcCtx.DB.First(&node, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "host not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": node})
}

func (h *Handler) Create(c *gin.Context) {
	var req createHostReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	if req.Username == "" {
		req.Username = "root"
	}
	if req.Port == 0 {
		req.Port = 22
	}
	status := req.Status
	if status == "" {
		status = "offline"
	}

	node := model.Node{Name: req.Name, IP: req.IP, Port: req.Port, SSHUser: req.Username, SSHPassword: req.Password, Status: status, Description: req.Description}
	if req.Password != "" {
		if cli, err := sshclient.NewSSHClient(req.Username, req.Password, req.IP, req.Port, ""); err == nil {
			if out, err := sshclient.RunCommand(cli, "hostname"); err == nil {
				node.Hostname = out
				node.Status = "online"
			}
			_ = cli.Close()
		}
	}

	if err := h.svcCtx.DB.Create(&node).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": node})
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid id"}})
		return
	}
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	if err := h.svcCtx.DB.Model(&model.Node{}).Where("id = ?", id).Updates(req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	var node model.Node
	_ = h.svcCtx.DB.First(&node, id).Error
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": node})
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid id"}})
		return
	}
	if err := h.svcCtx.DB.Delete(&model.Node{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) Action(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Action string `json:"action"`
	}
	_ = c.ShouldBindJSON(&req)
	_ = h.svcCtx.DB.Model(&model.Node{}).Where("id = ?", id).Update("status", strings.ToLower(req.Action)).Error
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"id": id, "action": req.Action}})
}

func (h *Handler) SSHCheck(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid id"}})
		return
	}
	var node model.Node
	if err := h.svcCtx.DB.First(&node, id).Error; err != nil {
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid id"}})
		return
	}
	var req struct {
		Command string `json:"command" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	var node model.Node
	if err := h.svcCtx.DB.First(&node, id).Error; err != nil {
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

func (h *Handler) Batch(c *gin.Context) {
	var req struct {
		HostIDs []uint   `json:"host_ids"`
		Action  string   `json:"action"`
		Tags    []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	if len(req.HostIDs) > 0 && req.Action != "" {
		_ = h.svcCtx.DB.Model(&model.Node{}).Where("id IN ?", req.HostIDs).Update("status", req.Action).Error
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) BatchExec(c *gin.Context) {
	var req struct {
		HostIDs []uint `json:"host_ids"`
		Command string `json:"command" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	results := map[string]any{}
	for _, id := range req.HostIDs {
		var node model.Node
		if err := h.svcCtx.DB.First(&node, id).Error; err != nil {
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

func (h *Handler) Facts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"os": "linux", "arch": "amd64", "source": "mvp"}})
}

func (h *Handler) Tags(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid id"}})
		return
	}
	var node model.Node
	if err := h.svcCtx.DB.First(&node, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "host not found"}})
		return
	}
	parts := strings.Split(node.Labels, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			out = append(out, p)
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": out})
}

func (h *Handler) AddTag(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Tag string `json:"tag" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	var node model.Node
	if err := h.svcCtx.DB.First(&node, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "host not found"}})
		return
	}
	if node.Labels == "" {
		node.Labels = req.Tag
	} else {
		node.Labels += "," + req.Tag
	}
	_ = h.svcCtx.DB.Save(&node).Error
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) RemoveTag(c *gin.Context) {
	id := c.Param("id")
	tag, _ := url.QueryUnescape(c.Param("tag"))
	var node model.Node
	if err := h.svcCtx.DB.First(&node, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "host not found"}})
		return
	}
	parts := strings.Split(node.Labels, ",")
	filtered := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" && p != tag {
			filtered = append(filtered, p)
		}
	}
	node.Labels = strings.Join(filtered, ",")
	_ = h.svcCtx.DB.Save(&node).Error
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
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
