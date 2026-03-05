package model

import "time"

// AIApprovalTicket stores third-party approval workflow state.
type AIApprovalTicket struct {
	ID                 string     `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	ConfirmationID     string     `gorm:"column:confirmation_id;type:varchar(64);index" json:"confirmation_id"`
	RequestUserID      uint64     `gorm:"column:request_user_id;index:idx_ai_approval_request_created" json:"request_user_id"`
	ApprovalToken      string     `gorm:"column:approval_token;type:varchar(128);uniqueIndex" json:"approval_token"`
	ToolName           string     `gorm:"column:tool_name;type:varchar(128);index" json:"tool_name"`
	TargetResourceType string     `gorm:"column:target_resource_type;type:varchar(64);index" json:"target_resource_type"`
	TargetResourceID   string     `gorm:"column:target_resource_id;type:varchar(128);index" json:"target_resource_id"`
	RiskLevel          string     `gorm:"column:risk_level;type:varchar(16);index" json:"risk_level"`
	Status             string     `gorm:"column:status;type:varchar(32);index" json:"status"`
	ApproverUserID     uint64     `gorm:"column:approver_user_id;index" json:"approver_user_id"`
	RejectReason       string     `gorm:"column:reject_reason;type:varchar(255)" json:"reject_reason"`
	ParamsJSON         string     `gorm:"column:params_json;type:longtext" json:"params_json"`
	PreviewJSON        string     `gorm:"column:preview_json;type:longtext" json:"preview_json"`
	ExpiresAt          time.Time  `gorm:"column:expires_at;index" json:"expires_at"`
	ApprovedAt         *time.Time `gorm:"column:approved_at" json:"approved_at,omitempty"`
	RejectedAt         *time.Time `gorm:"column:rejected_at" json:"rejected_at,omitempty"`
	CreatedAt          time.Time  `gorm:"column:created_at;autoCreateTime;index:idx_ai_approval_request_created" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AIApprovalTicket) TableName() string { return "ai_approval_tickets" }
