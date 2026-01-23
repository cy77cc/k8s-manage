package middleware

import (
	"fmt"
	"net/http"

	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/utils"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", config.CFG.Cors.AllowOrigins[0]) // 可将将 * 替换为指定的域名
		c.Header("Access-Control-Allow-Methods", utils.SlicesToString(config.CFG.Cors.AllowMethods, ","))
		c.Header("Access-Control-Allow-Headers", utils.SlicesToString(config.CFG.Cors.AllowHeaders, ","))
		c.Header("Access-Control-Expose-Headers", utils.SlicesToString(config.CFG.Cors.ExposeHeaders, ","))
		c.Header("Access-Control-Allow-Credentials", fmt.Sprintf("%t", config.CFG.Cors.AllowCredentials))
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}
