## Why

当前 AI 助手链路虽然已经具备 Rewrite、Planner、Executor、Experts、Summarizer 和 ThoughtChain，但系统默认由代码接管模型输出：Rewrite 会回退到代码理解，Planner 会重写模型计划，Summarizer 和 Final Renderer 会把模型结论替换成模板化内容。结果是用户看到的并不是“大模型理解并执行后的结果”，而是“代码猜测 + 模型素材 + 再次包装”的混合产物，既削弱了 AI 助手的核心价值，也让输出质量和一致性持续恶化。

这个变化要把 AI 助手重新收敛为“模型主导语义，代码负责边界”的系统：用户通过自然语言指挥，AI 负责理解、规划、执行和总结；当某一层模型或工具不可用时，系统明确告知“当前暂时不可用，请手动处理”，而不是再用大量纯代码兜底去伪装 AI 能力。Rewrite 仍然保留并增强，因为它是后续 RAG 检索增强的语义入口，但 Rewrite 之后的代码不应继续替模型做语义决策。

## What Changes

- 将 AI 助手重构为“分层兜底 + 明确告知”的模型优先架构：模型主导 `Rewrite -> Planner -> Expert -> Summarizer` 的语义输出，代码只负责协议校验、安全边界、运行时调度和错误告知。
- 移除当前链路中“代码替模型理解/改写/总结”的主路径，包括 Rewrite 的语义 fallback、Planner 的二次裁决与计划改写、Summarizer 的模板化骨架回填、Final Renderer 的语义接管。
- 保留 Rewrite 大模型，并把它升级为 RAG 语义入口：Rewrite 输出不仅服务 Planner，还要产出适合检索增强的规范化查询、关键词、范围和检索意图。
- 重构 Planner 为“模型规划 + 代码校验”模式：代码仅验证 schema、前置条件和安全边界，不再重写 `expert / intent / task / title / recommendation / conclusion` 等语义字段。
- 重构 Expert 执行契约为“结构化优先，半结构化可恢复”，避免因为模型没有完全命中严格 JSON shape 就被视为执行失败。
- 重构 Summarizer 为真正的模型总结层：总结失败时明确告知，而不是回退到“已完成步骤 N 个 / 收集证据 N 条 / 查看最终结论正文”这类低价值模板文本。
- 将 Final Renderer 降级为展示层，只负责排版、折叠、脱敏和证据呈现，不再重新组织答案语义。
- 调整前端 ThoughtChain 和最终回答展示：明确区分“AI 的过程”“AI 的结论”“原始执行证据”，避免前端继续压扁后端 richer output。
- 更新启动期和运行期的 AI 可用性策略：当 Rewrite、Planner、Expert 或 Summarizer 模型不可用时，系统必须向用户明确说明该层暂不可用，并建议手动操作，而不是继续依赖纯代码猜测。
- **BREAKING**: 移除现有“代码主导的语义 fallback”行为。部分当前依赖纯代码推断继续运行的请求，在模型不可用时将改为明确失败并提示用户手动处理。

## Capabilities

### New Capabilities
- `ai-model-first-runtime`: 定义 AI 助手的模型优先运行时契约，包括分层兜底、显式不可用告知、模型主导语义、代码只做边界与协议校验。
- `ai-rewrite-rag-bridge`: 定义 Rewrite 作为 RAG 语义入口的输出契约，包括规范化查询、检索关键词、范围、检索意图和面向下游 Planner/RAG 的统一中间层。

### Modified Capabilities
- `ai-module-architecture`: AI 模块整体职责边界调整为“模型主导语义、运行时代码负责边界与安全”，不再由代码主导语义 fallback。
- `ai-streaming-events`: 流事件需要支持更忠实地暴露模型阶段输出、显式不可用状态和原始执行证据，而不是只暴露代码包装后的摘要。
- `rag-knowledge-base`: RAG 接入点将从当前的通用自动检索，升级为依赖 Rewrite 结构化输出的检索增强链路。
- `agentic-multi-expert-architecture`: 多专家架构的主规范需要从“代码校正模型输出”调整为“模型规划与专家执行主导，代码仅执行 runtime control plane 与安全约束”。

## Impact

- Affected backend: `internal/ai/rewrite`, `internal/ai/planner`, `internal/ai/executor`, `internal/ai/summarizer`, `internal/ai/orchestrator`, `internal/ai/model`, `internal/service/ai`.
- Affected frontend: `web/src/components/AI/*`, `web/src/api/modules/ai.ts`, ThoughtChain 与最终回答展示链路。
- Affected APIs/events: AI chat SSE 事件语义、summary/final answer 契约、模型不可用时的用户可见错误语义。
- Affected docs/specs: AI 架构文档、前端 ThoughtChain 对接说明、RAG 接入说明、迁移文档。
- Affected tests: Rewrite/Planner/Executor/Summarizer 单测，端到端流式事件测试，模型不可用与半结构化恢复测试。
