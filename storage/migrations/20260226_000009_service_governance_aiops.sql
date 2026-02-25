-- +migrate Up
CREATE TABLE IF NOT EXISTS service_governance_policies (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service_id BIGINT UNSIGNED NOT NULL,
  env VARCHAR(32) NOT NULL,
  traffic_policy_json LONGTEXT,
  resilience_policy_json LONGTEXT,
  access_policy_json LONGTEXT,
  slo_policy_json LONGTEXT,
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_service_governance_env (service_id, env)
);

CREATE TABLE IF NOT EXISTS aiops_inspections (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  release_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  target_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  service_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  stage VARCHAR(16) NOT NULL,
  summary TEXT,
  findings_json LONGTEXT,
  suggestions_json LONGTEXT,
  status VARCHAR(32) NOT NULL DEFAULT 'done',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_aiops_release (release_id),
  INDEX idx_aiops_target (target_id),
  INDEX idx_aiops_service (service_id),
  INDEX idx_aiops_created_at (created_at)
);

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '服务治理查看', 'service:governance:read', 3, 'service', 'governance:read', '查看服务治理策略', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'service:governance:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '服务治理管理', 'service:governance:write', 3, 'service', 'governance:write', '管理服务治理策略', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'service:governance:write');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'AIOPS查看', 'aiops:read', 3, 'aiops', 'read', '查看AIOPS巡检', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'aiops:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'AIOPS执行', 'aiops:run', 3, 'aiops', 'run', '触发AIOPS巡检', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'aiops:run');

-- +migrate Down
DROP TABLE IF EXISTS aiops_inspections;
DROP TABLE IF EXISTS service_governance_policies;

DELETE FROM role_permissions WHERE permission_id IN (
  SELECT id FROM permissions WHERE code IN ('service:governance:read', 'service:governance:write', 'aiops:read', 'aiops:run')
);
DELETE FROM permissions WHERE code IN ('service:governance:read', 'service:governance:write', 'aiops:read', 'aiops:run');
