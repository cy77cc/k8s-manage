package model

import "time"

// AICheckPoint stores eino checkpoint blobs for interrupt/resume.
type AICheckPoint struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Key       string    `gorm:"column:key;type:varchar(255);uniqueIndex;not null" json:"key"`
	Value     []byte    `gorm:"column:value;type:mediumblob;not null" json:"value"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AICheckPoint) TableName() string { return "ai_checkpoints" }
