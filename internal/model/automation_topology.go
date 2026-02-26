package model

import "time"

type AutomationInventory struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(128);not null;index" json:"name"`
	HostsJSON string    `gorm:"column:hosts_json;type:longtext" json:"hosts_json"`
	CreatedBy uint      `gorm:"column:created_by;default:0;index" json:"created_by"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AutomationInventory) TableName() string { return "automation_inventories" }

type AutomationPlaybook struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	Name       string    `gorm:"column:name;type:varchar(128);not null;index" json:"name"`
	ContentYML string    `gorm:"column:content_yml;type:longtext" json:"content_yml"`
	RiskLevel  string    `gorm:"column:risk_level;type:varchar(32);not null;default:'medium'" json:"risk_level"`
	CreatedBy  uint      `gorm:"column:created_by;default:0;index" json:"created_by"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AutomationPlaybook) TableName() string { return "automation_playbooks" }

type AutomationRun struct {
	ID         string    `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	Action     string    `gorm:"column:action;type:varchar(128);not null;index" json:"action"`
	Status     string    `gorm:"column:status;type:varchar(32);not null;index" json:"status"`
	ResultJSON string    `gorm:"column:result_json;type:longtext" json:"result_json"`
	ParamsJSON string    `gorm:"column:params_json;type:longtext" json:"params_json"`
	Error      string    `gorm:"column:error;type:text" json:"error"`
	OperatorID uint      `gorm:"column:operator_id;default:0;index" json:"operator_id"`
	StartedAt  time.Time `gorm:"column:started_at;index" json:"started_at"`
	FinishedAt time.Time `gorm:"column:finished_at" json:"finished_at"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AutomationRun) TableName() string { return "automation_runs" }

type AutomationRunLog struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	RunID     string    `gorm:"column:run_id;type:varchar(64);not null;index" json:"run_id"`
	Level     string    `gorm:"column:level;type:varchar(16);not null;default:'info'" json:"level"`
	Message   string    `gorm:"column:message;type:text;not null" json:"message"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
}

func (AutomationRunLog) TableName() string { return "automation_run_logs" }

type AutomationExecutionAudit struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	RunID      string    `gorm:"column:run_id;type:varchar(64);not null;index" json:"run_id"`
	Action     string    `gorm:"column:action;type:varchar(128);not null;index" json:"action"`
	Status     string    `gorm:"column:status;type:varchar(32);not null;index" json:"status"`
	ActorID    uint      `gorm:"column:actor_id;default:0;index" json:"actor_id"`
	DetailJSON string    `gorm:"column:detail_json;type:longtext" json:"detail_json"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
}

func (AutomationExecutionAudit) TableName() string { return "automation_execution_audits" }

type TopologyAccessAudit struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	ActorID    uint      `gorm:"column:actor_id;default:0;index" json:"actor_id"`
	Action     string    `gorm:"column:action;type:varchar(64);not null;index" json:"action"`
	Scope      string    `gorm:"column:scope;type:varchar(128);not null;index" json:"scope"`
	FilterJSON string    `gorm:"column:filter_json;type:longtext" json:"filter_json"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
}

func (TopologyAccessAudit) TableName() string { return "topology_access_audits" }
