package model

import (
	"time"
)

// Cluster 集群表
type Cluster struct {
	ID          uint      `gorm:"primaryKey;column:id" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Description string    `gorm:"column:description;type:varchar(256)" json:"description"`
	Version     string    `gorm:"column:version;type:varchar(64)" json:"version"`
	Status      string    `gorm:"column:status;type:varchar(32);not null" json:"status"`
	Type        string    `gorm:"column:type;type:varchar(32);not null" json:"type"` // kubernetes / openshift
	Endpoint    string    `gorm:"column:endpoint;type:varchar(256);not null" json:"endpoint"`
	KubeConfig  string    `gorm:"column:kubeconfig;type:mediumtext" json:"kubeconfig"`
	CACert      string    `gorm:"column:ca_cert;type:text" json:"ca_cert"`
	Token       string    `gorm:"column:token;type:text" json:"token"`
	Nodes       string    `gorm:"column:nodes;type:json" json:"nodes"`
	AuthMethod  string    `gorm:"column:auth_method;type:varchar(32);not null" json:"auth_method"` // token / basic
	CreatedBy   string    `gorm:"column:created_by;type:varchar(64)" json:"created_by"`
	UpdatedBy   string    `gorm:"column:updated_by;type:varchar(64)" json:"updated_by"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Cluster) TableName() string {
	return "clusters"
}
