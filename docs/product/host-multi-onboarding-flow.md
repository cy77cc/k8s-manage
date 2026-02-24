# Host Multi-Onboarding Flow (Team-Product)

## 接入入口

1. SSH 接入：密码 / 密钥二选一。
2. 云平台导入：阿里云 / 腾讯云实例批量纳管。
3. KVM 虚拟化创建：从宿主机创建新虚拟机并纳管。
4. 密钥管理：创建/验证/删除 SSH 密钥。

## 页面路径

- `/hosts/onboarding`
- `/hosts/cloud-import`
- `/hosts/virtualization`
- `/hosts/keys`

## 交互约束

- SSH 仍走 probe -> confirm -> create。
- 云导入支持账号管理、实例查询、批量导入。
- KVM 支持预检查后创建，创建后返回任务ID和新主机记录。
- 密钥页面不展示明文私钥。
