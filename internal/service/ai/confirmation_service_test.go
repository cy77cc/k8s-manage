package ai

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
)

func newConfirmationTestService(t *testing.T) *ConfirmationService {
	t.Helper()
	h := newCommandTestHandler(t)
	if err := h.svcCtx.DB.AutoMigrate(&model.ConfirmationRequest{}); err != nil {
		t.Fatalf("auto migrate confirmation: %v", err)
	}
	svc := NewConfirmationService(h.svcCtx.DB)
	svc.pollInterval = 20 * time.Millisecond
	return svc
}

func TestConfirmationServiceRequestAndConfirm(t *testing.T) {
	svc := newConfirmationTestService(t)
	item, err := svc.RequestConfirmation(context.Background(), ConfirmationRequestInput{
		RequestUserID: 1,
		ToolName:      "host_batch_exec_apply",
		ToolMode:      "mutating",
		RiskLevel:     "high",
		Timeout:       2 * time.Second,
	})
	if err != nil {
		t.Fatalf("request confirmation: %v", err)
	}
	if item.Status != confirmationStatusPending {
		t.Fatalf("expected pending status, got %s", item.Status)
	}

	confirmed, err := svc.Confirm(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if confirmed.Status != confirmationStatusConfirmed {
		t.Fatalf("expected confirmed status, got %s", confirmed.Status)
	}
	if confirmed.ConfirmedAt == nil {
		t.Fatalf("expected confirmed_at set")
	}
}

func TestConfirmationServiceWaitForConfirmation(t *testing.T) {
	svc := newConfirmationTestService(t)
	item, err := svc.RequestConfirmation(context.Background(), ConfirmationRequestInput{
		RequestUserID: 1,
		ToolName:      "service_deploy_apply",
		Timeout:       2 * time.Second,
	})
	if err != nil {
		t.Fatalf("request confirmation: %v", err)
	}
	go func() {
		time.Sleep(60 * time.Millisecond)
		_, _ = svc.Confirm(context.Background(), item.ID)
	}()
	out, err := svc.WaitForConfirmation(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("wait for confirmation: %v", err)
	}
	if out.Status != confirmationStatusConfirmed {
		t.Fatalf("expected confirmed status, got %s", out.Status)
	}
}

func TestConfirmationServiceTimeoutExpire(t *testing.T) {
	svc := newConfirmationTestService(t)
	item, err := svc.RequestConfirmation(context.Background(), ConfirmationRequestInput{
		RequestUserID: 1,
		ToolName:      "service_deploy_apply",
		Timeout:       50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("request confirmation: %v", err)
	}
	time.Sleep(80 * time.Millisecond)
	out, err := svc.WaitForConfirmation(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("wait for confirmation: %v", err)
	}
	if out.Status != confirmationStatusExpired {
		t.Fatalf("expected expired status, got %s", out.Status)
	}
}

func TestConfirmationServiceCancel(t *testing.T) {
	svc := newConfirmationTestService(t)
	item, err := svc.RequestConfirmation(context.Background(), ConfirmationRequestInput{
		RequestUserID: 1,
		ToolName:      "host_batch_exec_apply",
		Timeout:       2 * time.Second,
	})
	if err != nil {
		t.Fatalf("request confirmation: %v", err)
	}
	out, err := svc.Cancel(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("cancel confirmation: %v", err)
	}
	if out.Status != confirmationStatusCancelled {
		t.Fatalf("expected cancelled status, got %s", out.Status)
	}
}

func TestConfirmationServiceWaitCancelledByContext(t *testing.T) {
	svc := newConfirmationTestService(t)
	item, err := svc.RequestConfirmation(context.Background(), ConfirmationRequestInput{
		RequestUserID: 1,
		ToolName:      "host_batch_exec_apply",
		Timeout:       2 * time.Second,
	})
	if err != nil {
		t.Fatalf("request confirmation: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	_, err = svc.WaitForConfirmation(ctx, item.ID)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}
