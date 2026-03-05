package experts

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestParsePrimaryDecision(t *testing.T) {
	o := NewOrchestrator(&fakeRegistry{}, NewResultAggregator(AggregationTemplate, nil))
	content := `{"need_helpers":true,"helper_requests":[{"expert_name":"k8s_expert","task":"检查pod状态"},{"expert_name":"unknown","task":"忽略我"}]}`
	decision := o.parsePrimaryDecision(content, []string{"k8s_expert", "topology_expert"})
	if !decision.NeedHelpers {
		t.Fatalf("expected NeedHelpers=true")
	}
	if len(decision.HelperRequests) != 1 {
		t.Fatalf("expected 1 helper request, got %d", len(decision.HelperRequests))
	}
	if decision.HelperRequests[0].ExpertName != "k8s_expert" {
		t.Fatalf("unexpected helper: %#v", decision.HelperRequests[0])
	}
}

func TestParsePrimaryDecisionFallbackDirectAnswer(t *testing.T) {
	o := NewOrchestrator(&fakeRegistry{}, NewResultAggregator(AggregationTemplate, nil))
	content := "直接给出结论"
	decision := o.parsePrimaryDecision(content, []string{"k8s_expert"})
	if decision.NeedHelpers {
		t.Fatalf("expected NeedHelpers=false")
	}
	if decision.DirectAnswer != content {
		t.Fatalf("unexpected direct answer: %q", decision.DirectAnswer)
	}
}

func TestBuildMessagesWithHistoryLimit(t *testing.T) {
	o := NewOrchestrator(&fakeRegistry{}, NewResultAggregator(AggregationTemplate, nil))
	history := make([]*schema.Message, 0, 12)
	for i := 0; i < 12; i++ {
		history = append(history, schema.UserMessage("h"))
	}
	messages := o.buildMessagesWithHistory(history, "current")
	if len(messages) != 11 {
		t.Fatalf("expected 11 messages (10 history + 1 current), got %d", len(messages))
	}
	if messages[len(messages)-1].Content != "current" {
		t.Fatalf("expected current message at end")
	}
}

func TestHelperExecutionPhaseEmitsProgress(t *testing.T) {
	reg := &fakeRegistry{experts: map[string]*Expert{"k8s_expert": {Name: "k8s_expert"}}}
	o := NewOrchestrator(reg, NewResultAggregator(AggregationTemplate, nil))
	events := make([]ExpertProgressEvent, 0)
	req := &ExecuteRequest{
		Message: "检查服务",
		History: []*schema.Message{},
		EventEmitter: func(event string, payload any) {
			if event != "expert_progress" {
				return
			}
			e, ok := payload.(ExpertProgressEvent)
			if ok {
				events = append(events, e)
			}
		},
	}
	results, _ := o.helperExecutionPhase(context.Background(), req, []HelperRequest{{ExpertName: "k8s_expert", Task: "检查pod"}})
	if len(results) != 1 {
		t.Fatalf("expected one helper result, got %d", len(results))
	}
	if len(events) != 2 {
		t.Fatalf("expected running/done events, got %d", len(events))
	}
	if events[0].Status != "running" || events[1].Status != "done" {
		t.Fatalf("unexpected statuses: %#v", events)
	}
}
