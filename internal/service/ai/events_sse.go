package ai

import (
	"encoding/json"
	"fmt"
	"net/http"

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
