// Package observability 提供 AI 编排层的可观测性能力。
//
// 本文件提供可观测性组件的初始化和集成功能，
// 支持 Prometheus 指标暴露和追踪数据存储。
package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
)

// Config 可观测性配置。
type Config struct {
	// TraceStore 追踪数据存储（数据库）
	TraceStore *TraceStore
	// EventHandler 事件处理器（可选）
	EventHandler EventHandler
}

// Handler 全局可观测性处理器。
var globalHandler *Handler

// Setup 初始化可观测性处理器。
// 此函数应该在应用启动时调用一次。
func Setup(cfg Config) *Handler {
	if globalHandler != nil {
		return globalHandler
	}

	globalHandler = NewHandler(cfg.TraceStore, cfg.EventHandler)
	return globalHandler
}

// SetupWithDB 使用数据库连接初始化可观测性处理器。
func SetupWithDB(db *gorm.DB, eventHandler EventHandler) *Handler {
	var traceStore *TraceStore
	if db != nil {
		traceStore = NewTraceStore(db)
	}
	return Setup(Config{
		TraceStore:   traceStore,
		EventHandler: eventHandler,
	})
}

// GetHandler 获取全局可观测性处理器。
func GetHandler() *Handler {
	return globalHandler
}

// PrometheusHandler 返回 Prometheus 指标端点的 HTTP handler。
// 用于暴露 /metrics 端点供 Prometheus 抓取。
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// RegisterPrometheusHandler 将 Prometheus 指标端点注册到指定的 ServeMux。
func RegisterPrometheusHandler(mux *http.ServeMux, path string) {
	mux.Handle(path, PrometheusHandler())
}
