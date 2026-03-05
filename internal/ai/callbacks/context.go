package callbacks

import "context"

type emitterKey struct{}

// WithEmitter injects an EventEmitter into context.
func WithEmitter(ctx context.Context, emitter EventEmitter) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if emitter == nil {
		emitter = NopEmitter
	}
	return context.WithValue(ctx, emitterKey{}, emitter)
}

// EmitterFromContext reads an EventEmitter from context.
func EmitterFromContext(ctx context.Context) EventEmitter {
	if ctx == nil {
		return NopEmitter
	}
	v := ctx.Value(emitterKey{})
	emitter, _ := v.(EventEmitter)
	if emitter == nil {
		return NopEmitter
	}
	return emitter
}

// HandlerFromContext builds a handler bound to context emitter.
func HandlerFromContext(ctx context.Context) *AIEventHandler {
	return NewAIEventHandler(EmitterFromContext(ctx))
}
