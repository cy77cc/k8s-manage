package utils

import (
	"errors"
	"time"

	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/golang-jwt/jwt/v5"
)

// MyClaims Id为微信openid
type MyClaims struct {
	Uid uint `json:"uid"`
	jwt.RegisteredClaims
}

// 常量
var (
	ErrTokenExpired     = errors.New("Token is expired")
	ErrTokenInvalid     = errors.New("Token is invalid")
	ErrTokenMalformed   = errors.New("Token is malformed")
	ErrTokenNotValidYet = errors.New("Token is not valid yet")
	ErrTokenNotValidId  = errors.New("Token is not valid id")
	ErrTokenSignature   = errors.New("Token signature is invalid")
	MySecret            = []byte(config.CFG.JWT.Secret)
)

// GenToken 使用hsa256加密生成token, 传递weixinopenid
func GenToken(id uint, isRefreshToken bool) (string, error) {

	var tokenExpireDuration = config.CFG.JWT.Expire

	if isRefreshToken {
		tokenExpireDuration = config.CFG.JWT.RefreshExpire
	}

	c := MyClaims{
		id,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpireDuration)),
			Issuer:    config.CFG.JWT.Issuer,
		},
	}
	// 指定加密方式
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// 使用密钥加密
	return token.SignedString(MySecret)
}

// ParseToken 解析token
func ParseToken(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&MyClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return MySecret, nil
		},
	)
	if err != nil {
		return nil, xcode.NewErrCode(xcode.TokenInvalid)
	}
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, xcode.NewErrCode(xcode.TokenInvalid)
}

// RefreshToken 刷新token
func RefreshToken(id uint, isRefreshToken bool) (string, error) {
	token, err := GenToken(id, isRefreshToken)
	if err != nil {
		return "", err
	}
	return token, nil
}
