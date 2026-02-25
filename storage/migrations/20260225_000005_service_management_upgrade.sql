-- +migrate Up
CREATE TABLE IF NOT EXISTS services (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  project_id BIGINT UNSIGNED NOT NULL,
  name VARCHAR(64) NOT NULL,
  type VARCHAR(32) NOT NULL,
  image VARCHAR(256) NOT NULL,
  replicas INT NOT NULL DEFAULT 1,
  service_port INT NOT NULL DEFAULT 0,
  container_port INT NOT NULL DEFAULT 0,
  node_port INT NOT NULL DEFAULT 0,
  env_vars LONGTEXT,
  resources LONGTEXT,
  yaml_content LONGTEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_services_project_id (project_id),
  INDEX idx_services_name (name),
  INDEX idx_services_type (type)
);

CREATE TABLE IF NOT EXISTS service_helm_releases (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service_id BIGINT UNSIGNED NOT NULL,
  chart_name VARCHAR(128) NOT NULL,
  chart_version VARCHAR(64) NOT NULL DEFAULT '',
  chart_ref VARCHAR(512) NOT NULL DEFAULT '',
  values_yaml LONGTEXT,
  rendered_yaml LONGTEXT,
  status VARCHAR(32) NOT NULL DEFAULT 'imported',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_service_helm_service (service_id),
  INDEX idx_service_helm_status (status)
);

CREATE TABLE IF NOT EXISTS service_render_snapshots (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service_id BIGINT UNSIGNED NOT NULL,
  target VARCHAR(16) NOT NULL,
  mode VARCHAR(16) NOT NULL,
  rendered_yaml LONGTEXT,
  diagnostics_json LONGTEXT,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_service_render_service (service_id),
  INDEX idx_service_render_target (target)
);

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'team_id');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN team_id BIGINT UNSIGNED NOT NULL DEFAULT 0', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'owner_user_id');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN owner_user_id BIGINT UNSIGNED NOT NULL DEFAULT 0', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'owner');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN owner VARCHAR(64) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'env');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN env VARCHAR(32) NOT NULL DEFAULT ''staging''', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'runtime_type');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN runtime_type VARCHAR(16) NOT NULL DEFAULT ''k8s''', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'config_mode');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN config_mode VARCHAR(16) NOT NULL DEFAULT ''standard''', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'service_kind');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN service_kind VARCHAR(32) NOT NULL DEFAULT ''web''', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'render_target');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN render_target VARCHAR(16) NOT NULL DEFAULT ''k8s''', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'labels_json');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN labels_json LONGTEXT', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'standard_config_json');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN standard_config_json LONGTEXT', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'custom_yaml');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN custom_yaml LONGTEXT', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'source_template_version');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN source_template_version VARCHAR(32) NOT NULL DEFAULT ''v1''', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND COLUMN_NAME = 'status');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE services ADD COLUMN status VARCHAR(32) NOT NULL DEFAULT ''draft''', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'services' AND INDEX_NAME = 'idx_services_project_team_env_runtime'
);
SET @sql := IF(@idx_exists = 0, 'CREATE INDEX idx_services_project_team_env_runtime ON services (project_id, team_id, env, runtime_type)', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '服务只读', 'service:read', 3, 'service', 'read', '查看服务', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'service:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '服务写入', 'service:write', 3, 'service', 'write', '创建/编辑服务', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'service:write');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '服务发布', 'service:deploy', 3, 'service', 'deploy', '服务部署', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'service:deploy');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '服务审批', 'service:approve', 3, 'service', 'approve', '生产部署审批', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'service:approve');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON p.code IN ('service:read')
WHERE r.code = 'viewer'
AND NOT EXISTS (
  SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission_id = p.id
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON p.code IN ('service:read', 'service:write', 'service:deploy')
WHERE r.code = 'operator'
AND NOT EXISTS (
  SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission_id = p.id
);

-- +migrate Down
DROP TABLE IF EXISTS service_render_snapshots;
DROP TABLE IF EXISTS service_helm_releases;

DELETE FROM role_permissions WHERE permission_id IN (
  SELECT id FROM permissions WHERE code IN ('service:read', 'service:write', 'service:deploy', 'service:approve')
);
DELETE FROM permissions WHERE code IN ('service:read', 'service:write', 'service:deploy', 'service:approve');
