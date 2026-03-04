package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TemplateVisibilityPrivate = "private"
	TemplateVisibilityPublic  = "public"
)

const (
	TemplateStatusDraft         = "draft"
	TemplateStatusPendingReview = "pending_review"
	TemplateStatusPublished     = "published"
	TemplateStatusRejected      = "rejected"
)

type ServiceTemplate struct {
	ID              uint      `gorm:"primaryKey;column:id" json:"id"`
	Name            string    `gorm:"column:name;type:varchar(128);not null;unique" json:"name"`
	DisplayName     string    `gorm:"column:display_name;type:varchar(256);not null" json:"display_name"`
	Description     string    `gorm:"column:description;type:text" json:"description"`
	Icon            string    `gorm:"column:icon;type:varchar(256)" json:"icon"`
	CategoryID      uint      `gorm:"column:category_id;not null;index" json:"category_id"`
	Version         string    `gorm:"column:version;type:varchar(32);default:'1.0.0'" json:"version"`
	OwnerID         uint64    `gorm:"column:owner_id;not null;index" json:"owner_id"`
	Visibility      string    `gorm:"column:visibility;type:varchar(16);default:'private'" json:"visibility"`
	Status          string    `gorm:"column:status;type:varchar(32);default:'draft';index" json:"status"`
	K8sTemplate     string    `gorm:"column:k8s_template;type:mediumtext" json:"k8s_template"`
	ComposeTemplate string    `gorm:"column:compose_template;type:mediumtext" json:"compose_template"`
	VariablesSchema string    `gorm:"column:variables_schema;type:json" json:"variables_schema"`
	Readme          string    `gorm:"column:readme;type:text" json:"readme"`
	Tags            string    `gorm:"column:tags;type:json" json:"tags"`
	DeployCount     int       `gorm:"column:deploy_count;default:0" json:"deploy_count"`
	ReviewNote      string    `gorm:"column:review_note;type:text" json:"review_note"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ServiceTemplate) TableName() string { return "service_templates" }

func (m *ServiceTemplate) Create(db *gorm.DB) error {
	return db.Create(m).Error
}

func (m *ServiceTemplate) GetByID(db *gorm.DB, id uint) error {
	return db.First(m, id).Error
}

func (m *ServiceTemplate) GetByName(db *gorm.DB, name string) error {
	return db.Where("name = ?", name).First(m).Error
}

func (m *ServiceTemplate) Update(db *gorm.DB, id uint, payload map[string]any) error {
	return db.Model(&ServiceTemplate{}).Where("id = ?", id).Updates(payload).Error
}

func (m *ServiceTemplate) Delete(db *gorm.DB, id uint) error {
	return db.Delete(&ServiceTemplate{}, id).Error
}

func ListServiceTemplates(db *gorm.DB, queryFn func(*gorm.DB) *gorm.DB) ([]ServiceTemplate, error) {
	q := db.Model(&ServiceTemplate{})
	if queryFn != nil {
		q = queryFn(q)
	}
	rows := make([]ServiceTemplate, 0, 32)
	err := q.Order("updated_at DESC").Find(&rows).Error
	return rows, err
}
