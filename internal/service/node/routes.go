package node

import (
	"github.com/cy77cc/k8s-manage/internal/middleware"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterNodeHandlers(r *gin.RouterGroup, serverCtx *svc.ServiceContext) {
	r.Use(middleware.JWTAuth())
}
