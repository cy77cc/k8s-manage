CREATE TABLE nodes (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '节点ID',
    name            VARCHAR(64) NOT NULL COMMENT '节点名称',
    hostname        VARCHAR(64) COMMENT '主机名',
    labels          JSON COMMENT '节点标签',
    description     VARCHAR(256) COMMENT '节点描述',

    ip              VARCHAR(45) NOT NULL COMMENT '节点IP',
    port            INT DEFAULT 22 COMMENT 'SSH端口',

    ssh_user        VARCHAR(64) NOT NULL DEFAULT 'root' COMMENT 'SSH用户名',
    ssh_key_id      BIGINT NOT NULL COMMENT 'SSH密钥ID',

    os              VARCHAR(64) COMMENT '操作系统',
    arch            VARCHAR(32) COMMENT '架构',
    kernel          VARCHAR(64) COMMENT '内核版本',

    cpu_cores       INT COMMENT 'CPU核心数',
    memory_mb       INT COMMENT '内存MB',
    disk_gb         INT COMMENT '磁盘GB',

    status          VARCHAR(32) NOT NULL COMMENT '节点状态',    
    role            VARCHAR(32) COMMENT '节点角色', -- master / worker / none
    cluster_id      BIGINT COMMENT '集群ID',      -- nullable

    last_check_at   TIMESTAMP NULL COMMENT '最后检查时间',
    created_at      TIMESTAMP NOT NULL COMMENT '创建时间',
    updated_at      TIMESTAMP NOT NULL COMMENT '更新时间',

    UNIQUE KEY uk_ip_port (ip, port)
);

CREATE TABLE ssh_keys (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'SSH密钥ID',
    name        VARCHAR(64) COMMENT '密钥名称',
    private_key TEXT NOT NULL COMMENT '私钥',
    passphrase  VARCHAR(128) COMMENT 'passphrase',
    created_at  TIMESTAMP COMMENT '创建时间'
);

CREATE TABLE node_events (
    id         BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '事件ID',
    node_id    BIGINT COMMENT '节点ID',
    type       VARCHAR(32) COMMENT '事件类型',
    message    TEXT COMMENT '事件消息',
    created_at TIMESTAMP COMMENT '创建时间'
);