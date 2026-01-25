package server

import (
	"fmt"

	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/service"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func Start() error {
	svcCtx := svc.MustNewServiceContext()
	r := gin.Default()
	service.Init(r, svcCtx)
	return r.Run(fmt.Sprintf("%s:%d", config.CFG.Server.Host, config.CFG.Server.Port))
}
