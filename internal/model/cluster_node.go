package model

import "time"

// ClusterNode 集群节点表，存储从 K8s API 同步的节点信息
type ClusterNode struct {
	ID               uint       `gorm:"primaryKey;column:id" json:"id"`
	ClusterID        uint       `gorm:"column:cluster_id;not null;index;uniqueIndex:uk_cluster_node,priority:1" json:"cluster_id"`
	HostID           *uint      `gorm:"column:host_id;index" json:"host_id"`
	Name             string     `gorm:"column:name;type:varchar(64);not null;uniqueIndex:uk_cluster_node,priority:2" json:"name"`
	IP               string     `gorm:"column:ip;type:varchar(45);not null" json:"ip"`
	Role             string     `gorm:"column:role;type:varchar(32);not null" json:"role"` // control-plane / worker / etcd
	Status           string     `gorm:"column:status;type:varchar(32);not null;index" json:"status"` // ready / notready / unknown
	KubeletVersion   string     `gorm:"column:kubelet_version;type:varchar(32)" json:"kubelet_version"`
	KubeProxyVersion string     `gorm:"column:kube_proxy_version;type:varchar(32)" json:"kube_proxy_version"`
	ContainerRuntime string     `gorm:"column:container_runtime;type:varchar(32)" json:"container_runtime"`
	OSImage          string     `gorm:"column:os_image;type:varchar(128)" json:"os_image"`
	KernelVersion    string     `gorm:"column:kernel_version;type:varchar(64)" json:"kernel_version"`
	AllocatableCPU   string     `gorm:"column:allocatable_cpu;type:varchar(16)" json:"allocatable_cpu"`
	AllocatableMem   string     `gorm:"column:allocatable_mem;type:varchar(16)" json:"allocatable_mem"`
	AllocatablePods  int        `gorm:"column:allocatable_pods;default:0" json:"allocatable_pods"`
	Labels           string     `gorm:"column:labels;type:json" json:"labels"`
	Taints           string     `gorm:"column:taints;type:json" json:"taints"`
	Conditions       string     `gorm:"column:conditions;type:json" json:"conditions"` // [{type, status, reason, message}]
	JoinedAt         *time.Time `gorm:"column:joined_at" json:"joined_at"`
	LastSeenAt       *time.Time `gorm:"column:last_seen_at" json:"last_seen_at"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ClusterNode) TableName() string {
	return "cluster_nodes"
}
