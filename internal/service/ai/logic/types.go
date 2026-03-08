package logic

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrConfirmationNotFound = errors.New("confirmation not found")

type AISession struct {
	ID        string    `json:"id"`
	UserID    uint64    `json:"user_id,omitempty"`
	Scene     string    `json:"scene"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type RecommendationRecord struct {
	ID             string    `json:"id"`
	UserID         uint64    `json:"user_id,omitempty"`
	Scene          string    `json:"scene"`
	Type           string    `json:"type"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	FollowupPrompt string    `json:"followup_prompt,omitempty"`
	Reasoning      string    `json:"reasoning,omitempty"`
	Relevance      float64   `json:"relevance,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
}

type ApprovalTicket struct {
	ID        string         `json:"id"`
	UserID    uint64         `json:"user_id,omitempty"`
	Tool      string         `json:"tool"`
	Status    string         `json:"status"`
	Params    map[string]any `json:"params,omitempty"`
	CreatedAt time.Time      `json:"created_at,omitempty"`
}

type ExecutionRecord struct {
	ID        string         `json:"id"`
	Tool      string         `json:"tool"`
	Status    string         `json:"status"`
	Result    map[string]any `json:"result,omitempty"`
	CreatedAt time.Time      `json:"created_at,omitempty"`
}

func ToString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprintf("%v", x)
	}
}

func NormalizeScene(scene string) string {
	normalized := strings.TrimSpace(strings.ToLower(scene))
	if normalized == "" {
		return "global"
	}
	return strings.TrimPrefix(normalized, "scene:")
}
