## Context

当前平台已具备 AI 对话、工具调用与多域业务能力（服务、部署、CI/CD、CMDB、监控），但用户仍需要在多个页面和接口之间切换完成实际操作。现状缺少统一命令抽象、跨域执行编排和可验证执行反馈，导致 AI 助手无法成为稳定的“系统操作桥梁”。

该变更是跨域改造，涉及 `internal/service/ai`、多业务域逻辑封装、RBAC/审批链路、前端 AI 页面交互和审计数据结构，属于高耦合且安全敏感的能力升级。

## Goals / Non-Goals

**Goals:**
- 建立 AI 助手命令中枢，支持命令解析、参数补全、意图路由和跨域执行编排。
- 提供“执行前可预览、执行中可确认、执行后可追溯”的完整闭环。
- 将命令执行接入 RBAC 与审批策略，高风险动作默认受控。
- 支持跨域聚合查询（如服务状态+发布记录+告警摘要）并输出结构化结果。

**Non-Goals:**
- 不引入新的底层 LLM 或向量检索基础设施。
- 不在本次变更中实现全量自治 Agent，仅支持受控命令编排。
- 不替代原有页面/接口的全部手工入口，保留现有操作链路作为兜底。

## Decisions

### Decision 1: 增加“命令编排层”而非直接在对话层调用各域逻辑
- Choice: 在 `internal/service/ai` 下新增命令编排组件（Parser + Planner + Executor）。
- Rationale: 将“理解”和“执行”解耦，便于权限校验、审批注入、重试与审计统一。
- Alternative considered: 直接在对话处理器中调用各域服务。缺点是流程不可复用，权限/审计容易分散。

### Decision 2: 采用“预览计划 + 显式确认”执行模型
- Choice: 命令先返回执行计划（目标资源、动作、参数、风险等级），用户确认后执行。
- Rationale: 降低误操作风险，满足高风险变更的可控要求。
- Alternative considered: 自动执行。缺点是风险不可接受，难以满足审批与审计要求。

### Decision 3: 统一命令结果协议
- Choice: 输出标准结果结构（status、summary、artifacts、next_actions、trace_id）。
- Rationale: 方便前端展示一致体验，并可用于审计/追溯。
- Alternative considered: 各命令自由返回。缺点是 UI 适配成本高，复盘困难。

### Decision 4: 审计模型扩展到“命令上下文”
- Choice: 在审计记录中增加 command_id、intent、plan_hash、approval_context、execution_summary。
- Rationale: 满足“AI 触发操作可追溯”的合规诉求。
- Alternative considered: 复用现有审计字段不扩展。缺点是无法定位 AI 命令与实际动作的映射关系。

## Risks / Trade-offs

- [Risk] 命令解析错误导致执行偏差。 → Mitigation: 强制预览+确认，参数校验失败不执行。
- [Risk] 跨域编排引入耦合与回归风险。 → Mitigation: 通过适配层封装域调用，增加契约测试。
- [Risk] 审批链路增加交互成本。 → Mitigation: 仅对高风险命令启用审批，低风险保持快速执行。
- [Risk] 聚合查询可能带来性能抖动。 → Mitigation: 增加聚合层超时、并发限制与缓存策略。

## Migration Plan

1. 新增 AI 命令编排层与统一结果协议，先接入只读查询命令。
2. 接入变更类命令（部署/回滚/配置）并串联 RBAC 与审批确认。
3. 扩展审计字段并落库，兼容旧审计查询接口。
4. 前端 AI 页面上线执行计划预览、确认弹窗和历史回放。
5. 灰度开放命令能力，观察失败率和误操作指标后全量。

Rollback strategy:
- 通过功能开关回退到旧 AI 对话模式（仅问答/建议，不执行变更）。
- 审计扩展字段保留，不影响旧链路读取。

## Open Questions

- 命令语法首期是否需要支持可组合管道（如先查后执行）？
- 高风险命令审批是否必须双人批准，还是沿用现有策略配置？
- 聚合查询结果默认时间窗口（例如 1h/24h）应如何统一？
