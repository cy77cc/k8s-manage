package model

import "time"

// AICommandExecution stores command bridge preview/execute history.
type AICommandExecution struct {
	ID               string    `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	UserID           uint64    `gorm:"column:user_id;index:idx_ai_cmd_user_created" json:"user_id"`
	Scene            string    `gorm:"column:scene;type:varchar(128);index" json:"scene"`
	CommandText      string    `gorm:"column:command_text;type:text" json:"command_text"`
	Intent           string    `gorm:"column:intent;type:varchar(128);index" json:"intent"`
	PlanHash         string    `gorm:"column:plan_hash;type:varchar(96);index" json:"plan_hash"`
	Risk             string    `gorm:"column:risk;type:varchar(16);index" json:"risk"`
	Status           string    `gorm:"column:status;type:varchar(32);index" json:"status"`
	TraceID          string    `gorm:"column:trace_id;type:varchar(96);index" json:"trace_id"`
	ParamsJSON       string    `gorm:"column:params_json;type:longtext" json:"params_json"`
	MissingJSON      string    `gorm:"column:missing_json;type:longtext" json:"missing_json"`
	PlanJSON         string    `gorm:"column:plan_json;type:longtext" json:"plan_json"`
	ResultJSON       string    `gorm:"column:result_json;type:longtext" json:"result_json"`
	ApprovalContext  string    `gorm:"column:approval_context;type:longtext" json:"approval_context"`
	ExecutionSummary string    `gorm:"column:execution_summary;type:text" json:"execution_summary"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime;index:idx_ai_cmd_user_created" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AICommandExecution) TableName() string { return "ai_command_executions" }
