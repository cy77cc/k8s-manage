package project

import (
	"github.com/cy77cc/k8s-manage/internal/service/project/handler"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterProjectHandlers(g *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	projectHandler := handler.NewProjectHandler(svcCtx)

	// Projects
	projects := g.Group("/projects")
	{
		projects.POST("", projectHandler.CreateProject)
		projects.GET("", projectHandler.ListProjects)
		projects.POST("/deploy", projectHandler.DeployProject)
	}
}
