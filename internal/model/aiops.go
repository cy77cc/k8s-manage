package model

import "time"

// RiskFinding 风险发现
type RiskFinding struct {
	ID          uint     `gorm:"primaryKey" json:"id"`
	Type        string   `gorm:"type:varchar(64);not null;index" json:"type"`
	Severity    string   `gorm:"type:varchar(16);not null;index" json:"severity"` // critical, high, medium, low
	Title       string   `gorm:"type:varchar(255);not null" json:"title"`
	Description string   `gorm:"type:text" json:"description"`
	ServiceID   uint     `gorm:"index" json:"service_id"`
	ServiceName string   `gorm:"type:varchar(255)" json:"service_name"`
	Metadata    string   `gorm:"type:text" json:"metadata"`
	CreatedAt   time.Time `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
}

func (RiskFinding) TableName() string {
	return "risk_findings"
}

// Anomaly 异常检测
type Anomaly struct {
	ID          uint     `gorm:"primaryKey" json:"id"`
	Type        string   `gorm:"type:varchar(64);not null;index" json:"type"`
	Metric      string   `gorm:"type:varchar(64);not null" json:"metric"`
	Value       float64  `json:"value"`
	Threshold   float64  `json:"threshold"`
	ServiceID   uint     `gorm:"index" json:"service_id"`
	ServiceName string   `gorm:"type:varchar(255)" json:"service_name"`
	DetectedAt  time.Time `json:"detected_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
}

func (Anomaly) TableName() string {
	return "anomalies"
}

// Suggestion 优化建议
type Suggestion struct {
	ID          uint     `gorm:"primaryKey" json:"id"`
	Type        string   `gorm:"type:varchar(64);not null;index" json:"type"`
	Title       string   `gorm:"type:varchar(255);not null" json:"title"`
	Description string   `gorm:"type:text" json:"description"`
	Impact      string   `gorm:"type:varchar(16);not null" json:"impact"` // high, medium, low
	ServiceID   uint     `gorm:"index" json:"service_id"`
	ServiceName string   `gorm:"type:varchar(255)" json:"service_name"`
	CreatedAt   time.Time `json:"created_at"`
	AppliedAt   *time.Time `json:"applied_at"`
}

func (Suggestion) TableName() string {
	return "suggestions"
}
