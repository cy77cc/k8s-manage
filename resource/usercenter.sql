CREATE TABLE `users`
(
    `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `username`         VARCHAR(64)     NOT NULL DEFAULT '' UNIQUE,
    `password_hash`    VARCHAR(255)    NOT NULL DEFAULT '',
    `email`            VARCHAR(128)    NOT NULL DEFAULT '',
    `phone`            VARCHAR(32)     NOT NULL DEFAULT '',
    `avatar`           VARCHAR(255)    NOT NULL DEFAULT '',
    `status`           TINYINT         NOT NULL DEFAULT 1,
    `create_time`      BIGINT          NOT NULL DEFAULT 0,
    `update_time`      BIGINT          NOT NULL DEFAULT 0,
    `last_login_time`  BIGINT          NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    KEY `idx_email` (`email`),
    KEY `idx_phone` (`phone`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE `roles`
(
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name`        VARCHAR(64)     NOT NULL DEFAULT '',
    `code`        VARCHAR(64)     NOT NULL DEFAULT '' UNIQUE,
    `description` VARCHAR(255)    NOT NULL DEFAULT '',
    `status`      TINYINT         NOT NULL DEFAULT 1,
    `create_time` BIGINT          NOT NULL DEFAULT 0,
    `update_time` BIGINT          NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_code` (`code`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE `permissions`
(
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name`        VARCHAR(64)     NOT NULL DEFAULT '',
    `code`        VARCHAR(128)    NOT NULL DEFAULT '' UNIQUE,
    `type`        TINYINT         NOT NULL DEFAULT 0,
    `resource`    VARCHAR(255)    NOT NULL DEFAULT '',
    `action`      VARCHAR(32)     NOT NULL DEFAULT '',
    `description` VARCHAR(255)    NOT NULL DEFAULT '',
    `status`      TINYINT         NOT NULL DEFAULT 1,
    `create_time` BIGINT          NOT NULL DEFAULT 0,
    `update_time` BIGINT          NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_code` (`code`),
    KEY `idx_resource_action` (`resource`, `action`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE `user_roles`
(
    `id`      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `role_id` BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_role` (`user_id`, `role_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_role_id` (`role_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE `role_permissions`
(
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `role_id`       BIGINT UNSIGNED NOT NULL,
    `permission_id` BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_role_perm` (`role_id`, `permission_id`),
    KEY `idx_role_id` (`role_id`),
    KEY `idx_permission_id` (`permission_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE `auth_refresh_tokens`
(
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id`     BIGINT UNSIGNED NOT NULL,
    `token`       VARCHAR(255)    NOT NULL DEFAULT '',
    `expires`     BIGINT          NOT NULL DEFAULT 0,
    `revoked`     TINYINT         NOT NULL DEFAULT 0,
    `create_time` BIGINT          NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_token` (`token`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

