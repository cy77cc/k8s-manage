package model

import "time"

// EnvironmentInstallJob tracks long-running runtime bootstrap work executed via SSH.
type EnvironmentInstallJob struct {
	ID              string     `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	Name            string     `gorm:"column:name;type:varchar(128);not null" json:"name"`
	RuntimeType     string     `gorm:"column:runtime_type;type:varchar(16);not null;index" json:"runtime_type"`
	TargetEnv       string     `gorm:"column:target_env;type:varchar(32);not null;default:'staging'" json:"target_env"`
	TargetID        uint       `gorm:"column:target_id;default:0;index" json:"target_id"`
	ClusterID       uint       `gorm:"column:cluster_id;default:0;index" json:"cluster_id"`
	Status          string     `gorm:"column:status;type:varchar(32);not null;default:'queued';index" json:"status"`
	PackageVersion  string     `gorm:"column:package_version;type:varchar(64);default:''" json:"package_version"`
	PackagePath     string     `gorm:"column:package_path;type:varchar(512);default:''" json:"package_path"`
	PackageChecksum string     `gorm:"column:package_checksum;type:varchar(128);default:''" json:"package_checksum"`
	StartedAt       *time.Time `gorm:"column:started_at" json:"started_at"`
	FinishedAt      *time.Time `gorm:"column:finished_at" json:"finished_at"`
	ErrorMessage    string     `gorm:"column:error_message;type:text" json:"error_message"`
	ResultJSON      string     `gorm:"column:result_json;type:longtext" json:"result_json"`
	CreatedBy       uint64     `gorm:"column:created_by;index" json:"created_by"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (EnvironmentInstallJob) TableName() string { return "environment_install_jobs" }

// EnvironmentInstallJobStep stores per-step logs and timing for bootstrap diagnostics.
type EnvironmentInstallJobStep struct {
	ID           uint       `gorm:"column:id;primaryKey" json:"id"`
	JobID        string     `gorm:"column:job_id;type:varchar(64);not null;index" json:"job_id"`
	StepName     string     `gorm:"column:step_name;type:varchar(64);not null" json:"step_name"`
	Phase        string     `gorm:"column:phase;type:varchar(32);default:''" json:"phase"`
	Status       string     `gorm:"column:status;type:varchar(32);not null;default:'queued';index" json:"status"`
	HostID       uint       `gorm:"column:host_id;default:0" json:"host_id"`
	Output       string     `gorm:"column:output;type:text" json:"output"`
	ErrorMessage string     `gorm:"column:error_message;type:text" json:"error_message"`
	StartedAt    *time.Time `gorm:"column:started_at" json:"started_at"`
	FinishedAt   *time.Time `gorm:"column:finished_at" json:"finished_at"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (EnvironmentInstallJobStep) TableName() string { return "environment_install_job_steps" }

// ClusterCredential stores encrypted credential materials for platform/external clusters.
type ClusterCredential struct {
	ID              uint       `gorm:"column:id;primaryKey" json:"id"`
	Name            string     `gorm:"column:name;type:varchar(128);not null" json:"name"`
	RuntimeType     string     `gorm:"column:runtime_type;type:varchar(16);not null;default:'k8s';index" json:"runtime_type"`
	Source          string     `gorm:"column:source;type:varchar(32);not null;index" json:"source"`
	ClusterID       uint       `gorm:"column:cluster_id;default:0;index" json:"cluster_id"`
	Endpoint        string     `gorm:"column:endpoint;type:varchar(256);default:''" json:"endpoint"`
	AuthMethod      string     `gorm:"column:auth_method;type:varchar(32);default:'kubeconfig'" json:"auth_method"`
	KubeconfigEnc   string     `gorm:"column:kubeconfig_enc;type:longtext" json:"-"`
	CACertEnc       string     `gorm:"column:ca_cert_enc;type:longtext" json:"-"`
	CertEnc         string     `gorm:"column:cert_enc;type:longtext" json:"-"`
	KeyEnc          string     `gorm:"column:key_enc;type:longtext" json:"-"`
	TokenEnc        string     `gorm:"column:token_enc;type:longtext" json:"-"`
	MetadataJSON    string     `gorm:"column:metadata_json;type:longtext" json:"metadata_json"`
	Status          string     `gorm:"column:status;type:varchar(32);default:'active'" json:"status"`
	LastTestAt      *time.Time `gorm:"column:last_test_at" json:"last_test_at"`
	LastTestStatus  string     `gorm:"column:last_test_status;type:varchar(32);default:''" json:"last_test_status"`
	LastTestMessage string     `gorm:"column:last_test_message;type:varchar(512);default:''" json:"last_test_message"`
	CreatedBy       uint64     `gorm:"column:created_by;index" json:"created_by"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ClusterCredential) TableName() string { return "cluster_credentials" }
