package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/cy77cc/OpsPilot/internal/model"
	"github.com/cy77cc/OpsPilot/internal/websocket"
	"gorm.io/gorm"
)

// NotificationIntegrator 通知集成服务
type NotificationIntegrator struct {
	db  *gorm.DB
	hub *websocket.Hub
}

// NewNotificationIntegrator 创建通知集成服务
func NewNotificationIntegrator(db *gorm.DB) *NotificationIntegrator {
	return &NotificationIntegrator{
		db:  db,
		hub: websocket.GetHub(),
	}
}

// CreateAlertNotification 创建告警通知
func (n *NotificationIntegrator) CreateAlertNotification(ctx context.Context, alert *model.AlertEvent) error {
	// 创建通知主体
	notif := model.Notification{
		Type:       "alert",
		Title:      alert.Title,
		Content:    alert.Message,
		Severity:   alert.Severity,
		Source:     alert.Source,
		SourceID:   fmt.Sprintf("%d", alert.ID),
		ActionURL:  fmt.Sprintf("/monitor?alert_id=%d", alert.ID),
		ActionType: "confirm",
	}

	// 获取所有应该通知的用户
	userIDs, err := n.getAlertNotificationUsers(ctx, alert)
	if err != nil {
		return err
	}

	if len(userIDs) == 0 {
		return nil
	}

	// 使用事务创建通知
	return n.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建通知
		if err := tx.Create(&notif).Error; err != nil {
			return err
		}

		// 创建用户通知关联并推送
		for _, userID := range userIDs {
			userNotif := model.UserNotification{
				UserID:         userID,
				NotificationID: notif.ID,
			}
			if err := tx.Create(&userNotif).Error; err != nil {
				return err
			}

			// 加载关联的通知用于推送
			userNotif.Notification = notif

			// 通过 WebSocket 推送
			go n.hub.PushNotification(userID, &userNotif)
		}

		return nil
	})
}

// CreateTaskNotification 创建任务通知
func (n *NotificationIntegrator) CreateTaskNotification(ctx context.Context, taskID, userID uint64, title, content, status string) error {
	notif := model.Notification{
		Type:      "task",
		Title:     title,
		Content:   content,
		Severity:  n.getTaskSeverity(status),
		Source:    "任务系统",
		SourceID:  fmt.Sprintf("%d", taskID),
		ActionURL: fmt.Sprintf("/tasks/%d", taskID),
	}

	return n.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&notif).Error; err != nil {
			return err
		}

		userNotif := model.UserNotification{
			UserID:         userID,
			NotificationID: notif.ID,
		}
		if err := tx.Create(&userNotif).Error; err != nil {
			return err
		}

		userNotif.Notification = notif
		go n.hub.PushNotification(userID, &userNotif)

		return nil
	})
}

// CreateSystemNotification 创建系统通知
func (n *NotificationIntegrator) CreateSystemNotification(ctx context.Context, title, content string, userIDs []uint64) error {
	notif := model.Notification{
		Type:     "system",
		Title:    title,
		Content:  content,
		Severity: "info",
		Source:   "系统",
	}

	return n.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&notif).Error; err != nil {
			return err
		}

		for _, userID := range userIDs {
			userNotif := model.UserNotification{
				UserID:         userID,
				NotificationID: notif.ID,
			}
			if err := tx.Create(&userNotif).Error; err != nil {
				return err
			}

			userNotif.Notification = notif
			go n.hub.PushNotification(userID, &userNotif)
		}

		return nil
	})
}

// getAlertNotificationUsers 获取应该接收告警通知的用户
func (n *NotificationIntegrator) getAlertNotificationUsers(ctx context.Context, alert *model.AlertEvent) ([]uint64, error) {
	// 目前简单实现：获取所有管理员用户
	// 后续可以根据告警规则的 channels 配置确定通知用户
	var users []struct {
		ID uint64 `gorm:"column:id"`
	}

	// 查询有监控查看权限的用户
	err := n.db.WithContext(ctx).
		Table("users").
		Where("role = ? OR role = ?", "admin", "super_admin").
		Pluck("id", &users).Error

	if err != nil {
		return nil, err
	}

	userIDs := make([]uint64, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	return userIDs, nil
}

// getTaskSeverity 根据任务状态获取通知严重级别
func (n *NotificationIntegrator) getTaskSeverity(status string) string {
	switch status {
	case "failed", "error":
		return "critical"
	case "completed", "success":
		return "info"
	default:
		return "warning"
	}
}

// PushNotificationUpdate 推送通知状态更新
func (n *NotificationIntegrator) PushNotificationUpdate(userID uint64, notifID uint, readAt, dismissedAt, confirmedAt *time.Time) {
	n.hub.PushUpdate(userID, notifID, readAt, dismissedAt, confirmedAt)
}
