-- +migrate Up
CREATE TABLE IF NOT EXISTS deployment_targets (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  target_type VARCHAR(16) NOT NULL,
  cluster_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  project_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  team_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  env VARCHAR(32) NOT NULL DEFAULT 'staging',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_deploy_target_type (target_type),
  INDEX idx_deploy_target_cluster (cluster_id),
  INDEX idx_deploy_target_project_team (project_id, team_id)
);

CREATE TABLE IF NOT EXISTS deployment_target_nodes (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  target_id BIGINT UNSIGNED NOT NULL,
  host_id BIGINT UNSIGNED NOT NULL,
  role VARCHAR(16) NOT NULL DEFAULT 'worker',
  weight INT NOT NULL DEFAULT 100,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_target_host (target_id, host_id),
  INDEX idx_target_nodes_target (target_id),
  INDEX idx_target_nodes_host (host_id)
);

CREATE TABLE IF NOT EXISTS deployment_releases (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service_id BIGINT UNSIGNED NOT NULL,
  target_id BIGINT UNSIGNED NOT NULL,
  namespace_or_project VARCHAR(128) NOT NULL DEFAULT '',
  runtime_type VARCHAR(16) NOT NULL,
  strategy VARCHAR(16) NOT NULL DEFAULT 'rolling',
  revision_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL DEFAULT 'created',
  manifest_snapshot LONGTEXT,
  checks_json LONGTEXT,
  warnings_json LONGTEXT,
  operator BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_deploy_release_service (service_id),
  INDEX idx_deploy_release_target (target_id),
  INDEX idx_deploy_release_created_at (created_at)
);

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'default_deployment_target_id');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN default_deployment_target_id BIGINT UNSIGNED NOT NULL DEFAULT 0', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'runtime_strategy_json');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN runtime_strategy_json LONGTEXT', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'clusters' AND COLUMN_NAME = 'management_mode');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE clusters ADD COLUMN management_mode VARCHAR(32) NOT NULL DEFAULT ''k8s-only''', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '部署目标查看', 'deploy:target:read', 3, 'deploy', 'target:read', '查看部署目标', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'deploy:target:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '部署目标管理', 'deploy:target:write', 3, 'deploy', 'target:write', '管理部署目标', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'deploy:target:write');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '发布查看', 'deploy:release:read', 3, 'deploy', 'release:read', '查看发布记录', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'deploy:release:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '发布执行', 'deploy:release:apply', 3, 'deploy', 'release:apply', '执行发布', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'deploy:release:apply');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '发布回滚', 'deploy:release:rollback', 3, 'deploy', 'release:rollback', '发布回滚', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'deploy:release:rollback');

-- +migrate Down
DROP TABLE IF EXISTS deployment_releases;
DROP TABLE IF EXISTS deployment_target_nodes;
DROP TABLE IF EXISTS deployment_targets;

DELETE FROM role_permissions WHERE permission_id IN (
  SELECT id FROM permissions WHERE code IN (
    'deploy:target:read', 'deploy:target:write',
    'deploy:release:read', 'deploy:release:apply', 'deploy:release:rollback'
  )
);
DELETE FROM permissions WHERE code IN (
  'deploy:target:read', 'deploy:target:write',
  'deploy:release:read', 'deploy:release:apply', 'deploy:release:rollback'
);
