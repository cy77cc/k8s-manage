package ai

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	coreai "github.com/cy77cc/OpsPilot/internal/ai"
	aistate "github.com/cy77cc/OpsPilot/internal/ai/state"
	"github.com/cy77cc/OpsPilot/internal/testutil"
	"github.com/gin-gonic/gin"
)

func TestBuildResumeRequestMapsLegacyCheckpointToStepIdentity(t *testing.T) {
	req := buildResumeRequest(approvalResponseRequest{
		SessionID:    "session-1",
		PlanID:       "plan-1",
		CheckpointID: "legacy-step",
		Approved:     true,
	})

	if req.SessionID != "session-1" || req.PlanID != "plan-1" {
		t.Fatalf("unexpected request: %#v", req)
	}
	if req.StepID != "legacy-step" {
		t.Fatalf("StepID = %q, want %q", req.StepID, "legacy-step")
	}
	if req.Target != "legacy-step" {
		t.Fatalf("Target = %q, want %q", req.Target, "legacy-step")
	}
}

func TestBuildResumeResponseMarksLegacyADKCompatibility(t *testing.T) {
	payload := buildResumeResponse(&coreai.ResumeResult{
		Resumed:   true,
		SessionID: "session-1",
		PlanID:    "plan-1",
		StepID:    "step-1",
		Status:    "approved",
		Message:   "审批已通过，待审批步骤会继续执行。",
	}, true)

	if payload["compat_mode"] != "legacy_adk_resume" {
		t.Fatalf("compat_mode = %#v", payload["compat_mode"])
	}
	if payload["deprecated"] != true {
		t.Fatalf("deprecated = %#v", payload["deprecated"])
	}
	msg, _ := payload["message"].(string)
	if msg == "" || msg == "审批已通过，待审批步骤会继续执行。" {
		t.Fatalf("legacy message = %q", msg)
	}
}

func TestSessionHandlersRespectSceneAndExposeThoughtChain(t *testing.T) {
	suite := testutil.NewIntegrationSuite(t)
	t.Cleanup(suite.Cleanup)

	handler := NewHTTPHandler(suite.SvcCtx)
	store := aistate.NewChatStore(suite.DB)
	ctx := t.Context()

	if err := store.AppendUserMessage(ctx, "session-global", 1, "global", "全局对话", "你好"); err != nil {
		t.Fatalf("AppendUserMessage(global) error = %v", err)
	}
	assistantID, err := store.CreateAssistantMessage(ctx, "session-global", 1, "global", "全局对话")
	if err != nil {
		t.Fatalf("CreateAssistantMessage(global) error = %v", err)
	}
	if err := store.UpdateAssistantMessage(ctx, "session-global", assistantID, aistate.ChatMessageRecord{
		Content: "回答",
		Status:  "completed",
		TraceID: "trace-1",
		ThoughtChain: []map[string]any{
			{"key": "rewrite", "title": "理解你的问题", "status": "success"},
		},
	}); err != nil {
		t.Fatalf("UpdateAssistantMessage(global) error = %v", err)
	}
	if err := store.AppendUserMessage(ctx, "session-k8s", 1, "scene:k8s", "K8s 对话", "查日志"); err != nil {
		t.Fatalf("AppendUserMessage(k8s) error = %v", err)
	}

	t.Run("CurrentSession filters by scene", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/ai/sessions/current?scene=scene:k8s", nil)
		c.Request = req
		c.Set("uid", uint64(1))

		handler.CurrentSession(c)

		var resp struct {
			Data map[string]any `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.Data["id"] != "session-k8s" {
			t.Fatalf("CurrentSession id = %#v, want session-k8s", resp.Data["id"])
		}
	})

	t.Run("GetSession returns persisted thoughtChain", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/ai/sessions/session-global?scene=global", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "session-global"}}
		c.Set("uid", uint64(1))

		handler.GetSession(c)

		var resp struct {
			Data struct {
				ID       string           `json:"id"`
				Messages []map[string]any `json:"messages"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.Data.ID != "session-global" {
			t.Fatalf("GetSession id = %q, want session-global", resp.Data.ID)
		}
		if len(resp.Data.Messages) != 2 {
			t.Fatalf("GetSession messages = %#v, want 2", resp.Data.Messages)
		}
		if _, ok := resp.Data.Messages[1]["thoughtChain"]; !ok {
			t.Fatalf("assistant message missing thoughtChain: %#v", resp.Data.Messages[1])
		}
	})
}
