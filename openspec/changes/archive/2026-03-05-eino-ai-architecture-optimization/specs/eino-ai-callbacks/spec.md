# Spec: Eino AI Callbacks

## 概述

统一 AI 助手事件处理，基于 eino Callbacks 机制实现切面编程。

## 接口定义

### EventEmitter

```go
// EventEmitter 事件发射器接口
type EventEmitter interface {
    Emit(event string, payload any) bool
}
```

### AIEventHandler

```go
// AIEventHandler 统一事件处理器
type AIEventHandler struct {
    emitter EventEmitter
}

// 工具调用生命周期
func (h *AIEventHandler) OnToolCallStart(ctx context.Context, tool, callID string, args map[string]any) context.Context
func (h *AIEventHandler) OnToolCallEnd(ctx context.Context, tool, callID string, result any, err error, duration time.Duration) context.Context

// 专家执行生命周期
func (h *AIEventHandler) OnExpertStart(ctx context.Context, expert, task string) context.Context
func (h *AIEventHandler) OnExpertEnd(ctx context.Context, expert string, duration time.Duration, err error) context.Context
```

## 事件类型

### ToolCallEvent

| 字段 | 类型 | 说明 |
|------|------|------|
| Tool | string | 工具名称 |
| CallID | string | 调用ID |
| Arguments | map[string]any | 调用参数 |
| Result | any | 返回结果 |
| Error | string | 错误信息 |
| Timestamp | time.Time | 时间戳 |
| Duration | time.Duration | 执行时长 |

### ExpertProgressEvent

| 字段 | 类型 | 说明 |
|------|------|------|
| Expert | string | 专家名称 |
| Status | string | 状态 (running/done/failed) |
| Task | string | 任务描述 |
| DurationMs | int64 | 执行时长(ms) |
| Error | string | 错误信息 |

## Context 集成

```go
// 注入 emitter
ctx = callbacks.WithEmitter(ctx, emitter)

// 获取 emitter
emitter := callbacks.EmitterFromContext(ctx)

// 构建 handler
handler := callbacks.HandlerFromContext(ctx)
```

## 使用示例

```go
// 在 chat_handler.go 中
emitter := callbacks.EventEmitterFunc(func(event string, payload any) bool {
    return emit(event, toPayloadMap(payload))
})
streamCtx := callbacks.WithEmitter(c.Request.Context(), emitter)

// 在专家执行中
handler := callbacks.HandlerFromContext(ctx)
handler.OnExpertStart(ctx, "k8s_expert", "analyze pod status")
// ... 执行 ...
handler.OnExpertEnd(ctx, "k8s_expert", duration, err)
```

## 迁移指南

### 替换 ProgressEmitter

```go
// 旧代码
ctx = experts.WithProgressEmitter(ctx, func(event string, payload any) {...})

// 新代码
ctx = callbacks.WithEmitter(ctx, callbacks.EventEmitterFunc(func(event string, payload any) bool {
    emit(event, payload)
    return true
}))
```

### 替换 ToolEventEmitter

```go
// 旧代码
ctx = tools.WithToolEventEmitter(ctx, func(event string, payload any) {...})

// 新代码 - 使用 callbacks handler
handler := callbacks.HandlerFromContext(ctx)
handler.OnToolCallStart(ctx, tool, callID, args)
```
