package model

import "time"

// Notification 通知主体
type Notification struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	Type       string    `gorm:"column:type;type:varchar(32);not null;index" json:"type"` // alert/task/system/approval
	Title      string    `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Content    string    `gorm:"column:content;type:text" json:"content"`
	Severity   string    `gorm:"column:severity;type:varchar(16);default:'info';index" json:"severity"` // critical/warning/info
	Source     string    `gorm:"column:source;type:varchar(128);index" json:"source"`
	SourceID   string    `gorm:"column:source_id;type:varchar(128);index" json:"source_id"`
	ActionURL  string    `gorm:"column:action_url;type:varchar(512)" json:"action_url"`
	ActionType string    `gorm:"column:action_type;type:varchar(32)" json:"action_type"` // confirm/approve/view
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
}

func (Notification) TableName() string { return "notifications" }

// UserNotification 用户通知关联
type UserNotification struct {
	ID             uint          `gorm:"primaryKey;column:id" json:"id"`
	UserID         uint64        `gorm:"column:user_id;not null;index:idx_user_notification" json:"user_id"`
	NotificationID uint          `gorm:"column:notification_id;not null;index:idx_user_notification" json:"notification_id"`
	ReadAt         *time.Time    `gorm:"column:read_at" json:"read_at"`
	DismissedAt    *time.Time    `gorm:"column:dismissed_at" json:"dismissed_at"`
	ConfirmedAt    *time.Time    `gorm:"column:confirmed_at" json:"confirmed_at"`
	Notification   Notification  `gorm:"foreignKey:NotificationID" json:"notification"`
	CreatedAt      time.Time     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (UserNotification) TableName() string { return "user_notifications" }
