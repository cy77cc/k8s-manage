CREATE TABLE `projects` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '项目ID',
    `name` VARCHAR(64) NOT NULL COMMENT '项目名称',
    `description` VARCHAR(256) COMMENT '项目描述',
    `owner_id` BIGINT COMMENT '负责人ID',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY `uk_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='项目表';

CREATE TABLE `services` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '服务ID',
    `project_id` BIGINT NOT NULL COMMENT '所属项目ID',
    `name` VARCHAR(64) NOT NULL COMMENT '服务名称',
    `type` VARCHAR(32) NOT NULL COMMENT '服务类型: stateless(Deployment)/stateful(StatefulSet)',
    `image` VARCHAR(256) NOT NULL COMMENT '镜像地址',
    `replicas` INT NOT NULL DEFAULT 1 COMMENT '副本数',
    `service_port` INT COMMENT '服务暴露端口',
    `container_port` INT COMMENT '容器端口',
    `node_port` INT COMMENT 'NodePort端口(可选)',
    `env_vars` JSON COMMENT '环境变量 JSON: [{"key":"k","value":"v"}]',
    `resources` JSON COMMENT '资源限制 JSON: {"limits":{"cpu":"1","memory":"1Gi"},"requests":{...}}',
    `yaml_content` MEDIUMTEXT COMMENT '生成的K8s YAML内容',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY `uk_project_service` (`project_id`, `name`),
    FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务表';
