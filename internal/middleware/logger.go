package middleware

import (
	"time"

	"github.com/cy77cc/k8s-manage/internal/logger"
	"github.com/cy77cc/k8s-manage/internal/xctx"
	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ctx := c.Request.Context()
		myCtx := xctx.FromContext(ctx)

		c.Next()

		logger.L().Info(
			"http request",
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.Int("status", c.Writer.Status()),
			logger.String("trace_id", myCtx.TraceID),
			logger.String("latency", time.Since(start).String()),
		)
	}
}
