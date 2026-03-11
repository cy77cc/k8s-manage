## MODIFIED Requirements

### Requirement: Rewrite Stage For Colloquial Input

系统 MUST 在正式规划前保留并强化 `Rewrite` 阶段，用于将用户口语化输入转换为稳定、可执行、可检索增强的任务表达。

`Rewrite` 输出 MUST 采用半结构化协议：

- 结构化字段用于下游 Planner 与 RAG
- narrative 用于解释和消歧
- retrieval 相关字段用于知识检索增强

`Rewrite` MUST NOT 在模型失败时退回到代码驱动的语义理解主路径。

#### Scenario: colloquial request is normalized before planning and retrieval
- **GIVEN** 用户请求 “帮我看看 payment-api 最近是不是有点慢，顺便查下是不是刚发版”
- **WHEN** Rewrite 阶段处理该输入
- **THEN** 系统 MUST 输出归一化后的调查目标
- **AND** MUST 提取 `service_name=payment-api` 作为资源 hint
- **AND** MUST 将任务模式标记为 `investigate`
- **AND** MUST 产出适合后续 RAG 的检索语义字段
- **AND** MUST 输出解释该改写含义的 `narrative`

#### Scenario: rewrite stage unavailable is reported explicitly
- **GIVEN** Rewrite runner 或模型不可用
- **WHEN** 用户发起新的 AI 请求
- **THEN** 系统 MUST 明确告知 AI 理解阶段当前不可用
- **AND** MUST NOT 使用纯代码构造替代性 normalized request 继续执行

### Requirement: Planner MUST Use Semi-Structured Output

Planner MUST 通过半结构化协议向 Orchestrator 输出决策，而不是依赖自然语言正文解析。

Planner 输出中的结构化字段用于运行时执行，`narrative` 用于解释和消歧。

运行时代码 MAY 校验 Planner 输出的 schema 和安全边界，但 MUST NOT 改写 Planner 的语义决策字段，例如：

- expert
- intent
- task
- recommendation semantics
- conclusion semantics

#### Scenario: planner emits structured plan with narrative
- **GIVEN** Rewrite 已将用户目标规整为调查任务
- **WHEN** Planner 完成资源解析与规划
- **THEN** Planner MUST 输出结构化 `ExecutionPlan`
- **AND** MUST 为计划补充自然语言 `narrative`
- **AND** MUST 为每个 step 提供机器可消费字段
- **AND** 运行时代码 MUST NOT 将一个语义有效的 plan 重写成新的 plan

#### Scenario: planner output fails explicit validation
- **GIVEN** Planner 输出结构不合法或缺少安全执行所必需的前置字段
- **WHEN** Orchestrator 校验该输出
- **THEN** 系统 MUST 明确标记该 planning 结果不可执行
- **AND** MUST NOT 使用代码自动改写其 expert、task、intent 或 step semantics 来继续执行

### Requirement: Summary MUST Be Separate From Process Display

总结阶段 MUST 与过程展示分离，并由模型主导最终结论内容。

- ThoughtChain 用于展示过程
- Summarizer 用于生成最终结论
- Final renderer 仅负责展示，不负责替换总结语义

系统 MUST NOT 以代码模板骨架作为最终结论主路径。

#### Scenario: summarizer drives final answer semantics
- **GIVEN** Executor 已完成步骤并生成执行证据
- **WHEN** Summarizer 生成最终结构化结论
- **THEN** 最终回答的语义 MUST 以 Summarizer 输出为准
- **AND** Renderer MUST 仅负责格式化展示

#### Scenario: summarizer unavailable falls back explicitly
- **GIVEN** 执行已经完成但 Summarizer 当前不可用
- **WHEN** 系统需要生成最终回答
- **THEN** 系统 MUST 明确告知 AI 总结阶段当前不可用
- **AND** MUST 提供原始执行结果或证据供用户查看
- **AND** MUST NOT 自动生成“已完成步骤 N 个 / 查看最终结论正文”一类模板化总结来冒充最终回答
