package ai

import (
	"context"
	"testing"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
)

func newApprovalNotifierTest(t *testing.T) *ApprovalNotifier {
	t.Helper()
	h := newCommandTestHandler(t)
	if err := h.svcCtx.DB.AutoMigrate(
		&model.Notification{},
		&model.UserNotification{},
		&model.AIApprovalTicket{},
	); err != nil {
		t.Fatalf("auto migrate notifier tables: %v", err)
	}
	return NewApprovalNotifier(h.svcCtx.DB)
}

func TestApprovalNotifierNotifyApprovers(t *testing.T) {
	n := newApprovalNotifierTest(t)
	ticket := &model.AIApprovalTicket{
		ID:            "ap-1",
		RequestUserID: 100,
		ApprovalToken: "ap-token-1",
		ToolName:      "service_deploy_apply",
		RiskLevel:     "high",
		Status:        "pending",
		ExpiresAt:     time.Now().Add(10 * time.Minute),
	}
	if err := n.db.Create(ticket).Error; err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	if err := n.NotifyApprovers(context.Background(), ticket, []uint64{2, 3}); err != nil {
		t.Fatalf("notify approvers: %v", err)
	}
	var count int64
	if err := n.db.Model(&model.UserNotification{}).Where("user_id IN ?", []uint64{2, 3}).Count(&count).Error; err != nil {
		t.Fatalf("count user notifications: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 user notifications, got %d", count)
	}
}

func TestApprovalNotifierCheckAndNotifyExpired(t *testing.T) {
	n := newApprovalNotifierTest(t)
	ticket := &model.AIApprovalTicket{
		ID:            "ap-expired",
		RequestUserID: 101,
		ApprovalToken: "ap-token-expired",
		ToolName:      "host_batch_exec_apply",
		RiskLevel:     "medium",
		Status:        "pending",
		ExpiresAt:     time.Now().Add(-1 * time.Minute),
	}
	if err := n.db.Create(ticket).Error; err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	changed, err := n.CheckAndNotifyExpired(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("check and notify expired: %v", err)
	}
	if changed != 1 {
		t.Fatalf("expected 1 changed ticket, got %d", changed)
	}
	var got model.AIApprovalTicket
	if err := n.db.Where("id = ?", ticket.ID).First(&got).Error; err != nil {
		t.Fatalf("query ticket: %v", err)
	}
	if got.Status != "expired" {
		t.Fatalf("expected ticket expired, got %s", got.Status)
	}
	var cnt int64
	if err := n.db.Model(&model.UserNotification{}).Where("user_id = ?", 101).Count(&cnt).Error; err != nil {
		t.Fatalf("count requester notifications: %v", err)
	}
	if cnt == 0 {
		t.Fatalf("expected timeout notification for requester")
	}
}

func TestApprovalNotifierReplayUnreadNotifications(t *testing.T) {
	n := newApprovalNotifierTest(t)
	notif := model.Notification{
		Type:       "approval",
		Title:      "test",
		Content:    "test",
		Severity:   "info",
		Source:     "ai_approval",
		SourceID:   "ap-2",
		ActionType: "approve",
	}
	if err := n.db.Create(&notif).Error; err != nil {
		t.Fatalf("create notification: %v", err)
	}
	userNotif := model.UserNotification{
		UserID:         200,
		NotificationID: notif.ID,
	}
	if err := n.db.Create(&userNotif).Error; err != nil {
		t.Fatalf("create user notification: %v", err)
	}
	if err := n.ReplayUnreadNotifications(context.Background(), 200, 10); err != nil {
		t.Fatalf("replay unread notifications: %v", err)
	}
}
