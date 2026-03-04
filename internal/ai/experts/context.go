package experts

import "context"

type progressEmitterKey struct{}

func WithProgressEmitter(ctx context.Context, emitter ProgressEmitter) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, progressEmitterKey{}, emitter)
}

func ProgressEmitterFromContext(ctx context.Context) ProgressEmitter {
	if ctx == nil {
		return nil
	}
	v := ctx.Value(progressEmitterKey{})
	emitter, _ := v.(ProgressEmitter)
	return emitter
}
