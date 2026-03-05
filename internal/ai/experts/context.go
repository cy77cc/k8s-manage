package experts

import (
	"context"

	aicallbacks "github.com/cy77cc/k8s-manage/internal/ai/callbacks"
)

func WithProgressEmitter(ctx context.Context, emitter ProgressEmitter) context.Context {
	if emitter == nil {
		return aicallbacks.WithEmitter(ctx, nil)
	}
	return aicallbacks.WithEmitter(ctx, aicallbacks.EventEmitterFunc(func(event string, payload any) bool {
		emitter(event, payload)
		return true
	}))
}

func ProgressEmitterFromContext(ctx context.Context) ProgressEmitter {
	emitter := aicallbacks.EmitterFromContext(ctx)
	if emitter == nil {
		return nil
	}
	return func(event string, payload any) {
		emitter.Emit(event, payload)
	}
}
