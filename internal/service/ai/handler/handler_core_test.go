package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	coreai "github.com/cy77cc/k8s-manage/internal/ai"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/service/ai/logic"
	"github.com/gin-gonic/gin"
)

type fakeOrchestrator struct {
	chatReq       coreai.ChatStreamRequest
	chatCalled    bool
	resumeCalled  bool
	resumePayload map[string]any
}

func (f *fakeOrchestrator) ChatStream(_ context.Context, req coreai.ChatStreamRequest, emit func(event string, payload map[string]any) bool) error {
	f.chatCalled = true
	f.chatReq = req
	emit("meta", map[string]any{"sessionId": "sess-test"})
	emit(coreai.EventPlanCreated, map[string]any{"plan_id": "plan-test", "objective": "检查磁盘"})
	emit("done", map[string]any{"stream_state": "ok"})
	return nil
}

func (f *fakeOrchestrator) ResumePayload(_ context.Context, _ string, _ map[string]any) (map[string]any, error) {
	f.resumeCalled = true
	return f.resumePayload, nil
}

type fakeGatewayRuntime struct {
	previewCalled      bool
	previewTool        string
	executeCalled      bool
	sessionListCalled  bool
	confirmCalled      bool
	executionCalled    bool
}

func (f *fakeGatewayRuntime) ListSessions(uid uint64, scene string) []*logic.AISession {
	f.sessionListCalled = true
	return []*logic.AISession{{ID: "sess-1", Scene: scene, Title: "test"}}
}
func (f *fakeGatewayRuntime) CurrentSession(uid uint64, scene string) (*logic.AISession, bool) {
	return &logic.AISession{ID: "sess-1", Scene: scene, Title: "test"}, true
}
func (f *fakeGatewayRuntime) GetSession(uid uint64, id string) (*logic.AISession, bool) {
	return &logic.AISession{ID: id, Scene: "global", Title: "test"}, true
}
func (f *fakeGatewayRuntime) BranchSession(uid uint64, sourceSessionID, anchorMessageID, title string) (*logic.AISession, error) {
	return &logic.AISession{ID: "sess-2", Scene: "global", Title: title}, nil
}
func (f *fakeGatewayRuntime) DeleteSession(uid uint64, id string) {}
func (f *fakeGatewayRuntime) UpdateSessionTitle(uid uint64, id, title string) (*logic.AISession, error) {
	return &logic.AISession{ID: id, Scene: "global", Title: title}, nil
}
func (f *fakeGatewayRuntime) PreviewTool(uid uint64, tool string, params map[string]any) (map[string]any, error) {
	f.previewCalled = true
	f.previewTool = tool
	return map[string]any{"tool": tool, "approval_required": true}, nil
}
func (f *fakeGatewayRuntime) ExecuteTool(ctx context.Context, uid uint64, tool string, params map[string]any, approvalToken string) (*logic.ExecutionRecord, error) {
	f.executeCalled = true
	return &logic.ExecutionRecord{ID: "exe-1", Tool: tool, Status: "succeeded"}, nil
}
func (f *fakeGatewayRuntime) GetExecution(id string) (*logic.ExecutionRecord, bool) {
	f.executionCalled = true
	return &logic.ExecutionRecord{ID: id, Tool: "host_exec", Status: "succeeded"}, true
}
func (f *fakeGatewayRuntime) CreateApproval(uid uint64, tool string, params map[string]any) (*logic.ApprovalTicket, error) {
	return &logic.ApprovalTicket{ID: "apv-1", Tool: tool, Status: "pending"}, nil
}
func (f *fakeGatewayRuntime) ConfirmApproval(uid uint64, id string, approve bool) (*logic.ApprovalTicket, error) {
	return &logic.ApprovalTicket{ID: id, Tool: "host_exec", Status: "approved"}, nil
}
func (f *fakeGatewayRuntime) ConfirmConfirmation(ctx context.Context, uid uint64, id string, approve bool) (*model.ConfirmationRequest, error) {
	f.confirmCalled = true
	return &model.ConfirmationRequest{ID: id, Status: "confirmed"}, nil
}

func newTestChatContext(t *testing.T, body any) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", bytes.NewReader(raw))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("uid", uint64(7))
	return ctx, recorder
}

func TestChatDelegatesToAICoreOrchestrator(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orch := &fakeOrchestrator{}
	h := &AIHandler{
		orchestrator: orch,
		gateway:      &fakeGatewayRuntime{},
		sessions:     logic.NewSessionStore(nil, nil),
		runtime:      logic.NewRuntimeStore(nil),
	}

	ctx, recorder := newTestChatContext(t, map[string]any{
		"message": "检查磁盘",
		"context": map[string]any{"scene": "global"},
	})
	h.chat(ctx)

	if !orch.chatCalled {
		t.Fatalf("expected orchestrator chat to be called")
	}
	if orch.chatReq.Message != "检查磁盘" {
		t.Fatalf("unexpected message: %q", orch.chatReq.Message)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "event: meta") || !strings.Contains(body, "event: plan_created") || !strings.Contains(body, "event: done") {
		t.Fatalf("expected SSE meta, plan_created and done events, got %q", body)
	}
}

func TestResumeApprovalDelegatesToAICoreOrchestrator(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orch := &fakeOrchestrator{resumePayload: map[string]any{"resumed": true, "content": "ok"}}
	h := &AIHandler{
		orchestrator: orch,
		gateway:      &fakeGatewayRuntime{},
		sessions:     logic.NewSessionStore(nil, nil),
		runtime:      logic.NewRuntimeStore(nil),
	}

	ctx, recorder := newTestChatContext(t, map[string]any{
		"checkpoint_id": "sess-1",
		"target":        "call-1",
		"data":          true,
	})
	h.resumeADKApproval(ctx)

	if !orch.resumeCalled {
		t.Fatalf("expected orchestrator resume to be called")
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

type fakeControlPlane struct{}

func (f *fakeControlPlane) ToolPolicy(context.Context, tools.ToolMeta, map[string]any) error { return nil }
func (f *fakeControlPlane) HasPermission(uint64, string) bool                                { return true }
func (f *fakeControlPlane) IsAdmin(uint64) bool                                              { return false }
func (f *fakeControlPlane) FindMeta(string) (tools.ToolMeta, bool)                           { return tools.ToolMeta{}, false }

func TestPreviewToolDelegatesToGatewayRuntime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gateway := &fakeGatewayRuntime{}
	h := &AIHandler{gateway: gateway, control: &fakeControlPlane{}}
	ctx, recorder := newTestChatContext(t, map[string]any{
		"tool": "host_batch_exec_apply",
		"params": map[string]any{"host_id": 1},
	})
	h.previewTool(ctx)
	if !gateway.previewCalled {
		t.Fatalf("expected preview runtime to be called")
	}
	if gateway.previewTool != "host_batch_exec_apply" {
		t.Fatalf("unexpected preview tool: %q", gateway.previewTool)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestListSessionsDelegatesToGatewayRuntime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gateway := &fakeGatewayRuntime{}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/ai/sessions?scene=global", nil)
	ctx.Set("uid", uint64(7))
	h := &AIHandler{gateway: gateway}
	h.listSessions(ctx)
	if !gateway.sessionListCalled {
		t.Fatalf("expected session listing to delegate to gateway runtime")
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestConfirmConfirmationDelegatesToGatewayRuntime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gateway := &fakeGatewayRuntime{}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	body := bytes.NewReader([]byte(`{"approve":true}`))
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/ai/confirmations/cfm-1/confirm", body)
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "id", Value: "cfm-1"}}
	ctx.Set("uid", uint64(7))
	h := &AIHandler{gateway: gateway}
	h.confirmConfirmation(ctx)
	if !gateway.confirmCalled {
		t.Fatalf("expected confirmation to delegate to gateway runtime")
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestExecuteToolDelegatesToGatewayRuntime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gateway := &fakeGatewayRuntime{}
	h := &AIHandler{gateway: gateway}
	ctx, recorder := newTestChatContext(t, map[string]any{
		"tool": "host_exec_readonly",
		"params": map[string]any{"host_id": 1, "command": "df -h"},
	})
	h.executeTool(ctx)
	if !gateway.executeCalled {
		t.Fatalf("expected execute tool to delegate to gateway runtime")
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestGetExecutionDelegatesToGatewayRuntime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gateway := &fakeGatewayRuntime{}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/ai/executions/exe-1", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "exe-1"}}
	ctx.Set("uid", uint64(7))
	h := &AIHandler{gateway: gateway}
	h.getExecution(ctx)
	if !gateway.executionCalled {
		t.Fatalf("expected getExecution to delegate to gateway runtime")
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}
