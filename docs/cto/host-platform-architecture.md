# Host Platform Architecture (Team-CTO)

## 目标

主机域升级为多入口接入平台：`manual_ssh`、`cloud_import`、`kvm_provision`。

## 本次实现

- `host` 领域新增入口 API：
  - `/api/v1/hosts/cloud/*`
  - `/api/v1/hosts/virtualization/*`
  - `/api/v1/credentials/ssh_keys*`
- 主机模型扩展：`source/provider/provider_instance_id/parent_host_id`。
- 密钥模型扩展：`fingerprint/algorithm/encrypted/usage_count`。
- 数据迁移：
  - `20260224_000003_host_platform_and_key_management.sql`

## 安全边界

- SSH 私钥加密落库（AES-GCM，密钥来源 `security.encryption_key`）。
- 私钥禁止回显；仅在创建时输入。
- 云账号密钥同样加密存储。

## 演进点

- 云厂商 API 当前为 MVP mock 适配层，后续替换 provider 实调用。
- KVM 当前为任务模型 + 预检查/创建闭环，后续接 libvirt 执行器。
