package ai

import (
	"bytes"
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	coreai "github.com/cy77cc/OpsPilot/internal/ai"
	"github.com/cy77cc/OpsPilot/internal/config"
	"github.com/cy77cc/OpsPilot/internal/svc"
	"github.com/cy77cc/OpsPilot/internal/testutil"
	"github.com/gin-gonic/gin"
)

type fakeRuntime struct {
	runFn          func(context.Context, coreai.RunRequest, coreai.StreamEmitter) error
	resumeFn       func(context.Context, coreai.ResumeRequest) (*coreai.ResumeResult, error)
	resumeStreamFn func(context.Context, coreai.ResumeRequest, coreai.StreamEmitter) (*coreai.ResumeResult, error)
}

func (f *fakeRuntime) Run(ctx context.Context, req coreai.RunRequest, emit coreai.StreamEmitter) error {
	return f.runFn(ctx, req, emit)
}

func (f *fakeRuntime) Resume(ctx context.Context, req coreai.ResumeRequest) (*coreai.ResumeResult, error) {
	return f.resumeFn(ctx, req)
}

func (f *fakeRuntime) ResumeStream(ctx context.Context, req coreai.ResumeRequest, emit coreai.StreamEmitter) (*coreai.ResumeResult, error) {
	return f.resumeStreamFn(ctx, req, emit)
}

func TestChatRoutesToAIV2WhenFeatureFlagEnabled(t *testing.T) {
	suite := testutil.NewIntegrationSuite(t)
	t.Cleanup(suite.Cleanup)
	prev := config.CFG.FeatureFlags.AIAssistantV2
	enabled := true
	config.CFG.FeatureFlags.AIAssistantV2 = &enabled
	t.Cleanup(func() {
		config.CFG.FeatureFlags.AIAssistantV2 = prev
	})

	handler := &HTTPHandler{
		svcCtx: &svc.ServiceContext{DB: suite.DB},
		orchestrator: &fakeRuntime{
			runFn: func(context.Context, coreai.RunRequest, coreai.StreamEmitter) error {
				t.Fatalf("legacy runtime should not be called")
				return nil
			},
			resumeFn:       func(context.Context, coreai.ResumeRequest) (*coreai.ResumeResult, error) { return nil, nil },
			resumeStreamFn: func(context.Context, coreai.ResumeRequest, coreai.StreamEmitter) (*coreai.ResumeResult, error) { return nil, nil },
		},
		aiv2: &fakeRuntime{
			runFn: func(_ context.Context, _ coreai.RunRequest, emit coreai.StreamEmitter) error {
				emit(coreai.StreamEvent{Type: "meta", Data: map[string]any{"session_id": "session-1"}})
				emit(coreai.StreamEvent{Type: "done", Data: map[string]any{"status": "completed"}})
				return nil
			},
			resumeFn:       func(context.Context, coreai.ResumeRequest) (*coreai.ResumeResult, error) { return nil, nil },
			resumeStreamFn: func(context.Context, coreai.ResumeRequest, coreai.StreamEmitter) (*coreai.ResumeResult, error) { return nil, nil },
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/ai/chat", bytes.NewBufferString(`{"message":"hello","context":{"scene":"global"}}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("uid", uint64(1))

	handler.Chat(c)

	if got := w.Header().Get("X-AI-Runtime-Mode"); got != "aiv2" {
		t.Fatalf("X-AI-Runtime-Mode = %q, want aiv2", got)
	}
	if !strings.Contains(w.Body.String(), "event: meta") || !strings.Contains(w.Body.String(), "event: done") {
		t.Fatalf("unexpected SSE body = %s", w.Body.String())
	}
}

func TestChatRoutesToLegacyWhenAIV2Disabled(t *testing.T) {
	suite := testutil.NewIntegrationSuite(t)
	t.Cleanup(suite.Cleanup)
	prev := config.CFG.FeatureFlags.AIAssistantV2
	disabled := false
	config.CFG.FeatureFlags.AIAssistantV2 = &disabled
	t.Cleanup(func() {
		config.CFG.FeatureFlags.AIAssistantV2 = prev
	})

	handler := &HTTPHandler{
		svcCtx: &svc.ServiceContext{DB: suite.DB},
		orchestrator: &fakeRuntime{
			runFn: func(_ context.Context, _ coreai.RunRequest, emit coreai.StreamEmitter) error {
				emit(coreai.StreamEvent{Type: "meta", Data: map[string]any{"session_id": "session-legacy"}})
				emit(coreai.StreamEvent{Type: "done", Data: map[string]any{"status": "completed"}})
				return nil
			},
			resumeFn:       func(context.Context, coreai.ResumeRequest) (*coreai.ResumeResult, error) { return nil, nil },
			resumeStreamFn: func(context.Context, coreai.ResumeRequest, coreai.StreamEmitter) (*coreai.ResumeResult, error) { return nil, nil },
		},
		aiv2: &fakeRuntime{
			runFn: func(context.Context, coreai.RunRequest, coreai.StreamEmitter) error {
				t.Fatalf("aiv2 runtime should not be called")
				return nil
			},
			resumeFn:       func(context.Context, coreai.ResumeRequest) (*coreai.ResumeResult, error) { return nil, nil },
			resumeStreamFn: func(context.Context, coreai.ResumeRequest, coreai.StreamEmitter) (*coreai.ResumeResult, error) { return nil, nil },
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/ai/chat", bytes.NewBufferString(`{"message":"hello","context":{"scene":"global"}}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("uid", uint64(1))

	handler.Chat(c)

	if got := w.Header().Get("X-AI-Runtime-Mode"); got == "aiv2" {
		t.Fatalf("X-AI-Runtime-Mode = %q, did not expect aiv2", got)
	}
	if !strings.Contains(w.Body.String(), "session-legacy") {
		t.Fatalf("unexpected SSE body = %s", w.Body.String())
	}
}
