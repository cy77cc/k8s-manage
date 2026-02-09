package model

import (
	"time"
)

// Project 项目表
type Project struct {
	ID          uint      `gorm:"primaryKey;column:id" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(64);not null;unique" json:"name"`
	Description string    `gorm:"column:description;type:varchar(256)" json:"description"`
	OwnerID     int64     `gorm:"column:owner_id" json:"owner_id"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	Services    []Service `gorm:"foreignKey:ProjectID" json:"services,omitempty"`
}

func (Project) TableName() string {
	return "projects"
}

// Service 服务表
type Service struct {
	ID            uint      `gorm:"primaryKey;column:id" json:"id"`
	ProjectID     uint      `gorm:"column:project_id;not null" json:"project_id"`
	Name          string    `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Type          string    `gorm:"column:type;type:varchar(32);not null" json:"type"` // stateless / stateful
	Image         string    `gorm:"column:image;type:varchar(256);not null" json:"image"`
	Replicas      int32     `gorm:"column:replicas;default:1" json:"replicas"`
	ServicePort   int32     `gorm:"column:service_port" json:"service_port"`
	ContainerPort int32     `gorm:"column:container_port" json:"container_port"`
	NodePort      int32     `gorm:"column:node_port" json:"node_port"`
	EnvVars       string    `gorm:"column:env_vars;type:json" json:"env_vars"`   // JSON string
	Resources     string    `gorm:"column:resources;type:json" json:"resources"` // JSON string
	YamlContent   string    `gorm:"column:yaml_content;type:mediumtext" json:"yaml_content"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Service) TableName() string {
	return "services"
}
