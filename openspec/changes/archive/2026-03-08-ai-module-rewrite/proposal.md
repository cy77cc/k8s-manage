# AI Module Rewrite

## Summary

完全重写 AI 模块，基于 Eino 框架实现新架构。核心目标是构建确定性、安全性与云原生基础设施深度集成的 AI-PaaS 平台。

## Motivation

现有 AI 模块代码结构混乱，缺乏清晰的架构分层。需要按照 `ai.md` 规范重新设计，实现：

1. **确定性工作流** - LLM 推理结果需经过严格校验才能执行
2. **安全性保障** - 风险操作需人工审批，权限校验前置
3. **可观测性** - 完整的追踪与审计能力

## Scope

### In Scope

- IntentRouter: 按工具域路由的多级意图分类
- ActionGraph: 基于 Workflow 的确定性执行流程
- Validation Layer: K8s OpenAPI 规范校验
- SecurityAspect: 权限检查 + Human-in-the-loop 中断
- Approval System: 双轨制审批系统（有权限确认/无权限任务审批）
- RAG System: 用户投喂 + 反馈收集的知识库

### Out of Scope

- 前端组件重写（仅在后端接口变化时调整）
- 审批流转系统的完整实现（Phase 2）

## Dependencies

- Eino v0.7.37 (compose.Graph, compose.Workflow, compose.Lambda, compose.ToolsNode)
- 现有 tools 模块（保留）
- Milvus 向量库（RAG）
- Redis（Checkpoint 存储）

## Risks

| Risk | Mitigation |
|------|------------|
| Eino API 变更 | 锁定版本，关注 release notes |
| Validation 层复杂度 | 从 Level 2 开始，按需扩展 |
| 审批系统复杂度 | Phase 1 仅实现核心流程 |

## Success Criteria

1. 所有工具调用经过权限校验
2. 高风险操作必须用户确认后执行
3. 无权限操作生成可执行的审批任务
4. Graph 执行支持 Interrupt/Resume
5. RAG 反馈有效解决问题后自动入库
