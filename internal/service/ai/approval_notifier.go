package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	aiws "github.com/cy77cc/k8s-manage/internal/websocket"
	"gorm.io/gorm"
)

type ApprovalNotifier struct {
	db           *gorm.DB
	hub          *aiws.Hub
	pollInterval time.Duration
}

func NewApprovalNotifier(db *gorm.DB) *ApprovalNotifier {
	return &ApprovalNotifier{
		db:           db,
		hub:          aiws.GetHub(),
		pollInterval: 2 * time.Second,
	}
}

func (n *ApprovalNotifier) NotifyApprovers(ctx context.Context, ticket *model.AIApprovalTicket, approvers []uint64) error {
	if n == nil || n.db == nil {
		return fmt.Errorf("approval notifier not initialized")
	}
	if ticket == nil {
		return fmt.Errorf("approval ticket is nil")
	}
	if len(approvers) == 0 {
		return nil
	}
	title := "AI审批请求"
	content := fmt.Sprintf("工具 %s 需要审批，风险级别 %s，过期时间 %s。", ticket.ToolName, ticket.RiskLevel, ticket.ExpiresAt.Format(time.RFC3339))
	notif := model.Notification{
		Type:       "approval",
		Title:      title,
		Content:    content,
		Severity:   severityByRisk(ticket.RiskLevel),
		Source:     "ai_approval",
		SourceID:   ticket.ID,
		ActionType: "approve",
		CreatedAt:  time.Now(),
	}
	if err := n.db.WithContext(ctx).Create(&notif).Error; err != nil {
		return err
	}
	for _, uid := range approvers {
		if uid == 0 {
			continue
		}
		userNotif := model.UserNotification{
			UserID:         uid,
			NotificationID: notif.ID,
			Notification:   notif,
			CreatedAt:      time.Now(),
		}
		if err := n.db.WithContext(ctx).Create(&userNotif).Error; err != nil {
			return err
		}
		if n.hub != nil {
			n.hub.PushNotification(uid, &userNotif)
		}
	}
	return nil
}

func (n *ApprovalNotifier) CheckAndNotifyExpired(ctx context.Context, now time.Time) (int64, error) {
	if n == nil || n.db == nil {
		return 0, fmt.Errorf("approval notifier not initialized")
	}
	var expired []model.AIApprovalTicket
	if err := n.db.WithContext(ctx).
		Where("status = ? AND expires_at <= ?", "pending", now).
		Find(&expired).Error; err != nil {
		return 0, err
	}
	if len(expired) == 0 {
		return 0, nil
	}
	ids := make([]string, 0, len(expired))
	for _, item := range expired {
		ids = append(ids, item.ID)
	}
	if err := n.db.WithContext(ctx).Model(&model.AIApprovalTicket{}).
		Where("id IN ?", ids).
		Updates(map[string]any{"status": "expired", "updated_at": now}).Error; err != nil {
		return 0, err
	}
	for _, item := range expired {
		_ = n.notifyRequesterExpired(ctx, &item)
	}
	return int64(len(expired)), nil
}

func (n *ApprovalNotifier) StartTimeoutMonitor(ctx context.Context) {
	if n == nil {
		return
	}
	ticker := time.NewTicker(n.pollInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, _ = n.CheckAndNotifyExpired(ctx, time.Now())
			}
		}
	}()
}

// ReplayUnreadNotifications pushes unread approval notifications for reconnecting clients.
func (n *ApprovalNotifier) ReplayUnreadNotifications(ctx context.Context, userID uint64, limit int) error {
	if n == nil || n.db == nil {
		return fmt.Errorf("approval notifier not initialized")
	}
	if n.hub == nil || userID == 0 {
		return nil
	}
	if limit <= 0 {
		limit = 20
	}
	var rows []model.UserNotification
	if err := n.db.WithContext(ctx).
		Preload("Notification").
		Where("user_id = ? AND read_at IS NULL", userID).
		Order("id DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return err
	}
	for i := range rows {
		n.hub.PushNotification(userID, &rows[i])
	}
	return nil
}

func (n *ApprovalNotifier) notifyRequesterExpired(ctx context.Context, ticket *model.AIApprovalTicket) error {
	if ticket == nil || ticket.RequestUserID == 0 {
		return nil
	}
	notif := model.Notification{
		Type:       "approval",
		Title:      "AI审批已超时",
		Content:    fmt.Sprintf("审批单 %s 已超时，工具: %s。", ticket.ID, strings.TrimSpace(ticket.ToolName)),
		Severity:   "warning",
		Source:     "ai_approval",
		SourceID:   ticket.ID,
		ActionType: "view",
		CreatedAt:  time.Now(),
	}
	if err := n.db.WithContext(ctx).Create(&notif).Error; err != nil {
		return err
	}
	userNotif := model.UserNotification{
		UserID:         ticket.RequestUserID,
		NotificationID: notif.ID,
		Notification:   notif,
		CreatedAt:      time.Now(),
	}
	if err := n.db.WithContext(ctx).Create(&userNotif).Error; err != nil {
		return err
	}
	if n.hub != nil {
		n.hub.PushNotification(ticket.RequestUserID, &userNotif)
	}
	return nil
}

func severityByRisk(risk string) string {
	switch strings.ToLower(strings.TrimSpace(risk)) {
	case "high":
		return "critical"
	case "medium":
		return "warning"
	default:
		return "info"
	}
}
