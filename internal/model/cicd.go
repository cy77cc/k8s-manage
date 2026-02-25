package model

import "time"

type CICDServiceCIConfig struct {
	ID             uint      `gorm:"primaryKey;column:id" json:"id"`
	ServiceID      uint      `gorm:"column:service_id;not null;index:idx_cicd_service_ci_service" json:"service_id"`
	RepoURL        string    `gorm:"column:repo_url;type:varchar(512);not null" json:"repo_url"`
	Branch         string    `gorm:"column:branch;type:varchar(128);default:'main'" json:"branch"`
	BuildStepsJSON string    `gorm:"column:build_steps_json;type:longtext" json:"build_steps_json"`
	ArtifactTarget string    `gorm:"column:artifact_target;type:varchar(512);not null" json:"artifact_target"`
	TriggerMode    string    `gorm:"column:trigger_mode;type:varchar(32);not null;default:'manual'" json:"trigger_mode"`
	Status         string    `gorm:"column:status;type:varchar(32);not null;default:'active'" json:"status"`
	UpdatedBy      uint      `gorm:"column:updated_by;not null;default:0" json:"updated_by"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (CICDServiceCIConfig) TableName() string { return "cicd_service_ci_configs" }

type CICDServiceCIRun struct {
	ID          uint      `gorm:"primaryKey;column:id" json:"id"`
	ServiceID   uint      `gorm:"column:service_id;not null;index:idx_cicd_ci_runs_service" json:"service_id"`
	CIConfigID  uint      `gorm:"column:ci_config_id;not null;index" json:"ci_config_id"`
	TriggerType string    `gorm:"column:trigger_type;type:varchar(32);not null" json:"trigger_type"`
	Status      string    `gorm:"column:status;type:varchar(32);not null;default:'queued'" json:"status"`
	Reason      string    `gorm:"column:reason;type:varchar(512);default:''" json:"reason"`
	TriggeredBy uint      `gorm:"column:triggered_by;not null;default:0;index" json:"triggered_by"`
	TriggeredAt time.Time `gorm:"column:triggered_at;not null" json:"triggered_at"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (CICDServiceCIRun) TableName() string { return "cicd_service_ci_runs" }

type CICDDeploymentCDConfig struct {
	ID                 uint      `gorm:"primaryKey;column:id" json:"id"`
	DeploymentID       uint      `gorm:"column:deployment_id;not null;index:uk_cicd_deploy_env,priority:1" json:"deployment_id"`
	Env                string    `gorm:"column:env;type:varchar(32);not null;index:uk_cicd_deploy_env_runtime,priority:2" json:"env"`
	RuntimeType        string    `gorm:"column:runtime_type;type:varchar(16);not null;default:'k8s';index:uk_cicd_deploy_env_runtime,priority:3" json:"runtime_type"`
	Strategy           string    `gorm:"column:strategy;type:varchar(32);not null;default:'rolling'" json:"strategy"`
	StrategyConfigJSON string    `gorm:"column:strategy_config_json;type:longtext" json:"strategy_config_json"`
	ApprovalRequired   bool      `gorm:"column:approval_required;not null;default:false" json:"approval_required"`
	UpdatedBy          uint      `gorm:"column:updated_by;not null;default:0" json:"updated_by"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (CICDDeploymentCDConfig) TableName() string { return "cicd_deployment_cd_configs" }

type CICDRelease struct {
	ID                    uint       `gorm:"primaryKey;column:id" json:"id"`
	ServiceID             uint       `gorm:"column:service_id;not null;index:idx_cicd_release_service" json:"service_id"`
	DeploymentID          uint       `gorm:"column:deployment_id;not null;index:idx_cicd_release_deployment" json:"deployment_id"`
	Env                   string     `gorm:"column:env;type:varchar(32);not null" json:"env"`
	RuntimeType           string     `gorm:"column:runtime_type;type:varchar(16);not null;default:'k8s';index:idx_cicd_release_runtime" json:"runtime_type"`
	Version               string     `gorm:"column:version;type:varchar(128);not null" json:"version"`
	Strategy              string     `gorm:"column:strategy;type:varchar(32);not null" json:"strategy"`
	Status                string     `gorm:"column:status;type:varchar(32);not null;index;default:'pending_approval'" json:"status"`
	TriggeredBy           uint       `gorm:"column:triggered_by;not null;default:0" json:"triggered_by"`
	ApprovedBy            uint       `gorm:"column:approved_by;not null;default:0" json:"approved_by"`
	ApprovalComment       string     `gorm:"column:approval_comment;type:varchar(1024);default:''" json:"approval_comment"`
	RollbackFromReleaseID uint       `gorm:"column:rollback_from_release_id;not null;default:0" json:"rollback_from_release_id"`
	DiagnosticsJSON       string     `gorm:"column:diagnostics_json;type:longtext" json:"diagnostics_json"`
	StartedAt             *time.Time `gorm:"column:started_at" json:"started_at"`
	FinishedAt            *time.Time `gorm:"column:finished_at" json:"finished_at"`
	CreatedAt             time.Time  `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (CICDRelease) TableName() string { return "cicd_releases" }

type CICDReleaseApproval struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	ReleaseID  uint      `gorm:"column:release_id;not null;index" json:"release_id"`
	ApproverID uint      `gorm:"column:approver_id;not null;default:0" json:"approver_id"`
	Decision   string    `gorm:"column:decision;type:varchar(32);not null" json:"decision"` // approved|rejected
	Comment    string    `gorm:"column:comment;type:varchar(1024);default:''" json:"comment"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
}

func (CICDReleaseApproval) TableName() string { return "cicd_release_approvals" }

type CICDAuditEvent struct {
	ID           uint      `gorm:"primaryKey;column:id" json:"id"`
	ServiceID    uint      `gorm:"column:service_id;not null;default:0;index" json:"service_id"`
	DeploymentID uint      `gorm:"column:deployment_id;not null;default:0;index" json:"deployment_id"`
	ReleaseID    uint      `gorm:"column:release_id;not null;default:0;index" json:"release_id"`
	EventType    string    `gorm:"column:event_type;type:varchar(64);not null;index" json:"event_type"`
	ActorID      uint      `gorm:"column:actor_id;not null;default:0;index" json:"actor_id"`
	PayloadJSON  string    `gorm:"column:payload_json;type:longtext" json:"payload_json"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
}

func (CICDAuditEvent) TableName() string { return "cicd_audit_events" }
