-- +migrate Up
CREATE TABLE IF NOT EXISTS cmdb_cis (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  ci_uid VARCHAR(160) NOT NULL,
  ci_type VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  source VARCHAR(64) NOT NULL DEFAULT 'manual',
  external_id VARCHAR(160) NOT NULL DEFAULT '',
  project_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  team_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  owner VARCHAR(128) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  tags_json LONGTEXT,
  attrs_json LONGTEXT,
  last_synced_at TIMESTAMP NULL,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  UNIQUE KEY uk_cmdb_ci_uid (ci_uid),
  KEY idx_cmdb_ci_type (ci_type),
  KEY idx_cmdb_status (status),
  KEY idx_cmdb_external_id (external_id),
  KEY idx_cmdb_project_team (project_id, team_id),
  KEY idx_cmdb_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS cmdb_relations (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  from_ci_id BIGINT UNSIGNED NOT NULL,
  to_ci_id BIGINT UNSIGNED NOT NULL,
  relation_type VARCHAR(64) NOT NULL,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_cmdb_relation (from_ci_id, to_ci_id, relation_type),
  KEY idx_cmdb_relation_from (from_ci_id),
  KEY idx_cmdb_relation_to (to_ci_id)
);

CREATE TABLE IF NOT EXISTS cmdb_sync_jobs (
  id VARCHAR(64) PRIMARY KEY,
  source VARCHAR(64) NOT NULL DEFAULT 'all',
  status VARCHAR(32) NOT NULL DEFAULT 'running',
  summary_json LONGTEXT,
  error_message TEXT,
  started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  finished_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_cmdb_sync_status (status),
  KEY idx_cmdb_sync_operator (operator_id)
);

CREATE TABLE IF NOT EXISTS cmdb_sync_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  job_id VARCHAR(64) NOT NULL,
  ci_uid VARCHAR(160) NOT NULL,
  action VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  diff_json LONGTEXT,
  error_message TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_cmdb_sync_record_job (job_id),
  KEY idx_cmdb_sync_record_ci (ci_uid),
  KEY idx_cmdb_sync_record_status (status)
);

CREATE TABLE IF NOT EXISTS cmdb_audits (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  ci_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  relation_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  action VARCHAR(64) NOT NULL,
  actor_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  before_json LONGTEXT,
  after_json LONGTEXT,
  detail TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_cmdb_audit_ci (ci_id),
  KEY idx_cmdb_audit_relation (relation_id),
  KEY idx_cmdb_audit_actor (actor_id),
  KEY idx_cmdb_audit_action (action),
  KEY idx_cmdb_audit_created (created_at)
);

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'CMDB查看', 'cmdb:read', 3, 'cmdb', 'read', '查看CMDB资产和拓扑', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'cmdb:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'CMDB管理', 'cmdb:write', 3, 'cmdb', 'write', '管理CMDB资产和关系', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'cmdb:write');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'CMDB同步', 'cmdb:sync', 3, 'cmdb', 'sync', '触发和重试CMDB同步任务', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'cmdb:sync');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'CMDB审计', 'cmdb:audit', 3, 'cmdb', 'audit', '查询CMDB审计与变更记录', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'cmdb:audit');

-- +migrate Down
DELETE FROM role_permissions WHERE permission_id IN (
  SELECT id FROM permissions WHERE code IN ('cmdb:read', 'cmdb:write', 'cmdb:sync', 'cmdb:audit')
);
DELETE FROM permissions WHERE code IN ('cmdb:read', 'cmdb:write', 'cmdb:sync', 'cmdb:audit');

DROP TABLE IF EXISTS cmdb_audits;
DROP TABLE IF EXISTS cmdb_sync_records;
DROP TABLE IF EXISTS cmdb_sync_jobs;
DROP TABLE IF EXISTS cmdb_relations;
DROP TABLE IF EXISTS cmdb_cis;
