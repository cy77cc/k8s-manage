package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	modelcomponent "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	aiplatform "github.com/cy77cc/k8s-manage/internal/ai"
	aitools "github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

type fakeE2EToolCallingModel struct{}

func (m *fakeE2EToolCallingModel) Generate(_ context.Context, input []*schema.Message, _ ...modelcomponent.Option) (*schema.Message, error) {
	last := ""
	for i := len(input) - 1; i >= 0; i-- {
		if input[i] != nil && input[i].Role == schema.User {
			last = strings.TrimSpace(input[i].Content)
			break
		}
	}
	if strings.Contains(last, "suggestion 智能体") {
		return schema.AssistantMessage("后续检查|查看会话记录|0.8|继续收敛本轮结果", nil), nil
	}
	return schema.AssistantMessage("已处理: "+last, nil), nil
}

func (m *fakeE2EToolCallingModel) Stream(_ context.Context, input []*schema.Message, _ ...modelcomponent.Option) (*schema.StreamReader[*schema.Message], error) {
	msg, err := m.Generate(context.Background(), input)
	if err != nil {
		return nil, err
	}
	sr, sw := schema.Pipe[*schema.Message](0)
	go func() {
		defer sw.Close()
		sw.Send(msg, nil)
	}()
	return sr, nil
}

func (m *fakeE2EToolCallingModel) WithTools(_ []*schema.ToolInfo) (modelcomponent.ToolCallingChatModel, error) {
	return m, nil
}

func newE2ETestHandler(t *testing.T) *handler {
	t.Helper()
	h := newCommandTestHandler(t)
	if err := h.svcCtx.DB.AutoMigrate(
		&model.AIChatSession{},
		&model.AIChatMessage{},
		&model.Cluster{},
	); err != nil {
		t.Fatalf("auto migrate e2e tables: %v", err)
	}
	seedPermissionTestData(t, h)
	if err := h.svcCtx.DB.Create(&model.Cluster{
		ID:      20,
		Name:    fmt.Sprintf("cluster-%s", strings.ReplaceAll(t.Name(), "/", "-")),
		Status:  "ready",
		Type:    "kubernetes",
		EnvType: "production",
	}).Error; err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	runner, err := aiplatform.NewPlatformRunner(
		context.Background(),
		&fakeE2EToolCallingModel{},
		aitools.PlatformDeps{DB: h.svcCtx.DB},
		&aiplatform.RunnerConfig{EnableStreaming: true},
	)
	if err != nil {
		t.Fatalf("new platform runner: %v", err)
	}
	h.svcCtx.AI = runner
	return h
}

func decodeSSEEvents(t *testing.T, body string) []map[string]any {
	t.Helper()
	chunks := strings.Split(strings.TrimSpace(body), "\n\n")
	out := make([]map[string]any, 0, len(chunks))
	for _, chunk := range chunks {
		lines := strings.Split(strings.TrimSpace(chunk), "\n")
		if len(lines) < 2 {
			continue
		}
		eventName := strings.TrimSpace(strings.TrimPrefix(lines[0], "event:"))
		raw := strings.TrimSpace(strings.TrimPrefix(lines[1], "data:"))
		if eventName == "" || raw == "" {
			continue
		}
		payload := map[string]any{"event": eventName}
		var data map[string]any
		if err := json.Unmarshal([]byte(raw), &data); err != nil {
			t.Fatalf("decode sse payload: %v", err)
		}
		for k, v := range data {
			payload[k] = v
		}
		out = append(out, payload)
	}
	return out
}

func newJSONContext(method, target string, body any, uid uint64) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, target, reader)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Set("uid", uid)
	return c, w
}

func TestUserConversationFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newE2ETestHandler(t)

	c, w := newJSONContext(http.MethodPost, "/api/v1/ai/chat", map[string]any{
		"message": "帮我总结当前会话状态",
		"context": map[string]any{"scene": "hosts"},
	}, 1)
	h.chat(c)

	body := w.Body.String()
	if !strings.Contains(body, "event: meta") || !strings.Contains(body, "event: done") {
		t.Fatalf("expected meta and done events, got: %s", body)
	}
	events := decodeSSEEvents(t, body)
	if len(events) == 0 {
		t.Fatalf("expected sse events")
	}
	var done map[string]any
	for _, item := range events {
		if item["event"] == "done" {
			done = item
			break
		}
	}
	if done == nil {
		t.Fatalf("expected done event in stream")
	}
	if done["stream_state"] != "ok" {
		t.Fatalf("expected ok stream state, got %#v", done["stream_state"])
	}

	session, ok := h.sessions.CurrentSession(1, "hosts")
	if !ok || session == nil {
		t.Fatalf("expected persisted current session")
	}
	if len(session.Messages) != 2 {
		t.Fatalf("expected user + assistant messages, got %d", len(session.Messages))
	}

	getCtx, getW := newJSONContext(http.MethodGet, "/api/v1/ai/sessions/"+session.ID, nil, 1)
	getCtx.Params = gin.Params{{Key: "id", Value: session.ID}}
	h.getSession(getCtx)

	var resp struct {
		Code xcode.Xcode `json:"code"`
		Data aiSession   `json:"data"`
	}
	if err := json.Unmarshal(getW.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode get session response: %v", err)
	}
	if resp.Code != xcode.Success {
		t.Fatalf("expected success, got body=%s", getW.Body.String())
	}
	if resp.Data.ID != session.ID {
		t.Fatalf("expected session id %s, got %s", session.ID, resp.Data.ID)
	}
}

func TestToolExecutionFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newE2ETestHandler(t)

	c, w := newJSONContext(http.MethodPost, "/api/v1/ai/tools/execute", map[string]any{
		"tool": "service_status",
		"params": map[string]any{
			"service_id": 10,
		},
	}, 1)
	h.executeTool(c)

	var resp struct {
		Code xcode.Xcode     `json:"code"`
		Data executionRecord `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode execute response: %v", err)
	}
	if resp.Code != xcode.Success {
		t.Fatalf("expected success, got body=%s", w.Body.String())
	}
	if resp.Data.Status != "succeeded" {
		t.Fatalf("expected succeeded execution, got %#v", resp.Data.Status)
	}
	if resp.Data.Result == nil || !resp.Data.Result.OK {
		t.Fatalf("expected successful tool result, got %#v", resp.Data.Result)
	}

	getCtx, getW := newJSONContext(http.MethodGet, "/api/v1/ai/executions/"+resp.Data.ID, nil, 1)
	getCtx.Params = gin.Params{{Key: "id", Value: resp.Data.ID}}
	h.getExecution(getCtx)

	var getResp struct {
		Code xcode.Xcode     `json:"code"`
		Data executionRecord `json:"data"`
	}
	if err := json.Unmarshal(getW.Body.Bytes(), &getResp); err != nil {
		t.Fatalf("decode get execution response: %v", err)
	}
	if getResp.Code != xcode.Success {
		t.Fatalf("expected success on get execution, got body=%s", getW.Body.String())
	}
	if getResp.Data.ID != resp.Data.ID {
		t.Fatalf("expected execution id %s, got %s", resp.Data.ID, getResp.Data.ID)
	}
}

func TestApprovalConfirmationFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newE2ETestHandler(t)

	createCtx, createW := newJSONContext(http.MethodPost, "/api/v1/ai/approvals", map[string]any{
		"tool": "service_deploy_apply",
		"params": map[string]any{
			"service_id": 10,
			"cluster_id": 20,
		},
	}, 1)
	h.createApproval(createCtx)

	var createResp struct {
		Code xcode.Xcode    `json:"code"`
		Data approvalTicket `json:"data"`
	}
	if err := json.Unmarshal(createW.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create approval response: %v", err)
	}
	if createResp.Code != xcode.Success {
		t.Fatalf("expected success, got body=%s", createW.Body.String())
	}
	if createResp.Data.Status != "pending" {
		t.Fatalf("expected pending approval, got %#v", createResp.Data.Status)
	}

	confirmCtx, confirmW := newJSONContext(http.MethodPost, "/api/v1/ai/approvals/"+createResp.Data.ID+"/confirm", map[string]any{
		"approve": true,
	}, 1)
	confirmCtx.Params = gin.Params{{Key: "id", Value: createResp.Data.ID}}
	h.confirmApproval(confirmCtx)

	var confirmResp struct {
		Code xcode.Xcode    `json:"code"`
		Data approvalTicket `json:"data"`
	}
	if err := json.Unmarshal(confirmW.Body.Bytes(), &confirmResp); err != nil {
		t.Fatalf("decode confirm approval response: %v", err)
	}
	if confirmResp.Code != xcode.Success {
		t.Fatalf("expected success, got body=%s", confirmW.Body.String())
	}
	if confirmResp.Data.Status != "approved" {
		t.Fatalf("expected approved status, got %#v", confirmResp.Data.Status)
	}
	if confirmResp.Data.ReviewUID != 1 {
		t.Fatalf("expected reviewer 1, got %d", confirmResp.Data.ReviewUID)
	}
}
