package ai

import (
	"context"
	"testing"
	"time"

	adkcore "github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	aitools "github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/service/ai/logic"
)

func TestPlannerBuildsHostPlanFromChatRequest(t *testing.T) {
	planner := NewPlanner()

	plan, err := planner.BuildPlan(context.Background(), PlanningRequest{
		SessionID: "sess-1",
		Message:   "帮我检查香港云服务器磁盘告警，并给出处置建议",
		Context: map[string]any{
			"scene": "hosts",
		},
	})
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	if plan.Objective.Summary == "" {
		t.Fatalf("expected objective summary")
	}
	if len(plan.Steps) < 2 {
		t.Fatalf("expected multiple steps, got %d", len(plan.Steps))
	}
	if plan.Steps[0].Domain != DomainHost {
		t.Fatalf("expected first step routed to host, got %q", plan.Steps[0].Domain)
	}
	if plan.Steps[0].Kind != StepKindHostIdentification {
		t.Fatalf("expected host identification step, got %q", plan.Steps[0].Kind)
	}
}

func TestExecutorRouterDispatchesByDomain(t *testing.T) {
	router := NewExecutorRouter()
	calls := map[Domain]int{}
	for _, domain := range []Domain{DomainHost, DomainK8s, DomainService, DomainMonitor} {
		router.Register(domain, DomainExecutorFunc(func(ctx context.Context, req ExecutionRequest) (ExecutionRecord, error) {
			calls[req.Step.Domain]++
			return ExecutionRecord{PlanID: req.Plan.PlanID, StepID: req.Step.StepID, Status: ExecutionStatusCompleted}, nil
		}))
	}

	for idx, domain := range []Domain{DomainHost, DomainK8s, DomainService, DomainMonitor} {
		_, err := router.Execute(context.Background(), ExecutionRequest{
			Plan: Plan{PlanID: "plan-1"},
			Step: PlanStep{StepID: "step-" + string(rune('a'+idx)), Domain: domain},
		})
		if err != nil {
			t.Fatalf("execute %s: %v", domain, err)
		}
	}

	for _, domain := range []Domain{DomainHost, DomainK8s, DomainService, DomainMonitor} {
		if calls[domain] != 1 {
			t.Fatalf("expected one dispatch for %s, got %d", domain, calls[domain])
		}
	}
}

func TestReplannerProducesFollowUpActions(t *testing.T) {
	replanner := NewReplanner()

	decision, err := replanner.Decide(context.Background(), ReplanRequest{
		Plan: Plan{
			PlanID: "plan-1",
			Steps: []PlanStep{{StepID: "step-1", Title: "检查磁盘", Domain: DomainHost}},
		},
		Execution: ExecutionRecord{
			StepID: "step-1",
			Status: ExecutionStatusCompleted,
			Evidence: []EvidenceItem{
				{
					EvidenceID: "ev-1",
					Type:       EvidenceTypeDiskUsage,
					Title:      "磁盘使用率告警",
					Summary:    "/data 使用率 92%",
					Severity:   SeverityWarning,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if decision.Outcome != ReplanOutcomeFinish {
		t.Fatalf("expected finish outcome, got %q", decision.Outcome)
	}
	if len(decision.FinalOutcome.NextActions) == 0 {
		t.Fatalf("expected next actions in final outcome")
	}
}

func TestOrchestratorChatStreamEmitsControlPlaneEvents(t *testing.T) {
	sessions := &fakeSessionStore{}
	runtime := logic.NewRuntimeStore(nil)
	runner := &fakeRunner{
		metas: []aitools.ToolMeta{{Name: "host_list_inventory"}},
		queryIter: eventIterator(&adkcore.AgentEvent{
			Output: &adkcore.AgentOutput{
				MessageOutput: &adkcore.MessageVariant{
					Message: schema.AssistantMessage("diagnosis complete", nil),
				},
			},
		}),
		generateReply: "检查建议|先做健康检查|0.8|降低风险",
	}
	control := NewControlPlane(nil, runtime, runner)
	orch := NewOrchestrator(runner, sessions, runtime, control)

	var events []string
	err := orch.ChatStream(context.Background(), ChatStreamRequest{
		UserID:  7,
		Message: "帮我检查香港云服务器磁盘告警",
		Context: map[string]any{"scene": "hosts"},
	}, func(event string, payload map[string]any) bool {
		events = append(events, event)
		return true
	})
	if err != nil {
		t.Fatalf("chat stream failed: %v", err)
	}

	required := map[string]bool{
		EventPlanCreated:    false,
		EventStepStatus:     false,
		EventSummary:        false,
		EventNextActions:    false,
		"done":              false,
	}
	for _, event := range events {
		if _, ok := required[event]; ok {
			required[event] = true
		}
	}
	for event, seen := range required {
		if !seen {
			t.Fatalf("expected %s to be emitted, got %#v", event, events)
		}
	}
}

func TestPlatformEventProjectorIncludesTurnLifecycleFields(t *testing.T) {
	projector := NewPlatformEventProjector()
	event := projector.PlanCreated(Plan{
		PlanID: "plan-1",
		Objective: Objective{
			Summary: "检查主机磁盘告警",
		},
		Steps: []PlanStep{
			{StepID: "step-1", Title: "确认主机", Domain: DomainHost, Kind: StepKindHostIdentification, Status: StepStatusReady},
		},
		CreatedAt: time.Now(),
	})

	if event.Type != EventPlanCreated {
		t.Fatalf("expected plan_created event, got %q", event.Type)
	}
	if event.PlanID != "plan-1" {
		t.Fatalf("expected plan id, got %q", event.PlanID)
	}
	if len(event.Payload["steps"].([]PlanStepView)) != 1 {
		t.Fatalf("expected one step view in payload")
	}
}

func TestOrchestratorChatStreamEmitsApprovalInterruptDuringControlPlaneRun(t *testing.T) {
	sessions := &fakeSessionStore{}
	runtime := logic.NewRuntimeStore(nil)
	runner := &fakeRunner{
		metas: []aitools.ToolMeta{{Name: "host_batch_exec_apply"}},
		queryIter: eventIterator(&adkcore.AgentEvent{
			Action: &adkcore.AgentAction{
				Interrupted: &adkcore.InterruptInfo{
					Data: &aitools.ApprovalInfo{
						ToolName:        "host_batch_exec_apply",
						ArgumentsInJSON: `{"host_ids":[1]}`,
						Risk:            aitools.ToolRiskHigh,
						Preview:         map[string]any{"target_count": 1},
					},
					InterruptContexts: []*adkcore.InterruptCtx{{ID: "call-1", IsRootCause: true}},
				},
			},
		}),
		generateReply: "执行建议|发起审批后再继续|0.9|审批是必要环节",
	}
	control := NewControlPlane(nil, runtime, runner)
	orch := NewOrchestrator(runner, sessions, runtime, control)

	var approvalPayload map[string]any
	err := orch.ChatStream(context.Background(), ChatStreamRequest{
		UserID:  7,
		Message: "帮我执行主机批量命令",
		Context: map[string]any{"scene": "hosts"},
	}, func(event string, payload map[string]any) bool {
		if event == "approval_required" {
			approvalPayload = payload
		}
		return true
	})
	if err != nil {
		t.Fatalf("chat stream failed: %v", err)
	}
	if approvalPayload == nil {
		t.Fatalf("expected approval_required event")
	}
	if approvalPayload["tool"] != "host_batch_exec_apply" {
		t.Fatalf("unexpected approval payload: %#v", approvalPayload)
	}
}
