## 1. Rewrite And RAG Bridge

- [x] 1.1 收缩 `internal/ai/rewrite` 的 fallback 语义，改为显式不可用告知而不是代码理解兜底
- [x] 1.2 扩展 Rewrite 输出契约，增加面向 RAG 的 retrieval 字段与语义范围
- [x] 1.3 调整 Planner/RAG 对 Rewrite 输出的消费方式，统一使用同一份语义中间层
- [x] 1.4 为 Rewrite 不可用、无效 JSON、RAG bridge 场景补充单测和集成测试

## 2. Planner And Expert Runtime

- [x] 2.1 将 Planner 的 normalize/canonicalize 逻辑收缩为协议修复与安全校验，移除语义改写主路径
- [x] 2.2 调整 Planner 失败语义，无法执行时明确标记 planning unavailable 或 invalid，而不是代码强改 plan
- [x] 2.3 放宽 Expert 结果契约为“结构化优先，半结构化可恢复”，实现最小恢复解析
- [x] 2.4 补充 host、k8s、service 等典型场景的回归测试，验证运行时代码不再重写模型语义

## 3. Summarizer And Final Answer

- [x] 3.1 删除 Summarizer 主路径中的模板化 base summary 骨架，仅保留显式失败兜底
- [x] 3.2 将 FinalAnswerRenderer 降级为展示层，移除 headline/findings/recommendations 的语义接管
- [x] 3.3 定义 Summarizer 不可用时的显式用户告知与原始证据展示策略
- [x] 3.4 补充最终回答稳定性、显式失败语义和原始证据展示测试

## 4. Streaming And Frontend Presentation

- [x] 4.1 调整 SSE 事件语义，确保 `stage_delta` 和 `delta` 忠实反映模型阶段输出
- [x] 4.2 前端拆分 ThoughtChain、Final Answer、Raw Evidence 三层展示结构
- [x] 4.3 更新会话持久化结构，支持保留 richer summary/evidence/result block
- [x] 4.4 补充前端渲染、刷新恢复、多轮对话和阶段不可用提示测试

## 5. Availability, Rollout, And Migration

- [x] 5.1 定义 Rewrite/Planner/Expert/Summarizer 分层健康检查与启动日志
- [x] 5.2 增加模型不可用时的统一错误码和用户可见文案
- [x] 5.3 设计灰度开关与回滚路径，允许逐步关闭旧的代码语义 fallback
- [x] 5.4 更新 AI 架构文档、RAG 接入文档、迁移说明和运维手册
