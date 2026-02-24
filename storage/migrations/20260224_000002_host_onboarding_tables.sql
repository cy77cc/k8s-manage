-- +migrate Up
CREATE TABLE IF NOT EXISTS host_probe_sessions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  token_hash VARCHAR(128) NOT NULL UNIQUE,
  name VARCHAR(128) NOT NULL,
  ip VARCHAR(64) NOT NULL,
  port INT NOT NULL DEFAULT 22,
  auth_type VARCHAR(32) NOT NULL,
  username VARCHAR(128) NOT NULL,
  ssh_key_id BIGINT UNSIGNED NULL,
  password_cipher TEXT,
  reachable TINYINT(1) NOT NULL DEFAULT 0,
  latency_ms BIGINT NOT NULL DEFAULT 0,
  facts_json LONGTEXT,
  warnings_json LONGTEXT,
  expires_at TIMESTAMP NOT NULL,
  consumed_at TIMESTAMP NULL,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_host_probe_expires (expires_at),
  INDEX idx_host_probe_creator (created_by)
);

-- +migrate Down
DROP TABLE IF EXISTS host_probe_sessions;
