package runtime

import "testing"

func TestSSEConverterApprovalRequiredIncludesCheckpointIdentity(t *testing.T) {
	converter := NewSSEConverter()
	events := converter.OnApprovalRequired(&PendingApproval{
		ID:       "approval-1",
		PlanID:   "plan-1",
		StepID:   "step-1",
		ToolName: "scale_deployment",
		Summary:  "scale nginx",
	}, "cp-1")

	if len(events) != 2 {
		t.Fatalf("event count = %d", len(events))
	}
	if got := events[1].Data["checkpoint_id"]; got != "cp-1" {
		t.Fatalf("checkpoint_id = %#v", got)
	}
}

func TestSSEConverterPlannerStartIncludesSemanticFields(t *testing.T) {
	converter := NewSSEConverter()
	events := converter.OnPlannerStart("sess-1", "plan-1", "turn-1")
	if len(events) != 2 {
		t.Fatalf("event count = %d", len(events))
	}

	stage := events[1]
	if got := stage.Data["title"]; got == nil || got == "" {
		t.Fatalf("title = %#v, want non-empty", got)
	}
	if got := stage.Data["description"]; got == nil || got == "" {
		t.Fatalf("description = %#v, want non-empty", got)
	}
}

func TestSSEConverterPlanCreatedIncludesSteps(t *testing.T) {
	converter := NewSSEConverter()
	evt := converter.OnPlanCreated("plan-1", "plan content", []string{"step a", "step b"})
	steps, ok := evt.Data["steps"].([]string)
	if !ok {
		t.Fatalf("steps type = %T, want []string", evt.Data["steps"])
	}
	if len(steps) != 2 {
		t.Fatalf("steps len = %d, want 2", len(steps))
	}
	if evt.Data["title"] == "" {
		t.Fatalf("title = %#v, want non-empty", evt.Data["title"])
	}
	if evt.Data["description"] == "" {
		t.Fatalf("description = %#v, want non-empty", evt.Data["description"])
	}
}

func TestSSEConverterExecuteStartIncludesToolContext(t *testing.T) {
	converter := NewSSEConverter()
	evt := converter.OnExecuteStart("step-1", "扩容部署", "scale_deployment", map[string]any{"replicas": 3})
	if evt.Type != EventStageDelta {
		t.Fatalf("type = %s, want %s", evt.Type, EventStageDelta)
	}
	if got := evt.Data["stage"]; got != "execute" {
		t.Fatalf("stage = %#v, want execute", got)
	}
	if got := evt.Data["tool_name"]; got != "scale_deployment" {
		t.Fatalf("tool_name = %#v, want scale_deployment", got)
	}
	if got := evt.Data["title"]; got != "扩容部署" {
		t.Fatalf("title = %#v, want 扩容部署", got)
	}
	params, ok := evt.Data["params"].(map[string]any)
	if !ok || params["replicas"] != 3 {
		t.Fatalf("params = %#v, want replicas=3", evt.Data["params"])
	}
}
