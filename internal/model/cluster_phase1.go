package model

import "time"

type ClusterNamespaceBinding struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	ClusterID uint      `gorm:"column:cluster_id;not null;index:idx_cluster_team_ns,priority:1" json:"cluster_id"`
	TeamID    uint      `gorm:"column:team_id;not null;index:idx_cluster_team_ns,priority:2" json:"team_id"`
	Namespace string    `gorm:"column:namespace;type:varchar(128);not null;index:idx_cluster_team_ns,priority:3" json:"namespace"`
	Env       string    `gorm:"column:env;type:varchar(32);default:''" json:"env"`
	Readonly  bool      `gorm:"column:readonly;not null;default:false" json:"readonly"`
	CreatedBy uint      `gorm:"column:created_by;not null;default:0" json:"created_by"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ClusterNamespaceBinding) TableName() string {
	return "cluster_namespace_bindings"
}

type ClusterReleaseRecord struct {
	ID          uint      `gorm:"primaryKey;column:id" json:"id"`
	ClusterID   uint      `gorm:"column:cluster_id;not null;index" json:"cluster_id"`
	Namespace   string    `gorm:"column:namespace;type:varchar(128);not null;index" json:"namespace"`
	App         string    `gorm:"column:app;type:varchar(128);not null" json:"app"`
	Strategy    string    `gorm:"column:strategy;type:varchar(32);not null;default:'rolling'" json:"strategy"`
	RolloutName string    `gorm:"column:rollout_name;type:varchar(128);not null;default:''" json:"rollout_name"`
	Revision    int       `gorm:"column:revision;not null;default:1" json:"revision"`
	Status      string    `gorm:"column:status;type:varchar(32);not null;default:'pending'" json:"status"`
	Operator    string    `gorm:"column:operator;type:varchar(64);not null;default:''" json:"operator"`
	PayloadJSON string    `gorm:"column:payload_json;type:longtext" json:"payload_json"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ClusterReleaseRecord) TableName() string {
	return "cluster_release_records"
}

type ClusterHPAPolicy struct {
	ID                uint      `gorm:"primaryKey;column:id" json:"id"`
	ClusterID         uint      `gorm:"column:cluster_id;not null;index:idx_cluster_ns_hpa,priority:1" json:"cluster_id"`
	Namespace         string    `gorm:"column:namespace;type:varchar(128);not null;index:idx_cluster_ns_hpa,priority:2" json:"namespace"`
	Name              string    `gorm:"column:name;type:varchar(128);not null;index:idx_cluster_ns_hpa,priority:3" json:"name"`
	TargetRefKind     string    `gorm:"column:target_ref_kind;type:varchar(64);not null" json:"target_ref_kind"`
	TargetRefName     string    `gorm:"column:target_ref_name;type:varchar(128);not null" json:"target_ref_name"`
	MinReplicas       int32     `gorm:"column:min_replicas;not null;default:1" json:"min_replicas"`
	MaxReplicas       int32     `gorm:"column:max_replicas;not null;default:1" json:"max_replicas"`
	CPUUtilization    *int32    `gorm:"column:cpu_utilization" json:"cpu_utilization,omitempty"`
	MemoryUtilization *int32    `gorm:"column:memory_utilization" json:"memory_utilization,omitempty"`
	RawPolicyJSON     string    `gorm:"column:raw_policy_json;type:longtext" json:"raw_policy_json"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ClusterHPAPolicy) TableName() string {
	return "cluster_hpa_policies"
}

type ClusterQuotaPolicy struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	ClusterID uint      `gorm:"column:cluster_id;not null;index:idx_cluster_ns_quota,priority:1" json:"cluster_id"`
	Namespace string    `gorm:"column:namespace;type:varchar(128);not null;index:idx_cluster_ns_quota,priority:2" json:"namespace"`
	Name      string    `gorm:"column:name;type:varchar(128);not null;index:idx_cluster_ns_quota,priority:3" json:"name"`
	Type      string    `gorm:"column:type;type:varchar(32);not null;default:'resourcequota'" json:"type"`
	SpecJSON  string    `gorm:"column:spec_json;type:longtext" json:"spec_json"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ClusterQuotaPolicy) TableName() string {
	return "cluster_quota_policies"
}

type ClusterDeployApproval struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	Ticket    string    `gorm:"column:ticket;type:varchar(96);not null;uniqueIndex" json:"ticket"`
	ClusterID uint      `gorm:"column:cluster_id;not null;index" json:"cluster_id"`
	Namespace string    `gorm:"column:namespace;type:varchar(128);not null" json:"namespace"`
	Action    string    `gorm:"column:action;type:varchar(32);not null" json:"action"`
	Status    string    `gorm:"column:status;type:varchar(32);not null;default:'pending'" json:"status"`
	RequestBy uint      `gorm:"column:request_by;not null;default:0" json:"request_by"`
	ReviewBy  uint      `gorm:"column:review_by;not null;default:0" json:"review_by"`
	ExpiresAt time.Time `gorm:"column:expires_at" json:"expires_at"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ClusterDeployApproval) TableName() string {
	return "cluster_deploy_approvals"
}

type ClusterOperationAudit struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	ClusterID  uint      `gorm:"column:cluster_id;not null;index" json:"cluster_id"`
	Namespace  string    `gorm:"column:namespace;type:varchar(128);not null;default:''" json:"namespace"`
	Action     string    `gorm:"column:action;type:varchar(64);not null;index" json:"action"`
	Resource   string    `gorm:"column:resource;type:varchar(64);not null;default:''" json:"resource"`
	ResourceID string    `gorm:"column:resource_id;type:varchar(128);not null;default:''" json:"resource_id"`
	Status     string    `gorm:"column:status;type:varchar(32);not null;default:'success'" json:"status"`
	Message    string    `gorm:"column:message;type:varchar(255);not null;default:''" json:"message"`
	OperatorID uint      `gorm:"column:operator_id;not null;default:0" json:"operator_id"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (ClusterOperationAudit) TableName() string {
	return "cluster_operation_audits"
}
