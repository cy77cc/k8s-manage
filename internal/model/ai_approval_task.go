package model

import (
	"encoding/json"
	"time"
)

type ExecutionStep struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type RiskAssessment struct {
	Level   string   `json:"level"`
	Summary string   `json:"summary,omitempty"`
	Items   []string `json:"items,omitempty"`
}

type TaskDetail struct {
	Summary        string          `json:"summary"`
	Steps          []ExecutionStep `json:"steps,omitempty"`
	RiskAssessment RiskAssessment  `json:"risk_assessment"`
	RollbackPlan   string          `json:"rollback_plan,omitempty"`
}

type ApprovalToolCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// AIApprovalTask stores executable approval work items for AI-governed operations.
type AIApprovalTask struct {
	ID                 string     `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	ConfirmationID     string     `gorm:"column:confirmation_id;type:varchar(64);index" json:"confirmation_id"`
	RequestUserID      uint64     `gorm:"column:request_user_id;index:idx_ai_approval_request_created" json:"request_user_id"`
	ApprovalToken      string     `gorm:"column:approval_token;type:varchar(128);uniqueIndex" json:"approval_token"`
	ToolName           string     `gorm:"column:tool_name;type:varchar(128);index" json:"tool_name"`
	TargetResourceType string     `gorm:"column:target_resource_type;type:varchar(64);index" json:"target_resource_type"`
	TargetResourceID   string     `gorm:"column:target_resource_id;type:varchar(128);index" json:"target_resource_id"`
	TargetResourceName string     `gorm:"column:target_resource_name;type:varchar(128);index" json:"target_resource_name"`
	RiskLevel          string     `gorm:"column:risk_level;type:varchar(16);index" json:"risk_level"`
	Status             string     `gorm:"column:status;type:varchar(32);index" json:"status"`
	ApproverUserID     uint64     `gorm:"column:approver_user_id;index" json:"approver_user_id"`
	RejectReason       string     `gorm:"column:reject_reason;type:varchar(255)" json:"reject_reason"`
	ParamsJSON         string     `gorm:"column:params_json;type:longtext" json:"params_json"`
	PreviewJSON        string     `gorm:"column:preview_json;type:longtext" json:"preview_json"`
	TaskDetailJSON     string     `gorm:"column:task_detail_json;type:longtext" json:"task_detail_json"`
	ToolCallsJSON      string     `gorm:"column:tool_calls_json;type:longtext" json:"tool_calls_json"`
	ExecutedAt         *time.Time `gorm:"column:executed_at" json:"executed_at,omitempty"`
	ExpiresAt          time.Time  `gorm:"column:expires_at;index" json:"expires_at"`
	ApprovedAt         *time.Time `gorm:"column:approved_at" json:"approved_at,omitempty"`
	RejectedAt         *time.Time `gorm:"column:rejected_at" json:"rejected_at,omitempty"`
	CreatedAt          time.Time  `gorm:"column:created_at;autoCreateTime;index:idx_ai_approval_request_created" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AIApprovalTask) TableName() string { return "ai_approval_tickets" }

func (m *AIApprovalTask) SetTaskDetail(detail TaskDetail) error {
	raw, err := json.Marshal(detail)
	if err != nil {
		return err
	}
	m.TaskDetailJSON = string(raw)
	return nil
}

func (m *AIApprovalTask) TaskDetail() (TaskDetail, error) {
	if m == nil || m.TaskDetailJSON == "" {
		return TaskDetail{}, nil
	}
	var detail TaskDetail
	err := json.Unmarshal([]byte(m.TaskDetailJSON), &detail)
	return detail, err
}

func (m *AIApprovalTask) SetToolCalls(calls []ApprovalToolCall) error {
	raw, err := json.Marshal(calls)
	if err != nil {
		return err
	}
	m.ToolCallsJSON = string(raw)
	return nil
}

func (m *AIApprovalTask) ToolCalls() ([]ApprovalToolCall, error) {
	if m == nil || m.ToolCallsJSON == "" {
		return nil, nil
	}
	var calls []ApprovalToolCall
	err := json.Unmarshal([]byte(m.ToolCallsJSON), &calls)
	return calls, err
}

type AIApprovalTicket = AIApprovalTask
