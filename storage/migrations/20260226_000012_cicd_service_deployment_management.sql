-- +migrate Up
CREATE TABLE IF NOT EXISTS cicd_service_ci_configs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service_id BIGINT UNSIGNED NOT NULL,
  repo_url VARCHAR(512) NOT NULL,
  branch VARCHAR(128) NOT NULL DEFAULT 'main',
  build_steps_json LONGTEXT,
  artifact_target VARCHAR(512) NOT NULL,
  trigger_mode VARCHAR(32) NOT NULL DEFAULT 'manual',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_cicd_ci_cfg_service (service_id)
);

CREATE TABLE IF NOT EXISTS cicd_service_ci_runs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service_id BIGINT UNSIGNED NOT NULL,
  ci_config_id BIGINT UNSIGNED NOT NULL,
  trigger_type VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'queued',
  reason VARCHAR(512) NOT NULL DEFAULT '',
  triggered_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  triggered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_cicd_ci_run_service (service_id),
  INDEX idx_cicd_ci_run_cfg (ci_config_id),
  INDEX idx_cicd_ci_run_triggered (triggered_at)
);

CREATE TABLE IF NOT EXISTS cicd_deployment_cd_configs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  deployment_id BIGINT UNSIGNED NOT NULL,
  env VARCHAR(32) NOT NULL,
  strategy VARCHAR(32) NOT NULL DEFAULT 'rolling',
  strategy_config_json LONGTEXT,
  approval_required TINYINT(1) NOT NULL DEFAULT 0,
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_cicd_deploy_env (deployment_id, env),
  INDEX idx_cicd_cd_cfg_deploy (deployment_id)
);

CREATE TABLE IF NOT EXISTS cicd_releases (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service_id BIGINT UNSIGNED NOT NULL,
  deployment_id BIGINT UNSIGNED NOT NULL,
  env VARCHAR(32) NOT NULL,
  version VARCHAR(128) NOT NULL,
  strategy VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending_approval',
  triggered_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  approved_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  approval_comment VARCHAR(1024) NOT NULL DEFAULT '',
  rollback_from_release_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  started_at TIMESTAMP NULL DEFAULT NULL,
  finished_at TIMESTAMP NULL DEFAULT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_cicd_release_service (service_id),
  INDEX idx_cicd_release_deploy (deployment_id),
  INDEX idx_cicd_release_status (status),
  INDEX idx_cicd_release_created (created_at)
);

CREATE TABLE IF NOT EXISTS cicd_release_approvals (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  release_id BIGINT UNSIGNED NOT NULL,
  approver_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  decision VARCHAR(32) NOT NULL,
  comment VARCHAR(1024) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_cicd_approval_release (release_id),
  INDEX idx_cicd_approval_created (created_at)
);

CREATE TABLE IF NOT EXISTS cicd_audit_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  deployment_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  release_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  event_type VARCHAR(64) NOT NULL,
  actor_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  payload_json LONGTEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_cicd_audit_service (service_id),
  INDEX idx_cicd_audit_deploy (deployment_id),
  INDEX idx_cicd_audit_release (release_id),
  INDEX idx_cicd_audit_type (event_type),
  INDEX idx_cicd_audit_created (created_at)
);

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'CI配置查看', 'cicd:ci:read', 3, 'cicd', 'ci:read', '查看服务CI配置', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'cicd:ci:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'CI配置管理', 'cicd:ci:write', 3, 'cicd', 'ci:write', '管理服务CI配置', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'cicd:ci:write');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'CI运行触发', 'cicd:ci:run', 3, 'cicd', 'ci:run', '触发服务CI运行', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'cicd:ci:run');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'CD配置查看', 'cicd:cd:read', 3, 'cicd', 'cd:read', '查看部署CD配置', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'cicd:cd:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'CD配置管理', 'cicd:cd:write', 3, 'cicd', 'cd:write', '管理部署CD配置', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'cicd:cd:write');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '发布触发', 'cicd:release:run', 3, 'cicd', 'release:run', '触发发布', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'cicd:release:run');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '发布审批', 'cicd:release:approve', 3, 'cicd', 'release:approve', '审批发布', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'cicd:release:approve');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '发布回滚', 'cicd:release:rollback', 3, 'cicd', 'release:rollback', '执行发布回滚', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'cicd:release:rollback');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '发布审计查看', 'cicd:audit:read', 3, 'cicd', 'audit:read', '查看CI/CD审计事件', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'cicd:audit:read');

-- +migrate Down
DROP TABLE IF EXISTS cicd_audit_events;
DROP TABLE IF EXISTS cicd_release_approvals;
DROP TABLE IF EXISTS cicd_releases;
DROP TABLE IF EXISTS cicd_deployment_cd_configs;
DROP TABLE IF EXISTS cicd_service_ci_runs;
DROP TABLE IF EXISTS cicd_service_ci_configs;

DELETE FROM role_permissions WHERE permission_id IN (
  SELECT id FROM permissions WHERE code IN (
    'cicd:ci:read', 'cicd:ci:write', 'cicd:ci:run',
    'cicd:cd:read', 'cicd:cd:write',
    'cicd:release:run', 'cicd:release:approve', 'cicd:release:rollback',
    'cicd:audit:read'
  )
);
DELETE FROM permissions WHERE code IN (
  'cicd:ci:read', 'cicd:ci:write', 'cicd:ci:run',
  'cicd:cd:read', 'cicd:cd:write',
  'cicd:release:run', 'cicd:release:approve', 'cicd:release:rollback',
  'cicd:audit:read'
);
