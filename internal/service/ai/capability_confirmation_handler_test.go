package ai

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

func TestConfirmConfirmation_ApproveByOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newCommandTestHandler(t)
	if err := h.svcCtx.DB.AutoMigrate(&model.ConfirmationRequest{}); err != nil {
		t.Fatalf("auto migrate confirmation: %v", err)
	}
	svc := NewConfirmationService(h.svcCtx.DB)
	item, err := svc.RequestConfirmation(httptest.NewRequest(http.MethodGet, "/", nil).Context(), ConfirmationRequestInput{
		RequestUserID: 1,
		ToolName:      "host_batch_exec_apply",
		ToolMode:      "mutating",
		RiskLevel:     "medium",
		Timeout:       2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("request confirmation: %v", err)
	}

	raw, _ := json.Marshal(map[string]any{"approve": true})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/confirmations/"+item.ID+"/confirm", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: item.ID}}
	c.Set("uid", uint64(1))

	h.confirmConfirmation(c)

	var resp struct {
		Code xcode.Xcode `json:"code"`
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != xcode.Success {
		t.Fatalf("expected success, got code=%d body=%s", resp.Code, w.Body.String())
	}
	if resp.Data.ID != item.ID {
		t.Fatalf("expected id=%s, got %s", item.ID, resp.Data.ID)
	}
	if resp.Data.Status != confirmationStatusConfirmed {
		t.Fatalf("expected status=%s, got %s", confirmationStatusConfirmed, resp.Data.Status)
	}
}

func TestConfirmConfirmation_ForbiddenForNonOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newCommandTestHandler(t)
	if err := h.svcCtx.DB.AutoMigrate(&model.ConfirmationRequest{}); err != nil {
		t.Fatalf("auto migrate confirmation: %v", err)
	}
	svc := NewConfirmationService(h.svcCtx.DB)
	item, err := svc.RequestConfirmation(httptest.NewRequest(http.MethodGet, "/", nil).Context(), ConfirmationRequestInput{
		RequestUserID: 1,
		ToolName:      "host_batch_exec_apply",
		ToolMode:      "mutating",
		RiskLevel:     "medium",
		Timeout:       2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("request confirmation: %v", err)
	}

	raw, _ := json.Marshal(map[string]any{"approve": true})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/confirmations/"+item.ID+"/confirm", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: item.ID}}
	c.Set("uid", uint64(2))

	h.confirmConfirmation(c)

	var resp struct {
		Code xcode.Xcode `json:"code"`
		Msg  string      `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != xcode.Forbidden {
		t.Fatalf("expected forbidden, got code=%d body=%s", resp.Code, w.Body.String())
	}
}

