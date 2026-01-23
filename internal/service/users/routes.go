package users

import (
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterHandlers(r *gin.RouterGroup, serverCtx *svc.ServiceContext) {
	r.Group("auth", )
}