## Context

当前 AI 模块已经具备 `internal/ai` 编排宿主、SSE 输出、审批恢复、会话持久化和前端抽屉渲染能力，但三个关键边界仍然错位：

- 后端以“阶段完成后给稳定结果”为主，而不是以“turn 持续生长”为主。
- 前端以 `content + thinking + tools` 的单消息模型消费流，而不是消费结构化 turn/block 生命周期。
- 持久化层以 `ai_chat_messages` 覆盖写整条 assistant 消息为主，无法自然表达局部流式、审批中断、恢复续写和历史回放。

这次变更跨越 `internal/ai`、`internal/service/ai`、`internal/ai/state`、`internal/ai/runtime`、`internal/model`、`storage/migration` 与 `web/src/components/AI/**`，同时涉及 SSE 协议兼容、数据库迁移和前端交互重构，属于典型的跨域架构改造。

## Goals / Non-Goals

**Goals:**
- 将一次 assistant 响应重构为稳定的 `turn -> blocks` 生命周期模型，并使同一 turn 能从开始、审批中断到恢复执行持续增长。
- 为后端建立 turn/block 投影层，使 rewrite、plan、execute、summary 的阶段结果能够以用户可消费的 block 语义流式输出。
- 为前端建立 reducer 驱动的 turn/block 状态机，实现真正的流式打字机效果、工具卡片、审批卡片和证据卡片。
- 为数据库建立 turn/block 持久化模型，使流式历史、审批恢复和消息回放与实时体验一致。
- 在 rollout 期间保留现有 `meta/delta/tool_call/...` SSE 兼容事件，避免一次性破坏旧前端逻辑。
- 固定 resume、session replay、rollout flag 的对外契约，避免实现阶段再次分叉。

**Non-Goals:**
- 不在本次变更中重写整个 ADK 编排核心为全新的 graph-native runtime。
- 不在本次变更中重新设计专家能力、工具权限或领域执行路由本身。
- 不默认向普通用户展示完整 reasoning、原始 tool JSON 或 planner 内部中间态。
- 不要求本次变更完成所有 UI 美术重构，本次重点是交互模型、信息结构和流式体验。

## Decisions

### 1. 用 `turn/block` 替代“单条 assistant 文本消息”作为主渲染对象

**Decision**

将 assistant 输出抽象为一个 `turn`，其中包含多个可独立更新的 `block`。典型 block 类型包括：

- `status`
- `plan`
- `tool`
- `approval`
- `evidence`
- `thinking`
- `text`
- `error`

前端不再把一个 assistant turn 只理解为 `content` 字段，而是按 block 顺序和状态渲染。

**Why**

- `delta` 只适合最终正文，不足以表达计划、工具、审批和证据这些非文本 UI 语义。
- 同一轮 turn 在审批恢复后继续生长时，单条文本消息模型会丢失局部上下文和状态定位。
- block 模型更适合 progressive disclosure，能够区分默认展示与按需展开的信息。

**Alternatives considered**

- 继续使用现有 `ChatMessage`，在 `metadata_json` 中不断塞更多字段。
  - 否决原因：会让前端继续依赖弱类型对象和条件渲染，复杂度继续外溢。
- 只在前端引入 block 概念，后端保持当前事件结构不变。
  - 否决原因：前端仍需大量猜测事件语义，审批恢复和回放一致性难以保证。

### 2. 保留 SSE 兼容层，但新增 turn/block 原生事件

**Decision**

SSE 协议采用“双轨制”：

- 兼容轨：继续发 `meta`、`delta`、`thinking_delta`、`tool_call`、`tool_result`、`approval_required`、`done`、`error`
- 原生轨：新增 `turn_started`、`block_open`、`block_delta`、`block_replace`、`block_close`、`turn_state`、`turn_done`

兼容轨在 rollout 期间继续服务现有消费者，原生轨作为新的语义源。

**Why**

- 当前前端和部分测试已经绑定旧事件族，直接切换会放大风险。
- 原生轨能让 turn/block UI 与持久化模型严格对齐，避免兼容轨表达力不足。
- 保留兼容轨有助于分阶段灰度前端改造。

**Alternatives considered**

- 直接删除旧事件，全面切到 block 事件。
  - 否决原因：风险过高，且会阻断平滑迁移。
- 继续只发旧事件，在前端猜出 block。
  - 否决原因：无法解决语义归属、恢复续写和历史回放问题。

### 3. 引入 Turn Projector，隔离 AI 控制平面与 UI 语义投影

**Decision**

在 `internal/ai` 中新增 turn projector 层：

- 编排层输出领域结果：rewrite 输出、planner decision、execution record、summary output
- projector 负责把这些结果映射为 turn/block 事件和兼容 SSE 事件
- gateway 仍只负责 transport framing 和 auth/session shell

**Why**

- 避免 handler 或 orchestrator 直接拼用户文案与 UI 语义，降低耦合。
- 让同一份领域结果既可投影为新 block 事件，也可投影为旧兼容事件。
- 便于未来替换编排细节而不重做前端协议。

**Alternatives considered**

- 继续在 `orchestrator.go` 中直接发所有 UI 事件。
  - 否决原因：阶段逻辑、领域状态和 UI 语义已经开始缠绕，后续只会更难维护。

### 4. 数据库新增 turn/block 表，旧 message 表在兼容期保留

**Decision**

新增至少三张表：

- `ai_chat_turns`
- `ai_chat_blocks`
- `ai_chat_events`（轻量事件日志，可选查询，不要求每次回放都依赖它）

现有 `ai_chat_sessions` 继续保留。
现有 `ai_chat_messages` 在兼容期继续写入，但逐步退化为兼容投影，而非唯一事实来源。

**Why**

- 单表覆盖写无法表达一个 turn 的多个 block 及局部更新。
- turn/block 存储能自然承接审批、恢复、部分完成、局部失败和回放。
- `ai_chat_events` 可用于调试、补偿和回放校验，但不要求前端直接依赖事件表。

**Alternatives considered**

- 继续只用 `ai_chat_messages.metadata_json` 嵌套所有 block。
  - 否决原因：查询、迁移、索引和部分更新都很差，后续成本更高。
- 完全事件溯源，只存事件不存快照。
  - 否决原因：对当前系统来说过重，恢复 UI 需要大量重放，实施成本不必要。

### 5. Redis 执行态增加 turn 关联，resume 必须继续原 turn

**Decision**

`internal/ai/runtime.ExecutionState` 增加 turn 关联字段，如：

- `turn_id`
- `active_block_ids`
- `resume_stream_token`（如需要）

审批恢复后继续更新原 assistant turn，而不是返回一个孤立 JSON 结果或开启新消息。

**Why**

- 当前 resume 语义只能恢复执行，不保证前端展示续在同一个对象上。
- 将 turn identity 带入执行态后，resume 与 UI 生命周期才是一条链路。

**Alternatives considered**

- 仅靠 `session_id + step_id` 在恢复时让前端自己猜当前消息。
  - 否决原因：前端无法可靠定位正确的渲染对象，尤其在多轮并发和刷新恢复场景下。

### 6. 恢复流使用独立的流式端点，旧 JSON 恢复接口保留兼容

**Decision**

保持现有 `/api/v1/ai/resume/step` JSON 恢复接口不变，用于兼容旧前端与纯控制流调用；新增 `/api/v1/ai/resume/step/stream` 作为 turn/block 模型下的标准恢复续流端点。

**Why**

- 现有桥接规范要求已存在的 gateway 契约稳定，直接把 `/resume/step` 从 JSON 改成 SSE 风险过高。
- 恢复后的用户体验需要继续流式输出，单独的 stream 端点能避免在同一接口上混用两种传输模式。
- 新端点让前端能清晰地区分“提交审批决策”和“继续接收这次 turn 的续流”。

**Alternatives considered**

- 直接把 `/api/v1/ai/resume/step` 改成 SSE。
  - 否决原因：会破坏现有调用语义，与稳定桥接要求冲突。
- 在 `/api/v1/ai/resume/step` 上通过 header 或 query 协商 JSON/SSE。
  - 否决原因：语义含混，测试与客户端实现复杂度更高。

### 7. 固定 rollout 配置名和默认行为

**Decision**

新增配置 `ai.use_turn_block_streaming`，默认 `false`。开启后：

- 后端开始产出 turn/block 原生事件并写入 turn/block 持久化模型
- 前端允许切换到 turn/block renderer
- 兼容 SSE 事件与旧 message 投影仍保留

关闭时保留当前消息模型与兼容事件路径。

**Why**

- 任务里已经要求灰度能力，但没有固定配置名会导致实现随意发明 flag。
- 与现有 `ai.use_multi_domain_arch` 一样，turn/block 体验改造也需要独立 rollout 面。

**Alternatives considered**

- 复用 `ai.use_multi_domain_arch`。
  - 否决原因：运行时编排切换与 UI/持久化切换不是同一维度。
- 不做配置，直接一次性切换。
  - 否决原因：前后端、SSE、存储同步改动太大，不适合硬切。

### 8. 打字机效果只对真正需要连续阅读的文本开启

**Decision**

渲染策略分层：

- `text` block：使用真实 token/chunk 流 + 前端微缓冲实现打字机
- `status` block：句子级或状态级更新
- `plan` block：步骤级更新
- `tool` / `approval` / `evidence` block：卡片状态更新，不做逐字动画
- `thinking` block：默认折叠，仅在需要时展开

**Why**

- 所有文本都逐字打印会制造噪音，降低信息密度。
- 用户真正期待强打字机体验的是“最终回答”，而不是内部状态标签。

**Alternatives considered**

- 对所有块统一逐字流。
  - 否决原因：视觉噪音大，执行状态卡片会变得迟滞且难读。

### 9. 用户可见内容采用默认展示、折叠展示、调试专用三层策略

**Decision**

前端消息块默认按三层信息密度展示：

- 默认展示：阶段状态、计划摘要、工具关键动作、审批卡片、最终回答、关键证据摘要
- 折叠展示：步骤细节、原始证据片段、扩展计划详情
- 调试专用：完整 thinking、原始 tool arguments/result JSON、内部中间输出

普通模式默认不自动展开 `thinking`，也不直接展示原始 tool JSON。

**Why**

- 提升体验的关键不是“显示更多”，而是“显示恰当”。
- 如果不把默认展示策略写死，实现时很容易把内部 agent 输出直接暴露给用户。

**Alternatives considered**

- 让前端在实现阶段自行决定展示策略。
  - 否决原因：会导致行为漂移，无法测试，也难以统一产品体验。

### 10. 展示模式由抽屉内显式用户偏好驱动，而不是由 rollout flag 隐式决定

**Decision**

AI 抽屉必须提供明确的展示模式切换，至少包含：

- `normal`：默认用户模式
- `debug`：显式开启的深度信息模式

模式来源规则如下：

- 初次进入抽屉时默认使用 `normal`
- 用户只能通过抽屉内明确的显示设置切换到 `debug`
- 当前模式必须持久化为前端本地用户偏好，以便下次打开同一抽屉时复用
- `ai.use_turn_block_streaming` 只控制 turn/block 管线是否启用，不直接决定用户看到 `normal` 还是 `debug`

**Why**

- “展示模式”和“功能 rollout”属于两个维度，混在一起会让开关语义失真。
- 如果不固定模式来源，前端实现时就会出现 URL、环境变量、开发态开关等多种不一致入口。

**Alternatives considered**

- 把 `debug` 视为开发环境专属模式。
  - 否决原因：运维和专家用户在生产环境也可能需要查看证据和调试细节。
- 让模式完全由后端配置决定。
  - 否决原因：这是纯前端信息密度选择，不应依赖后端发布开关。

### 11. 动效与交互必须服从可访问性、触控和 reduced-motion 约束

**Decision**

前端 turn/block 渲染器必须满足以下交互边界：

- 所有审批按钮、展开按钮、跳转按钮和工具卡片交互入口必须支持键盘聚焦和触发
- 主要可点击目标在触控模式下必须达到最小触控尺寸
- 活跃 assistant turn 的增量更新必须对辅助技术保持可感知，但不能因高频 chunk 更新造成读屏噪音失控
- 当系统启用 `prefers-reduced-motion` 时：
  - 关闭逐字打字机动画，直接按 chunk 追加文本
  - 禁用非必要的平滑滚动和卡片过渡动画
  - 保留状态变化和内容追加本身，不以视觉动画作为唯一反馈

**Why**

- 这次改造引入更多动态内容和卡片交互，如果不把 a11y/touch/reduced-motion 写进设计，最终只会得到“看起来更活跃，但更难操作”的 UI。

**Alternatives considered**

- 先只做视觉体验，后续再补无障碍。
  - 否决原因：流式和卡片化 UI 的结构一旦定型，再补无障碍和交互边界成本更高。

### 12. 流式聊天采用“条件自动跟随 + 明确回到底部”的滚动策略

**Decision**

AI 抽屉的滚动行为采用条件自动跟随：

- 当用户仍停留在列表底部附近时，当前活跃 turn 的增量更新可自动跟随到底部
- 当用户主动上滑离开底部时，后续流式 block 更新不得强制抢回滚动位置
- 用户离开底部期间，抽屉必须显示明确的“跳转到最新”入口
- 用户点击“跳转到最新”或重新滚回底部附近后，自动跟随恢复
- 在 reduced-motion 模式下，自动跟随应使用非平滑方式，避免额外运动感

**Why**

- turn/block 模型下更新频率显著提高，沿用“每次消息变化都 smooth scroll 到底”的策略会引发明显抖动和抢焦点。

**Alternatives considered**

- 每次 block 更新都强制滚到底部。
  - 否决原因：会打断用户查看上文、工具结果和审批卡片，体验显著变差。

## Risks / Trade-offs

- [兼容期事件双写复杂度上升] → 通过 projector 集中产生原生轨与兼容轨，避免各阶段手工重复发事件。
- [数据库迁移期间旧接口和新接口并存] → 先新增表和双写，再切读路径，最后再考虑收缩旧表职责。
- [前端状态机重构可能引入历史会话回放差异] → 先实现 turn/block 读取适配层，将旧会话投影为最小 block 集，再逐步切换真实新表读取。
- [resume 续流改造影响审批路径稳定性] → 先保留现有 `/resume/step` JSON 兼容接口，再新增 stream-resume 路径做灰度。
- [把更多中间态展示给用户可能造成噪音] → 明确默认展示与折叠展示策略，只暴露用户真正需要的状态、证据和确认动作。
- [高频流式更新可能造成滚动抖动和读屏噪音] → 采用条件自动跟随、明确的 jump-to-latest 入口，以及 reduced-motion / 辅助技术专用降噪策略。

## Migration Plan

1. 新增数据库表与 migration，保持旧表不变。
2. 在 `internal/ai` 引入 turn projector，后端开始双发兼容 SSE 与 turn/block SSE。
3. 在 `chatRecorder` 与存储层开始双写：
   - 旧 `ai_chat_messages`
   - 新 `ai_chat_turns` / `ai_chat_blocks`
4. 前端新增 turn/block reducer 和 renderer，通过 feature flag 或兼容适配逐步接管 AI 抽屉。
5. 新增 `/api/v1/ai/resume/step/stream`，使审批恢复继续原 turn，同时保留现有 `/api/v1/ai/resume/step` JSON 兼容接口。
6. 更新 `api/ai/v1` 会话契约，新增 turn replay 字段并保留旧形态兼容字段。
7. 会话详情接口优先从新表构建响应，保留旧形态兼容字段。
8. 观察稳定后，收缩旧 message 覆盖写路径，将其降级为兼容投影或历史兼容读取。

**Rollback**

- 关闭 turn/block 前端入口，回退到现有 `useAIChat` 文本消息消费路径。
- 后端继续保留旧 SSE 兼容轨，不影响既有 chat 基本可用性。
- 关闭 `ai.use_turn_block_streaming` 后，新的 stream-resume 和 turn/block 读取路径停止对外暴露或停止被前端使用。
- 数据库新增表可保留但不切读路径；双写异常时可只保留旧表写入。

## Open Questions

- `ai_chat_events` 是否需要作为正式查询接口暴露给调试视图，还是仅保留内部运维用途。
- 历史旧消息是否需要一次性回填为 turn/block 结构，还是按读取时实时投影即可。
