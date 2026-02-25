-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(64) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  email VARCHAR(128) NOT NULL DEFAULT '',
  phone VARCHAR(32) NOT NULL DEFAULT '',
  avatar VARCHAR(255) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1,
  create_time BIGINT NOT NULL DEFAULT 0,
  update_time BIGINT NOT NULL DEFAULT 0,
  last_login_time BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS roles (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(64) NOT NULL DEFAULT '',
  code VARCHAR(64) NOT NULL UNIQUE,
  description VARCHAR(255) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1,
  create_time BIGINT NOT NULL DEFAULT 0,
  update_time BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS permissions (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(64) NOT NULL DEFAULT '',
  code VARCHAR(128) NOT NULL UNIQUE,
  type TINYINT NOT NULL DEFAULT 0,
  resource VARCHAR(255) NOT NULL DEFAULT '',
  action VARCHAR(32) NOT NULL DEFAULT '',
  description VARCHAR(255) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1,
  create_time BIGINT NOT NULL DEFAULT 0,
  update_time BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS user_roles (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  role_id BIGINT NOT NULL,
  UNIQUE KEY uk_user_role (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS role_permissions (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  role_id BIGINT NOT NULL,
  permission_id BIGINT NOT NULL,
  UNIQUE KEY uk_role_permission (role_id, permission_id)
);

INSERT INTO roles (name, code, description, status, create_time, update_time)
SELECT '管理员', 'admin', '系统管理员', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'admin');

INSERT INTO roles (name, code, description, status, create_time, update_time)
SELECT '运维', 'operator', '运维角色', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'operator');

INSERT INTO roles (name, code, description, status, create_time, update_time)
SELECT '访客', 'viewer', '只读角色', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'viewer');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '主机查看', 'host:read', 3, 'host', 'read', '查看主机', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'host:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '主机管理', 'host:write', 3, 'host', 'write', '管理主机', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'host:write');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'K8s查看', 'kubernetes:read', 3, 'kubernetes', 'read', '查看K8s资源', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'kubernetes:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'K8s管理', 'kubernetes:write', 3, 'kubernetes', 'write', '管理K8s资源', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'kubernetes:write');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '任务查看', 'task:read', 3, 'task', 'read', '查看任务', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'task:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '任务管理', 'task:write', 3, 'task', 'write', '管理任务', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'task:write');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '配置查看', 'config:read', 3, 'config', 'read', '查看配置', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'config:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '配置管理', 'config:write', 3, 'config', 'write', '管理配置', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'config:write');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '监控查看', 'monitoring:read', 3, 'monitoring', 'read', '查看监控', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'monitoring:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '监控管理', 'monitoring:write', 3, 'monitoring', 'write', '管理监控', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'monitoring:write');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'RBAC查看', 'rbac:read', 3, 'rbac', 'read', '查看权限中心', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'rbac:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'RBAC管理', 'rbac:write', 3, 'rbac', 'write', '管理权限中心', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'rbac:write');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '自动化', 'automation:*', 3, 'automation', '*', '自动化能力', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'automation:*');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'CI/CD', 'cicd:*', 3, 'cicd', '*', 'CI/CD能力', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'cicd:*');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT 'CMDB', 'cmdb:*', 3, 'cmdb', '*', 'CMDB能力', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'cmdb:*');

-- +migrate Down
DELETE FROM permissions WHERE code IN (
  'host:read', 'host:write',
  'kubernetes:read', 'kubernetes:write',
  'task:read', 'task:write',
  'config:read', 'config:write',
  'monitoring:read', 'monitoring:write',
  'rbac:read', 'rbac:write',
  'automation:*', 'cicd:*', 'cmdb:*'
);
DELETE FROM roles WHERE code IN ('admin', 'operator', 'viewer');
