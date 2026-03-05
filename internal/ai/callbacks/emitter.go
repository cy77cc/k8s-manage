package callbacks

// EventEmitter emits a structured event payload.
type EventEmitter interface {
	Emit(event string, payload any) bool
}

// EventEmitterFunc adapts a function to EventEmitter.
type EventEmitterFunc func(event string, payload any) bool

func (f EventEmitterFunc) Emit(event string, payload any) bool {
	if f == nil {
		return false
	}
	return f(event, payload)
}

type nopEmitter struct{}

func (nopEmitter) Emit(string, any) bool { return true }

// NopEmitter is a no-op emitter used when no emitter is provided.
var NopEmitter EventEmitter = nopEmitter{}
