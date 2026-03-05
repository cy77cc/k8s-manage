package model

import "time"

// ConfirmationRequest stores user confirmation state for mutating AI actions.
type ConfirmationRequest struct {
	ID            string     `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	RequestUserID uint64     `gorm:"column:request_user_id;index:idx_ai_confirmation_user_created" json:"request_user_id"`
	TraceID       string     `gorm:"column:trace_id;type:varchar(96);index" json:"trace_id"`
	ToolName      string     `gorm:"column:tool_name;type:varchar(128);index" json:"tool_name"`
	ToolMode      string     `gorm:"column:tool_mode;type:varchar(32);index" json:"tool_mode"`
	RiskLevel     string     `gorm:"column:risk_level;type:varchar(16);index" json:"risk_level"`
	ParamsJSON    string     `gorm:"column:params_json;type:longtext" json:"params_json"`
	PreviewJSON   string     `gorm:"column:preview_json;type:longtext" json:"preview_json"`
	Status        string     `gorm:"column:status;type:varchar(32);index" json:"status"`
	Reason        string     `gorm:"column:reason;type:varchar(255)" json:"reason"`
	ExpiresAt     time.Time  `gorm:"column:expires_at;index" json:"expires_at"`
	ConfirmedAt   *time.Time `gorm:"column:confirmed_at" json:"confirmed_at,omitempty"`
	CancelledAt   *time.Time `gorm:"column:cancelled_at" json:"cancelled_at,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime;index:idx_ai_confirmation_user_created" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ConfirmationRequest) TableName() string { return "ai_confirmations" }
