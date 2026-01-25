package middleware

import (
	"net/http"

	"github.com/cy77cc/k8s-manage/internal/response"
	"github.com/cy77cc/k8s-manage/internal/utils"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenH := c.Request.Header.Get("Authorization")
		resp := response.NewResp(xcode.TokenInvalid, "请求未携带 token，无法访问", nil)
		if tokenH == "" {
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
		}

		token, err := utils.ParseToken(tokenH)

		if err != nil {
			resp.Msg = "token 无效请重新登录"
			c.JSON(http.StatusUnauthorized, resp)
		}

		c.Set("uid", token.Uid)

		c.Next()
	}
}
