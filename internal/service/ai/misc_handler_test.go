package ai

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/gin-gonic/gin"
)

func TestBranchSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newCommandTestHandler(t)
	if err := h.svcCtx.DB.AutoMigrate(&model.AIChatSession{}, &model.AIChatMessage{}); err != nil {
		t.Fatalf("auto migrate ai chat tables: %v", err)
	}
	baseSessionID := "sess-source-1"
	if _, err := h.store.appendMessage(1, "services:list", baseSessionID, map[string]any{
		"id":        "u-1",
		"role":      "user",
		"content":   "先看服务列表",
		"timestamp": time.Now(),
	}); err != nil {
		t.Fatalf("append user message: %v", err)
	}
	if _, err := h.store.appendMessage(1, "services:list", baseSessionID, map[string]any{
		"id":        "a-1",
		"role":      "assistant",
		"content":   "这是服务列表结果",
		"timestamp": time.Now(),
	}); err != nil {
		t.Fatalf("append assistant message: %v", err)
	}

	body := map[string]any{
		"messageId": "u-1",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/sessions/sess-source-1/branch", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: baseSessionID}}
	c.Set("uid", uint64(1))

	h.branchSession(c)
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Code int `json:"code"`
		Data struct {
			ID       string `json:"id"`
			Messages []struct {
				ID string `json:"id"`
			} `json:"messages"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != 1000 {
		t.Fatalf("unexpected business code: %d", resp.Code)
	}
	if resp.Data.ID == "" || resp.Data.ID == baseSessionID {
		t.Fatalf("expected new branch session id, got %q", resp.Data.ID)
	}
	if len(resp.Data.Messages) != 1 {
		t.Fatalf("expected 1 message in branch, got %d", len(resp.Data.Messages))
	}
	if resp.Data.Messages[0].ID == "u-1" {
		t.Fatalf("expected cloned message id, got source id")
	}
}
