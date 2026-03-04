# Proposal: 多专家协作机制重构

## 问题背景

在Hybrid MOE架构重构后，多专家协作存在以下严重问题：

### 问题1: 多专家输出未汇总

**现象**: 用户看到碎片化的输出，每个专家的输出都直接展示在前端。

```
用户: "分析这个服务的状态"

AI: [service_expert] 找到服务 payment-api...
AI: [k8s_expert] Pod 状态正常...        ← 用户困惑：为什么又有输出？
AI: [topology_expert] 依赖关系...        ← 继续困惑
AI: "--- 综合分析 ---" 根据以上...       ← 这是什么？
```

**根因**: `orchestrator.StreamExecute()` 直接将每个专家的输出转发到前端，而非先汇总再输出。

### 问题2: 不需要的专家被强制调用

**现象**: 配置固定的 `helper_experts` 每次都被调用，即使不需要。

```yaml
# scene_mappings.yaml
services:detail:
  primary_expert: service_expert
  helper_experts: [k8s_workload_expert, topology_expert]  # 每次都调用
  strategy: sequential
```

用户只想简单查询服务信息，但 k8s 和 topology 专家也被拉进来，造成延迟。

### 问题3: 上下文记忆丢失

**现象**: 多轮对话失去连贯性。

```
用户: "查看主机A的CPU"
AI: "找到主机A，CPU使用率50%"
用户: "那个集群呢？"
AI: "什么集群？"  ← 丢失了上下文
```

**根因**: `ExecuteRequest.History` 字段存在，但 `executor.ExecuteStep()` 没有传递给专家。

```go
// executor.go: 当前实现
exp.Agent.Stream(ctx, []*schema.Message{
    schema.UserMessage(msg),  // 只有当前消息，没有历史！
})
```

### 问题4: Tool调用链过长

**现象**: 每个专家独立执行 ReAct 循环，工具调用叠加。

```
service_expert: Tool Call 1, 2, 3...
k8s_expert: Tool Call 4, 5, 6...
topology_expert: Tool Call 7, 8...
```

用户等待时间 = 所有专家调用时间之和。

## 目标

### 主要目标

1. **输出汇总**: 助手专家静默执行，主专家流式输出汇总结果
2. **智能调用**: 主专家决定是否需要助手，而非配置固定
3. **上下文传递**: 所有专家都能看到历史对话
4. **进度动画**: 助手执行期间显示进度动画

### 非目标

- 不改变现有的工具实现
- 不改变专家注册表结构
- 不引入新的依赖

## 范围

### 修改的文件

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `configs/scene_mappings.yaml` | 修改 | `helper_experts` → `optional_helpers` |
| `internal/ai/experts/types.go` | 修改 | 添加相关类型 |
| `internal/ai/experts/router.go` | 修改 | 解析 `optional_helpers` |
| `internal/ai/experts/orchestrator.go` | 重写 | 主从协作流程 |
| `internal/ai/experts/executor.go` | 修改 | 传递历史上下文 |
| `internal/ai/platform_agent.go` | 修改 | 适配新流程 |
| `web/src/components/AI/ChatInterface.tsx` | 修改 | 支持 expert_progress 事件 |

### 接口变更

```go
// SceneMapping 结构变更
type SceneMapping struct {
    PrimaryExpert    string   `yaml:"primary_expert"`
    OptionalHelpers  []string `yaml:"optional_helpers"`  // 原 helper_experts
    Strategy         string   `yaml:"strategy"`
    // ...
}

// RouteDecision 新增字段
type RouteDecision struct {
    PrimaryExpert   string
    OptionalHelpers []string  // 可选助手，主专家决定是否调用
    Strategy        ExecutionStrategy
    // ...
}

// ExecuteRequest 确保 History 传递
type ExecuteRequest struct {
    Message        string
    Decision       *RouteDecision
    RuntimeContext map[string]any
    History        []*schema.Message  // 必须传递给所有专家
}

// 新增 SSE 事件类型
type ExpertProgressEvent struct {
    Expert      string `json:"expert"`
    Status      string `json:"status"`  // "running" | "done"
    Task        string `json:"task,omitempty"`
    DurationMs  int64  `json:"duration_ms,omitempty"`
}
```

## 解决方案概述

### 核心流程

```
用户消息 + 历史
      │
      ▼
路由决策 → 主专家 + 可选助手列表
      │
      ▼
┌─────────────────────────────────────────┐
│ 主专家第一轮 (决策)                       │
│ - 分析意图                               │
│ - 决定是否需要助手                        │
│ - 输出: [REQUEST_HELPER: xxx] 或直接回答 │
└─────────────────────────────────────────┘
      │
      ├─ 直接回答 → 流式输出，结束
      │
      ▼ 需要助手
┌─────────────────────────────────────────┐
│ emit("expert_progress", {running})      │
│ 助手执行 (静默，可并行)                   │
│ emit("expert_progress", {done})         │
└─────────────────────────────────────────┘
      │
      ▼
┌─────────────────────────────────────────┐
│ 主专家第二轮 (汇总)                       │
│ - 流式输出最终回答                        │
└─────────────────────────────────────────┘
```

### 关键设计决策

1. **助手输出**: 纯动画，不发送内容到前端（节省上下文）
2. **助手调用**: 配置提供候选，主专家决定是否调用
3. **上下文传递**: 所有专家都收到完整历史
4. **助手执行方式**: 使用 `Generate` 而非 `Stream`，因为助手输出不需要流式传输，`Generate` 更快

## 风险评估

| 风险 | 等级 | 缓解措施 |
|------|------|----------|
| 主专家决策延迟 | 中 | 设置超时，默认不调用助手 |
| 助手执行失败 | 低 | 优雅降级，主专家仍可输出 |
| 前端动画兼容 | 低 | 渐进增强，不支持时降级 |

## 验收标准

1. 用户只看到一条连贯的汇总输出
2. 主专家能自主决定是否需要助手
3. 多轮对话保持上下文连贯
4. 助手执行期间前端显示进度动画
5. 单专家场景不受影响
