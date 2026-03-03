-- +migrate Up
-- Add env_type column to clusters table for deployment environment constraint
SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'clusters'
);

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'clusters' AND COLUMN_NAME = 'env_type'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists = 0,
  'ALTER TABLE clusters ADD COLUMN env_type VARCHAR(32) NOT NULL DEFAULT ''development'' AFTER source',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'clusters' AND INDEX_NAME = 'idx_cluster_env_type'
);
SET @sql := IF(@tbl_exists = 1 AND @idx_exists = 0,
  'CREATE INDEX idx_cluster_env_type ON clusters (env_type)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- +migrate Down
SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'clusters' AND INDEX_NAME = 'idx_cluster_env_type'
);
SET @sql := IF(@idx_exists > 0,
  'DROP INDEX idx_cluster_env_type ON clusters',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'clusters'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'clusters' AND COLUMN_NAME = 'env_type'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists > 0,
  'ALTER TABLE clusters DROP COLUMN env_type',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
