package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func writeSSE(c *gin.Context, flusher http.Flusher, event string, payload any) bool {
	raw, err := json.Marshal(payload)
	if err != nil {
		return false
	}
	if _, err = fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, raw); err != nil {
		return false
	}
	flusher.Flush()
	return true
}

type sseEvent struct {
	event   string
	payload gin.H
}

type sseWriter struct {
	c       *gin.Context
	flusher http.Flusher
	turnID  string
	ch      chan sseEvent
	done    chan struct{}
	mu      sync.RWMutex
	err     error
}

func newSSEWriter(c *gin.Context, flusher http.Flusher, turnID string) *sseWriter {
	w := &sseWriter{
		c:       c,
		flusher: flusher,
		turnID:  turnID,
		ch:      make(chan sseEvent, 64),
		done:    make(chan struct{}),
	}
	go w.loop()
	return w
}

func (w *sseWriter) loop() {
	defer close(w.done)
	for evt := range w.ch {
		payload := evt.payload
		if payload == nil {
			payload = gin.H{}
		}
		payload["turn_id"] = w.turnID
		if !writeSSE(w.c, w.flusher, evt.event, payload) {
			w.setErr(fmt.Errorf("stream write failed"))
			return
		}
	}
}

func (w *sseWriter) Emit(event string, payload gin.H) bool {
	if w.hasErr() {
		return false
	}
	select {
	case <-w.done:
		return false
	case w.ch <- sseEvent{event: event, payload: payload}:
		return true
	}
}

func (w *sseWriter) Close() error {
	close(w.ch)
	<-w.done
	return w.Err()
}

func (w *sseWriter) Err() error {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.err
}

func (w *sseWriter) hasErr() bool {
	return w.Err() != nil
}

func (w *sseWriter) setErr(err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.err == nil {
		w.err = err
	}
}

func heartbeatLoop(stop <-chan struct{}, emit func(event string, payload gin.H) bool) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if !emit("heartbeat", gin.H{"status": "alive"}) {
				return
			}
		case <-stop:
			return
		}
	}
}
