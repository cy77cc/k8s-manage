-- +migrate Up
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'preview_context_hash'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE deployment_releases ADD COLUMN preview_context_hash VARCHAR(128) NOT NULL DEFAULT '''' AFTER target_revision',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'preview_token_hash'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE deployment_releases ADD COLUMN preview_token_hash VARCHAR(128) NOT NULL DEFAULT '''' AFTER preview_context_hash',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'preview_expires_at'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE deployment_releases ADD COLUMN preview_expires_at TIMESTAMP NULL DEFAULT NULL AFTER preview_token_hash',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND INDEX_NAME = 'idx_deploy_release_preview_expires'
);
SET @sql := IF(@idx_exists = 0,
  'CREATE INDEX idx_deploy_release_preview_expires ON deployment_releases (preview_expires_at)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_release_audits' AND COLUMN_NAME = 'correlation_id'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE deployment_release_audits ADD COLUMN correlation_id VARCHAR(96) NOT NULL DEFAULT '''' AFTER release_id',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_release_audits' AND COLUMN_NAME = 'trace_id'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE deployment_release_audits ADD COLUMN trace_id VARCHAR(96) NOT NULL DEFAULT '''' AFTER correlation_id',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_release_audits' AND INDEX_NAME = 'idx_deploy_release_audit_correlation'
);
SET @sql := IF(@idx_exists = 0,
  'CREATE INDEX idx_deploy_release_audit_correlation ON deployment_release_audits (correlation_id)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_release_audits' AND INDEX_NAME = 'idx_deploy_release_audit_trace'
);
SET @sql := IF(@idx_exists = 0,
  'CREATE INDEX idx_deploy_release_audit_trace ON deployment_release_audits (trace_id)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- +migrate Down
SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_release_audits' AND INDEX_NAME = 'idx_deploy_release_audit_trace'
);
SET @sql := IF(@idx_exists > 0,
  'DROP INDEX idx_deploy_release_audit_trace ON deployment_release_audits',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_release_audits' AND INDEX_NAME = 'idx_deploy_release_audit_correlation'
);
SET @sql := IF(@idx_exists > 0,
  'DROP INDEX idx_deploy_release_audit_correlation ON deployment_release_audits',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND INDEX_NAME = 'idx_deploy_release_preview_expires'
);
SET @sql := IF(@idx_exists > 0,
  'DROP INDEX idx_deploy_release_preview_expires ON deployment_releases',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_release_audits' AND COLUMN_NAME = 'trace_id'
);
SET @sql := IF(@col_exists > 0,
  'ALTER TABLE deployment_release_audits DROP COLUMN trace_id',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_release_audits' AND COLUMN_NAME = 'correlation_id'
);
SET @sql := IF(@col_exists > 0,
  'ALTER TABLE deployment_release_audits DROP COLUMN correlation_id',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'preview_expires_at'
);
SET @sql := IF(@col_exists > 0,
  'ALTER TABLE deployment_releases DROP COLUMN preview_expires_at',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'preview_token_hash'
);
SET @sql := IF(@col_exists > 0,
  'ALTER TABLE deployment_releases DROP COLUMN preview_token_hash',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'preview_context_hash'
);
SET @sql := IF(@col_exists > 0,
  'ALTER TABLE deployment_releases DROP COLUMN preview_context_hash',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
