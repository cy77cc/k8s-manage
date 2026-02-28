package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

func TestResponseTypedNilCodeErrorDoesNotPanic(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	var codeErr *xcode.CodeError
	var err error = codeErr

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Response panicked with typed nil error: %v", r)
		}
	}()

	Response(ctx, nil, err)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected http status: got=%d want=%d", recorder.Code, http.StatusOK)
	}

	var resp Resp
	if unmarshalErr := json.Unmarshal(recorder.Body.Bytes(), &resp); unmarshalErr != nil {
		t.Fatalf("failed to parse response body: %v", unmarshalErr)
	}

	if resp.Code != xcode.ServerError {
		t.Fatalf("unexpected business code: got=%d want=%d", resp.Code, xcode.ServerError)
	}
}
