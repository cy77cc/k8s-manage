package model

import "time"

// Job 任务调度定义
type Job struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"type:varchar(255);not null" json:"name"`
	Type        string     `gorm:"type:varchar(32);not null;default:'shell'" json:"type"` // shell, script
	Command     string     `gorm:"type:text" json:"command"`
	HostIDs     string     `gorm:"type:text" json:"host_ids"`
	Cron        string     `gorm:"type:varchar(64)" json:"cron"`
	Status      string     `gorm:"type:varchar(32);not null;default:'pending'" json:"status"` // pending, running, success, failed
	Timeout     int        `gorm:"default:300" json:"timeout"`
	Priority    int        `gorm:"default:0" json:"priority"`
	Description string     `gorm:"type:text" json:"description"`
	LastRun     *time.Time `json:"last_run"`
	NextRun     *time.Time `json:"next_run"`
	CreatedBy   uint       `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Job) TableName() string {
	return "jobs"
}

// JobExecution 任务执行记录
type JobExecution struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	JobID     uint       `gorm:"not null;index" json:"job_id"`
	HostID    uint       `json:"host_id"`
	HostIP    string     `gorm:"type:varchar(64)" json:"host_ip"`
	Status    string     `gorm:"type:varchar(32);not null;default:'pending'" json:"status"` // pending, running, success, failed
	ExitCode  int        `json:"exit_code"`
	Output    string     `gorm:"type:text" json:"output"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	CreatedAt time.Time  `json:"created_at"`
}

func (JobExecution) TableName() string {
	return "job_executions"
}

// JobLog 任务执行日志
type JobLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	JobID       uint      `gorm:"not null;index" json:"job_id"`
	ExecutionID uint      `gorm:"index" json:"execution_id"`
	Level       string    `gorm:"type:varchar(16);default:'info'" json:"level"` // info, warn, error
	Message     string    `gorm:"type:text" json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}

func (JobLog) TableName() string {
	return "job_logs"
}
