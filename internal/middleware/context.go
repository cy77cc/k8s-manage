package middleware

import (
	"time"

	"github.com/cy77cc/k8s-manage/internal/xctx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 初始化自定义 Context
		myCtx := xctx.NewContext()
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.NewString()
		}
		myCtx.TraceID = traceID
		myCtx.StartTime = time.Now().Unix()

		// 2. 将其注入到标准 Context 中
		// 这一点至关重要：这样后续的 Service 层即使只拿到了 context.Context，也能取出 myCtx
		ctx := xctx.WithContext(c.Request.Context(), myCtx)

		// 3. 更新 Gin 的 Request，以便后续 c.Request.Context() 能拿到包含 myCtx 的 context
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
