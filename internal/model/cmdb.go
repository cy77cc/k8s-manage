package model

import (
	"time"

	"gorm.io/gorm"
)

type CMDBCI struct {
	ID           uint           `gorm:"primaryKey;column:id" json:"id"`
	CIUID        string         `gorm:"column:ci_uid;type:varchar(160);not null;uniqueIndex" json:"ci_uid"`
	CIType       string         `gorm:"column:ci_type;type:varchar(64);not null;index" json:"ci_type"`
	Name         string         `gorm:"column:name;type:varchar(128);not null;index" json:"name"`
	Source       string         `gorm:"column:source;type:varchar(64);not null;default:'manual';index" json:"source"`
	ExternalID   string         `gorm:"column:external_id;type:varchar(160);not null;default:'';index" json:"external_id"`
	ProjectID    uint           `gorm:"column:project_id;default:0;index" json:"project_id"`
	TeamID       uint           `gorm:"column:team_id;default:0;index" json:"team_id"`
	Owner        string         `gorm:"column:owner;type:varchar(128);not null;default:''" json:"owner"`
	Status       string         `gorm:"column:status;type:varchar(32);not null;default:'active';index" json:"status"`
	TagsJSON     string         `gorm:"column:tags_json;type:longtext" json:"tags_json"`
	AttrsJSON    string         `gorm:"column:attrs_json;type:longtext" json:"attrs_json"`
	LastSyncedAt *time.Time     `gorm:"column:last_synced_at" json:"last_synced_at,omitempty"`
	CreatedBy    uint           `gorm:"column:created_by;default:0" json:"created_by"`
	UpdatedBy    uint           `gorm:"column:updated_by;default:0" json:"updated_by"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (CMDBCI) TableName() string { return "cmdb_cis" }

type CMDBRelation struct {
	ID           uint      `gorm:"primaryKey;column:id" json:"id"`
	FromCIID     uint      `gorm:"column:from_ci_id;not null;index:idx_cmdb_relation_from_to,priority:1" json:"from_ci_id"`
	ToCIID       uint      `gorm:"column:to_ci_id;not null;index:idx_cmdb_relation_from_to,priority:2" json:"to_ci_id"`
	RelationType string    `gorm:"column:relation_type;type:varchar(64);not null;index:idx_cmdb_relation_from_to,priority:3" json:"relation_type"`
	CreatedBy    uint      `gorm:"column:created_by;default:0" json:"created_by"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (CMDBRelation) TableName() string { return "cmdb_relations" }

type CMDBSyncJob struct {
	ID           string    `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	Source       string    `gorm:"column:source;type:varchar(64);not null;default:'all';index" json:"source"`
	Status       string    `gorm:"column:status;type:varchar(32);not null;default:'running';index" json:"status"`
	SummaryJSON  string    `gorm:"column:summary_json;type:longtext" json:"summary_json"`
	ErrorMessage string    `gorm:"column:error_message;type:text" json:"error_message"`
	StartedAt    time.Time `gorm:"column:started_at" json:"started_at"`
	FinishedAt   time.Time `gorm:"column:finished_at" json:"finished_at"`
	OperatorID   uint      `gorm:"column:operator_id;default:0;index" json:"operator_id"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (CMDBSyncJob) TableName() string { return "cmdb_sync_jobs" }

type CMDBSyncRecord struct {
	ID           uint      `gorm:"primaryKey;column:id" json:"id"`
	JobID        string    `gorm:"column:job_id;type:varchar(64);not null;index" json:"job_id"`
	CIUID        string    `gorm:"column:ci_uid;type:varchar(160);not null;index" json:"ci_uid"`
	Action       string    `gorm:"column:action;type:varchar(32);not null;index" json:"action"`
	Status       string    `gorm:"column:status;type:varchar(32);not null;index" json:"status"`
	DiffJSON     string    `gorm:"column:diff_json;type:longtext" json:"diff_json"`
	ErrorMessage string    `gorm:"column:error_message;type:text" json:"error_message"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (CMDBSyncRecord) TableName() string { return "cmdb_sync_records" }

type CMDBAudit struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	CIID       uint      `gorm:"column:ci_id;default:0;index" json:"ci_id"`
	RelationID uint      `gorm:"column:relation_id;default:0;index" json:"relation_id"`
	Action     string    `gorm:"column:action;type:varchar(64);not null;index" json:"action"`
	ActorID    uint      `gorm:"column:actor_id;default:0;index" json:"actor_id"`
	BeforeJSON string    `gorm:"column:before_json;type:longtext" json:"before_json"`
	AfterJSON  string    `gorm:"column:after_json;type:longtext" json:"after_json"`
	Detail     string    `gorm:"column:detail;type:text" json:"detail"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
}

func (CMDBAudit) TableName() string { return "cmdb_audits" }
