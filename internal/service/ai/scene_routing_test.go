package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSceneToolsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newCommandTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/scene/services:list/tools", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "scene", Value: "services:list"}}
	c.Set("uid", uint64(1))

	h.sceneTools(c)
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", w.Code)
	}
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Scene string `json:"scene"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != 1000 {
		t.Fatalf("unexpected business code: %d", resp.Code)
	}
	if resp.Data.Scene != "services:list" {
		t.Fatalf("unexpected scene: %s", resp.Data.Scene)
	}
}
