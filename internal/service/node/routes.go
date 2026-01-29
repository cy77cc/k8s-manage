package node

import (
	"github.com/cy77cc/k8s-manage/internal/middleware"
	"github.com/cy77cc/k8s-manage/internal/service/node/handler"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterNodeHandlers(r *gin.RouterGroup, serverCtx *svc.ServiceContext) {
	g := r.Group("node", middleware.JWTAuth())
	handler := handler.NewNodeHandler(serverCtx)
	g.POST("add", handler.Add)

}
