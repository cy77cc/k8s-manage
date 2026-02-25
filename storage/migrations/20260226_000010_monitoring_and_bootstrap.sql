-- +migrate Up
CREATE TABLE IF NOT EXISTS alert_rules (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  metric VARCHAR(64) NOT NULL,
  operator VARCHAR(8) NOT NULL DEFAULT 'gt',
  threshold DECIMAL(14,4) NOT NULL DEFAULT 0,
  duration_sec INT NOT NULL DEFAULT 300,
  severity VARCHAR(16) NOT NULL DEFAULT 'warning',
  source VARCHAR(32) NOT NULL DEFAULT 'system',
  scope VARCHAR(32) NOT NULL DEFAULT 'global',
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_alert_rules_metric (metric),
  INDEX idx_alert_rules_enabled (enabled)
);

CREATE TABLE IF NOT EXISTS alerts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  rule_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  title VARCHAR(255) NOT NULL,
  message TEXT,
  metric VARCHAR(64) NOT NULL DEFAULT '',
  value DECIMAL(14,4) NOT NULL DEFAULT 0,
  threshold DECIMAL(14,4) NOT NULL DEFAULT 0,
  severity VARCHAR(16) NOT NULL DEFAULT 'warning',
  source VARCHAR(128) NOT NULL DEFAULT '',
  status VARCHAR(16) NOT NULL DEFAULT 'firing',
  triggered_at TIMESTAMP NULL,
  resolved_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_alerts_rule (rule_id),
  INDEX idx_alerts_metric (metric),
  INDEX idx_alerts_source (source),
  INDEX idx_alerts_status (status),
  INDEX idx_alerts_created (created_at)
);

CREATE TABLE IF NOT EXISTS metric_points (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  metric VARCHAR(64) NOT NULL,
  source VARCHAR(128) NOT NULL DEFAULT '',
  value DECIMAL(14,4) NOT NULL DEFAULT 0,
  collected_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_metric_points_metric_time (metric, collected_at),
  INDEX idx_metric_points_source (source)
);

CREATE TABLE IF NOT EXISTS cluster_bootstrap_tasks (
  id VARCHAR(64) PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  control_plane_host_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  worker_ids_json LONGTEXT,
  cni VARCHAR(32) NOT NULL DEFAULT 'flannel',
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  result_json LONGTEXT,
  error_message TEXT,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_cluster_bootstrap_control (control_plane_host_id),
  INDEX idx_cluster_bootstrap_status (status),
  INDEX idx_cluster_bootstrap_created_by (created_by)
);

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '监控查看', 'monitoring:read', 3, 'monitoring', 'read', '查看监控与告警', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'monitoring:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '监控管理', 'monitoring:write', 3, 'monitoring', 'write', '管理监控规则', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'monitoring:write');

-- +migrate Down
DELETE FROM role_permissions WHERE permission_id IN (
  SELECT id FROM permissions WHERE code IN ('monitoring:read', 'monitoring:write')
);
DELETE FROM permissions WHERE code IN ('monitoring:read', 'monitoring:write');

DROP TABLE IF EXISTS cluster_bootstrap_tasks;
DROP TABLE IF EXISTS metric_points;
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS alert_rules;
