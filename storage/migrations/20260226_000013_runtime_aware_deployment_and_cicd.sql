-- +migrate Up
SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets' AND COLUMN_NAME = 'runtime_type');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE deployment_targets ADD COLUMN runtime_type VARCHAR(16) NOT NULL DEFAULT ''k8s'' AFTER target_type', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE deployment_targets SET runtime_type = target_type WHERE runtime_type = '' OR runtime_type IS NULL;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets' AND INDEX_NAME = 'idx_deploy_target_runtime'
);
SET @sql := IF(@idx_exists = 0, 'CREATE INDEX idx_deploy_target_runtime ON deployment_targets (runtime_type)', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'source_release_id');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE deployment_releases ADD COLUMN source_release_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER revision_id', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'target_revision');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE deployment_releases ADD COLUMN target_revision VARCHAR(128) NOT NULL DEFAULT '''' AFTER source_release_id', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'runtime_context_json');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE deployment_releases ADD COLUMN runtime_context_json LONGTEXT AFTER manifest_snapshot', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'diagnostics_json');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE deployment_releases ADD COLUMN diagnostics_json LONGTEXT AFTER warnings_json', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'verification_json');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE deployment_releases ADD COLUMN verification_json LONGTEXT AFTER diagnostics_json', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'updated_at');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE deployment_releases ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER created_at', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND INDEX_NAME = 'idx_deploy_release_runtime'
);
SET @sql := IF(@idx_exists = 0, 'CREATE INDEX idx_deploy_release_runtime ON deployment_releases (runtime_type, status)', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND INDEX_NAME = 'idx_deploy_release_service_target_runtime'
);
SET @sql := IF(@idx_exists = 0, 'CREATE INDEX idx_deploy_release_service_target_runtime ON deployment_releases (service_id, target_id, runtime_type)', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_deployment_cd_configs' AND COLUMN_NAME = 'runtime_type');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE cicd_deployment_cd_configs ADD COLUMN runtime_type VARCHAR(16) NOT NULL DEFAULT ''k8s'' AFTER env', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_deployment_cd_configs' AND INDEX_NAME = 'uk_cicd_deploy_env_runtime'
);
SET @sql := IF(@idx_exists = 0, 'ALTER TABLE cicd_deployment_cd_configs ADD UNIQUE KEY uk_cicd_deploy_env_runtime (deployment_id, env, runtime_type)', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_releases' AND COLUMN_NAME = 'runtime_type');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE cicd_releases ADD COLUMN runtime_type VARCHAR(16) NOT NULL DEFAULT ''k8s'' AFTER env', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_releases' AND COLUMN_NAME = 'diagnostics_json');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE cicd_releases ADD COLUMN diagnostics_json LONGTEXT AFTER rollback_from_release_id', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_releases' AND INDEX_NAME = 'idx_cicd_release_runtime'
);
SET @sql := IF(@idx_exists = 0, 'CREATE INDEX idx_cicd_release_runtime ON cicd_releases (runtime_type, status)', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'K8s发布读取', 'deploy:k8s:read', 3, 'deploy', 'k8s:read', '读取K8s发布记录与诊断', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'deploy:k8s:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'K8s发布执行', 'deploy:k8s:apply', 3, 'deploy', 'k8s:apply', '执行K8s发布', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'deploy:k8s:apply');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'K8s发布回滚', 'deploy:k8s:rollback', 3, 'deploy', 'k8s:rollback', '执行K8s发布回滚', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'deploy:k8s:rollback');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'K8s发布审批', 'deploy:k8s:approve', 3, 'deploy', 'k8s:approve', '审批K8s发布', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'deploy:k8s:approve');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'Compose发布读取', 'deploy:compose:read', 3, 'deploy', 'compose:read', '读取Compose发布记录与诊断', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'deploy:compose:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'Compose发布执行', 'deploy:compose:apply', 3, 'deploy', 'compose:apply', '执行Compose发布', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'deploy:compose:apply');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'Compose发布回滚', 'deploy:compose:rollback', 3, 'deploy', 'compose:rollback', '执行Compose发布回滚', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'deploy:compose:rollback');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'Compose发布审批', 'deploy:compose:approve', 3, 'deploy', 'compose:approve', '审批Compose发布', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'deploy:compose:approve');

-- +migrate Down
SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets' AND INDEX_NAME = 'idx_deploy_target_runtime'
);
SET @sql := IF(@idx_exists > 0, 'DROP INDEX idx_deploy_target_runtime ON deployment_targets', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND INDEX_NAME = 'idx_deploy_release_runtime'
);
SET @sql := IF(@idx_exists > 0, 'DROP INDEX idx_deploy_release_runtime ON deployment_releases', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND INDEX_NAME = 'idx_deploy_release_service_target_runtime'
);
SET @sql := IF(@idx_exists > 0, 'DROP INDEX idx_deploy_release_service_target_runtime ON deployment_releases', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_deployment_cd_configs' AND INDEX_NAME = 'uk_cicd_deploy_env_runtime'
);
SET @sql := IF(@idx_exists > 0, 'ALTER TABLE cicd_deployment_cd_configs DROP INDEX uk_cicd_deploy_env_runtime', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_releases' AND INDEX_NAME = 'idx_cicd_release_runtime'
);
SET @sql := IF(@idx_exists > 0, 'DROP INDEX idx_cicd_release_runtime ON cicd_releases', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_targets' AND COLUMN_NAME = 'runtime_type');
SET @sql := IF(@col_exists > 0, 'ALTER TABLE deployment_targets DROP COLUMN runtime_type', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'source_release_id');
SET @sql := IF(@col_exists > 0, 'ALTER TABLE deployment_releases DROP COLUMN source_release_id', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'target_revision');
SET @sql := IF(@col_exists > 0, 'ALTER TABLE deployment_releases DROP COLUMN target_revision', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'runtime_context_json');
SET @sql := IF(@col_exists > 0, 'ALTER TABLE deployment_releases DROP COLUMN runtime_context_json', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'diagnostics_json');
SET @sql := IF(@col_exists > 0, 'ALTER TABLE deployment_releases DROP COLUMN diagnostics_json', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'verification_json');
SET @sql := IF(@col_exists > 0, 'ALTER TABLE deployment_releases DROP COLUMN verification_json', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'deployment_releases' AND COLUMN_NAME = 'updated_at');
SET @sql := IF(@col_exists > 0, 'ALTER TABLE deployment_releases DROP COLUMN updated_at', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_deployment_cd_configs' AND COLUMN_NAME = 'runtime_type');
SET @sql := IF(@col_exists > 0, 'ALTER TABLE cicd_deployment_cd_configs DROP COLUMN runtime_type', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_releases' AND COLUMN_NAME = 'runtime_type');
SET @sql := IF(@col_exists > 0, 'ALTER TABLE cicd_releases DROP COLUMN runtime_type', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_releases' AND COLUMN_NAME = 'diagnostics_json');
SET @sql := IF(@col_exists > 0, 'ALTER TABLE cicd_releases DROP COLUMN diagnostics_json', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

DELETE FROM role_permissions WHERE permission_id IN (
  SELECT id FROM permissions WHERE code IN (
    'deploy:k8s:read', 'deploy:k8s:apply', 'deploy:k8s:rollback', 'deploy:k8s:approve',
    'deploy:compose:read', 'deploy:compose:apply', 'deploy:compose:rollback', 'deploy:compose:approve'
  )
);
DELETE FROM permissions WHERE code IN (
  'deploy:k8s:read', 'deploy:k8s:apply', 'deploy:k8s:rollback', 'deploy:k8s:approve',
  'deploy:compose:read', 'deploy:compose:apply', 'deploy:compose:rollback', 'deploy:compose:approve'
);
