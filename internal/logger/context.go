package logger

import "context"

type ctxKey string

const traceIDKey ctxKey = "trace_id"

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func (z *zapLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return z
	}

	if v := ctx.Value(traceIDKey); v != nil {
		return z.With(String("trace_id", v.(string)))
	}
	return z
}
