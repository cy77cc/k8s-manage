-- +migrate Up
CREATE TABLE IF NOT EXISTS service_revisions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service_id BIGINT UNSIGNED NOT NULL,
  revision_no INT UNSIGNED NOT NULL DEFAULT 1,
  config_mode VARCHAR(16) NOT NULL,
  render_target VARCHAR(16) NOT NULL,
  standard_config_json LONGTEXT,
  custom_yaml LONGTEXT,
  variable_schema_json LONGTEXT,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_service_revisions_service (service_id),
  INDEX idx_service_revisions_created_at (created_at)
);

CREATE TABLE IF NOT EXISTS service_variable_sets (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service_id BIGINT UNSIGNED NOT NULL,
  env VARCHAR(32) NOT NULL,
  values_json LONGTEXT,
  secret_keys_json LONGTEXT,
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_service_env (service_id, env),
  INDEX idx_service_variable_sets_service (service_id)
);

CREATE TABLE IF NOT EXISTS service_deploy_targets (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service_id BIGINT UNSIGNED NOT NULL,
  cluster_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  namespace VARCHAR(128) NOT NULL DEFAULT 'default',
  deploy_target VARCHAR(16) NOT NULL DEFAULT 'k8s',
  policy_json LONGTEXT,
  is_default TINYINT(1) NOT NULL DEFAULT 1,
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_service_deploy_targets_service (service_id),
  INDEX idx_service_deploy_targets_cluster_ns (cluster_id, namespace)
);

CREATE TABLE IF NOT EXISTS service_release_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service_id BIGINT UNSIGNED NOT NULL,
  revision_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  cluster_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  namespace VARCHAR(128) NOT NULL DEFAULT 'default',
  env VARCHAR(32) NOT NULL DEFAULT 'staging',
  deploy_target VARCHAR(16) NOT NULL DEFAULT 'k8s',
  status VARCHAR(32) NOT NULL DEFAULT 'created',
  rendered_yaml LONGTEXT,
  variables_snapshot_json LONGTEXT,
  error LONGTEXT,
  operator BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_service_release_records_service (service_id),
  INDEX idx_service_release_records_env (env),
  INDEX idx_service_release_records_created_at (created_at)
);

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'last_revision_id');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN last_revision_id BIGINT UNSIGNED NOT NULL DEFAULT 0', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'default_target_id');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN default_target_id BIGINT UNSIGNED NOT NULL DEFAULT 0', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'template_engine_version');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN template_engine_version VARCHAR(16) NOT NULL DEFAULT ''v1''', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- +migrate Down
DROP TABLE IF EXISTS service_release_records;
DROP TABLE IF EXISTS service_deploy_targets;
DROP TABLE IF EXISTS service_variable_sets;
DROP TABLE IF EXISTS service_revisions;
