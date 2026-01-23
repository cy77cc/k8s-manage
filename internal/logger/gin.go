package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		traceID := uuid.NewString()
		ctx := WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		l := L()

		l.Info("xxxx")

		L().Info(
			"http request",
			String("method", c.Request.Method),
			String("path", c.Request.URL.Path),
			Int("status", c.Writer.Status()),
			String("trace_id", traceID),
			String("latency", time.Since(start).String()),
		)
	}
}
