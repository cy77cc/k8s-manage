// Package observability 提供 AI 编排层的可观测性能力。
//
// 本文件提供可观测性组件的初始化和集成功能，
// 支持全局回调注册和事件处理。
package observability

import (
	"sync"

	"github.com/cloudwego/eino/callbacks"
)

var (
	globalHandler     *Handler
	globalHandlerOnce sync.Once
)

// Config 可观测性配置。
type Config struct {
	// EnableTracing 是否启用链路追踪
	EnableTracing bool
	// EnableMetrics 是否启用指标收集
	EnableMetrics bool
	// EventHandler 事件处理器
	EventHandler EventHandler
}

// Setup 初始化全局可观测性处理器。
// 此函数应该在应用启动时调用一次。
// 返回 Handler 实例，可用于获取指标和追踪信息。
func Setup(cfg Config) *Handler {
	globalHandlerOnce.Do(func() {
		globalHandler = NewHandler(cfg.EventHandler)

		// 注册全局回调
		callbacks.AppendGlobalHandlers(globalHandler.BuildCallbackHandler())
	})
	return globalHandler
}

// GetHandler 获取全局可观测性处理器。
func GetHandler() *Handler {
	return globalHandler
}

// Snapshot 返回当前全局指标快照。
func Snapshot() MetricsSnapshot {
	if globalHandler == nil {
		return MetricsSnapshot{}
	}
	return globalHandler.Snapshot()
}
