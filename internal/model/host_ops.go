package model

import "time"

// HostHealthSnapshot stores computed host health diagnostics.
type HostHealthSnapshot struct {
	ID                 uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	HostID             uint64    `gorm:"column:host_id;index" json:"host_id"`
	State              string    `gorm:"column:state;type:varchar(32);index" json:"state"`
	ConnectivityStatus string    `gorm:"column:connectivity_status;type:varchar(32)" json:"connectivity_status"`
	ResourceStatus     string    `gorm:"column:resource_status;type:varchar(32)" json:"resource_status"`
	SystemStatus       string    `gorm:"column:system_status;type:varchar(32)" json:"system_status"`
	LatencyMS          int64     `gorm:"column:latency_ms" json:"latency_ms"`
	CpuLoad            float64   `gorm:"column:cpu_load" json:"cpu_load"`
	MemoryUsedMB       int       `gorm:"column:memory_used_mb" json:"memory_used_mb"`
	MemoryTotalMB      int       `gorm:"column:memory_total_mb" json:"memory_total_mb"`
	DiskUsedPct        float64   `gorm:"column:disk_used_pct" json:"disk_used_pct"`
	InodeUsedPct       float64   `gorm:"column:inode_used_pct" json:"inode_used_pct"`
	SummaryJSON        string    `gorm:"column:summary_json;type:longtext" json:"summary_json"`
	ErrorMessage       string    `gorm:"column:error_message;type:text" json:"error_message"`
	CheckedAt          time.Time `gorm:"column:checked_at;index" json:"checked_at"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (HostHealthSnapshot) TableName() string { return "host_health_snapshots" }

// AIHostExecutionRecord stores per-host execution result for AI command/script tasks.
type AIHostExecutionRecord struct {
	ID          uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ExecutionID string     `gorm:"column:execution_id;type:varchar(64);index" json:"execution_id"`
	CommandID   string     `gorm:"column:command_id;type:varchar(64);index" json:"command_id"`
	HostID      uint64     `gorm:"column:host_id;index" json:"host_id"`
	HostIP      string     `gorm:"column:host_ip;type:varchar(64)" json:"host_ip"`
	HostName    string     `gorm:"column:host_name;type:varchar(128)" json:"host_name"`
	CommandText string     `gorm:"column:command_text;type:text" json:"command_text"`
	ScriptPath  string     `gorm:"column:script_path;type:varchar(256)" json:"script_path"`
	Status      string     `gorm:"column:status;type:varchar(32);index" json:"status"`
	StdoutText  string     `gorm:"column:stdout_text;type:longtext" json:"stdout_text"`
	StderrText  string     `gorm:"column:stderr_text;type:longtext" json:"stderr_text"`
	ExitCode    int        `gorm:"column:exit_code" json:"exit_code"`
	StartedAt   *time.Time `gorm:"column:started_at" json:"started_at"`
	FinishedAt  *time.Time `gorm:"column:finished_at" json:"finished_at"`
	PolicyJSON  string     `gorm:"column:policy_json;type:longtext" json:"policy_json"`
	CreatedBy   uint64     `gorm:"column:created_by" json:"created_by"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AIHostExecutionRecord) TableName() string { return "ai_host_execution_records" }
