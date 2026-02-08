package model

import (
	"time"
)

type NodeID uint

type Node struct {
	ID          NodeID    `gorm:"primaryKey;column:id" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Hostname    string    `gorm:"column:hostname;type:varchar(64)" json:"hostname"`
	Labels      string    `gorm:"column:labels;type:json" json:"labels"`
	Description string    `gorm:"column:description;type:varchar(256)" json:"description"`
	IP          string    `gorm:"column:ip;type:varchar(45);not null" json:"ip"`
	Port        int       `gorm:"column:port;default:22" json:"port"`
	SSHUser     string    `gorm:"column:ssh_user;type:varchar(64);not null;default:root" json:"ssh_user"`
	SSHPassword string    `gorm:"column:ssh_password;type:varchar(256)" json:"ssh_password"`
	SSHKeyID    *NodeID   `gorm:"column:ssh_key_id" json:"ssh_key_id"`
	OS          string    `gorm:"column:os;type:varchar(64)" json:"os"`
	Arch        string    `gorm:"column:arch;type:varchar(32)" json:"arch"`
	Kernel      string    `gorm:"column:kernel;type:varchar(64)" json:"kernel"`
	CpuCores    int       `gorm:"column:cpu_cores" json:"cpu_cores"`
	MemoryMB    int       `gorm:"column:memory_mb" json:"memory_mb"`
	DiskGB      int       `gorm:"column:disk_gb" json:"disk_gb"`
	Status      string    `gorm:"column:status;type:varchar(32);not null" json:"status"`
	Role        string    `gorm:"column:role;type:varchar(32)" json:"role"`
	ClusterID   uint      `gorm:"column:cluster_id" json:"cluster_id"`
	LastCheckAt time.Time `gorm:"column:last_check_at" json:"last_check_at"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (n *Node) TableName() string {
	return "nodes"
}

type SSHKey struct {
	ID         NodeID    `gorm:"primaryKey;column:id" json:"id"`
	Name       string    `gorm:"column:name;type:varchar(64)" json:"name"`
	PublicKey  string    `gorm:"column:public_key;type:text;not null" json:"public_key"`
	PrivateKey string    `gorm:"column:private_key;type:text;not null" json:"private_key"`
	Passphrase string    `gorm:"column:passphrase;type:varchar(128)" json:"passphrase"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
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
