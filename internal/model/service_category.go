package model

import (
	"time"

	"gorm.io/gorm"
)

type ServiceCategory struct {
	ID          uint      `gorm:"primaryKey;column:id" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(64);not null;unique" json:"name"`
	DisplayName string    `gorm:"column:display_name;type:varchar(128);not null" json:"display_name"`
	Icon        string    `gorm:"column:icon;type:varchar(256)" json:"icon"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	SortOrder   int       `gorm:"column:sort_order;default:0" json:"sort_order"`
	IsSystem    bool      `gorm:"column:is_system;default:false" json:"is_system"`
	CreatedBy   uint64    `gorm:"column:created_by;default:0" json:"created_by"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ServiceCategory) TableName() string { return "service_categories" }

func (m *ServiceCategory) Create(db *gorm.DB) error {
	return db.Create(m).Error
}

func (m *ServiceCategory) GetByID(db *gorm.DB, id uint) error {
	return db.First(m, id).Error
}

func (m *ServiceCategory) GetByName(db *gorm.DB, name string) error {
	return db.Where("name = ?", name).First(m).Error
}

func (m *ServiceCategory) Update(db *gorm.DB, id uint, payload map[string]any) error {
	return db.Model(&ServiceCategory{}).Where("id = ?", id).Updates(payload).Error
}

func (m *ServiceCategory) Delete(db *gorm.DB, id uint) error {
	return db.Delete(&ServiceCategory{}, id).Error
}

func ListServiceCategories(db *gorm.DB) ([]ServiceCategory, error) {
	rows := make([]ServiceCategory, 0, 16)
	err := db.Order("sort_order ASC, id ASC").Find(&rows).Error
	return rows, err
}
