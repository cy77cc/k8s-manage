package service

import (
	"net/http"

	_ "github.com/cy77cc/k8s-manage/docs"
	"github.com/cy77cc/k8s-manage/internal/logger"
	"github.com/cy77cc/k8s-manage/internal/middleware"
	"github.com/cy77cc/k8s-manage/internal/service/user"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
)

func Init(r *gin.Engine, serverCtx *svc.ServiceContext) {
	r.Use(logger.GinLogger(), middleware.Cors())
	r.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	v1 := r.Group("/api/v1")
	user.RegisterHandlers(v1, serverCtx)
}
