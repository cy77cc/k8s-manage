package response

import (
	"net/http"
	"time"

	"github.com/cy77cc/k8s-manage/internal/utils"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

type Resp struct {
	Code      xcode.Xcode `json:"code"`
	Msg       string      `json:"msg"`
	Data      any         `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

func NewResp(code xcode.Xcode, msg string, data any) *Resp {
	return &Resp{
		Code:      code,
		Msg:       msg,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// Error returns an error response
func errorResp(code xcode.Xcode, msg string) *Resp {
	return &Resp{
		Code:      code,
		Msg:       msg,
		Data:      nil,
		Timestamp: utils.GetTimestamp(),
	}
}

// Success returns a success response
func success(data interface{}) *Resp {
	return &Resp{
		Code:      xcode.Success,
		Msg:       xcode.Success.Msg(),
		Data:      data,
		Timestamp: utils.GetTimestamp(),
	}
}

// Response unifies HTTP response handling
func Response(c *gin.Context, resp interface{}, err error) {
	if err != nil {
		// Convert error to CodeError
		codeErr := xcode.FromError(err)

		// Set HTTP Status Code based on Xcode
		// We can choose to always return 200 and let client check Code,
		// or return semantic HTTP status codes.
		// For this implementation, let's try to return semantic codes if they are 4xx/5xx,
		// but typically for "business errors" (like 4004 Password Error), many systems return 200 OK.
		// However, xcode.HttpStatus() handles this mapping.
		// If we want to strictly follow REST, we use w.WriteHeader.
		// But httpx.WriteJson writes 200 by default unless we use WriteHeader before.
		// httpx.OkJson writes 200.

		// Let's use 200 for everything to simplify client parsing, unless it's a severe infrastructure error?
		// User requested "improve code".
		// Best practice: Use 200 for business logic errors (Code > 0), use 500 for crashes.
		// But let's stick to what xcode.HttpStatus() returns if we want semantic.
		// For now, I will return 200 OK with the JSON body containing the error code.
		// This is consistent with many microservice frontends.

		c.JSON(http.StatusOK, errorResp(codeErr.Code, codeErr.Msg))
	} else {
		c.JSON(http.StatusOK, success(resp))
	}
}
