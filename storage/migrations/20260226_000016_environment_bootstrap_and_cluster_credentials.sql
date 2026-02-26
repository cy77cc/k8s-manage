-- +migrate Up
CREATE TABLE IF NOT EXISTS environment_install_jobs (
  id VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  runtime_type VARCHAR(16) NOT NULL,
  target_env VARCHAR(32) NOT NULL DEFAULT 'staging',
  target_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  cluster_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL DEFAULT 'queued',
  package_version VARCHAR(64) NOT NULL DEFAULT '',
  package_path VARCHAR(512) NOT NULL DEFAULT '',
  package_checksum VARCHAR(128) NOT NULL DEFAULT '',
  started_at TIMESTAMP NULL DEFAULT NULL,
  finished_at TIMESTAMP NULL DEFAULT NULL,
  error_message TEXT,
  result_json LONGTEXT,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX idx_env_install_status (status),
  INDEX idx_env_install_runtime (runtime_type),
  INDEX idx_env_install_target (target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS environment_install_job_steps (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  job_id VARCHAR(64) NOT NULL,
  step_name VARCHAR(64) NOT NULL,
  phase VARCHAR(32) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT 'queued',
  host_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  output TEXT,
  error_message TEXT,
  started_at TIMESTAMP NULL DEFAULT NULL,
  finished_at TIMESTAMP NULL DEFAULT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX idx_env_step_job (job_id),
  INDEX idx_env_step_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS cluster_credentials (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(128) NOT NULL,
  runtime_type VARCHAR(16) NOT NULL DEFAULT 'k8s',
  source VARCHAR(32) NOT NULL,
  cluster_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  endpoint VARCHAR(256) NOT NULL DEFAULT '',
  auth_method VARCHAR(32) NOT NULL DEFAULT 'kubeconfig',
  kubeconfig_enc LONGTEXT,
  ca_cert_enc LONGTEXT,
  cert_enc LONGTEXT,
  key_enc LONGTEXT,
  token_enc LONGTEXT,
  metadata_json LONGTEXT,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  last_test_at TIMESTAMP NULL DEFAULT NULL,
  last_test_status VARCHAR(32) NOT NULL DEFAULT '',
  last_test_message VARCHAR(512) NOT NULL DEFAULT '',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX idx_cluster_credential_source (source),
  INDEX idx_cluster_credential_cluster (cluster_id),
  INDEX idx_cluster_credential_runtime (runtime_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets' AND COLUMN_NAME = 'cluster_source'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists = 0,
  'ALTER TABLE deployment_targets ADD COLUMN cluster_source VARCHAR(32) NOT NULL DEFAULT ''platform_managed'' AFTER cluster_id',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets' AND COLUMN_NAME = 'credential_id'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists = 0,
  'ALTER TABLE deployment_targets ADD COLUMN credential_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER cluster_source',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets' AND COLUMN_NAME = 'bootstrap_job_id'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists = 0,
  'ALTER TABLE deployment_targets ADD COLUMN bootstrap_job_id VARCHAR(64) NOT NULL DEFAULT '''' AFTER credential_id',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets' AND COLUMN_NAME = 'readiness_status'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists = 0,
  'ALTER TABLE deployment_targets ADD COLUMN readiness_status VARCHAR(32) NOT NULL DEFAULT ''unknown'' AFTER status',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets' AND INDEX_NAME = 'idx_deploy_target_credential'
);
SET @sql := IF(@idx_exists = 0,
  'CREATE INDEX idx_deploy_target_credential ON deployment_targets (credential_id)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets' AND INDEX_NAME = 'idx_deploy_target_cluster_source'
);
SET @sql := IF(@idx_exists = 0,
  'CREATE INDEX idx_deploy_target_cluster_source ON deployment_targets (cluster_source)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '部署凭据读取', 'deploy:credential:read', 3, 'deploy', 'credential:read', '读取部署凭据元数据和连通性结果', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'deploy:credential:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '部署凭据写入', 'deploy:credential:write', 3, 'deploy', 'credential:write', '创建导入部署凭据和注册平台凭据', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'deploy:credential:write');

-- +migrate Down
DELETE FROM role_permissions WHERE permission_id IN (
  SELECT id FROM permissions WHERE code IN ('deploy:credential:read', 'deploy:credential:write')
);
DELETE FROM permissions WHERE code IN ('deploy:credential:read', 'deploy:credential:write');

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets' AND INDEX_NAME = 'idx_deploy_target_cluster_source'
);
SET @sql := IF(@idx_exists > 0,
  'DROP INDEX idx_deploy_target_cluster_source ON deployment_targets',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets' AND INDEX_NAME = 'idx_deploy_target_credential'
);
SET @sql := IF(@idx_exists > 0,
  'DROP INDEX idx_deploy_target_credential ON deployment_targets',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets' AND COLUMN_NAME = 'readiness_status'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists > 0,
  'ALTER TABLE deployment_targets DROP COLUMN readiness_status',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets' AND COLUMN_NAME = 'bootstrap_job_id'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists > 0,
  'ALTER TABLE deployment_targets DROP COLUMN bootstrap_job_id',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets' AND COLUMN_NAME = 'credential_id'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists > 0,
  'ALTER TABLE deployment_targets DROP COLUMN credential_id',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tbl_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets'
);
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets' AND COLUMN_NAME = 'cluster_source'
);
SET @sql := IF(@tbl_exists = 1 AND @col_exists > 0,
  'ALTER TABLE deployment_targets DROP COLUMN cluster_source',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

DROP TABLE IF EXISTS environment_install_job_steps;
DROP TABLE IF EXISTS environment_install_jobs;
DROP TABLE IF EXISTS cluster_credentials;
