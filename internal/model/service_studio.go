package model

import "time"

type ServiceRevision struct {
	ID             uint      `gorm:"primaryKey;column:id" json:"id"`
	ServiceID      uint      `gorm:"column:service_id;not null;index" json:"service_id"`
	RevisionNo     uint      `gorm:"column:revision_no;not null" json:"revision_no"`
	ConfigMode     string    `gorm:"column:config_mode;type:varchar(16);not null" json:"config_mode"`
	RenderTarget   string    `gorm:"column:render_target;type:varchar(16);not null" json:"render_target"`
	StandardConfig string    `gorm:"column:standard_config_json;type:longtext" json:"standard_config_json"`
	CustomYAML     string    `gorm:"column:custom_yaml;type:longtext" json:"custom_yaml"`
	VariableSchema string    `gorm:"column:variable_schema_json;type:longtext" json:"variable_schema_json"`
	CreatedBy      uint      `gorm:"column:created_by;default:0" json:"created_by"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (ServiceRevision) TableName() string {
	return "service_revisions"
}

type ServiceVariableSet struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	ServiceID  uint      `gorm:"column:service_id;not null;index:idx_service_env,priority:1" json:"service_id"`
	Env        string    `gorm:"column:env;type:varchar(32);not null;index:idx_service_env,priority:2" json:"env"`
	ValuesJSON string    `gorm:"column:values_json;type:longtext" json:"values_json"`
	SecretKeys string    `gorm:"column:secret_keys_json;type:longtext" json:"secret_keys_json"`
	UpdatedBy  uint      `gorm:"column:updated_by;default:0" json:"updated_by"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ServiceVariableSet) TableName() string {
	return "service_variable_sets"
}

type ServiceDeployTarget struct {
	ID           uint      `gorm:"primaryKey;column:id" json:"id"`
	ServiceID    uint      `gorm:"column:service_id;not null;index:idx_target_service_default,priority:1" json:"service_id"`
	ClusterID    uint      `gorm:"column:cluster_id;not null;default:0" json:"cluster_id"`
	Namespace    string    `gorm:"column:namespace;type:varchar(128);not null;default:'default'" json:"namespace"`
	DeployTarget string    `gorm:"column:deploy_target;type:varchar(16);not null;default:'k8s'" json:"deploy_target"`
	PolicyJSON   string    `gorm:"column:policy_json;type:longtext" json:"policy_json"`
	IsDefault    bool      `gorm:"column:is_default;not null;default:true;index:idx_target_service_default,priority:2" json:"is_default"`
	UpdatedBy    uint      `gorm:"column:updated_by;default:0" json:"updated_by"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ServiceDeployTarget) TableName() string {
	return "service_deploy_targets"
}

type ServiceReleaseRecord struct {
	ID                uint      `gorm:"primaryKey;column:id" json:"id"`
	ServiceID         uint      `gorm:"column:service_id;not null;index" json:"service_id"`
	RevisionID        uint      `gorm:"column:revision_id;not null;default:0" json:"revision_id"`
	ClusterID         uint      `gorm:"column:cluster_id;not null;default:0" json:"cluster_id"`
	Namespace         string    `gorm:"column:namespace;type:varchar(128);not null;default:'default'" json:"namespace"`
	Env               string    `gorm:"column:env;type:varchar(32);not null;default:'staging'" json:"env"`
	DeployTarget      string    `gorm:"column:deploy_target;type:varchar(16);not null;default:'k8s'" json:"deploy_target"`
	Status            string    `gorm:"column:status;type:varchar(32);not null;default:'created'" json:"status"`
	RenderedYAML      string    `gorm:"column:rendered_yaml;type:longtext" json:"rendered_yaml"`
	VariablesSnapshot string    `gorm:"column:variables_snapshot_json;type:longtext" json:"variables_snapshot_json"`
	Error             string    `gorm:"column:error;type:longtext" json:"error"`
	Operator          uint      `gorm:"column:operator;not null;default:0" json:"operator"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (ServiceReleaseRecord) TableName() string {
	return "service_release_records"
}
