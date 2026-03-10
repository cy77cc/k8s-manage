package state

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
)

const defaultTTL = 24 * time.Hour

type SessionSnapshot struct {
	SessionID string          `json:"session_id"`
	Title     string          `json:"title,omitempty"`
	Messages  []StoredMessage `json:"messages"`
	Context   map[string]any  `json:"context,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type StoredMessage struct {
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	ToolCallID string    `json:"tool_call_id,omitempty"`
	ToolName   string    `json:"tool_name,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

type SessionState struct {
	client redis.UniversalClient
	prefix string
	ttl    time.Duration
}

type CheckpointStore struct {
	client redis.UniversalClient
	prefix string
	ttl    time.Duration
}

func NewSessionState(client redis.UniversalClient, prefix string) *SessionState {
	if prefix == "" {
		prefix = "ai:session:"
	}
	return &SessionState{client: client, prefix: prefix, ttl: defaultTTL}
}

func NewCheckpointStore(client redis.UniversalClient, prefix string) *CheckpointStore {
	if prefix == "" {
		prefix = "ai:checkpoint:"
	}
	return &CheckpointStore{client: client, prefix: prefix, ttl: defaultTTL}
}

func (s *SessionState) Save(ctx context.Context, snapshot SessionSnapshot) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("session state is not initialized")
	}
	now := time.Now().UTC()
	if snapshot.CreatedAt.IsZero() {
		snapshot.CreatedAt = now
	}
	snapshot.UpdatedAt = now
	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("marshal session snapshot: %w", err)
	}
	return s.client.Set(ctx, s.key(snapshot.SessionID), data, s.ttl).Err()
}

func (s *SessionState) Load(ctx context.Context, sessionID string) (*SessionSnapshot, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("session state is not initialized")
	}
	raw, err := s.client.Get(ctx, s.key(sessionID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var snapshot SessionSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, fmt.Errorf("unmarshal session snapshot: %w", err)
	}
	return &snapshot, nil
}

func (s *SessionState) AppendMessage(ctx context.Context, sessionID string, msg *schema.Message) error {
	snapshot, err := s.Load(ctx, sessionID)
	if err != nil {
		return err
	}
	if snapshot == nil {
		snapshot = &SessionSnapshot{
			SessionID: sessionID,
			Messages:  make([]StoredMessage, 0, 4),
			Context:   map[string]any{},
			CreatedAt: time.Now().UTC(),
		}
	}
	if msg != nil {
		snapshot.Messages = append(snapshot.Messages, StoredMessage{
			Role:       string(msg.Role),
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
			ToolName:   msg.ToolName,
			Timestamp:  time.Now().UTC(),
		})
	}
	return s.Save(ctx, *snapshot)
}

func (s *SessionState) EnsureTitle(ctx context.Context, sessionID, title string) error {
	snapshot, err := s.Load(ctx, sessionID)
	if err != nil {
		return err
	}
	if snapshot == nil {
		snapshot = &SessionSnapshot{
			SessionID: sessionID,
			Messages:  make([]StoredMessage, 0, 4),
			Context:   map[string]any{},
			CreatedAt: time.Now().UTC(),
		}
	}
	if strings.TrimSpace(snapshot.Title) == "" {
		snapshot.Title = strings.TrimSpace(title)
	}
	return s.Save(ctx, *snapshot)
}

func (s *SessionState) UpdateTitle(ctx context.Context, sessionID, title string) error {
	snapshot, err := s.Load(ctx, sessionID)
	if err != nil {
		return err
	}
	if snapshot == nil {
		return fmt.Errorf("session snapshot not found")
	}
	snapshot.Title = strings.TrimSpace(title)
	return s.Save(ctx, *snapshot)
}

func (s *SessionState) Delete(ctx context.Context, sessionID string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("session state is not initialized")
	}
	return s.client.Del(ctx, s.key(sessionID)).Err()
}

func (s *SessionState) Clone(ctx context.Context, fromID, toID, title string) (*SessionSnapshot, error) {
	snapshot, err := s.Load(ctx, fromID)
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		return nil, nil
	}
	out := *snapshot
	out.SessionID = toID
	if strings.TrimSpace(title) != "" {
		out.Title = strings.TrimSpace(title)
	}
	out.CreatedAt = time.Now().UTC()
	out.UpdatedAt = out.CreatedAt
	if err := s.Save(ctx, out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SessionState) List(ctx context.Context) ([]SessionSnapshot, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("session state is not initialized")
	}
	keys, err := s.client.Keys(ctx, s.prefix+"*").Result()
	if err != nil {
		return nil, err
	}
	out := make([]SessionSnapshot, 0, len(keys))
	for _, key := range keys {
		raw, err := s.client.Get(ctx, key).Bytes()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, err
		}
		var snapshot SessionSnapshot
		if err := json.Unmarshal(raw, &snapshot); err != nil {
			return nil, fmt.Errorf("unmarshal session snapshot: %w", err)
		}
		out = append(out, snapshot)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out, nil
}

func (c *CheckpointStore) Get(ctx context.Context, checkPointID string) ([]byte, bool, error) {
	if c == nil || c.client == nil {
		return nil, false, fmt.Errorf("checkpoint store is not initialized")
	}
	raw, err := c.client.Get(ctx, c.key(checkPointID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, err
	}
	return raw, true, nil
}

func (c *CheckpointStore) Set(ctx context.Context, checkPointID string, checkPoint []byte) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("checkpoint store is not initialized")
	}
	return c.client.Set(ctx, c.key(checkPointID), checkPoint, c.ttl).Err()
}

func (c *CheckpointStore) ComposeStore() compose.CheckPointStore {
	return c
}

func (s *SessionState) key(sessionID string) string {
	return s.prefix + sessionID
}

func (c *CheckpointStore) key(checkpointID string) string {
	return c.prefix + checkpointID
}
