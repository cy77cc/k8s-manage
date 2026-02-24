package handler

import (
	"net/http"
	"strconv"
	"time"

	v1 "github.com/cy77cc/k8s-manage/api/project/v1"
	"github.com/cy77cc/k8s-manage/internal/service/project/logic"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

type ServiceHandler struct {
	logic *logic.ServiceLogic
}

func NewServiceHandler(svcCtx *svc.ServiceContext) *ServiceHandler {
	return &ServiceHandler{logic: logic.NewServiceLogic(svcCtx)}
}

func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req v1.CreateServiceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	resp, err := h.logic.CreateService(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *ServiceHandler) ListServices(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	var projectID uint
	if projectIDStr != "" {
		pid, err := strconv.ParseUint(projectIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
			return
		}
		projectID = uint(pid)
	}
	resp, err := h.logic.ListServices(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp, "total": len(resp)})
}

func (h *ServiceHandler) GetService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	resp, err := h.logic.GetService(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req v1.CreateServiceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.logic.UpdateService(c.Request.Context(), uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.logic.DeleteService(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "Service deleted successfully", "data": nil})
}

func (h *ServiceHandler) DeployService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		ClusterID uint `json:"cluster_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	if req.ClusterID == 0 {
		req.ClusterID = 1
	}
	if err := h.logic.DeployService(c.Request.Context(), v1.DeployServiceReq{ServiceID: uint(id), ClusterID: req.ClusterID}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "Service deployed successfully", "data": nil})
}

func (h *ServiceHandler) RollbackService(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "rollback is not implemented in MVP", "data": nil})
}

func (h *ServiceHandler) GetEvents(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": []gin.H{{"id": 1, "service_id": c.Param("id"), "type": "deploy", "level": "info", "message": "service created", "created_at": time.Now()}}})
}

func (h *ServiceHandler) GetQuota(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"cpuLimit": 8000, "memoryLimit": 16384, "cpuUsed": 1200, "memoryUsed": 2048}})
}
