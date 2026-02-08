package middleware

import (
	"net/http"
	"strings"

	"github.com/cy77cc/k8s-manage/internal/response"
	"github.com/cy77cc/k8s-manage/internal/utils"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessTokenH := c.Request.Header.Get("Authorization")
		resp := response.NewResp(xcode.TokenInvalid, "请求未携带 token，无法访问", nil)
		if accessTokenH == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, xcode.NewErrCode(xcode.Unauthorized))
			return
		}

		parts := strings.SplitN(accessTokenH, " ", 2)

		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, xcode.NewErrCode(xcode.Unauthorized))
			return
		}

		tokenStr := parts[1]

		accessToken, err := utils.ParseToken(tokenStr)

		if err != nil {
			resp.Msg = "token 无效请重新登录"
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
		}

		c.Set("uid", accessToken.Uid)

		c.Next()
	}
}
