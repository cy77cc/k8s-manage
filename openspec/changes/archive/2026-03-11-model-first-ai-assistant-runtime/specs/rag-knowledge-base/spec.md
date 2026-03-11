## MODIFIED Requirements

### Requirement: 知识检索增强

系统 SHALL 在 AI 对话时基于 Rewrite 的结构化语义结果自动检索相关知识并增强 Prompt。

RAG 检索输入 MUST 优先使用 Rewrite 输出中的：

- retrieval intent
- normalized goal
- normalized targets
- resource hints
- retrieval queries
- retrieval keywords
- knowledge scope

系统 MUST NOT 让 RAG 与 Planner 各自独立重写原始用户输入并产生冲突语义。

#### Scenario: 自动知识注入
- **GIVEN** 用户发送消息
- **WHEN** AI 准备处理请求
- **THEN** 系统 MUST 先消费 Rewrite 语义结果
- **AND** MUST 基于该结果自动检索相关知识
- **AND** MUST 将检索结果注入到后续 Prompt 中

#### Scenario: 多 Collection 并行检索
- **GIVEN** 需要检索多个知识源
- **WHEN** 执行检索操作
- **THEN** 系统 MUST 基于统一 Rewrite 语义并行检索所有 Collection
- **AND** 检索结果 MUST 在合理时间内合并

#### Scenario: 检索结果去重
- **GIVEN** 检索返回多个相似结果
- **WHEN** 构建增强 Prompt
- **THEN** 系统 MUST 去除重复或高度相似的结果
- **AND** 保留结果 MUST 具有多样性

#### Scenario: rewrite unavailable blocks semantic retrieval
- **GIVEN** Rewrite 阶段不可用或返回无效结构化结果
- **WHEN** 系统尝试执行知识检索增强
- **THEN** 系统 MUST 明确标记当前无法进行可靠的语义检索
- **AND** MUST NOT 退回到与 Planner 无关的代码关键词拼接来伪造检索语义
