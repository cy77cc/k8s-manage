## Context

当前 AI 助手链路已经具备 `Rewrite -> Planner -> Executor Runtime -> Experts -> Summarizer -> ThoughtChain` 的基本分层，但在实现上存在明显的“代码接管模型语义”问题：

- Rewrite 在模型失败或输出不完整时，会退回到代码生成的标准化结果。
- Planner 的 `normalizeDecision / canonicalizePlan / validatePlanPrerequisites` 不仅校验格式，还会重写 `expert / intent / task / title`，甚至把模型 plan 打回 clarify。
- Expert 结果依赖严格 JSON 协议，模型即使得出正确结论，只要格式偏移也会被视为失败。
- Summarizer 以硬编码模板骨架为主，模型输出字段不足时会被回灌为“已完成步骤 N 个 / 查看最终结论正文”这类低价值文本。
- Final renderer 在后端重新组织答案结构，前端继续把 richer result 压扁为固定块。

这导致 AI 助手的实际行为与产品目标相反：用户本来希望用自然语言代替繁琐的点击和手动检索，但系统实际输出却更接近“代码猜测 + 模型素材 + 再包装”。同时，后续计划接入 RAG，要求 Rewrite 产出稳定、可检索、可复用的规范化语义中间层，这进一步要求 Rewrite 必须由大模型主导，而不能再由代码 fallback 冒充理解能力。

## Goals / Non-Goals

**Goals:**
- 让大模型重新成为 AI 助手的语义主导者，覆盖 Rewrite、Planner、Expert、Summarizer 四个核心阶段。
- 把代码职责收缩到协议校验、安全边界、运行时调度、状态持久化、显式失败告知。
- 保留并增强 Rewrite，因为它既是规划入口，也是后续 RAG 检索增强的语义桥。
- 将“兜底”统一改成分层显式失败：某层模型不可用时直接告诉用户该层不可用，不再用纯代码伪装 AI 能力。
- 保持半结构化协议作为链路通信媒介，但禁止代码通过协议回填和改写重新决定模型语义。
- 让前端清晰区分 AI 过程、AI 最终结论和原始执行证据。

**Non-Goals:**
- 不在本次变化中引入新的基础模型 provider 或替换 Eino/ADK。
- 不在本次变化中改写所有领域工具本身的业务能力。
- 不把安全边界降级为“完全信任模型”；审批、RBAC、风险分级仍由代码主导。
- 不在本次变化中完成全部 RAG 能力建设，只定义 Rewrite 与 RAG 的语义桥和检索增强接入点。

## Decisions

### 1. 模型主导语义，代码主导边界

系统将明确采用“模型主导语义、代码主导边界”的分层约束：

- Rewrite/Planner/Expert/Summarizer 负责理解、规划、执行判断和结论表达。
- Runtime 代码负责 schema 校验、审批、权限、状态机、事件流和错误归类。
- 代码不得再擅自重写模型的 `intent / expert / task / recommendation / conclusion`。

为什么：
- 这与 AI 助手“自然语言指挥，AI 理解并执行”的产品定位一致。
- 当前最大的问题不是模型不够强，而是代码层层重写模型语义。

备选方案：
- 继续保持“模型输出 + 代码回填兜底”。否决，因为这会持续制造低质量但貌似稳定的垃圾结果。

### 2. Rewrite 保留并增强，成为 RAG 语义入口

Rewrite 大模型保留，并扩展为同时服务 Planner 与 RAG 的统一语义层。Rewrite 输出需要新增面向检索增强的字段，例如：

- `retrieval_intent`
- `retrieval_queries`
- `retrieval_keywords`
- `knowledge_scope`
- `requires_rag`

为什么：
- 用户口语化输入直接进入 RAG 检索质量不稳定。
- Rewrite 的语义规范化是更适合与 Milvus / 知识库检索结合的入口。

备选方案：
- 取消 Rewrite，用代码直推 query。否决，因为这会直接降低检索语义质量。

### 3. Normalization 只允许修协议，不允许改语义

Rewrite、Planner、Summarizer 各阶段的 `normalize*` 函数只允许：

- 修复类型
- 补充非语义型默认值
- 验证 schema
- 标记错误

禁止：

- 改写专家选择
- 改写任务描述
- 改写标题和建议
- 把模型 plan 自动打回 clarify，除非严格的安全/执行前置条件明确失败

为什么：
- 当前 normalize 实际上是二次决策器，是整个系统最严重的能力限幅点。

备选方案：
- 只收缩部分 normalize。否决，因为局部收缩仍会保留“代码篡改模型语义”的架构问题。

### 4. 分层兜底统一为显式不可用告知

兜底策略统一改成：

- Rewrite 模型不可用：告知“AI 理解模块当前不可用”
- Planner 模型不可用：告知“AI 规划模块当前不可用”
- Expert 模型不可用：告知“AI 执行专家当前不可用”
- Summarizer 模型不可用：告知“AI 总结模块当前不可用，可查看原始执行结果”

为什么：
- 纯代码 fallback 的输出通常低质量且误导用户。
- 对运维场景而言，“明确不能用”比“给出看似合理的垃圾结果”更安全。

备选方案：
- 保留现有 fallback 作为兼容。否决，因为这会持续模糊 AI 可用性边界。

### 5. Expert 输出改为“结构化优先，半结构化可恢复”

专家仍优先输出结构化结果，但执行层必须允许对以下情况做恢复解析：

- 合法 JSON 但字段缺失
- 半结构化文本
- 明显包含总结、观察事实、建议的 plain text

只有在无法恢复、无法提取最小结果时才判定为 expert failure。

为什么：
- 现在严格的 `emit_expert_result` 协议让模型格式偏移直接变成系统失败。
- 半结构化协议的目的本来是便于通信，不是限制模型能力。

备选方案：
- 继续强制严格 JSON。否决，因为这会把大量“理解正确但格式不完美”的结果误判成失败。

### 6. Summarizer 改为模型总结层，Renderer 改为展示层

Summarizer 负责生成最终结构化结论；Final renderer 只负责：

- 排版
- 分段
- 折叠原始证据
- 脱敏
- 可选 markdown 格式化

Renderer 不再决定 headline、findings 和 recommendations。

为什么：
- 当前 renderer 已经成为第二个 summarizer，导致最终答案并不是模型真正的输出。

备选方案：
- 保留 renderer 的语义组织能力。否决，因为这会继续让代码替模型回答。

### 7. 前端分离“过程 / 结论 / 证据”

前端展示改为三层：

- ThoughtChain：过程
- Final Answer：AI 最终结论
- Raw Evidence：工具输出 / 原始结构化数据（折叠）

为什么：
- 当前 markdown block 与 ThoughtChain 混合显示，用户很难区分哪些是 AI 结论，哪些是执行素材。

备选方案：
- 继续把所有内容压到 markdown block。否决，因为这会持续损失后端 richer result。

## Risks / Trade-offs

- [Risk] 模型不可用时用户会更频繁看到“当前不可用”  
  → Mitigation：在启动期与运行期做更明确的模型健康检查，并在 UI 上明确建议手动操作路径。

- [Risk] 去掉代码语义 fallback 后，短期内会暴露更多真实模型质量问题  
  → Mitigation：为 Rewrite/Planner/Expert/Summarizer 分阶段增加评估指标和回归样本，优先修复模型契约和 prompt，而不是重新加回代码猜测。

- [Risk] 放宽 Expert 结果协议会增加执行层解析复杂度  
  → Mitigation：采用“结构化优先，有限恢复”的策略，只恢复最小结果字段，不做无限制文本猜测。

- [Risk] Frontend 展示 richer result 会增加渲染复杂度  
  → Mitigation：先保留现有 ThoughtChain 组件，只扩展结果块结构，不一次性重写整个聊天 UI。

- [Risk] RAG 接入 Rewrite 后，中间层字段变化会影响现有 Planner  
  → Mitigation：以新增字段为主，先保证 Planner 可兼容旧字段，再逐步切换到新字段。

## Migration Plan

1. 先在 OpenSpec 层定义新的模型优先契约和 RAG bridge 契约。
2. 收缩 Rewrite/Planner/Summarizer 的 normalize/fallback 行为，保留日志与指标，暂不立即删除全部旧代码路径。
3. 引入分层显式不可用语义，并更新 SSE 事件与前端提示。
4. 重构 Expert 结果协议为“结构化优先，半结构化可恢复”。
5. 将 Final renderer 降级为展示层，同时更新前端 block 渲染结构。
6. 最后删除不再使用的语义 fallback 代码和模板骨架。

回滚策略：
- 保留 feature flag，允许在灰度期间回退到现有实现。
- 任何阶段若模型稳定性不足，可先回滚到“旧链路 + 新提示文案”，但不应再新增更多代码语义 fallback。

## Open Questions

- Rewrite 输出中面向 RAG 的字段最终是否需要独立子对象，例如 `retrieval`，还是保持扁平字段即可？
- Expert 半结构化恢复解析是否需要统一抽象为共享恢复器，还是由 executor 单独处理？
- 前端是否需要把原始执行证据持久化为独立 block，而不是继续挂在 ThoughtChain detail 下？
