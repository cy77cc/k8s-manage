## Phase 1: 基础设施 (Week 1)

### Task 1.1: 数据库表迁移
- [x] 创建 `ai_checkpoints` 表迁移文件
- [x] 定义表结构（key, value, created_at, updated_at）
- [x] 编写 Down 迁移脚本
- [x] 本地测试迁移

**文件:**
- `migrations/YYYYMMDDHHMMSS_create_ai_checkpoints_table.up.sql`
- `migrations/YYYYMMDDHHMMSS_create_ai_checkpoints_table.down.sql`

**验收标准:**
- 迁移可正常执行和回滚
- 表结构符合设计

---

### Task 1.2: 实现 DBCheckPointStore
- [x] 创建 `internal/ai/store.go`
- [x] 实现 `compose.CheckPointStore` 接口
- [x] 实现 `Set(ctx, key, value)` 方法
- [x] 实现 `Get(ctx, key)` 方法
- [x] 添加错误处理和日志
- [x] 编写单元测试

**文件:**
- `internal/ai/store.go`
- `internal/ai/store_test.go`

**代码示例:**
```go
type DBCheckPointStore struct {
    db *gorm.DB
}

func NewDBCheckPointStore(db *gorm.DB) compose.CheckPointStore {
    return &DBCheckPointStore{db: db}
}

func (s *DBCheckPointStore) Set(ctx context.Context, key string, value []byte) error {
    return s.db.Clauses(clause.OnConflict{
        Columns:   []clause.Column{{Name: "key"}},
        DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
    }).Create(&model.AICheckPoint{
        Key:       key,
        Value:     value,
        UpdatedAt: time.Now(),
    }).Error
}

func (s *DBCheckPointStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
    var cp model.AICheckPoint
    err := s.db.Where("key = ?", key).First(&cp).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, false, nil
    }
    if err != nil {
        return nil, false, err
    }
    return cp.Value, true, nil
}
```

**验收标准:**
- 单元测试覆盖率 > 90%
- 可正常读写 CheckPoint

---

### Task 1.3: 创建审批包装器
- [x] 创建 `internal/ai/tools/wrapper.go`
- [x] 定义 `ApprovalInfo` 结构体
- [x] 定义 `ApprovalResult` 结构体
- [x] 注册 schema 类型
- [x] 实现 `ApprovableTool` 包装器
- [x] 编写单元测试

**文件:**
- `internal/ai/tools/wrapper.go`
- `internal/ai/tools/wrapper_test.go`

**代码示例:**
```go
type ApprovalInfo struct {
    ToolName        string         `json:"tool_name"`
    ArgumentsInJSON string         `json:"arguments"`
    Risk            ToolRisk       `json:"risk"`
    Preview         map[string]any `json:"preview"`
}

type ApprovalResult struct {
    Approved         bool    `json:"approved"`
    DisapproveReason *string `json:"disapprove_reason,omitempty"`
}

func init() {
    schema.Register[*ApprovalInfo]()
    schema.Register[*ApprovalResult]()
}

type ApprovableTool struct {
    tool.InvokableTool
    risk       ToolRisk
    previewFn  func(ctx context.Context, args string) (map[string]any, error)
}

func (t *ApprovableTool) InvokableRun(ctx context.Context, args string, opts ...tool.Option) (string, error) {
    info, _ := t.Info(ctx)
    wasInterrupted, _, storedArgs := tool.GetInterruptState[string](ctx)

    if !wasInterrupted {
        preview, _ := t.previewFn(ctx, args)
        return "", tool.StatefulInterrupt(ctx, &ApprovalInfo{
            ToolName:        info.Name,
            ArgumentsInJSON: args,
            Risk:            t.risk,
            Preview:         preview,
        }, args)
    }

    _, hasData, result := tool.GetResumeContext[*ApprovalResult](ctx)
    if !hasData {
        return "", fmt.Errorf("missing approval result")
    }
    if !result.Approved {
        if result.DisapproveReason != nil {
            return fmt.Sprintf("审批拒绝: %s", *result.DisapproveReason), nil
        }
        return "审批已拒绝", nil
    }

    return t.InvokableTool.InvokableRun(ctx, storedArgs, opts...)
}
```

**验收标准:**
- 中断流程正确触发
- 恢复流程正确处理
- 单元测试覆盖率 > 90%

---

### Task 1.4: 创建审核包装器
- [x] 定义 `ReviewEditInfo` 结构体
- [x] 定义 `ReviewEditResult` 结构体
- [x] 实现 `ReviewableTool` 包装器
- [x] 编写单元测试

**文件:**
- `internal/ai/tools/wrapper.go` (追加)
- `internal/ai/tools/wrapper_test.go` (追加)

**验收标准:**
- 用户可编辑参数
- 可拒绝执行
- 单元测试覆盖率 > 90%

---

## Phase 2: Agent 重构 (Week 2)

### Task 2.1: 创建新的 Agent 构建器
- [x] 创建 `internal/ai/agent.go`
- [x] 实现 `NewPlatformAgent()` 函数
- [x] 使用 `planexecute.NewPlanner()`
- [x] 使用 `planexecute.NewExecutor()`
- [x] 使用 `planexecute.NewReplanner()`
- [x] 使用 `planexecute.New()` 组合
- [x] 编写集成测试

**文件:**
- `internal/ai/agent.go`
- `internal/ai/agent_test.go`

**代码示例:**
```go
func NewPlatformAgent(ctx context.Context, chatModel model.ToolCallingChatModel, deps tools.PlatformDeps) (adk.Agent, error) {
    // 构建工具集
    allTools, err := tools.BuildAllTools(ctx, deps)
    if err != nil {
        return nil, fmt.Errorf("build tools: %w", err)
    }

    // Planner
    planner, err := planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
        ToolCallingChatModel: chatModel,
    })
    if err != nil {
        return nil, fmt.Errorf("create planner: %w", err)
    }

    // Executor
    executor, err := planexecute.NewExecutor(ctx, &planexecute.ExecutorConfig{
        Model: chatModel,
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: allTools,
            },
        },
    })
    if err != nil {
        return nil, fmt.Errorf("create executor: %w", err)
    }

    // Replanner
    replanner, err := planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
        ChatModel: chatModel,
    })
    if err != nil {
        return nil, fmt.Errorf("create replanner: %w", err)
    }

    // Plan-Execute-Replan Agent
    return planexecute.New(ctx, &planexecute.Config{
        Planner:       planner,
        Executor:      executor,
        Replanner:     replanner,
        MaxIterations: 20,
    })
}
```

**验收标准:**
- Agent 可正常创建
- 简单查询可正常响应
- 多步骤任务可正常执行

---

### Task 2.2: 拆分工具注册代码
- [x] 创建 `internal/ai/tools/ops.go`
- [x] 创建 `internal/ai/tools/k8s.go`
- [x] 创建 `internal/ai/tools/service.go`
- [x] 创建 `internal/ai/tools/deployment.go`
- [x] 创建 `internal/ai/tools/cicd.go`
- [x] 创建 `internal/ai/tools/monitor.go`
- [x] 创建 `internal/ai/tools/config.go`
- [x] 创建 `internal/ai/tools/governance.go`
- [x] 创建 `internal/ai/tools/inventory.go`
- [x] 创建统一的 `BuildAllTools()` 函数

**文件:**
- `internal/ai/tools/ops.go`
- `internal/ai/tools/k8s.go`
- `internal/ai/tools/service.go`
- `internal/ai/tools/deployment.go`
- `internal/ai/tools/cicd.go`
- `internal/ai/tools/monitor.go`
- `internal/ai/tools/config.go`
- `internal/ai/tools/governance.go`
- `internal/ai/tools/inventory.go`
- `internal/ai/tools/builder.go`

**验收标准:**
- 所有现有工具保留
- 工具按类别组织
- 构建函数正确添加包装器

---

### Task 2.3: 为工具添加风险分级
- [x] 审核所有工具的风险等级
- [x] 更新工具元数据
- [x] 为高风险工具添加预览函数
- [x] 更新构建逻辑

**风险分级:**
| 工具 | 风险等级 | 包装类型 |
|-----|---------|---------|
| os_get_cpu_mem | Low | 无 |
| os_get_disk_fs | Low | 无 |
| host_ssh_exec_readonly | Medium | ReviewEdit |
| host_batch_exec_apply | High | Approvable |
| service_deploy_apply | High | Approvable |
| cicd_pipeline_trigger | High | Approvable |
| ... | ... | ... |

**验收标准:**
- 所有工具有明确风险等级
- 高风险工具有预览函数
- 构建逻辑正确应用包装

---

### Task 2.4: 简化模型创建
- [x] 创建 `internal/ai/model.go`
- [x] 移植现有模型创建逻辑
- [x] 支持 qwen/ollama 配置
- [x] 添加模型健康检查

**文件:**
- `internal/ai/model.go`

**验收标准:**
- 模型可正常创建
- 配置可正确读取

---

## Phase 3: HTTP 层简化 (Week 3)

### Task 3.1: 重写 HTTP 处理器
- [x] 创建新的 `internal/service/ai/handler.go`
- [x] 使用 `adk.NewRunner()`
- [x] 使用 `runner.Query()`
- [x] 实现事件迭代处理
- [x] 保持 API 签名不变

**文件:**
- `internal/service/ai/handler.go`

**代码示例:**
```go
func (h *handler) chat(c *gin.Context) {
    var req chatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        httpx.BindErr(c, err)
        return
    }

    runner := adk.NewRunner(c.Request.Context(), adk.RunnerConfig{
        EnableStreaming: true,
        Agent:           h.svcCtx.AI,
    })

    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    flusher, _ := c.Writer.(http.Flusher)

    iter := runner.Query(c.Request.Context(), req.Message)

    for {
        event, ok := iter.Next()
        if !ok {
            break
        }
        h.processEvent(c, flusher, event)
    }
}

func (h *handler) processEvent(c *gin.Context, flusher http.Flusher, event *adk.AgentEvent) {
    if event.Err != nil {
        if isInterrupted(event) {
            h.handleInterrupt(c, flusher, event)
        } else {
            emitError(c, flusher, event.Err)
        }
        return
    }

    if event.Output != nil && event.Output.MessageStream != nil {
        h.handleStream(c, flusher, event)
    }

    if event.Action != nil {
        h.handleAction(c, flusher, event)
    }
}
```

**验收标准:**
- API 签名不变
- SSE 事件格式兼容
- 代码行数 < 200

---

### Task 3.2: 实现 SSE 事件适配器
- [x] 创建 `internal/service/ai/events.go`
- [x] 定义事件格式转换逻辑
- [x] 实现 `handleStream()`
- [x] 实现 `handleInterrupt()`
- [x] 实现 `handleAction()`

**文件:**
- `internal/service/ai/events.go`

**验收标准:**
- 事件格式与现有前端兼容
- 审批事件正确发送
- 错误事件正确发送

---

### Task 3.3: 适配审批恢复处理
- [x] 实现审批恢复 API
- [x] 使用 `adk.Runner` 恢复执行
- [x] 传递审批结果上下文

**验收标准:**
- 审批通过后继续执行
- 审批拒绝后返回结果

---

## Phase 4: 清理与验证 (Week 4)

### Task 4.1: 删除旧代码
- [x] 删除 `internal/ai/experts/` 目录
- [x] 删除 `internal/ai/graph/` 目录
- [x] 删除 `internal/ai/callbacks/` 目录
- [x] 删除旧的 `internal/ai/platform_agent.go`
- [x] 删除旧的 `internal/ai/client.go`
- [x] 更新所有 import 引用

**验收标准:**
- 无编译错误
- 无遗留引用

---

### Task 4.2: 更新 ServiceContext
- [x] 更新 `internal/svc/service_context.go`
- [x] 使用新的 Agent 类型
- [x] 初始化 CheckPointStore

**验收标准:**
- 服务可正常启动
- Agent 可正常创建

---

### Task 4.3: 集成测试
- [x] 编写端到端测试
- [x] 测试简单查询
- [x] 测试多步骤任务
- [x] 测试审批流程
- [x] 测试错误恢复

**文件:**
- `internal/ai/integration_test.go`

**验收标准:**
- 所有关键流程通过
- 性能无明显下降

---

### Task 4.4: 性能基准测试
- [x] 编写性能测试
- [x] 对比新旧实现延迟
- [x] 监控内存占用
- [x] 监控数据库写入

**文件:**
- `internal/ai/benchmark_test.go`

**验收标准:**
- 延迟无明显增加
- 内存无泄漏

---

### Task 4.5: 文档更新
- [x] 更新 API 文档
- [x] 更新架构文档
- [x] 更新部署文档

**验收标准:**
- 文档与代码一致

---

## Dependency Graph

```
Phase 1 (基础设施)
    Task 1.1 (数据库表) ──┐
                         ├──> Task 1.3 (审批包装器)
    Task 1.2 (CheckPoint) ─┘
                         │
                         └──> Task 1.4 (审核包装器)

Phase 2 (Agent重构)
    Task 2.2 (拆分工具) ──┐
                         ├──> Task 2.1 (Agent构建)
    Task 2.3 (风险分级) ──┘
                         │
    Task 2.4 (模型创建) ──┘

Phase 3 (HTTP简化)
    Task 2.1 ────────────────> Task 3.1 (HTTP处理器)
                              │
                              ├──> Task 3.2 (事件适配)
                              │
                              └──> Task 3.3 (审批恢复)

Phase 4 (清理验证)
    Task 3.1 ────────────────> Task 4.1 (删除旧代码)
                              │
                              ├──> Task 4.2 (更新ServiceContext)
                              │
                              ├──> Task 4.3 (集成测试)
                              │
                              ├──> Task 4.4 (性能测试)
                              │
                              └──> Task 4.5 (文档更新)
```

## Estimated Effort

| Phase | Tasks | Estimated Days |
|-------|-------|---------------|
| Phase 1 | 4 | 5 |
| Phase 2 | 4 | 5 |
| Phase 3 | 3 | 4 |
| Phase 4 | 5 | 4 |
| **Total** | **16** | **18** |

## Success Criteria

1. **功能完整性**
   - 所有现有功能保持不变
   - 审批流程正常工作
   - 多步骤任务正常执行

2. **代码质量**
   - 核心代码量减少 80%+
   - 单元测试覆盖率 > 80%
   - 无遗留自建框架代码

3. **性能**
   - 延迟无明显增加
   - 内存无明显泄漏
   - 数据库写入可控

4. **可维护性**
   - 拥抱 eino ADK 标准
   - 代码结构清晰
   - 文档完整
