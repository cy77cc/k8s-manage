# Host Release Test Plan (Team-QA)

## 核心回归

1. SSH 密钥管理
- 创建密钥成功
- 无 `security.encryption_key` 时创建失败
- key 在被主机引用时删除失败
- 指定目标IP验证成功/失败路径

2. 云平台导入
- 新增账号成功
- 查询实例返回列表
- 批量导入后 `nodes.source=cloud_import`
- 导入任务状态可查询

3. KVM 虚拟化
- 预检查成功
- 创建任务成功并生成主机 `source=kvm_provision`
- 虚拟化任务可查询

4. 兼容与编译
- `/node/add` 仍可用
- `go test ./...` 通过
- `npm run build` 通过
