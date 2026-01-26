package xctx

import (
	"context"
	"sync"
)

// ctxKey 是私有类型，防止外部包的 key 冲突
// 使用空结构体是为了不占用内存
type ctxKey struct{}

type Context struct {
	TraceID   string
	UID       string
	Role      string
	ClientIP  string
	RequestID string
	StartTime int64
	EndTime   int64
	Latency   int64
	Token     string
	data      map[any]any
	mu        sync.RWMutex
}

func NewContext(opts ...func(ctx *Context)) *Context {
	ctx := &Context{
		data: make(map[any]any),
	}
	for _, opt := range opts {
		opt(ctx)
	}
	return ctx
}

// WithValue 将自定义 Context 注入到标准 context.Context 中
func WithContext(parent context.Context, c *Context) context.Context {
	return context.WithValue(parent, ctxKey{}, c)
}

// FromContext 从标准 context.Context 中提取自定义 Context
func FromContext(ctx context.Context) *Context {
	if ctx == nil {
		return nil
	}
	c, ok := ctx.Value(ctxKey{}).(*Context)
	if !ok {
		return nil
	}
	return c
}

func (c *Context) Get(key any) any {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[key]
}

func (c *Context) Set(key, value any) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}
