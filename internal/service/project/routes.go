package project

import (
	"github.com/cy77cc/k8s-manage/internal/service/project/handler"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterProjectHandlers(g *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	projectHandler := handler.NewProjectHandler(svcCtx)
	serviceHandler := handler.NewServiceHandler(svcCtx)

	// Projects
	projects := g.Group("/projects")
	{
		projects.POST("", projectHandler.CreateProject)
		projects.GET("", projectHandler.ListProjects)
		projects.POST("/deploy", projectHandler.DeployProject)
	}

	// Services
	services := g.Group("/services")
	{
		services.POST("", serviceHandler.CreateService)
		services.GET("", serviceHandler.ListServices)
		services.GET("/:id", serviceHandler.GetService)
		services.PUT("/:id", serviceHandler.UpdateService)
		services.DELETE("/:id", serviceHandler.DeleteService)
		services.POST("/deploy", serviceHandler.DeployService)
	}
}
