package model

import "time"

// HostProbeSession stores one-time onboarding probe result.
type HostProbeSession struct {
	ID             uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TokenHash      string     `gorm:"column:token_hash;type:varchar(128);uniqueIndex;not null" json:"-"`
	Name           string     `gorm:"column:name;type:varchar(128);not null" json:"name"`
	IP             string     `gorm:"column:ip;type:varchar(64);not null" json:"ip"`
	Port           int        `gorm:"column:port;not null;default:22" json:"port"`
	AuthType       string     `gorm:"column:auth_type;type:varchar(32);not null" json:"auth_type"`
	Username       string     `gorm:"column:username;type:varchar(128);not null" json:"username"`
	SSHKeyID       *uint64    `gorm:"column:ssh_key_id" json:"ssh_key_id"`
	PasswordCipher string     `gorm:"column:password_cipher;type:text" json:"-"`
	Reachable      bool       `gorm:"column:reachable;not null;default:false" json:"reachable"`
	LatencyMS      int64      `gorm:"column:latency_ms;not null;default:0" json:"latency_ms"`
	FactsJSON      string     `gorm:"column:facts_json;type:longtext" json:"facts_json"`
	WarningsJSON   string     `gorm:"column:warnings_json;type:longtext" json:"warnings_json"`
	ExpiresAt      time.Time  `gorm:"column:expires_at;index" json:"expires_at"`
	ConsumedAt     *time.Time `gorm:"column:consumed_at" json:"consumed_at"`
	CreatedBy      uint64     `gorm:"column:created_by;index" json:"created_by"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (HostProbeSession) TableName() string { return "host_probe_sessions" }
