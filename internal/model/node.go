package model

import (
	"time"
)

type NodeID uint

type Node struct {
	ID           NodeID    `gorm:"primaryKey;column:id" json:"id"`
	Name         string    `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Hostname     string    `gorm:"column:hostname;type:varchar(64)" json:"hostname"`
	Labels       string    `gorm:"column:labels;type:json" json:"labels"`
	Description  string    `gorm:"column:description;type:varchar(256)" json:"description"`
	IP           string    `gorm:"column:ip;type:varchar(45);not null" json:"ip"`
	Port         int       `gorm:"column:port;default:22" json:"port"`
	SSHUser      string    `gorm:"column:ssh_user;type:varchar(64);not null;default:root" json:"ssh_user"`
	SSHPassword  string    `gorm:"column:ssh_password;type:varchar(256)" json:"ssh_password"`
	SSHKeyID     *NodeID   `gorm:"column:ssh_key_id" json:"ssh_key_id"`
	OS           string    `gorm:"column:os;type:varchar(64)" json:"os"`
	Arch         string    `gorm:"column:arch;type:varchar(32)" json:"arch"`
	Kernel       string    `gorm:"column:kernel;type:varchar(64)" json:"kernel"`
	CpuCores     int       `gorm:"column:cpu_cores" json:"cpu_cores"`
	MemoryMB     int       `gorm:"column:memory_mb" json:"memory_mb"`
	DiskGB       int       `gorm:"column:disk_gb" json:"disk_gb"`
	Status       string    `gorm:"column:status;type:varchar(32);not null" json:"status"`
	Role         string    `gorm:"column:role;type:varchar(32)" json:"role"`
	ClusterID    uint      `gorm:"column:cluster_id" json:"cluster_id"`
	Source       string    `gorm:"column:source;type:varchar(32);default:manual_ssh" json:"source"`
	Provider     string    `gorm:"column:provider;type:varchar(32)" json:"provider"`
	ProviderID   string    `gorm:"column:provider_instance_id;type:varchar(128)" json:"provider_instance_id"`
	ParentHostID *NodeID   `gorm:"column:parent_host_id" json:"parent_host_id"`
	LastCheckAt  time.Time `gorm:"column:last_check_at" json:"last_check_at"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (n *Node) TableName() string {
	return "nodes"
}

type SSHKey struct {
	ID          NodeID    `gorm:"primaryKey;column:id" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(64)" json:"name"`
	PublicKey   string    `gorm:"column:public_key;type:text;not null" json:"public_key"`
	PrivateKey  string    `gorm:"column:private_key;type:longtext;not null" json:"private_key"`
	Passphrase  string    `gorm:"column:passphrase;type:varchar(128)" json:"passphrase"`
	Fingerprint string    `gorm:"column:fingerprint;type:varchar(128)" json:"fingerprint"`
	Algorithm   string    `gorm:"column:algorithm;type:varchar(32)" json:"algorithm"`
	Encrypted   bool      `gorm:"column:encrypted;default:false" json:"encrypted"`
	UsageCount  int       `gorm:"column:usage_count;default:0" json:"usage_count"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (s *SSHKey) TableName() string {
	return "ssh_keys"
}

type NodeEvent struct {
	ID        NodeID    `gorm:"primaryKey;column:id" json:"id"`
	NodeID    uint      `gorm:"column:node_id" json:"node_id"`
	Type      string    `gorm:"column:type;type:varchar(32)" json:"type"`
	Message   string    `gorm:"column:message;type:text" json:"message"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (n *NodeEvent) TableName() string {
	return "node_events"
}

type HostCloudAccount struct {
	ID                 uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Provider           string    `gorm:"column:provider;type:varchar(32);not null;index" json:"provider"`
	AccountName        string    `gorm:"column:account_name;type:varchar(128);not null" json:"account_name"`
	AccessKeyID        string    `gorm:"column:access_key_id;type:varchar(256);not null" json:"access_key_id"`
	AccessKeySecretEnc string    `gorm:"column:access_key_secret_enc;type:longtext;not null" json:"-"`
	RegionDefault      string    `gorm:"column:region_default;type:varchar(64)" json:"region_default"`
	Status             string    `gorm:"column:status;type:varchar(32);default:active" json:"status"`
	CreatedBy          uint64    `gorm:"column:created_by;index" json:"created_by"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (HostCloudAccount) TableName() string { return "host_cloud_accounts" }

type HostImportTask struct {
	ID           string    `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	Provider     string    `gorm:"column:provider;type:varchar(32);not null;index" json:"provider"`
	AccountID    uint64    `gorm:"column:account_id;index" json:"account_id"`
	RequestJSON  string    `gorm:"column:request_json;type:longtext" json:"request_json"`
	ResultJSON   string    `gorm:"column:result_json;type:longtext" json:"result_json"`
	Status       string    `gorm:"column:status;type:varchar(32);index" json:"status"`
	ErrorMessage string    `gorm:"column:error_message;type:text" json:"error_message"`
	CreatedBy    uint64    `gorm:"column:created_by;index" json:"created_by"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (HostImportTask) TableName() string { return "host_import_tasks" }

type HostVirtualizationTask struct {
	ID           string    `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	HostID       uint64    `gorm:"column:host_id;index" json:"host_id"`
	Hypervisor   string    `gorm:"column:hypervisor;type:varchar(32);not null" json:"hypervisor"`
	RequestJSON  string    `gorm:"column:request_json;type:longtext" json:"request_json"`
	VMName       string    `gorm:"column:vm_name;type:varchar(128)" json:"vm_name"`
	VMIP         string    `gorm:"column:vm_ip;type:varchar(64)" json:"vm_ip"`
	Status       string    `gorm:"column:status;type:varchar(32);index" json:"status"`
	ErrorMessage string    `gorm:"column:error_message;type:text" json:"error_message"`
	CreatedBy    uint64    `gorm:"column:created_by;index" json:"created_by"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (HostVirtualizationTask) TableName() string { return "host_virtualization_tasks" }
