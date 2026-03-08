package approval

import (
	"context"
	"errors"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/model"
)

type fakeRunner struct {
	result tools.ToolResult
	err    error
}

func (f *fakeRunner) RunTool(_ context.Context, _ string, _ map[string]any) (tools.ToolResult, error) {
	return f.result, f.err
}

type fakePublisher struct {
	updates []any
}

func (f *fakePublisher) Publish(update any, _ ...uint64) {
	f.updates = append(f.updates, update)
}

func TestApprovalExecutorExecuteSuccess(t *testing.T) {
	task := &model.AIApprovalTask{
		ID:            "apv-1",
		RequestUserID: 7,
		ApprovalToken: "tok-1",
		ToolName:      "service_restart",
		ParamsJSON:    `{"service_id":1}`,
		Status:        "approved",
	}
	pub := &fakePublisher{}
	executor := NewApprovalExecutor(nil, &fakeRunner{result: tools.ToolResult{OK: true, Source: "tool"}}, pub)

	outcome, err := executor.Execute(context.Background(), task)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if outcome.Record.Status != "succeeded" {
		t.Fatalf("expected succeeded, got %s", outcome.Record.Status)
	}
	if task.Status != "executed" {
		t.Fatalf("expected executed status, got %s", task.Status)
	}
	if len(pub.updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(pub.updates))
	}
}

func TestApprovalExecutorExecuteFailure(t *testing.T) {
	task := &model.AIApprovalTask{
		ID:            "apv-2",
		RequestUserID: 9,
		ApprovalToken: "tok-2",
		ToolName:      "service_delete",
		ParamsJSON:    `{"service_id":1}`,
		Status:        "approved",
	}
	executor := NewApprovalExecutor(nil, &fakeRunner{err: errors.New("boom")}, &fakePublisher{})

	outcome, err := executor.Execute(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error")
	}
	if outcome.Record.Status != "failed" {
		t.Fatalf("expected failed record, got %s", outcome.Record.Status)
	}
	if task.Status != "failed" {
		t.Fatalf("expected failed status, got %s", task.Status)
	}
}
