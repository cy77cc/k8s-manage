# AI Tool Usage Guide

## 1. 基础流程

1. 在 AI 面板输入问题或命令。
2. 先用只读工具确认现状（inventory/list/detail）。
3. 如需变更，先预览，再审批，再执行。
4. 根据 `tool_result` 与恢复建议决定重试或降级。

## 2. 常用模式

### 2.1 资源盘点

- 主机：`host_list_inventory`
- 集群：`cluster_list_inventory`
- 服务：`service_list_inventory`
- 部署目标：`deployment_target_list`

### 2.2 参数提示

调用：`GET /api/v1/ai/tools/:name/params/hints`

适用场景：

- 不确定 `*_id` 的可选值
- 不确定参数默认值
- 不确定参数说明

### 2.3 场景推荐

调用：`GET /api/v1/ai/scene/:scene/tools`

示例场景：

- `services:list`
- `deployment:targets`
- `governance:permissions`

## 3. 命令中心实践

### 3.1 自动补全

在命令输入框直接输入关键字，系统会按 `scene + q` 返回建议。

### 3.2 别名

- 内置别名：`hst/svc/cls/pl/job/cfg/alert/topo`
- 自定义别名：
  - `POST /api/v1/ai/commands/aliases`
  - 作用域：`user + scene`

### 3.3 参数模板

- 保存模板：`POST /api/v1/ai/commands/templates`
- 预览时传 `template=<name>` 自动注入模板参数

## 4. 错误恢复

工具失败时关注 `execution_error`：

- `code`
- `recoverable`
- `suggestions`
- `hint_action`

推荐动作：

1. 按 `hint_action` 修正参数。
2. 调用 inventory/list 工具确认资源 ID。
3. 缩小查询范围（`limit`、`time_range`）后重试。
4. 变更工具失败时先检查审批状态。

## 5. 安全约束

- 高风险变更必须审批。
- 主机批量命令默认拦截危险命令。
- 先只读诊断，再执行变更。
