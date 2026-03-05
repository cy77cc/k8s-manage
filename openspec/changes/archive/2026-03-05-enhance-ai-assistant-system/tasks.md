# Tasks: AI助手系统优化

## 阶段1: Eino框架废弃方法迁移 (3天)

### 1.1 Graph Runner 增强
- [x] 扩展 `internal/ai/graph/types.go` 添加流式执行类型
- [x] 修改 `internal/ai/graph/builder.go` 添加 `BuildStream` 方法
- [x] 完善 `internal/ai/graph/runners.go` 实现流式执行器
- [x] 添加 `internal/ai/graph/stream_test.go` 单元测试

### 1.2 移除废弃代码
- [x] 删除 `internal/ai/experts/orchestrator.go`
- [x] 删除 `internal/ai/experts/orchestrator_test.go`
- [x] 删除 `internal/ai/experts/orchestrator_primary_led_test.go`
- [x] 更新所有引用 Orchestrator 的代码

### 1.3 PlatformAgent 重构
- [x] 移除 `platform_agent.go` 中的 orchestrator 字段
- [x] 统一使用 graphRunner 和 streamRunner
- [x] 添加回退到默认 Agent 的逻辑
- [x] 更新集成测试

### 1.4 React Agent 配置更新
- [x] 更新 `internal/ai/experts/registry.go` 的 buildAgent 方法
- [x] 使用 `react.WithTools` 选项函数
- [x] 确保所有配置项有对应实现
- [x] 验证 Agent 功能不变

---

## 阶段2: 确认-执行审批机制 (4天)

### 2.1 数据模型
- [x] 创建 `internal/model/ai_confirmation.go` 定义 ConfirmationRequest
- [x] 创建 `internal/model/ai_approval.go` 定义 AIApprovalTicket
- [x] 添加数据库迁移脚本
- [x] 创建 `ai_confirmations` 和 `ai_approval_tickets` 表

### 2.2 操作预览生成
- [x] 创建 `internal/service/ai/preview_builder.go`
- [x] 实现 BuildPreview 方法生成操作预览
- [x] 实现 extractTargetResources 提取目标资源
- [x] 实现 generateImpactScope 生成影响范围描述
- [x] 为部署类操作实现差异预览（可选）

### 2.3 用户确认服务
- [x] 创建 `internal/service/ai/confirmation_service.go`
- [x] 实现 RequestConfirmation 请求用户确认
- [x] 实现 WaitForConfirmation 等待确认结果
- [x] 实现 Confirm/Cancel 确认/取消操作
- [x] 实现确认超时机制

### 2.4 权限检查服务
- [x] 创建 `internal/service/ai/permission_checker.go`
- [x] 实现 ToolResourceMapping 工具-资源映射表
- [x] 实现 CheckPermission 方法
- [x] 实现 FindApprovers 方法（查找审批人）
- [x] 添加单元测试

### 2.5 审批通知服务
- [x] 创建 `internal/service/ai/approval_notifier.go`
- [x] 集成 WebSocket Hub
- [x] 实现 NotifyApprovers 推送审批通知
- [x] 实现审批超时监控
- [x] 添加重连支持

### 2.6 风险分级配置
- [x] 创建 `configs/approval_config.yaml`
- [x] 实现风险分级逻辑（low/medium/high）
- [x] 实现配置加载和验证
- [x] 支持工具级别的风险和超时配置

### 2.7 集成到工具执行流程
- [x] 修改 `internal/service/ai/policy.go` 实现"确认-执行"流程
- [x] 修改 `internal/service/ai/chat_handler.go` 处理确认/审批事件
- [x] 添加 confirmation_required SSE 事件
- [x] 添加 approval_required SSE 事件
- [x] 添加确认和审批超时错误处理

---

## 阶段3: RAG知识库 (5天)

### 3.1 Milvus 集成
- [x] 添加 Milvus 客户端依赖到 go.mod
- [x] 创建 `internal/rag/milvus_client.go` 连接管理
- [x] 实现健康检查和重连机制
- [x] 配置环境变量支持

### 3.2 Collection Schema
- [x] 创建 `internal/rag/milvus_schema.go`
- [x] 定义 tool_examples Collection
- [x] 定义 platform_assets Collection
- [x] 定义 troubleshooting_cases Collection
- [x] 实现自动创建 Collection 逻辑

### 3.3 Embedding 服务
- [x] 创建 `internal/rag/embedder.go`
- [x] 支持 OpenAI text-embedding-3-small
- [x] 添加本地模型支持（可选）
- [x] 实现批量化 Embedding

### 3.4 数据摄入服务
- [x] 创建 `internal/rag/ingestion.go`
- [x] 实现 IngestToolExamples 方法
- [x] 实现 IngestPlatformAssets 方法
- [x] 实现 IngestTroubleshootingCases 方法
- [x] 添加增量更新支持

### 3.5 检索服务
- [x] 创建 `internal/rag/retriever.go`
- [x] 实现 Retrieve 方法（多 Collection 并行）
- [x] 实现 BuildAugmentedPrompt 方法
- [x] 添加结果去重和排序

### 3.6 定时任务
- [x] 创建 `internal/rag/scheduler.go`
- [x] 配置工具示例增量更新任务（每小时）
- [x] 配置平台资产全量同步任务（每天）
- [x] 添加任务监控和告警

### 3.7 AI 集成
- [x] 修改 `platform_agent.go` 集成 RAG 检索
- [x] 在消息处理前执行知识检索
- [x] 将检索结果注入 Prompt

---

## 阶段4: SKILLS支持 (3天)

### 4.1 技能配置
- [x] 创建 `configs/skills.yaml` 示例配置
- [x] 定义至少 3 个常用技能（deploy_service, diagnose_host, batch_exec）

### 4.2 技能注册表
- [x] 创建 `internal/ai/skills/registry.go`
- [x] 定义 Skill、SkillParameter、SkillStep 结构体
- [x] 实现 LoadSkills 配置加载
- [x] 实现 MatchSkill 匹配逻辑
- [x] 添加配置验证

### 4.3 技能执行器
- [x] 创建 `internal/ai/skills/executor.go`
- [x] 实现 Execute 主流程
- [x] 实现 executeToolStep
- [x] 实现 executeApprovalStep
- [x] 实现 executeResolverStep（可选）

### 4.4 参数处理
- [x] 创建 `internal/ai/skills/params.go`
- [x] 实现 validateParams 参数验证
- [x] 实现 extractParams 参数提取
- [x] 实现 renderParamsTemplate 模板渲染
- [x] 支持 enum_source 解析

### 4.5 PlatformAgent 集成
- [x] 在 PlatformAgent 添加 skillRegistry 和 skillExecutor
- [x] 修改 Stream 方法优先匹配技能
- [x] 实现 executeSkillStream 流式执行
- [x] 添加技能执行结果格式化

### 4.6 API 支持
- [x] 创建 `internal/service/ai/skill_handler.go`
- [x] 添加 GET /ai/skills 接口（列出所有技能）
- [x] 添加 POST /ai/skills/:name/execute 接口（直接执行技能）
- [x] 添加 POST /ai/skills/reload 接口（重载配置）

---

## 测试与验收

### 单元测试
- [x] Graph Runner 流式执行测试
- [x] 权限检查逻辑测试
- [x] 审批超时机制测试
- [x] Milvus 检索测试
- [x] 技能匹配和执行测试

### 集成测试
- [ ] AI 对话端到端测试
- [ ] 审批流程端到端测试
- [ ] RAG 知识增强测试
- [ ] 技能执行端到端测试

### 性能测试
- [ ] Graph Runner 性能对比测试
- [ ] Milvus 检索延迟测试
- [ ] 高并发审批请求测试

---

## 部署清单

### 基础设施
- [x] Milvus 部署（Docker Compose 或 Kubernetes）
- [x] Embedding 服务配置

### 配置文件
- [x] configs/approval_config.yaml
- [x] configs/skills.yaml
- [x] Milvus 连接配置

### 数据库迁移
- [x] ai_approval_tickets 表创建

### 环境变量
- [x] MILVUS_HOST
- [x] MILVUS_PORT
- [x] EMBEDDING_MODEL / OPENAI_API_KEY
