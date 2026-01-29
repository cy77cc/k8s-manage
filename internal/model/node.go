package model

import (
	"time"
)

type NodeID uint

type Node struct {
	ID          NodeID    `gorm:"primaryKey;column:id" json:"id"`
	Name        string    `gorm:"column:name" json:"name"`
	Hostname    string    `gorm:"column:hostname" json:"hostname"`
	Labels      string    `gorm:"column:labels" json:"labels"`
	Description string    `gorm:"column:description" json:"description"`
	IP          string    `gorm:"column:ip" json:"ip"`
	Port        int       `gorm:"column:port" json:"port"`
	SshUser     string    `gorm:"column:ssh_user" json:"ssh_user"`
	SshKeyID    uint      `gorm:"column:ssh_key_id" json:"ssh_key_id"`
	OS          string    `gorm:"column:os" json:"os"`
	Arch        string    `gorm:"column:arch" json:"arch"`
	Kernel      string    `gorm:"column:kernel" json:"kernel"`
	CpuCores    int       `gorm:"column:cpu_cores" json:"cpu_cores"`
	MemoryMB    int       `gorm:"column:memory_mb" json:"memory_mb"`
	DiskGB      int       `gorm:"column:disk_gb" json:"disk_gb"`
	Status      string    `gorm:"column:status" json:"status"`
	Role        string    `gorm:"column:role" json:"role"`
	ClusterID   uint      `gorm:"column:cluster_id" json:"cluster_id"`
	LastCheckAt time.Time `gorm:"column:last_check_at" json:"last_check_at"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (n *Node) TableName() string {
	return "nodes"
}

type SSHKey struct {
	ID        NodeID    `gorm:"primaryKey;column:id" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`
	PublicKey string    `gorm:"column:public_key" json:"public_key"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (s *SSHKey) TableName() string {
	return "ssh_keys"
}

type NodeEvent struct {
	ID        NodeID    `gorm:"primaryKey;column:id" json:"id"`
	NodeID    uint      `gorm:"column:node_id" json:"node_id"`
	Type      string    `gorm:"column:type" json:"type"`
	Message   string    `gorm:"column:message" json:"message"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (n *NodeEvent) TableName() string {
	return "node_events"
}
