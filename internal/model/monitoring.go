package model

import "time"

type AlertRule struct {
	ID          uint      `gorm:"primaryKey;column:id" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(128);not null" json:"name"`
	Metric      string    `gorm:"column:metric;type:varchar(64);not null;index" json:"metric"`
	Operator    string    `gorm:"column:operator;type:varchar(8);default:'gt'" json:"operator"`
	Threshold   float64   `gorm:"column:threshold;type:decimal(12,4);default:0" json:"threshold"`
	DurationSec int       `gorm:"column:duration_sec;default:300" json:"duration_sec"`
	Severity    string    `gorm:"column:severity;type:varchar(16);default:'warning'" json:"severity"`
	Source      string    `gorm:"column:source;type:varchar(32);default:'system'" json:"source"`
	Scope       string    `gorm:"column:scope;type:varchar(32);default:'global'" json:"scope"`
	Enabled     bool      `gorm:"column:enabled;default:true" json:"enabled"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AlertRule) TableName() string { return "alert_rules" }

type AlertEvent struct {
	ID          uint      `gorm:"primaryKey;column:id" json:"id"`
	RuleID      uint      `gorm:"column:rule_id;index" json:"rule_id"`
	Title       string    `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Message     string    `gorm:"column:message;type:text" json:"message"`
	Metric      string    `gorm:"column:metric;type:varchar(64);index" json:"metric"`
	Value       float64   `gorm:"column:value;type:decimal(14,4);default:0" json:"value"`
	Threshold   float64   `gorm:"column:threshold;type:decimal(14,4);default:0" json:"threshold"`
	Severity    string    `gorm:"column:severity;type:varchar(16);default:'warning'" json:"severity"`
	Source      string    `gorm:"column:source;type:varchar(128);index" json:"source"`
	Status      string    `gorm:"column:status;type:varchar(16);default:'firing';index" json:"status"`
	TriggeredAt time.Time `gorm:"column:triggered_at;index" json:"triggered_at"`
	ResolvedAt  time.Time `gorm:"column:resolved_at" json:"resolved_at"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AlertEvent) TableName() string { return "alerts" }

type MetricPoint struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	Metric    string    `gorm:"column:metric;type:varchar(64);index:idx_metric_time,priority:1" json:"metric"`
	Source    string    `gorm:"column:source;type:varchar(128);index" json:"source"`
	Value     float64   `gorm:"column:value;type:decimal(14,4);default:0" json:"value"`
	Collected time.Time `gorm:"column:collected_at;index:idx_metric_time,priority:2" json:"collected_at"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (MetricPoint) TableName() string { return "metric_points" }

type ClusterBootstrapTask struct {
	ID             string    `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	Name           string    `gorm:"column:name;type:varchar(128);not null" json:"name"`
	ControlPlaneID uint      `gorm:"column:control_plane_host_id;index" json:"control_plane_host_id"`
	WorkerIDsJSON  string    `gorm:"column:worker_ids_json;type:longtext" json:"worker_ids_json"`
	CNI            string    `gorm:"column:cni;type:varchar(32);default:'flannel'" json:"cni"`
	Status         string    `gorm:"column:status;type:varchar(32);index" json:"status"`
	ResultJSON     string    `gorm:"column:result_json;type:longtext" json:"result_json"`
	ErrorMessage   string    `gorm:"column:error_message;type:text" json:"error_message"`
	CreatedBy      uint64    `gorm:"column:created_by;index" json:"created_by"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ClusterBootstrapTask) TableName() string { return "cluster_bootstrap_tasks" }
