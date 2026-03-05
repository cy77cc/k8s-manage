package ai

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSSEWriterEmitAndFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	w := newSSEWriter(c, rec, "turn-test-1")
	if ok := w.Emit("meta", gin.H{"sessionId": "s1"}); !ok {
		t.Fatalf("expected meta emit ok")
	}
	if ok := w.Emit("delta", gin.H{"contentChunk": "hello"}); !ok {
		t.Fatalf("expected delta emit ok")
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close sse writer: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "event: meta") {
		t.Fatalf("missing meta event: %s", body)
	}
	if !strings.Contains(body, "event: delta") {
		t.Fatalf("missing delta event: %s", body)
	}
	if !strings.Contains(body, "\"turn_id\":\"turn-test-1\"") {
		t.Fatalf("missing turn_id in payload: %s", body)
	}
	if !strings.Contains(body, "\"contentChunk\":\"hello\"") {
		t.Fatalf("missing delta payload content: %s", body)
	}
}
