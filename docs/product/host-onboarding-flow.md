# Host Onboarding Flow (Product)

## 目标

把“新增主机”改为可解释、可重试的三步流程，降低失败率和误入库。

## 三步流程

1. 连接信息
- 输入：`name/ip/port/auth_type/username/password|ssh_key_id`
- 操作：点击“执行探测”

2. 探测结果确认
- 展示：`reachable/latency/facts/warnings/error_code`
- 行为：失败可返回上一步修改重试；成功进入下一步

3. 入库确认
- 输入：`description/labels/role/cluster_id`
- admin 可见 `force create`
- 提交：`POST /api/v1/hosts`（携带 `probe_token`）

## 错误模型

- `validation_error`
- `connect_error`
- `auth_error`
- `timeout_error`
- `probe_expired`
- `probe_not_found`

## 交互原则

- 探测失败不直接入库（非 admin）。
- 所有失败都提供“返回修改并重试”路径。
- 探测结果中 facts/warnings 必须可见，便于操作员判断。
