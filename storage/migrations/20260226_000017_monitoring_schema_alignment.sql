-- +migrate Up
SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules' AND COLUMN_NAME = 'window_sec'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists = 0,
  'ALTER TABLE alert_rules ADD COLUMN window_sec INT NOT NULL DEFAULT 3600 AFTER duration_sec',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules' AND COLUMN_NAME = 'granularity_sec'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists = 0,
  'ALTER TABLE alert_rules ADD COLUMN granularity_sec INT NOT NULL DEFAULT 60 AFTER window_sec',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules' AND COLUMN_NAME = 'dimensions_json'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists = 0,
  'ALTER TABLE alert_rules ADD COLUMN dimensions_json LONGTEXT AFTER granularity_sec',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules' AND COLUMN_NAME = 'state'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists = 0,
  'ALTER TABLE alert_rules ADD COLUMN state VARCHAR(16) NOT NULL DEFAULT ''enabled'' AFTER scope',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules' AND INDEX_NAME = 'idx_alert_rules_state'
);
SET @sql := IF(@idx_exists = 0,
  'CREATE INDEX idx_alert_rules_state ON alert_rules (state)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'metric_points'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'metric_points' AND COLUMN_NAME = 'dimensions_json'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists = 0,
  'ALTER TABLE metric_points ADD COLUMN dimensions_json LONGTEXT AFTER source',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS alert_rule_evaluations (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  rule_id BIGINT UNSIGNED NOT NULL,
  metric VARCHAR(64) NOT NULL DEFAULT '',
  operator VARCHAR(8) NOT NULL DEFAULT 'gt',
  value DECIMAL(14,4) NOT NULL DEFAULT 0,
  threshold DECIMAL(14,4) NOT NULL DEFAULT 0,
  triggered TINYINT(1) NOT NULL DEFAULT 0,
  prev_state VARCHAR(16) NOT NULL DEFAULT 'normal',
  next_state VARCHAR(16) NOT NULL DEFAULT 'normal',
  evaluated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_alert_rule_evaluations_rule (rule_id),
  INDEX idx_alert_rule_evaluations_metric (metric),
  INDEX idx_alert_rule_evaluations_triggered (triggered),
  INDEX idx_alert_rule_evaluations_evaluated (evaluated_at),
  INDEX idx_alert_rule_evaluations_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS alert_notification_channels (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  type VARCHAR(32) NOT NULL,
  provider VARCHAR(64) NOT NULL DEFAULT '',
  target VARCHAR(512) NOT NULL DEFAULT '',
  config_json LONGTEXT,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_alert_notification_channels_type (type),
  INDEX idx_alert_notification_channels_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS alert_notification_deliveries (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  alert_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  rule_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  channel_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  channel_type VARCHAR(32) NOT NULL DEFAULT '',
  target VARCHAR(512) NOT NULL DEFAULT '',
  status VARCHAR(16) NOT NULL DEFAULT 'sent',
  error_message TEXT,
  delivered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_alert_notification_deliveries_alert (alert_id),
  INDEX idx_alert_notification_deliveries_rule (rule_id),
  INDEX idx_alert_notification_deliveries_channel (channel_id),
  INDEX idx_alert_notification_deliveries_channel_type (channel_type),
  INDEX idx_alert_notification_deliveries_status (status),
  INDEX idx_alert_notification_deliveries_delivered (delivered_at),
  INDEX idx_alert_notification_deliveries_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +migrate Down
DROP TABLE IF EXISTS alert_notification_deliveries;
DROP TABLE IF EXISTS alert_notification_channels;
DROP TABLE IF EXISTS alert_rule_evaluations;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules' AND INDEX_NAME = 'idx_alert_rules_state'
);
SET @sql := IF(@idx_exists > 0,
  'DROP INDEX idx_alert_rules_state ON alert_rules',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'metric_points'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'metric_points' AND COLUMN_NAME = 'dimensions_json'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists > 0,
  'ALTER TABLE metric_points DROP COLUMN dimensions_json',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules' AND COLUMN_NAME = 'state'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists > 0,
  'ALTER TABLE alert_rules DROP COLUMN state',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules' AND COLUMN_NAME = 'dimensions_json'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists > 0,
  'ALTER TABLE alert_rules DROP COLUMN dimensions_json',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules' AND COLUMN_NAME = 'granularity_sec'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists > 0,
  'ALTER TABLE alert_rules DROP COLUMN granularity_sec',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'alert_rules' AND COLUMN_NAME = 'window_sec'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists > 0,
  'ALTER TABLE alert_rules DROP COLUMN window_sec',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
