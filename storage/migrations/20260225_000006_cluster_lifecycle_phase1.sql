-- +migrate Up
CREATE TABLE IF NOT EXISTS cluster_namespace_bindings (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  cluster_id BIGINT UNSIGNED NOT NULL,
  team_id BIGINT UNSIGNED NOT NULL,
  namespace VARCHAR(128) NOT NULL,
  env VARCHAR(32) NOT NULL DEFAULT '',
  readonly TINYINT(1) NOT NULL DEFAULT 0,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_cluster_team_namespace (cluster_id, team_id, namespace),
  KEY idx_cluster_ns_cluster (cluster_id),
  KEY idx_cluster_ns_team (team_id)
);

CREATE TABLE IF NOT EXISTS cluster_release_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  cluster_id BIGINT UNSIGNED NOT NULL,
  namespace VARCHAR(128) NOT NULL,
  app VARCHAR(128) NOT NULL,
  strategy VARCHAR(32) NOT NULL DEFAULT 'rolling',
  rollout_name VARCHAR(128) NOT NULL DEFAULT '',
  revision INT NOT NULL DEFAULT 1,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  operator VARCHAR(64) NOT NULL DEFAULT '',
  payload_json LONGTEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_cluster_release_cluster (cluster_id),
  KEY idx_cluster_release_ns (namespace)
);

CREATE TABLE IF NOT EXISTS cluster_hpa_policies (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  cluster_id BIGINT UNSIGNED NOT NULL,
  namespace VARCHAR(128) NOT NULL,
  name VARCHAR(128) NOT NULL,
  target_ref_kind VARCHAR(64) NOT NULL,
  target_ref_name VARCHAR(128) NOT NULL,
  min_replicas INT NOT NULL DEFAULT 1,
  max_replicas INT NOT NULL DEFAULT 1,
  cpu_utilization INT NULL,
  memory_utilization INT NULL,
  raw_policy_json LONGTEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_cluster_hpa (cluster_id, namespace, name),
  KEY idx_cluster_hpa_cluster (cluster_id)
);

CREATE TABLE IF NOT EXISTS cluster_quota_policies (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  cluster_id BIGINT UNSIGNED NOT NULL,
  namespace VARCHAR(128) NOT NULL,
  name VARCHAR(128) NOT NULL,
  type VARCHAR(32) NOT NULL DEFAULT 'resourcequota',
  spec_json LONGTEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_cluster_quota (cluster_id, namespace, name, type),
  KEY idx_cluster_quota_cluster (cluster_id)
);

CREATE TABLE IF NOT EXISTS cluster_deploy_approvals (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  ticket VARCHAR(96) NOT NULL,
  cluster_id BIGINT UNSIGNED NOT NULL,
  namespace VARCHAR(128) NOT NULL,
  action VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  request_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  review_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  expires_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_cluster_approval_ticket (ticket),
  KEY idx_cluster_approval_cluster (cluster_id),
  KEY idx_cluster_approval_status (status)
);

CREATE TABLE IF NOT EXISTS cluster_operation_audits (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  cluster_id BIGINT UNSIGNED NOT NULL,
  namespace VARCHAR(128) NOT NULL DEFAULT '',
  action VARCHAR(64) NOT NULL,
  resource VARCHAR(64) NOT NULL DEFAULT '',
  resource_id VARCHAR(128) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT 'success',
  message VARCHAR(255) NOT NULL DEFAULT '',
  operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_cluster_audit_cluster (cluster_id),
  KEY idx_cluster_audit_action (action)
);

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'K8s只读', 'k8s:read', 3, 'k8s', 'read', '查看集群与资源', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'k8s:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'K8s写入', 'k8s:write', 3, 'k8s', 'write', '管理命名空间与发布配置', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'k8s:write');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'K8s部署', 'k8s:deploy', 3, 'k8s', 'deploy', '部署发布', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'k8s:deploy');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'K8s回滚', 'k8s:rollback', 3, 'k8s', 'rollback', '回滚发布', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'k8s:rollback');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'K8s弹性策略', 'k8s:hpa', 3, 'k8s', 'hpa', '管理HPA策略', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'k8s:hpa');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'K8s配额策略', 'k8s:quota', 3, 'k8s', 'quota', '管理Quota/LimitRange', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'k8s:quota');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'K8s审批', 'k8s:approve', 3, 'k8s', 'approve', '生产环境部署审批', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'k8s:approve');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'K8s命名空间绑定', 'k8s:namespace:bind', 3, 'k8s', 'namespace:bind', '管理team与namespace绑定', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'k8s:namespace:bind');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p ON p.code IN ('k8s:read')
WHERE r.code = 'viewer'
AND NOT EXISTS (SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission_id = p.id);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p ON p.code IN ('k8s:read', 'k8s:write', 'k8s:deploy', 'k8s:hpa', 'k8s:quota')
WHERE r.code = 'operator'
AND NOT EXISTS (SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission_id = p.id);

-- +migrate Down
DROP TABLE IF EXISTS cluster_operation_audits;
DROP TABLE IF EXISTS cluster_deploy_approvals;
DROP TABLE IF EXISTS cluster_quota_policies;
DROP TABLE IF EXISTS cluster_hpa_policies;
DROP TABLE IF EXISTS cluster_release_records;
DROP TABLE IF EXISTS cluster_namespace_bindings;

DELETE FROM role_permissions WHERE permission_id IN (
  SELECT id FROM permissions WHERE code IN ('k8s:read', 'k8s:write', 'k8s:deploy', 'k8s:rollback', 'k8s:hpa', 'k8s:quota', 'k8s:approve', 'k8s:namespace:bind')
);
DELETE FROM permissions WHERE code IN ('k8s:read', 'k8s:write', 'k8s:deploy', 'k8s:rollback', 'k8s:hpa', 'k8s:quota', 'k8s:approve', 'k8s:namespace:bind');
