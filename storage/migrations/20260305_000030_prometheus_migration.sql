-- +migrate Up

ALTER TABLE alert_rules
  ADD COLUMN promql_expr VARCHAR(512) DEFAULT '' COMMENT 'PromQL expression (optional, overrides generated query)',
  ADD COLUMN labels_json LONGTEXT COMMENT 'Prometheus labels JSON',
  ADD COLUMN annotations_json LONGTEXT COMMENT 'Prometheus annotations JSON';

CREATE TABLE IF NOT EXISTS alert_silences (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  silence_id VARCHAR(64) NOT NULL COMMENT 'Alertmanager silence ID',
  matchers_json LONGTEXT NOT NULL COMMENT 'silence matchers JSON',
  starts_at TIMESTAMP NOT NULL,
  ends_at TIMESTAMP NOT NULL,
  created_by BIGINT UNSIGNED NOT NULL,
  comment VARCHAR(512) DEFAULT '',
  status VARCHAR(16) DEFAULT 'active' COMMENT 'active, expired',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_alert_silences_silence_id (silence_id),
  INDEX idx_alert_silences_time (starts_at, ends_at),
  INDEX idx_alert_silences_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='alert silences';

DROP TABLE IF EXISTS metric_points;
DROP TABLE IF EXISTS alert_rule_evaluations;

-- +migrate Down

DROP TABLE IF EXISTS alert_silences;

CREATE TABLE IF NOT EXISTS metric_points (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  metric VARCHAR(64) NOT NULL,
  source VARCHAR(128) DEFAULT '',
  dimensions_json LONGTEXT,
  value DECIMAL(14,4) DEFAULT 0,
  collected_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_metric_time (metric, collected_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS alert_rule_evaluations (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  rule_id BIGINT UNSIGNED NOT NULL,
  metric VARCHAR(64) NOT NULL,
  operator VARCHAR(8) DEFAULT 'gt',
  value DECIMAL(14,4) DEFAULT 0,
  threshold DECIMAL(14,4) DEFAULT 0,
  triggered TINYINT(1) DEFAULT 0,
  prev_state VARCHAR(16) DEFAULT 'normal',
  next_state VARCHAR(16) DEFAULT 'normal',
  evaluated_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_rule_eval_rule_id (rule_id),
  INDEX idx_rule_eval_metric (metric),
  INDEX idx_rule_eval_triggered (triggered)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

ALTER TABLE alert_notification_channels
  DROP COLUMN config_json,
  DROP COLUMN provider;

ALTER TABLE alert_rules
  DROP COLUMN annotations_json,
  DROP COLUMN labels_json,
  DROP COLUMN promql_expr;
