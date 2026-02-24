package node

import (
	"github.com/cy77cc/k8s-manage/internal/middleware"
	hostlogic "github.com/cy77cc/k8s-manage/internal/service/host/logic"
	"github.com/cy77cc/k8s-manage/internal/service/node/handler"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterNodeHandlers(r *gin.RouterGroup, serverCtx *svc.ServiceContext) {
	g := r.Group("node", middleware.JWTAuth())
	h := handler.NewNodeHandler(serverCtx)
	g.Use(func(c *gin.Context) {
		c.Header("Deprecation", "true")
		c.Header("Sunset", hostlogic.NodeSunsetDateRFC)
		c.Next()
	})
	// Add Node permission check
	g.POST("add", middleware.CasbinAuth(serverCtx.CasbinEnforcer, "node:add"), h.Add)
}
