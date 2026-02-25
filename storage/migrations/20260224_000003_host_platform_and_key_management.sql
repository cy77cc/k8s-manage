-- +migrate Up
ALTER TABLE ssh_keys
  ADD COLUMN fingerprint VARCHAR(128) NULL,
  ADD COLUMN algorithm VARCHAR(32) NULL,
  ADD COLUMN encrypted TINYINT(1) NOT NULL DEFAULT 0,
  ADD COLUMN usage_count INT NOT NULL DEFAULT 0;

ALTER TABLE nodes
  ADD COLUMN source VARCHAR(32) NOT NULL DEFAULT 'manual_ssh',
  ADD COLUMN provider VARCHAR(32) NULL,
  ADD COLUMN provider_instance_id VARCHAR(128) NULL,
  ADD COLUMN parent_host_id BIGINT UNSIGNED NULL;

CREATE TABLE IF NOT EXISTS host_cloud_accounts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  provider VARCHAR(32) NOT NULL,
  account_name VARCHAR(128) NOT NULL,
  access_key_id VARCHAR(256) NOT NULL,
  access_key_secret_enc LONGTEXT NOT NULL,
  region_default VARCHAR(64),
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_host_cloud_provider (provider),
  INDEX idx_host_cloud_creator (created_by)
);

CREATE TABLE IF NOT EXISTS host_import_tasks (
  id VARCHAR(64) PRIMARY KEY,
  provider VARCHAR(32) NOT NULL,
  account_id BIGINT UNSIGNED NULL,
  request_json LONGTEXT,
  result_json LONGTEXT,
  status VARCHAR(32) NOT NULL,
  error_message TEXT,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_host_import_provider (provider),
  INDEX idx_host_import_status (status),
  INDEX idx_host_import_creator (created_by)
);

CREATE TABLE IF NOT EXISTS host_virtualization_tasks (
  id VARCHAR(64) PRIMARY KEY,
  host_id BIGINT UNSIGNED NOT NULL,
  hypervisor VARCHAR(32) NOT NULL,
  request_json LONGTEXT,
  vm_name VARCHAR(128),
  vm_ip VARCHAR(64),
  status VARCHAR(32) NOT NULL,
  error_message TEXT,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_host_virt_host (host_id),
  INDEX idx_host_virt_status (status),
  INDEX idx_host_virt_creator (created_by)
);

-- +migrate Down
DROP TABLE IF EXISTS host_virtualization_tasks;
DROP TABLE IF EXISTS host_import_tasks;
DROP TABLE IF EXISTS host_cloud_accounts;
ALTER TABLE nodes
  DROP COLUMN parent_host_id,
  DROP COLUMN provider_instance_id,
  DROP COLUMN provider,
  DROP COLUMN source;
ALTER TABLE ssh_keys
  DROP COLUMN usage_count,
  DROP COLUMN encrypted,
  DROP COLUMN algorithm,
  DROP COLUMN fingerprint;
