package service

import (
	"net/http"

	_ "github.com/cy77cc/k8s-manage/docs"
	"github.com/cy77cc/k8s-manage/internal/middleware"
	"github.com/cy77cc/k8s-manage/internal/service/ai"
	"github.com/cy77cc/k8s-manage/internal/service/node"
	"github.com/cy77cc/k8s-manage/internal/service/project"
	"github.com/cy77cc/k8s-manage/internal/service/user"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
)

func Init(r *gin.Engine, serverCtx *svc.ServiceContext) {
	r.Use(middleware.ContextMiddleware(), middleware.Cors(), middleware.Logger())
	r.GET("/swagger/*any", func(c *gin.Context) {
		if c.Param("any") == "" || c.Param("any") == "/" {
			c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
			return
		}
		gs.WrapHandler(swaggerFiles.Handler)(c)
	})
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	v1 := r.Group("/api/v1")
	user.RegisterUserHandlers(v1, serverCtx)
	node.RegisterNodeHandlers(v1, serverCtx)
	project.RegisterProjectHandlers(v1, serverCtx)

	// AI routes
	v1.POST("/ai/chat", ai.ChatHandler(serverCtx))
}
