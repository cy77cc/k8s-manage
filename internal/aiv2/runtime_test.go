package aiv2

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/OpsPilot/internal/ai/events"
	legacyai "github.com/cy77cc/OpsPilot/internal/ai"
	"github.com/redis/go-redis/v9"
)

type memoryCheckpointStore struct {
	data map[string][]byte
}

func newMemoryCheckpointStore() *memoryCheckpointStore {
	return &memoryCheckpointStore{data: make(map[string][]byte)}
}

func (s *memoryCheckpointStore) Get(_ context.Context, id string) ([]byte, bool, error) {
	value, ok := s.data[id]
	return value, ok, nil
}

func (s *memoryCheckpointStore) Set(_ context.Context, id string, value []byte) error {
	s.data[id] = value
	return nil
}

func newTestRuntime(t *testing.T) *Runtime {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
		mr.Close()
	})
	return &Runtime{
		store:      NewPendingStore(client, "ai:v2:test:pending:", time.Hour),
		checkpoint: newRedisCheckpointStore(client, "ai:v2:test:checkpoint:", time.Hour),
		metrics:    &Metrics{},
	}
}

func TestConsumePersistsPendingApprovalOnInterrupt(t *testing.T) {
	runtime := newTestRuntime(t)
	iter, gen := adk.NewAsyncIteratorPair[*adk.AgentEvent]()
	go func() {
		gen.Send(&adk.AgentEvent{
			Action: &adk.AgentAction{
				Interrupted: &adk.InterruptInfo{
					InterruptContexts: []*adk.InterruptCtx{
						{
							ID:          "interrupt-1",
							Info:        &ApprovalInterruptInfo{ToolName: "host_batch_exec_apply", Summary: "Approve host mutation", Risk: "high", Mode: "mutating"},
							IsRootCause: true,
						},
					},
				},
			},
		})
		gen.Close()
	}()

	answer, thinking, interrupted, interruptID, err := runtime.consume(
		context.Background(),
		iter,
		events.EventMeta{SessionID: "session-1", TurnID: "turn-1", TraceID: "trace-1"},
		nil,
		"session-1",
		"turn-1",
		"trace-1",
		"checkpoint-1",
		map[string]ToolPolicy{"host_batch_exec_apply": {Name: "host_batch_exec_apply", Expert: "hostops", Risk: "high", Mode: "mutating"}},
	)
	if err != nil {
		t.Fatalf("consume error = %v", err)
	}
	if answer != "" || thinking != "" {
		t.Fatalf("unexpected answer/thinking = %q / %q", answer, thinking)
	}
	if !interrupted || interruptID != "interrupt-1" {
		t.Fatalf("interrupt = %v, interruptID = %q", interrupted, interruptID)
	}
	pending, err := runtime.store.Get(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("store.Get error = %v", err)
	}
	if pending == nil || pending.ToolName != "host_batch_exec_apply" || pending.CheckPointID != "checkpoint-1" {
		t.Fatalf("unexpected pending approval = %#v", pending)
	}
}

func TestConsumeProjectsSummaryAnswer(t *testing.T) {
	runtime := newTestRuntime(t)
	iter, gen := adk.NewAsyncIteratorPair[*adk.AgentEvent]()
	go func() {
		gen.Send(&adk.AgentEvent{
			Output: &adk.AgentOutput{
				MessageOutput: &adk.MessageVariant{
					Message: schema.AssistantMessage("## Final Result\n\nDone.", nil),
				},
			},
		})
		gen.Close()
	}()

	var emitted []legacyai.StreamEvent
	answer, thinking, interrupted, _, err := runtime.consume(
		context.Background(),
		iter,
		events.EventMeta{SessionID: "session-1", TurnID: "turn-1", TraceID: "trace-1"},
		func(evt legacyai.StreamEvent) bool {
			emitted = append(emitted, evt)
			return true
		},
		"session-1",
		"turn-1",
		"trace-1",
		"checkpoint-1",
		nil,
	)
	if err != nil {
		t.Fatalf("consume error = %v", err)
	}
	if interrupted {
		t.Fatalf("unexpected interrupt")
	}
	if thinking != "" {
		t.Fatalf("unexpected thinking = %q", thinking)
	}
	if answer != "## Final Result\n\nDone." {
		t.Fatalf("answer = %q", answer)
	}
	if len(emitted) == 0 || emitted[len(emitted)-1].Type != events.Delta {
		t.Fatalf("expected delta events, got %#v", emitted)
	}
}

type resumableApprovalAgent struct{}

func (a *resumableApprovalAgent) Name(context.Context) string        { return "aiv2-test" }
func (a *resumableApprovalAgent) Description(context.Context) string { return "test" }

func (a *resumableApprovalAgent) Run(ctx context.Context, _ *adk.AgentInput, _ ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	it, gen := adk.NewAsyncIteratorPair[*adk.AgentEvent]()
	go func() {
		gen.Send(adk.StatefulInterrupt(ctx, &ApprovalInterruptInfo{
			ToolName:  "host_batch_exec_apply",
			Summary:   "Approve host mutation",
			SessionID: "session-1",
			TurnID:    "turn-1",
		}, "state"))
		gen.Close()
	}()
	return it
}

func (a *resumableApprovalAgent) Resume(ctx context.Context, info *adk.ResumeInfo, _ ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	it, gen := adk.NewAsyncIteratorPair[*adk.AgentEvent]()
	go func() {
		decision, _ := info.ResumeData.(*ApprovalDecision)
		content := "## Final Result\n\nApproved."
		if decision != nil && !decision.Approved {
			content = "## Final Result\n\nRejected."
		}
		gen.Send(&adk.AgentEvent{
			Output: &adk.AgentOutput{
				MessageOutput: &adk.MessageVariant{Message: schema.AssistantMessage(content, nil)},
			},
		})
		gen.Close()
	}()
	return it
}

func TestResumeApprovedAndRejectedTerminalCompletion(t *testing.T) {
	runtime := newTestRuntime(t)
	store := newMemoryCheckpointStore()
	agent := &resumableApprovalAgent{}
	runner := adk.NewRunner(context.Background(), adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: true,
		CheckPointStore: store,
	})
	iter := runner.Run(context.Background(), []adk.Message{schema.UserMessage("run")}, adk.WithCheckPointID("cp-resume"))
	var interruptID string
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Action != nil && event.Action.Interrupted != nil && len(event.Action.Interrupted.InterruptContexts) > 0 {
			interruptID = event.Action.Interrupted.InterruptContexts[0].ID
		}
	}
	if interruptID == "" {
		t.Fatalf("expected interrupt id from initial run")
	}
	if err := runtime.store.Save(context.Background(), PendingApproval{
		CheckPointID: "cp-resume",
		InterruptID:  interruptID,
		SessionID:    "session-1",
		TurnID:       "turn-1",
		TraceID:      "trace-1",
		ToolName:     "host_batch_exec_apply",
		Summary:      "Approve host mutation",
		CreatedAt:    time.Now().UTC(),
	}); err != nil {
		t.Fatalf("save pending error = %v", err)
	}
	runtime.buildFn = func(ctx context.Context, sessionID, turnID string, runtimeCtx map[string]any) (adk.Agent, *adk.Runner, map[string]ToolPolicy, error) {
		return agent, adk.NewRunner(ctx, adk.RunnerConfig{
			Agent:           agent,
			EnableStreaming: true,
			CheckPointStore: store,
		}), nil, nil
	}

	answer, approved, err := runtime.resume(context.Background(), legacyai.ResumeRequest{SessionID: "session-1", Approved: true}, nil)
	if err != nil {
		t.Fatalf("approved resume error = %v", err)
	}
	if approved == nil || approved.Status != "approved" || answer == "" {
		t.Fatalf("unexpected approved resume result = %#v, answer = %q", approved, answer)
	}

	if err := runtime.store.Save(context.Background(), PendingApproval{
		CheckPointID: "cp-resume",
		InterruptID:  interruptID,
		SessionID:    "session-1",
		TurnID:       "turn-1",
		TraceID:      "trace-1",
		ToolName:     "host_batch_exec_apply",
		Summary:      "Approve host mutation",
		CreatedAt:    time.Now().UTC(),
	}); err != nil {
		t.Fatalf("save pending error = %v", err)
	}
	answer, rejected, err := runtime.resume(context.Background(), legacyai.ResumeRequest{SessionID: "session-1", Approved: false, Reason: "cancel"}, nil)
	if err != nil {
		t.Fatalf("rejected resume error = %v", err)
	}
	if rejected == nil || rejected.Status != "rejected" || answer == "" {
		t.Fatalf("unexpected rejected resume result = %#v, answer = %q", rejected, answer)
	}
}

func TestObservabilityMiddlewareCountsToolCalls(t *testing.T) {
	metrics := &Metrics{}
	middleware := NewObservabilityMiddleware(metrics)
	endpoint, err := middleware.WrapInvokableToolCall(context.Background(), func(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
		return "ok", nil
	}, &adk.ToolContext{Name: "k8s_query", CallID: "call-1"})
	if err != nil {
		t.Fatalf("WrapInvokableToolCall error = %v", err)
	}
	if _, err := endpoint(context.Background(), `{"cluster_id":1}`); err != nil {
		t.Fatalf("endpoint error = %v", err)
	}
	if metrics.ToolCalls.Load() != 1 {
		t.Fatalf("ToolCalls = %d, want 1", metrics.ToolCalls.Load())
	}
}
