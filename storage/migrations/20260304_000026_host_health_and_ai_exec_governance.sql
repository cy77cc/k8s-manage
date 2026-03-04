-- +migrate Up
ALTER TABLE nodes
  ADD COLUMN health_state VARCHAR(32) NOT NULL DEFAULT 'unknown',
  ADD COLUMN maintenance_reason VARCHAR(512) DEFAULT '',
  ADD COLUMN maintenance_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  ADD COLUMN maintenance_started_at TIMESTAMP NULL DEFAULT NULL,
  ADD COLUMN maintenance_until TIMESTAMP NULL DEFAULT NULL;

CREATE INDEX idx_nodes_health_state ON nodes(health_state);
CREATE INDEX idx_nodes_maintenance_until ON nodes(maintenance_until);

CREATE TABLE IF NOT EXISTS host_health_snapshots (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  host_id BIGINT UNSIGNED NOT NULL,
  state VARCHAR(32) NOT NULL DEFAULT 'unknown',
  connectivity_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
  resource_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
  system_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
  latency_ms BIGINT NOT NULL DEFAULT 0,
  cpu_load DOUBLE NOT NULL DEFAULT 0,
  memory_used_mb INT NOT NULL DEFAULT 0,
  memory_total_mb INT NOT NULL DEFAULT 0,
  disk_used_pct DOUBLE NOT NULL DEFAULT 0,
  inode_used_pct DOUBLE NOT NULL DEFAULT 0,
  summary_json LONGTEXT,
  error_message TEXT,
  checked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX idx_host_health_host (host_id),
  INDEX idx_host_health_state (state),
  INDEX idx_host_health_checked_at (checked_at)
);

CREATE TABLE IF NOT EXISTS ai_host_execution_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  execution_id VARCHAR(64) NOT NULL,
  command_id VARCHAR(64) DEFAULT '',
  host_id BIGINT UNSIGNED NOT NULL,
  host_ip VARCHAR(64) DEFAULT '',
  host_name VARCHAR(128) DEFAULT '',
  command_text TEXT,
  script_path VARCHAR(256) DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT 'running',
  stdout_text LONGTEXT,
  stderr_text LONGTEXT,
  exit_code INT NOT NULL DEFAULT 0,
  started_at TIMESTAMP NULL DEFAULT NULL,
  finished_at TIMESTAMP NULL DEFAULT NULL,
  policy_json LONGTEXT,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX idx_ai_host_exec_execution (execution_id),
  INDEX idx_ai_host_exec_command (command_id),
  INDEX idx_ai_host_exec_host (host_id),
  INDEX idx_ai_host_exec_status (status),
  INDEX idx_ai_host_exec_created (created_at)
);

-- +migrate Down
DROP TABLE IF EXISTS ai_host_execution_records;
DROP TABLE IF EXISTS host_health_snapshots;

DROP INDEX idx_nodes_maintenance_until ON nodes;
DROP INDEX idx_nodes_health_state ON nodes;

ALTER TABLE nodes
  DROP COLUMN maintenance_until,
  DROP COLUMN maintenance_started_at,
  DROP COLUMN maintenance_by,
  DROP COLUMN maintenance_reason,
  DROP COLUMN health_state;
