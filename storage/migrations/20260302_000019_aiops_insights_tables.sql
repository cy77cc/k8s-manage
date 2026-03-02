-- +migrate Up
-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    action_type VARCHAR(64) NOT NULL COMMENT '操作类型: release_apply, release_approve, target_create, etc.',
    resource_type VARCHAR(64) NOT NULL COMMENT '资源类型: release, target, cluster, credential',
    resource_id BIGINT UNSIGNED NOT NULL COMMENT '资源ID',
    actor_id BIGINT UNSIGNED NOT NULL COMMENT '操作者ID',
    actor_name VARCHAR(255) NOT NULL COMMENT '操作者名称',
    detail JSON COMMENT '操作详情',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_action_type (action_type),
    INDEX idx_resource (resource_type, resource_id),
    INDEX idx_actor (actor_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计日志表';

-- Policies table
CREATE TABLE IF NOT EXISTS policies (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(32) NOT NULL COMMENT 'traffic, resilience, access, slo',
    target_id BIGINT UNSIGNED COMMENT '关联的部署目标ID',
    config JSON COMMENT '策略配置',
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_type (type),
    INDEX idx_target (target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='策略配置表';

-- Risk findings table
CREATE TABLE IF NOT EXISTS risk_findings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    type VARCHAR(64) NOT NULL COMMENT 'risk type',
    severity VARCHAR(16) NOT NULL COMMENT 'critical, high, medium, low',
    title VARCHAR(255) NOT NULL,
    description TEXT,
    service_id BIGINT UNSIGNED,
    service_name VARCHAR(255),
    metadata TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP NULL,
    INDEX idx_severity (severity),
    INDEX idx_service (service_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='风险发现表';

-- Anomalies table
CREATE TABLE IF NOT EXISTS anomalies (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    type VARCHAR(64) NOT NULL,
    metric VARCHAR(64) NOT NULL,
    value DOUBLE NOT NULL,
    threshold DOUBLE NOT NULL,
    service_id BIGINT UNSIGNED,
    service_name VARCHAR(255),
    detected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP NULL,
    INDEX idx_type (type),
    INDEX idx_service (service_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='异常检测表';

-- Suggestions table
CREATE TABLE IF NOT EXISTS suggestions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    type VARCHAR(64) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    impact VARCHAR(16) NOT NULL COMMENT 'high, medium, low',
    service_id BIGINT UNSIGNED,
    service_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    applied_at TIMESTAMP NULL,
    INDEX idx_impact (impact),
    INDEX idx_service (service_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优化建议表';

-- +migrate Down
DROP TABLE IF EXISTS suggestions;
DROP TABLE IF EXISTS anomalies;
DROP TABLE IF EXISTS risk_findings;
DROP TABLE IF EXISTS policies;
DROP TABLE IF EXISTS audit_logs;
