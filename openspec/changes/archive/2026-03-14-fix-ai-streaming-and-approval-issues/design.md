# Design: 修复 AI 流式输出和审批功能问题

## Overview

本设计文档描述如何修复 AI 模块的 SSE 流式输出、思维链显示、对话持久化和审批功能的多个问题。

## Architecture

### 当前架构问题

```
┌──────────────────────────────────────────────────────────────────────┐
│                          Current Flow                                 │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ADK Runner                                                           │
│       │                                                               │
│       ▼                                                               │
│  orchestrator.streamExecution()                                       │
│       │                                                               │
│       │  msg.Content = "完整内容" (累积)                              │
│       │                                                               │
│       ▼                                                               │
│  OnTextDelta(完整内容)  ────────────►  SSE Event                      │
│       │                                    │                          │
│       │                                    ▼                          │
│       │                              前端追加内容                       │
│       │                              (导致重复或大chunk)               │
│       │                                                               │
└───────┴───────────────────────────────────────────────────────────────┘
```

### 修复后架构

```
┌──────────────────────────────────────────────────────────────────────┐
│                          Fixed Flow                                   │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ADK Runner                                                           │
│       │                                                               │
│       ▼                                                               │
│  orchestrator.streamExecution()                                       │
│       │                                                               │
│       │  msg.Content = "完整内容" (累积)                              │
│       │  lastContent = "上次发送的内容"                               │
│       │  chunk = text[len(lastContent):]  // 计算增量                 │
│       │                                                               │
│       ▼                                                               │
│  OnTextDelta(chunk)  ──────────────►  SSE Event                      │
│       │                                    │                          │
│       │                                    ▼                          │
│       │                              前端追加增量                       │
│       │                              (正确逐字显示)                    │
│       │                                                               │
└───────┴──────────────────────────────────────────────────────────────┘
```

## Component Design

### 1. 流式增量输出

**文件**: `internal/ai/orchestrator.go`

```go
func (o *Orchestrator) streamExecution(...) {
    var lastContent string  // 新增：跟踪已发送内容

    for {
        // ... 获取事件 ...

        if msg != nil && strings.TrimSpace(msg.Content) != "" {
            text := strings.TrimSpace(msg.Content)
            // 计算增量
            if len(text) > len(lastContent) {
                chunk := text[len(lastContent):]
                lastContent = text
                emit(o.converter.OnTextDelta(chunk))
            }
        }
    }
}
```

### 2. 思维链事件增强

**文件**: `internal/ai/runtime/sse_converter.go`

```go
func (c *SSEConverter) OnPlannerStart(sessionID, planID, turnID string) []StreamEvent {
    return []StreamEvent{
        {Type: EventTurnStarted, Data: map[string]any{
            "turn_id":    turnID,
            "session_id": sessionID,
        }},
        {Type: EventStageDelta, Data: map[string]any{
            "stage":       "plan",
            "status":      "loading",
            "plan_id":     planID,
            "title":       "整理执行步骤",
            "description": "正在根据你的需求整理执行步骤",
        }},
    }
}

func (c *SSEConverter) OnPlanCreated(planID, content string, steps []StepInfo) StreamEvent {
    return StreamEvent{Type: EventStageDelta, Data: map[string]any{
        "stage":       "plan",
        "status":      "success",
        "plan_id":     planID,
        "content":     strings.TrimSpace(content),
        "title":       "执行步骤已整理",
        "description": "已生成可执行步骤",
        "steps":       steps,  // 新增：步骤列表
    }}
}
```

### 3. 用户消息持久化修复

**文件**: `internal/service/ai/session_recorder.go`

```go
func (r *chatRecorder) handleMeta(ctx context.Context, payload map[string]any) {
    r.sessionID = firstString(payload["session_id"], payload["sessionId"])
    r.assistantTurnID = firstString(payload["turn_id"], payload["turnId"])
    r.assistant.TraceID = firstString(payload["trace_id"], payload["traceId"])

    // 确保 sessionID 存在
    if r.sessionID == "" {
        r.sessionID = uuid.NewString()
    }

    // 立即持久化用户消息
    _ = r.store.EnsureSession(ctx, r.sessionID, r.userID, r.scene, r.title)
    _ = r.store.AppendUserMessage(ctx, r.sessionID, r.userID, r.scene, r.title, r.prompt)

    // ... 创建助手消息 ...
}
```

### 4. 审批门包装修复

**文件**: `internal/ai/tools/tools.go`

```go
func NewAllTools(ctx context.Context, deps common.PlatformDeps) []tool.BaseTool {
    // ...
    for _, current := range base {
        invokable, ok := current.(tool.InvokableTool)
        if !ok {
            log.Debug("tool does not implement InvokableTool")
            out = append(out, current)
            continue
        }

        info, err := invokable.Info(ctx)
        if err != nil || info == nil {
            log.Warn("tool info unavailable", "error", err)
            out = append(out, current)
            continue
        }

        spec, ok := registry.Get(info.Name)
        if !ok {
            log.Warn("tool not found in registry", "name", info.Name)
            // 尝试根据工具名称推断模式
            mode := inferToolMode(info.Name)
            if mode == "mutating" {
                // 包装审批门
                out = append(out, approvaltools.NewGate(invokable, ...))
            } else {
                out = append(out, current)
            }
            continue
        }

        out = append(out, approvaltools.NewGate(invokable, ...))
    }
    return out
}

func inferToolMode(name string) string {
    mutatingPatterns := []string{"_apply", "_exec", "_delete", "_update", "_create", "_restart", "_scale"}
    for _, pattern := range mutatingPatterns {
        if strings.Contains(name, pattern) {
            return "mutating"
        }
    }
    return "readonly"
}
```

## Error Handling

- 流式输出失败时记录日志，返回已发送内容
- 持久化失败时记录日志，但不中断流程
- 审批门包装失败时回退到基础工具

## Testing Strategy

1. **流式输出测试**
   - 验证 delta 事件包含增量内容
   - 验证前端逐字显示

2. **思维链测试**
   - 验证 stage_delta 事件包含 title 和 description
   - 验证前端动态显示

3. **持久化测试**
   - 发送消息后验证数据库记录
   - 刷新页面后验证消息恢复

4. **审批测试**
   - 执行高风险工具验证审批面板显示
   - 验证审批确认后继续执行
