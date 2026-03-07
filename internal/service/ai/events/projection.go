package events

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type ProjectedEvent struct {
	Name    string
	Payload gin.H
}

// ProjectCompatibilityEvents keeps AI core events as the semantic source of truth,
// while normalizing payloads so legacy SSE consumers can continue to function.
func ProjectCompatibilityEvents(event string, payload gin.H) []ProjectedEvent {
	normalized := clonePayload(payload)
	switch event {
	case "approval_required":
		normalized["approval_required"] = true
		if _, ok := normalized["previewDiff"]; !ok {
			if preview, ok := normalized["preview"].(map[string]any); ok {
				if diff, ok := preview["preview_diff"]; ok && strings.TrimSpace(toString(diff)) != "" {
					normalized["previewDiff"] = diff
				}
			}
		}
	case "confirmation_required":
		if token, ok := normalized["token"]; ok && normalized["confirmation_token"] == nil {
			normalized["confirmation_token"] = token
		}
		if expires, ok := normalized["expires_at"]; ok && normalized["expiresAt"] == nil {
			normalized["expiresAt"] = expires
		}
	}
	return []ProjectedEvent{{Name: event, Payload: normalized}}
}

func clonePayload(payload gin.H) gin.H {
	if payload == nil {
		return gin.H{}
	}
	out := gin.H{}
	for k, v := range payload {
		out[k] = v
	}
	return out
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	default:
		return ""
	}
}
