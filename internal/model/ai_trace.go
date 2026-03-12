// Package model 提供数据库模型定义。
//
// 本文件定义 AI 可观测性追踪数据的数据模型。
// 用于存储 LLM 调用、工具调用、Agent 运行的详细追踪信息。
package model

import "time"

// AITraceSpan 是 AI 追踪跨度表模型，记录每次 LLM/工具/Agent 调用的详细信息。
//
// 表名: ai_trace_spans
// 用途:
//   - 调试：查看某次会话的完整调用链
//   - 问题排查：定位具体哪个调用出错
//   - 性能分析：分析各阶段耗时
//
// 数据保留策略建议：
//   - 热数据（7天内）：完整保留
//   - 温数据（30天内）：保留摘要，清理 input/output
//   - 冷数据（超过30天）：可归档或删除
type AITraceSpan struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                        // 主键
	SpanID       string    `gorm:"column:span_id;type:varchar(64);uniqueIndex" json:"span_id"`          // 跨度唯一标识
	SpanType     string    `gorm:"column:span_type;type:varchar(32);index" json:"span_type"`            // 跨度类型: llm/tool/agent
	Name         string    `gorm:"column:name;type:varchar(200);index" json:"name"`                     // 跨度名称 (如: llm:planner-stage)
	SessionID    string    `gorm:"column:session_id;type:varchar(64);index" json:"session_id"`          // 关联的会话 ID
	TraceID      string    `gorm:"column:trace_id;type:varchar(64);index" json:"trace_id"`              // 追踪 ID (同一次请求共享)
	ParentSpanID string    `gorm:"column:parent_span_id;type:varchar(64)" json:"parent_span_id"`        // 父跨度 ID (用于构建调用树)
	StartTime    time.Time `gorm:"column:start_time;index" json:"start_time"`                           // 开始时间
	EndTime      time.Time `gorm:"column:end_time" json:"end_time"`                                     // 结束时间
	DurationMs   int64     `gorm:"column:duration_ms" json:"duration_ms"`                               // 持续时间（毫秒）
	Status       string    `gorm:"column:status;type:varchar(32);default:'success'" json:"status"`      // 状态: success/error
	ErrorMsg     string    `gorm:"column:error_msg;type:text" json:"error_msg"`                         // 错误信息
	Input        string    `gorm:"column:input;type:longtext" json:"input"`                             // 输入内容（截断）
	Output       string    `gorm:"column:output;type:longtext" json:"output"`                           // 输出内容（截断）
	Tokens       int64     `gorm:"column:tokens;default:0" json:"tokens"`                               // Token 数量（仅 LLM 有效）
	MetadataJSON string    `gorm:"column:metadata_json;type:text" json:"metadata_json"`                 // 扩展元数据 (JSON)
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`                  // 创建时间
}

// TableName 返回 AI 追踪跨度表名。
func (AITraceSpan) TableName() string { return "ai_trace_spans" }

// SpanType 常量定义。
const (
	SpanTypeLLM   = "llm"   // LLM 调用
	SpanTypeTool  = "tool"  // 工具调用
	SpanTypeAgent = "agent" // Agent 运行
)

// Status 常量定义。
const (
	SpanStatusSuccess = "success" // 成功
	SpanStatusError   = "error"   // 错误
)
