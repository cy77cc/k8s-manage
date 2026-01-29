package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/logger"
	"github.com/cy77cc/k8s-manage/internal/service"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/cy77cc/k8s-manage/storage"
	"github.com/gin-gonic/gin"
)

// Start 启动 HTTP 服务器
func Start(ctx context.Context) error {
	go startServer(ctx)
	<-ctx.Done()
	logger.L().Info("Shutting Down...........")
	return nil
}

// startServer 启动 Gin 服务
func startServer(ctx context.Context) {
	svcCtx := svc.MustNewServiceContext()
	storage.MustMigrate(svcCtx.DB)
	r := gin.Default()
	r.Use(gin.Recovery())
	service.Init(r, svcCtx)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.CFG.Server.Host, config.CFG.Server.Port),
		Handler: r,
	}

	go func() {
		<-ctx.Done()

		logger.L().Info("http server shutting down")

		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.L().Error("http shutdown error", logger.Error(err))
		}
	}()
	logger.L().Info("http server started", logger.String("addr", srv.Addr))

	// 阻塞监听
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return 
	}
}
