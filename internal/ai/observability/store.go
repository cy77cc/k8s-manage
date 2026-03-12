// Package observability 提供 AI 编排层的可观测性能力。
//
// 本文件实现追踪数据的数据库存储。
package observability

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/cy77cc/OpsPilot/internal/model"
	"gorm.io/gorm"
)

// TraceStore 追踪数据存储。
type TraceStore struct {
	db *gorm.DB
}

// NewTraceStore 创建追踪数据存储实例。
func NewTraceStore(db *gorm.DB) *TraceStore {
	if db == nil {
		return nil
	}
	return &TraceStore{db: db}
}

// SaveSpan 保存追踪跨度到数据库。
func (s *TraceStore) SaveSpan(ctx context.Context, span *model.AITraceSpan) error {
	if s == nil || s.db == nil || span == nil {
		return nil
	}
	return s.db.WithContext(ctx).Create(span).Error
}

// SaveSpanAsync 异步保存追踪跨度。
func (s *TraceStore) SaveSpanAsync(span *model.AITraceSpan) {
	if s == nil || s.db == nil || span == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.SaveSpan(ctx, span)
	}()
}

// GetSpansBySession 获取指定会话的所有追踪跨度。
func (s *TraceStore) GetSpansBySession(ctx context.Context, sessionID string) ([]model.AITraceSpan, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	var spans []model.AITraceSpan
	err := s.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("start_time ASC").
		Find(&spans).Error
	return spans, err
}

// GetSpansByTraceID 获取指定追踪 ID 的所有跨度。
func (s *TraceStore) GetSpansByTraceID(ctx context.Context, traceID string) ([]model.AITraceSpan, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	var spans []model.AITraceSpan
	err := s.db.WithContext(ctx).
		Where("trace_id = ?", traceID).
		Order("start_time ASC").
		Find(&spans).Error
	return spans, err
}

// DeleteOldSpans 删除超过指定天数的追踪数据。
func (s *TraceStore) DeleteOldSpans(ctx context.Context, retentionDays int) error {
	if s == nil || s.db == nil {
		return nil
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)
	return s.db.WithContext(ctx).
		Where("created_at < ?", cutoff).
		Delete(&model.AITraceSpan{}).Error
}

// PurgeSpanDetails 清理指定天数前的详细数据（保留元数据）。
// 用于减少存储占用，同时保留基本统计信息。
func (s *TraceStore) PurgeSpanDetails(ctx context.Context, retentionDays int) error {
	if s == nil || s.db == nil {
		return nil
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)
	return s.db.WithContext(ctx).
		Model(&model.AITraceSpan{}).
		Where("created_at < ?", cutoff).
		Updates(map[string]any{
			"input":         "",
			"output":        "",
			"metadata_json": "",
		}).Error
}

// BuildSpan 从回调数据构建追踪跨度模型。
func BuildSpan(spanID, spanType, name, sessionID, traceID, parentSpanID string,
	startTime time.Time, durationMs int64, status, errorMsg, input, output string,
	tokens int64, metadata map[string]any) *model.AITraceSpan {

	span := &model.AITraceSpan{
		SpanID:       spanID,
		SpanType:     spanType,
		Name:         name,
		SessionID:    sessionID,
		TraceID:      traceID,
		ParentSpanID: parentSpanID,
		StartTime:    startTime.UTC(),
		EndTime:      startTime.Add(time.Duration(durationMs) * time.Millisecond),
		DurationMs:   durationMs,
		Status:       status,
		ErrorMsg:     truncateString(errorMsg, 2000),
		Input:        truncateString(input, 4000),
		Output:       truncateString(output, 4000),
		Tokens:       tokens,
		CreatedAt:    time.Now().UTC(),
	}

	if len(metadata) > 0 {
		data, _ := json.Marshal(metadata)
		span.MetadataJSON = string(data)
	}

	return span
}

func truncateString(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
