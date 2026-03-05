# rag-knowledge-base Specification

## Purpose

构建基于 Milvus 向量数据库的 RAG（检索增强生成）知识库，为 AI 助手提供知识检索能力，提升响应准确率和问题解决效率。

## Requirements

### Requirement: Milvus 向量数据库集成

系统 SHALL 集成 Milvus 作为向量存储和检索引擎。

#### Scenario: Milvus 连接管理
- **GIVEN** 系统启动时
- **WHEN** 初始化 RAG 服务
- **THEN** 系统 MUST 成功连接到 Milvus 实例
- **AND** 系统 MUST 创建所需的 Collections

#### Scenario: Collection 自动创建
- **GIVEN** 目标 Collection 不存在
- **WHEN** 系统首次运行
- **THEN** 系统 MUST 自动创建 Collection
- **AND** Collection MUST 配置正确的向量维度（1536）和索引类型

### Requirement: 工具调用示例 Collection

系统 SHALL 维护工具调用示例的向量化存储。

#### Scenario: 成功案例摄入
- **GIVEN** AI 命令执行成功完成
- **WHEN** 定时摄入任务运行
- **THEN** 系统 MUST 提取工具名称、意图、参数、结果摘要
- **AND** 系统 MUST 将意图文本向量化后存储

#### Scenario: 工具示例检索
- **GIVEN** 用户查询涉及工具使用
- **WHEN** AI 需要确定如何调用工具
- **THEN** 系统 MUST 检索相似的工具调用示例
- **AND** 检索结果 MUST 按相似度排序

#### Scenario: 定时增量更新
- **GIVEN** 新的工具执行记录产生
- **WHEN** 定时任务执行（每小时）
- **THEN** 系统 MUST 增量摄入新的成功案例
- **AND** 系统 MUST 更新已有记录的向量

### Requirement: 平台资产索引 Collection

系统 SHALL 维护平台资源元数据的向量化索引。

#### Scenario: 资产全量同步
- **GIVEN** 定时同步任务触发（每天）
- **WHEN** 系统执行资产同步
- **THEN** 系统 MUST 同步所有主机、服务、集群元数据
- **AND** 系统 MUST 更新或插入资产记录

#### Scenario: 资产名称检索
- **GIVEN** 用户查询涉及资源名称（如"香港服务器"）
- **WHEN** AI 需要确定资源 ID
- **THEN** 系统 MUST 通过向量检索匹配资源
- **AND** 检索结果 MUST 包含资源类型、ID、名称

#### Scenario: 资产状态过滤
- **GIVEN** 用户查询特定状态的资源
- **WHEN** 执行向量检索
- **THEN** 系统 MUST 支持按状态字段过滤
- **AND** 过滤 MUST 在向量检索前执行

### Requirement: 故障排查案例 Collection

系统 SHALL 支持故障排查案例的存储和检索。

#### Scenario: 案例存储
- **GIVEN** 运维人员录入故障案例
- **WHEN** 系统接收案例数据
- **THEN** 系统 MUST 存储标题、症状、诊断过程、解决方案
- **AND** 系统 MUST 将症状文本向量化

#### Scenario: 相似故障检索
- **GIVEN** 用户描述故障现象
- **WHEN** AI 需要诊断建议
- **THEN** 系统 MUST 检索相似的历史故障案例
- **AND** 检索结果 MUST 包含解决方案建议

### Requirement: 知识检索增强

系统 SHALL 在 AI 对话时自动检索相关知识并增强 Prompt。

#### Scenario: 自动知识注入
- **GIVEN** 用户发送消息
- **WHEN** AI 准备处理请求
- **THEN** 系统 MUST 自动检索相关知识
- **AND** 系统 MUST 将知识注入到 Prompt 中

#### Scenario: 多 Collection 并行检索
- **GIVEN** 需要检索多个知识源
- **WHEN** 执行检索操作
- **THEN** 系统 MUST 并行检索所有 Collection
- **AND** 检索结果 MUST 在合理时间内合并

#### Scenario: 检索结果去重
- **GIVEN** 检索返回多个相似结果
- **WHEN** 构建增强 Prompt
- **THEN** 系统 MUST 去除重复或高度相似的结果
- **AND** 保留结果 MUST 具有多样性

## Constraints

- 向量维度 MUST 为 1536（与 Embedding 模型匹配）
- 单次检索延迟 MUST < 500ms
- 定时任务 MUST 不影响主服务性能
- 数据摄入 MUST 支持增量更新

## Dependencies

- Milvus 2.4+ 部署
- Embedding 模型（OpenAI text-embedding-3-small 或本地模型）
- 定时任务调度器
