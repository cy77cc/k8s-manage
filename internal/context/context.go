package context

import "sync"

type Context struct {
	TraceID   string
	UID    string
	Role      string
	ClientIP  string
	RequestID string
	StartTime int64
	EndTime   int64
	Latency   int64
	Token     string
	data map[any]any
	mu sync.RWMutex
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
