package handler

import (
	"net/http"

	v1 "github.com/cy77cc/k8s-manage/api/project/v1"
	"github.com/cy77cc/k8s-manage/internal/service/project/logic"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	logic *logic.ProjectLogic
}

func NewProjectHandler(svcCtx *svc.ServiceContext) *ProjectHandler {
	return &ProjectHandler{
		logic: logic.NewProjectLogic(svcCtx),
	}
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req v1.CreateProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.logic.CreateProject(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ProjectHandler) ListProjects(c *gin.Context) {
	resp, err := h.logic.ListProjects(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ProjectHandler) DeployProject(c *gin.Context) {
	var req v1.DeployProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.logic.DeployProject(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deployed successfully"})
}
