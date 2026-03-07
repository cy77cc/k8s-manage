package events

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestProjectCompatibilityEventsPreservesCardNativeEvents(t *testing.T) {
	projected := ProjectCompatibilityEvents("plan_created", gin.H{
		"plan_id":   "plan-1",
		"objective": "检查主机磁盘告警",
	})
	if len(projected) != 1 {
		t.Fatalf("expected one projected event, got %d", len(projected))
	}
	if projected[0].Name != "plan_created" {
		t.Fatalf("expected original event name, got %q", projected[0].Name)
	}
	if projected[0].Payload["plan_id"] != "plan-1" {
		t.Fatalf("expected plan payload to be preserved")
	}
}

func TestProjectCompatibilityEventsNormalizesApprovalPayload(t *testing.T) {
	projected := ProjectCompatibilityEvents("approval_required", gin.H{
		"tool": "host_batch_exec_apply",
		"preview": map[string]any{
			"preview_diff": "delete old logs",
		},
	})
	if len(projected) != 1 {
		t.Fatalf("expected one projected event, got %d", len(projected))
	}
	payload := projected[0].Payload
	if payload["approval_required"] != true {
		t.Fatalf("expected approval_required marker, got %#v", payload)
	}
	if payload["previewDiff"] != "delete old logs" {
		t.Fatalf("expected previewDiff to be normalized, got %#v", payload["previewDiff"])
	}
}

func TestProjectCompatibilityEventsNormalizesConfirmationPayload(t *testing.T) {
	projected := ProjectCompatibilityEvents("confirmation_required", gin.H{
		"token":      "cfm-1",
		"expires_at": "2026-03-07T10:00:00Z",
	})
	payload := projected[0].Payload
	if payload["confirmation_token"] != "cfm-1" {
		t.Fatalf("expected confirmation token alias, got %#v", payload)
	}
	if payload["expiresAt"] != "2026-03-07T10:00:00Z" {
		t.Fatalf("expected expiresAt alias, got %#v", payload)
	}
}
