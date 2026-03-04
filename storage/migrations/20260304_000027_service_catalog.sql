-- +migrate Up
CREATE TABLE IF NOT EXISTS service_categories (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL,
  display_name VARCHAR(128) NOT NULL,
  icon VARCHAR(256) DEFAULT '',
  description TEXT,
  sort_order INT NOT NULL DEFAULT 0,
  is_system TINYINT(1) NOT NULL DEFAULT 0,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_service_categories_name (name),
  KEY idx_service_categories_sort (sort_order)
);

CREATE TABLE IF NOT EXISTS service_templates (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(128) NOT NULL,
  display_name VARCHAR(256) NOT NULL,
  description TEXT,
  icon VARCHAR(256) DEFAULT '',
  category_id BIGINT UNSIGNED NOT NULL,
  version VARCHAR(32) NOT NULL DEFAULT '1.0.0',
  owner_id BIGINT UNSIGNED NOT NULL,
  visibility VARCHAR(16) NOT NULL DEFAULT 'private',
  status VARCHAR(32) NOT NULL DEFAULT 'draft',
  k8s_template MEDIUMTEXT,
  compose_template MEDIUMTEXT,
  variables_schema JSON,
  readme TEXT,
  tags JSON,
  deploy_count INT NOT NULL DEFAULT 0,
  review_note TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_service_templates_name (name),
  KEY idx_service_templates_category (category_id),
  KEY idx_service_templates_owner (owner_id),
  KEY idx_service_templates_status (status),
  KEY idx_service_templates_visibility_status (visibility, status)
);

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '目录读取', 'catalog:read', 3, 'catalog', 'read', '浏览服务目录', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'catalog:read');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '目录写入', 'catalog:write', 3, 'catalog', 'write', '创建和编辑模板', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'catalog:write');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '目录审核', 'catalog:approve', 3, 'catalog', 'approve', '审核并发布模板', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'catalog:approve');

INSERT INTO permissions (name, code, type, resource, action, description, status, create_time, update_time)
SELECT '目录管理', 'catalog:manage', 3, 'catalog', 'manage', '管理目录分类', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'catalog:manage');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON p.code IN ('catalog:read')
WHERE r.code = 'viewer'
AND NOT EXISTS (
  SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission_id = p.id
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON p.code IN ('catalog:read', 'catalog:write')
WHERE r.code = 'operator'
AND NOT EXISTS (
  SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission_id = p.id
);

-- +migrate Down
ALTER TABLE service_templates DROP FOREIGN KEY fk_service_templates_category;
DROP TABLE IF EXISTS service_templates;
DROP TABLE IF EXISTS service_categories;

DELETE FROM role_permissions WHERE permission_id IN (
  SELECT id FROM permissions WHERE code IN ('catalog:read', 'catalog:write', 'catalog:approve', 'catalog:manage')
);
DELETE FROM permissions WHERE code IN ('catalog:read', 'catalog:write', 'catalog:approve', 'catalog:manage');
