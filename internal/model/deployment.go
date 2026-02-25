package model

import "time"

type DeploymentTarget struct {
	ID          uint      `gorm:"primaryKey;column:id" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(128);not null" json:"name"`
	TargetType  string    `gorm:"column:target_type;type:varchar(16);not null;index" json:"target_type"` // k8s|compose
	RuntimeType string    `gorm:"column:runtime_type;type:varchar(16);not null;default:'k8s';index" json:"runtime_type"`
	ClusterID   uint      `gorm:"column:cluster_id;default:0;index" json:"cluster_id"`
	ProjectID   uint      `gorm:"column:project_id;default:0;index" json:"project_id"`
	TeamID      uint      `gorm:"column:team_id;default:0;index" json:"team_id"`
	Env         string    `gorm:"column:env;type:varchar(32);default:'staging';index" json:"env"`
	Status      string    `gorm:"column:status;type:varchar(32);default:'active'" json:"status"`
	CreatedBy   uint      `gorm:"column:created_by;default:0" json:"created_by"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (DeploymentTarget) TableName() string { return "deployment_targets" }

type DeploymentTargetNode struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	TargetID  uint      `gorm:"column:target_id;not null;index:idx_target_host,priority:1" json:"target_id"`
	HostID    uint      `gorm:"column:host_id;not null;index:idx_target_host,priority:2" json:"host_id"`
	Role      string    `gorm:"column:role;type:varchar(16);default:'worker'" json:"role"` // manager|worker
	Weight    int       `gorm:"column:weight;default:100" json:"weight"`
	Status    string    `gorm:"column:status;type:varchar(32);default:'active'" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (DeploymentTargetNode) TableName() string { return "deployment_target_nodes" }

type DeploymentRelease struct {
	ID                 uint      `gorm:"primaryKey;column:id" json:"id"`
	ServiceID          uint      `gorm:"column:service_id;not null;index" json:"service_id"`
	TargetID           uint      `gorm:"column:target_id;not null;index" json:"target_id"`
	NamespaceOrProject string    `gorm:"column:namespace_or_project;type:varchar(128);default:''" json:"namespace_or_project"`
	RuntimeType        string    `gorm:"column:runtime_type;type:varchar(16);not null;index" json:"runtime_type"` // k8s|compose
	Strategy           string    `gorm:"column:strategy;type:varchar(16);default:'rolling'" json:"strategy"`
	RevisionID         uint      `gorm:"column:revision_id;default:0;index" json:"revision_id"`
	SourceReleaseID    uint      `gorm:"column:source_release_id;default:0;index" json:"source_release_id"`
	TargetRevision     string    `gorm:"column:target_revision;type:varchar(128);default:''" json:"target_revision"`
	Status             string    `gorm:"column:status;type:varchar(32);default:'pending_approval';index" json:"status"`
	ManifestSnapshot   string    `gorm:"column:manifest_snapshot;type:longtext" json:"manifest_snapshot"`
	RuntimeContextJSON string    `gorm:"column:runtime_context_json;type:longtext" json:"runtime_context_json"`
	ChecksJSON         string    `gorm:"column:checks_json;type:longtext" json:"checks_json"`
	WarningsJSON       string    `gorm:"column:warnings_json;type:longtext" json:"warnings_json"`
	DiagnosticsJSON    string    `gorm:"column:diagnostics_json;type:longtext" json:"diagnostics_json"`
	VerificationJSON   string    `gorm:"column:verification_json;type:longtext" json:"verification_json"`
	Operator           uint      `gorm:"column:operator;default:0;index" json:"operator"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (DeploymentRelease) TableName() string { return "deployment_releases" }

type ServiceGovernancePolicy struct {
	ID                   uint      `gorm:"primaryKey;column:id" json:"id"`
	ServiceID            uint      `gorm:"column:service_id;not null;index:idx_service_env_governance,priority:1" json:"service_id"`
	Env                  string    `gorm:"column:env;type:varchar(32);not null;index:idx_service_env_governance,priority:2" json:"env"`
	TrafficPolicyJSON    string    `gorm:"column:traffic_policy_json;type:longtext" json:"traffic_policy_json"`
	ResiliencePolicyJSON string    `gorm:"column:resilience_policy_json;type:longtext" json:"resilience_policy_json"`
	AccessPolicyJSON     string    `gorm:"column:access_policy_json;type:longtext" json:"access_policy_json"`
	SLOPolicyJSON        string    `gorm:"column:slo_policy_json;type:longtext" json:"slo_policy_json"`
	UpdatedBy            uint      `gorm:"column:updated_by;default:0" json:"updated_by"`
	UpdatedAt            time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ServiceGovernancePolicy) TableName() string { return "service_governance_policies" }

type AIOPSInspection struct {
	ID              uint      `gorm:"primaryKey;column:id" json:"id"`
	ReleaseID       uint      `gorm:"column:release_id;default:0;index" json:"release_id"`
	TargetID        uint      `gorm:"column:target_id;default:0;index" json:"target_id"`
	ServiceID       uint      `gorm:"column:service_id;default:0;index" json:"service_id"`
	Stage           string    `gorm:"column:stage;type:varchar(16);not null"` // pre|post|periodic
	Summary         string    `gorm:"column:summary;type:text" json:"summary"`
	FindingsJSON    string    `gorm:"column:findings_json;type:longtext" json:"findings_json"`
	SuggestionsJSON string    `gorm:"column:suggestions_json;type:longtext" json:"suggestions_json"`
	Status          string    `gorm:"column:status;type:varchar(32);default:'done'" json:"status"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
}

func (AIOPSInspection) TableName() string { return "aiops_inspections" }
