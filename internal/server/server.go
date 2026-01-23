package server

import (
	"fmt"

	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/service"
	"github.com/gin-gonic/gin"
)

func Start() error {
	r := gin.Default()
	service.Init(r)
	return r.Run(fmt.Sprintf("%s:%d", config.CFG.Server.Host, config.CFG.Server.Port))
}
