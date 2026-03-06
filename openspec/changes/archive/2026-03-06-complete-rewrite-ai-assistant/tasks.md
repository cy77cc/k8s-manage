# Tasks: AI 助手完全重构

## Phase 1: 基础架构 (3 天)

### 1.1 后端路由简化
- [x] 创建新的 `internal/service/ai/v2/` 目录结构
- [x] 实现 6 个核心端点：
  - `POST /chat` - SSE 流式对话
  - `POST /chat/respond` - 响应确认请求
  - `GET /sessions` - 会话列表
  - `GET /sessions/:id` - 会话详情
  - `DELETE /sessions/:id` - 删除会话
  - `GET /tools` - 工具列表
- [x] 添加 feature flag 控制新旧版本切换
- [x] 保持旧路由兼容，标记为 deprecated

### 1.2 存储层重构
- [x] 创建数据库模型 `AIChatSession`, `AIChatMessage`
- [x] 实现 `SessionStore` (PostgreSQL + Redis 缓存)
- [x] 实现 `CheckPointStore` (Redis)
- [x] 添加数据库迁移脚本
- [x] 移除旧的内存 `memoryStore`

### 1.3 Agent 框架搭建
- [x] 实现 `IntentClassifier` 意图分类器
- [x] 实现 `SimpleChatMode` 简单问答模式
- [x] 实现 `AgenticMode` Agent 执行模式
- [x] 实现 `HybridAgent` 混合入口
- [x] 编写单元测试

---

## Phase 2: 前端重写 (4 天)

### 2.1 项目结构
- [x] 创建 `web/src/pages/AIChat/` 目录
- [x] 安装 `@ant-design/x` 及相关依赖
- [x] 配置 TypeScript 类型定义

### 2.2 核心组件开发
- [x] `ChatPage` - 页面容器
- [x] `ConversationSidebar` - 会话列表侧边栏
  - [ ] `ConversationList` - 会话列表
  - [ ] `NewChatButton` - 新建会话按钮
- [x] `ChatMain` - 主聊天区域
  - [ ] `MessageList` - 消息列表
  - [ ] `UserBubble` - 用户消息气泡
  - [ ] `AssistantBubble` - AI 消息气泡
  - [ ] `MarkdownContent` - Markdown 渲染
  - [ ] `ToolExecutionCard` - 工具执行卡片
  - [ ] `ConfirmationPanel` - 交互式确认面板
  - [ ] `InputArea` - 输入区域

### 2.3 Hooks 开发
- [x] `useChatSession` - 会话状态管理
- [x] `useSSEConnection` - SSE 连接管理
- [x] `useConfirmation` - 确认交互状态

### 2.4 样式和交互
- [x] 定义主题变量
- [x] 实现响应式布局
- [x] 添加加载状态和动画
- [x] 实现键盘快捷键

---

## Phase 3: 工具重构 (2 天)

### 3.1 工具注册表
- [x] 实现 `ToolRegistry` 工具注册表
- [x] 定义 `ToolDefinition` 结构
- [x] 实现工具分类（K8s/Host/Service/Monitor/MCP）

### 3.2 K8s 工具实现
- [x] `k8s_query` - 统一资源查询
  - 合并 `k8s_list_pods`, `k8s_list_services`, `k8s_list_deployments`, `k8s_list_nodes`
- [x] `k8s_logs` - Pod 日志查询
- [x] `k8s_events` - K8s 事件查询

### 3.3 主机工具实现
- [x] `host_exec` - 单机命令执行
- [x] `host_batch` - 批量主机操作（高风险）

### 3.4 服务工具实现
- [x] `service_deploy` - 服务部署（高风险）
- [x] `service_status` - 服务状态查询

### 3.5 监控工具实现
- [x] `monitor_alert` - 告警查询
- [x] `monitor_metric` - 指标查询

### 3.6 MCP 适配
- [x] 重构 `MCPClientManager`
- [x] 实现 MCP 工具代理
- [x] 添加 MCP 工具前缀处理

---

## Phase 4: 集成测试 (2 天)

### 4.1 单元测试
- [x] Agent 模块测试
  - [x] `IntentClassifier` 测试
  - [x] `HybridAgent` 测试
  - [x] `SimpleChatMode` 测试
  - [x] `AgenticMode` 测试
- [x] 存储层测试
  - [x] `SessionStore` 测试
  - [x] `CheckPointStore` 测试
- [x] 工具测试
  - [x] K8s 工具测试
  - [x] 主机工具测试
  - [x] 服务工具测试
  - [x] 监控工具测试

### 4.2 集成测试
- [x] SSE 流式响应测试
- [x] 交互式确认流程测试
- [x] 会话持久化测试
- [x] MCP 工具调用测试

### 4.3 端到端测试
- [x] 用户对话流程测试
- [x] 工具执行流程测试
- [x] 审批确认流程测试

### 4.4 性能测试
- [x] 响应时间测试
- [x] 并发测试
- [x] 内存使用测试

---

## Phase 5: 清理和文档 (1 天)

### 5.1 代码清理
- [x] 移除旧的 Agent 实现
- [x] 移除旧的前端组件
- [x] 清理未使用的路由
- [x] 移除 deprecated 标记的代码

### 5.2 文档更新
- [x] API 文档更新
- [x] 组件文档更新
- [x] 部署文档更新
- [x] 迁移指南

---

## 任务依赖关系

```
Phase 1 (基础架构)
├─ 1.1 路由简化 ─────┐
├─ 1.2 存储层重构 ───┼──▶ Phase 3 (工具重构)
└─ 1.3 Agent 框架 ──┘         │
                               ▼
Phase 2 (前端重写) ◀────────────┘
│
▼
Phase 4 (集成测试)
│
▼
Phase 5 (清理文档)
```

## 风险和阻塞点

| 风险 | 缓解措施 | 负责人 |
|------|----------|--------|
| Ant Design X 组件不满足需求 | 预先调研，准备自定义组件方案 | 前端 |
| 意图分类准确率低 | 添加 fallback 机制，支持用户手动切换 | 后端 |
| 工具合并影响现有功能 | 保留旧工具 API，标记 deprecated | 后端 |
| 测试覆盖率不足 | 边开发边写测试，设置覆盖率门槛 | 全员 |

## 验收标准

- [x] 所有单元测试通过
- [x] 集成测试通过
- [ ] 代码覆盖率 > 80%
- [x] 前端包大小 < 200KB (gzip)
- [x] 简单问答响应时间 < 1s
- [x] Agent 执行首次响应 < 2s
- [x] 交互式确认流程正常工作
- [x] 所有核心工具功能正常
