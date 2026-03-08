package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type ProjectedEvent struct {
	Name    string
	Payload gin.H
}

type SSEWriter struct {
	ctx     *gin.Context
	flusher interface{ Flush() }
	turnID  string
}

func NewSSEWriter(ctx *gin.Context, flusher interface{ Flush() }, turnID string) *SSEWriter {
	return &SSEWriter{ctx: ctx, flusher: flusher, turnID: turnID}
}

func (w *SSEWriter) Emit(event string, payload gin.H) bool {
	if payload == nil {
		payload = gin.H{}
	}
	payload["turnId"] = w.turnID
	raw, err := json.Marshal(payload)
	if err != nil {
		return false
	}
	if _, err := fmt.Fprintf(w.ctx.Writer, "event: %s\ndata: %s\n\n", event, raw); err != nil {
		return false
	}
	w.flusher.Flush()
	return true
}

func (w *SSEWriter) Close() {}

func ProjectCompatibilityEvents(event string, payload gin.H) []ProjectedEvent {
	return []ProjectedEvent{{Name: event, Payload: payload}}
}

func HeartbeatLoop(stop <-chan struct{}, emit func(event string, payload gin.H) bool) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if !emit("heartbeat", gin.H{"ts": time.Now().UTC().Format(time.RFC3339Nano)}) {
				return
			}
		}
	}
}
