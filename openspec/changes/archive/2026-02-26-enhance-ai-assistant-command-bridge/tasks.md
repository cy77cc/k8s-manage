## 1. Command Bridge Core

- [x] 1.1 在 `internal/service/ai` 增加命令桥接核心结构（parser/planner/executor）与统一命令上下文对象。
- [x] 1.2 实现命令意图解析与参数标准化，支持缺失参数识别与补全提示。
- [x] 1.3 实现跨域路由注册机制，接入 service/deployment/cicd/cmdb/monitoring 的动作映射。
- [x] 1.4 定义统一命令执行结果协议（status/summary/artifacts/trace_id/next_actions）并在 handler 中返回。

## 2. Controlled Execution and Security

- [x] 2.1 为命令执行增加风险分级规则（只读/低风险变更/高风险变更）。
- [x] 2.2 实现执行前计划预览接口，返回目标资源、执行步骤、参数与风险等级。
- [x] 2.3 在 mutating 命令链路接入 RBAC 校验，确保按命令动作与资源鉴权。
- [x] 2.4 对高风险命令接入审批门禁，未审批通过时禁止执行。

## 3. Audit and Traceability

- [x] 3.1 扩展审计数据结构与迁移（Up/Down），增加 command_id/intent/plan_hash/approval_context/execution_summary 字段。
- [x] 3.2 在 AI 命令触发的发布/回滚/配置变更中写入增强审计上下文。
- [x] 3.3 增加审计查询适配，支持按 trace_id 或 command_id 追溯命令执行链路。

## 4. Aggregated Query Experience

- [x] 4.1 实现跨域聚合查询执行器，支持单命令汇总服务状态、发布记录、告警与资产关系。
- [x] 4.2 增加聚合查询超时与并发控制，确保命令响应稳定性。
- [x] 4.3 为聚合结果提供统一摘要与分域明细输出结构。

## 5. API and Frontend UX

- [x] 5.1 新增/调整 AI 命令相关 API 合同（命令预览、确认执行、执行历史、回放详情）。
- [x] 5.2 更新 `web/src/api/modules` 中 AI 相关请求/响应类型与调用封装。
- [x] 5.3 更新 `web/src/pages/AI`，增加命令建议、执行计划预览、确认弹窗与结果面板。
- [x] 5.4 增加命令历史与执行回放视图，支持查看输入、计划、时间线与结果摘要。

## 6. Validation and Rollout

- [x] 6.1 增加后端测试：命令解析、参数补全、路由分发、审批门禁、RBAC 拒绝、审计写入。
- [x] 6.2 增加前端交互测试：命令建议、预览确认、执行结果展示、历史回放。
- [ ] 6.3 在预发环境执行高风险命令审批链路与回滚演练验证。
- [x] 6.4 运行 `openspec validate --changes --json` 并修复所有问题。
