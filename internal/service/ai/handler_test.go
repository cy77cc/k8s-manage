package ai

import (
	"testing"

	coreai "github.com/cy77cc/OpsPilot/internal/ai"
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
