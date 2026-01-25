package user

import (
	"github.com/cy77cc/k8s-manage/internal/middleware"
	userHandler "github.com/cy77cc/k8s-manage/internal/service/user/handler"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterHandlers(r *gin.RouterGroup, serverCtx *svc.ServiceContext) {
	authGroup := r.Group("auth")

	userHandler := userHandler.NewuserHandler(serverCtx)

	{
		authGroup.POST("login", userHandler.Login)
		authGroup.POST("logout", userHandler.Logout)
		authGroup.POST("refresh", userHandler.Refresh)
		authGroup.POST("register", userHandler.Register)
	}
	userGroup := r.Group("user", middleware.JWTAuth())
	{
		userGroup.POST("/")
	}
}
