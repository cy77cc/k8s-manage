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
		accessTokenH := c.Request.Header.Get("Authorization")
		resp := response.NewResp(xcode.TokenInvalid, "请求未携带 token，无法访问", nil)
		if accessTokenH == "" {
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
		}

		accessToken, err := utils.ParseToken(accessTokenH)

		if err != nil {
			resp.Msg = "token 无效请重新登录"
			c.JSON(http.StatusUnauthorized, resp)
		}

		c.Set("uid", accessToken.Uid)

		c.Next()
	}
}
