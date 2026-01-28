CREATE TABLE `users`
(
    `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `username`         VARCHAR(64)     NOT NULL DEFAULT '' UNIQUE COMMENT '用户名',
    `password_hash`    VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '密码哈希值',
    `email`            VARCHAR(128)    NOT NULL DEFAULT '' COMMENT '邮箱',
    `phone`            VARCHAR(32)     NOT NULL DEFAULT '' COMMENT '手机号',
    `avatar`           VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '头像URL',
    `status`           TINYINT         NOT NULL DEFAULT 1 COMMENT '状态（1：正常；2：禁用）',
    `create_time`      BIGINT          NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time`      BIGINT          NOT NULL DEFAULT 0 COMMENT '更新时间',
    `last_login_time`  BIGINT          NOT NULL DEFAULT 0 COMMENT '最后登录时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    KEY `idx_email` (`email`),
    KEY `idx_phone` (`phone`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE `roles`
(
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '角色ID',
    `name`        VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '角色名称',
    `code`        VARCHAR(64)     NOT NULL DEFAULT '' UNIQUE COMMENT '角色代码',
    `description` VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '角色描述',
    `status`      TINYINT         NOT NULL DEFAULT 1 COMMENT '状态（1：正常；2：禁用）',
    `create_time` BIGINT          NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` BIGINT          NOT NULL DEFAULT 0 COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_code` (`code`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE `permissions`
(
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '权限ID',
    `name`        VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '权限名称',
    `code`        VARCHAR(128)    NOT NULL DEFAULT '' UNIQUE COMMENT '权限代码',
    `type`        TINYINT         NOT NULL DEFAULT 0 COMMENT '权限类型（0：菜单；1：接口）',
    `resource`    VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '资源路径',
    `action`      VARCHAR(32)     NOT NULL DEFAULT '' COMMENT '操作方法',
    `description` VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '权限描述',
    `status`      TINYINT         NOT NULL DEFAULT 1 COMMENT '状态（1：正常；2：禁用）',
    `create_time` BIGINT          NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` BIGINT          NOT NULL DEFAULT 0 COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_code` (`code`),
    KEY `idx_resource_action` (`resource`, `action`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE `user_roles`
(
    `id`      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户角色ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_role` (`user_id`, `role_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_role_id` (`role_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE `role_permissions`
(
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '角色权限ID',
    `role_id`       BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    `permission_id` BIGINT UNSIGNED NOT NULL COMMENT '权限ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_role_perm` (`role_id`, `permission_id`),
    KEY `idx_role_id` (`role_id`),
    KEY `idx_permission_id` (`permission_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE `auth_refresh_tokens`
(
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '刷新令牌ID',
    `user_id`     BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `token`       VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '刷新令牌值',
    `expires`     BIGINT          NOT NULL DEFAULT 0 COMMENT '过期时间',
    `revoked`     TINYINT         NOT NULL DEFAULT 0 COMMENT '是否已撤销（0：未撤销；1：已撤销）',
    `create_time` BIGINT          NOT NULL DEFAULT 0 COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_token` (`token`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE `jwt_blacklist`
(
    `jti`        VARCHAR(128) PRIMARY KEY COMMENT 'JWT ID',
    `expired_at` TIMESTAMP COMMENT '过期时间'
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;
