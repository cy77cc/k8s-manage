-- +migrate Up
SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases'
);

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'trigger_source'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists = 0,
  'ALTER TABLE deployment_releases ADD COLUMN trigger_source VARCHAR(32) NOT NULL DEFAULT ''manual'' AFTER strategy',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'trigger_context_json'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists = 0,
  'ALTER TABLE deployment_releases ADD COLUMN trigger_context_json LONGTEXT AFTER runtime_context_json',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'ci_run_id'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists = 0,
  'ALTER TABLE deployment_releases ADD COLUMN ci_run_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER operator',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND INDEX_NAME = 'idx_deploy_release_trigger_source'
);
SET @sql := IF(@tbl_exists = 1 AND @idx_exists = 0,
  'CREATE INDEX idx_deploy_release_trigger_source ON deployment_releases (trigger_source, status)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND INDEX_NAME = 'idx_deploy_release_ci_run'
);
SET @sql := IF(@tbl_exists = 1 AND @idx_exists = 0,
  'CREATE INDEX idx_deploy_release_ci_run ON deployment_releases (ci_run_id)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- +migrate Down
SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases'
);

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND INDEX_NAME = 'idx_deploy_release_ci_run'
);
SET @sql := IF(@tbl_exists = 1 AND @idx_exists > 0,
  'DROP INDEX idx_deploy_release_ci_run ON deployment_releases',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND INDEX_NAME = 'idx_deploy_release_trigger_source'
);
SET @sql := IF(@tbl_exists = 1 AND @idx_exists > 0,
  'DROP INDEX idx_deploy_release_trigger_source ON deployment_releases',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'ci_run_id'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists > 0,
  'ALTER TABLE deployment_releases DROP COLUMN ci_run_id',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'trigger_context_json'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists > 0,
  'ALTER TABLE deployment_releases DROP COLUMN trigger_context_json',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'trigger_source'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists > 0,
  'ALTER TABLE deployment_releases DROP COLUMN trigger_source',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
