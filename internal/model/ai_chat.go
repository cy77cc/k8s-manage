package model

import "time"

// AIChatSession stores conversation metadata by user and scene.
type AIChatSession struct {
	ID        string    `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	UserID    uint64    `gorm:"column:user_id;index:idx_ai_session_user_scene" json:"user_id"`
	Scene     string    `gorm:"column:scene;type:varchar(128);index:idx_ai_session_user_scene" json:"scene"`
	Title     string    `gorm:"column:title;type:varchar(128)" json:"title"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AIChatSession) TableName() string { return "ai_chat_sessions" }

// AIChatMessage stores each message item under a session.
type AIChatMessage struct {
	ID        string    `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	SessionID string    `gorm:"column:session_id;type:varchar(64);index:idx_ai_msg_session_created" json:"session_id"`
	Role      string    `gorm:"column:role;type:varchar(32);index" json:"role"`
	Content   string    `gorm:"column:content;type:longtext" json:"content"`
	Thinking  string    `gorm:"column:thinking;type:longtext" json:"thinking"`
	CreatedAt time.Time `gorm:"column:created_at;index:idx_ai_msg_session_created;autoCreateTime" json:"created_at"`
}

func (AIChatMessage) TableName() string { return "ai_chat_messages" }
