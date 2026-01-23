package service

import (
	"net/http"

	"github.com/cy77cc/k8s-manage/internal/logger"
	"github.com/cy77cc/k8s-manage/internal/middleware"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine, svcCtx *svc.ServiceContext) {
	r.Use(logger.GinLogger(), middleware.Cors())
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("favicon.ico", func(c *gin.Context) {
		c.Data(
			http.StatusOK,
			"image/x-icon",
			[]byte("xxxx"),
		)
	})

}
