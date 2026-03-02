package httpx

import (
	"net/http"

	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

type response struct {
	Code xcode.Xcode `json:"code"`
	Msg  string      `json:"msg"`
	Data any         `json:"data,omitempty"`
}

// OK writes a successful response (code 1000) with the given data.
// Always returns HTTP 200.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, response{
		Code: xcode.Success,
		Msg:  xcode.Success.Msg(),
		Data: data,
	})
}

// Fail writes an error response with the given xcode.
// msg overrides the default message when non-empty.
// Always returns HTTP 200.
func Fail(c *gin.Context, code xcode.Xcode, msg string) {
	if msg == "" {
		msg = code.Msg()
	}
	c.JSON(http.StatusOK, response{
		Code: code,
		Msg:  msg,
	})
}

// BindErr writes a parameter-binding error response (code 2000).
// Always returns HTTP 200.
func BindErr(c *gin.Context, err error) {
	Fail(c, xcode.ParamError, err.Error())
}

// ServerErr writes a server error response (code 3000).
// Always returns HTTP 200.
func ServerErr(c *gin.Context, err error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	Fail(c, xcode.ServerError, msg)
}

// NotFound writes a not found error response.
// Always returns HTTP 200.
func NotFound(c *gin.Context, msg string) {
	if msg == "" {
		msg = "Resource not found"
	}
	Fail(c, xcode.NotFound, msg)
}

// BadRequest writes a bad request error response.
// Always returns HTTP 200.
func BadRequest(c *gin.Context, msg string) {
	Fail(c, xcode.ParamError, msg)
}
