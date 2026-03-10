package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cy77cc/OpsPilot/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const defaultScene = "global"

type ChatStore struct {
	db *gorm.DB
}

type ChatMessageMetadata struct {
	ThoughtChain    []map[string]any `json:"thoughtChain,omitempty"`
	TraceID         string           `json:"traceId,omitempty"`
	Recommendations []map[string]any `json:"recommendations,omitempty"`
}

type ChatMessageRecord struct {
	ID              string
	Role            string
	Content         string
	Thinking        string
	Status          string
	ThoughtChain    []map[string]any
	TraceID         string
	Recommendations []map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ChatSessionRecord struct {
	ID        string
	Scene     string
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
	Messages  []ChatMessageRecord
}

func NewChatStore(db *gorm.DB) *ChatStore {
	if db == nil {
		return nil
	}
	return &ChatStore{db: db}
}

func (s *ChatStore) EnsureSession(ctx context.Context, sessionID string, userID uint64, scene, title string) error {
	if s == nil || s.db == nil {
		return nil
	}
	scene = normalizeScene(scene)
	title = strings.TrimSpace(title)

	var existing model.AIChatSession
	err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", sessionID, userID).Take(&existing).Error
	switch {
	case err == nil:
		updates := map[string]any{
			"scene":      scene,
			"updated_at": time.Now().UTC(),
		}
		if existing.Title == "" && title != "" {
			updates["title"] = title
		}
		return s.db.WithContext(ctx).Model(&model.AIChatSession{}).
			Where("id = ? AND user_id = ?", sessionID, userID).
			Updates(updates).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return s.db.WithContext(ctx).Create(&model.AIChatSession{
			ID:        sessionID,
			UserID:    userID,
			Scene:     scene,
			Title:     title,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}).Error
	default:
		return err
	}
}

func (s *ChatStore) AppendUserMessage(ctx context.Context, sessionID string, userID uint64, scene, title, content string) error {
	if err := s.EnsureSession(ctx, sessionID, userID, scene, title); err != nil {
		return err
	}
	msg := model.AIChatMessage{
		ID:        uuid.NewString(),
		SessionID: sessionID,
		Role:      "user",
		Content:   strings.TrimSpace(content),
		Status:    "completed",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := s.db.WithContext(ctx).Create(&msg).Error; err != nil {
		return err
	}
	return s.touchSession(ctx, sessionID)
}

func (s *ChatStore) CreateAssistantMessage(ctx context.Context, sessionID string, userID uint64, scene, title string) (string, error) {
	if err := s.EnsureSession(ctx, sessionID, userID, scene, title); err != nil {
		return "", err
	}
	id := uuid.NewString()
	msg := model.AIChatMessage{
		ID:        id,
		SessionID: sessionID,
		Role:      "assistant",
		Status:    "streaming",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := s.db.WithContext(ctx).Create(&msg).Error; err != nil {
		return "", err
	}
	return id, s.touchSession(ctx, sessionID)
}

func (s *ChatStore) UpdateAssistantMessage(ctx context.Context, sessionID, messageID string, patch ChatMessageRecord) error {
	if s == nil || s.db == nil || strings.TrimSpace(messageID) == "" {
		return nil
	}
	metaBytes, err := json.Marshal(ChatMessageMetadata{
		ThoughtChain:    patch.ThoughtChain,
		TraceID:         strings.TrimSpace(patch.TraceID),
		Recommendations: patch.Recommendations,
	})
	if err != nil {
		return fmt.Errorf("marshal assistant metadata: %w", err)
	}
	updates := map[string]any{
		"content":       patch.Content,
		"thinking":      patch.Thinking,
		"status":        strings.TrimSpace(patch.Status),
		"metadata_json": string(metaBytes),
		"updated_at":    time.Now().UTC(),
	}
	if err := s.db.WithContext(ctx).Model(&model.AIChatMessage{}).
		Where("id = ? AND session_id = ?", messageID, sessionID).
		Updates(updates).Error; err != nil {
		return err
	}
	return s.touchSession(ctx, sessionID)
}

func (s *ChatStore) UpdateTitle(ctx context.Context, userID uint64, sessionID, title string) error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.WithContext(ctx).Model(&model.AIChatSession{}).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Updates(map[string]any{
			"title":      strings.TrimSpace(title),
			"updated_at": time.Now().UTC(),
		}).Error
}

func (s *ChatStore) Delete(ctx context.Context, userID uint64, sessionID string) error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("session_id = ?", sessionID).Delete(&model.AIChatMessage{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND user_id = ?", sessionID, userID).Delete(&model.AIChatSession{}).Error
	})
}

func (s *ChatStore) Clone(ctx context.Context, userID uint64, fromID, toID, title string) (*ChatSessionRecord, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	row, err := s.GetSession(ctx, userID, "", fromID, true)
	if err != nil || row == nil {
		return row, err
	}
	clone := model.AIChatSession{
		ID:        toID,
		UserID:    userID,
		Scene:     row.Scene,
		Title:     firstNonEmpty(strings.TrimSpace(title), row.Title),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&clone).Error; err != nil {
			return err
		}
		for _, msg := range row.Messages {
			metaBytes, err := json.Marshal(ChatMessageMetadata{
				ThoughtChain:    msg.ThoughtChain,
				TraceID:         msg.TraceID,
				Recommendations: msg.Recommendations,
			})
			if err != nil {
				return err
			}
			copyMsg := model.AIChatMessage{
				ID:           uuid.NewString(),
				SessionID:    toID,
				Role:         msg.Role,
				Content:      msg.Content,
				Thinking:     msg.Thinking,
				Status:       msg.Status,
				MetadataJSON: string(metaBytes),
				CreatedAt:    msg.CreatedAt,
				UpdatedAt:    time.Now().UTC(),
			}
			if err := tx.Create(&copyMsg).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetSession(ctx, userID, row.Scene, toID, true)
}

func (s *ChatStore) ListSessions(ctx context.Context, userID uint64, scene string) ([]ChatSessionRecord, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	scene = normalizeScene(scene)
	var rows []model.AIChatSession
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND scene = ?", userID, scene).
		Order("updated_at DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]ChatSessionRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, ChatSessionRecord{
			ID:        row.ID,
			Scene:     row.Scene,
			Title:     firstNonEmpty(strings.TrimSpace(row.Title), "新对话"),
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

func (s *ChatStore) CurrentSession(ctx context.Context, userID uint64, scene string, includeMessages bool) (*ChatSessionRecord, error) {
	rows, err := s.ListSessions(ctx, userID, scene)
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return s.GetSession(ctx, userID, normalizeScene(scene), rows[0].ID, includeMessages)
}

func (s *ChatStore) GetSession(ctx context.Context, userID uint64, scene, sessionID string, includeMessages bool) (*ChatSessionRecord, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	var row model.AIChatSession
	query := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", sessionID, userID)
	if strings.TrimSpace(scene) != "" {
		query = query.Where("scene = ?", normalizeScene(scene))
	}
	if err := query.Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	out := &ChatSessionRecord{
		ID:        row.ID,
		Scene:     normalizeScene(row.Scene),
		Title:     firstNonEmpty(strings.TrimSpace(row.Title), "新对话"),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if !includeMessages {
		return out, nil
	}
	var msgs []model.AIChatMessage
	if err := s.db.WithContext(ctx).
		Where("session_id = ?", row.ID).
		Order("created_at ASC").
		Find(&msgs).Error; err != nil {
		return nil, err
	}
	out.Messages = make([]ChatMessageRecord, 0, len(msgs))
	for _, msg := range msgs {
		meta, err := parseChatMessageMetadata(msg.MetadataJSON)
		if err != nil {
			return nil, err
		}
		out.Messages = append(out.Messages, ChatMessageRecord{
			ID:              msg.ID,
			Role:            msg.Role,
			Content:         msg.Content,
			Thinking:        msg.Thinking,
			Status:          msg.Status,
			ThoughtChain:    meta.ThoughtChain,
			TraceID:         meta.TraceID,
			Recommendations: meta.Recommendations,
			CreatedAt:       msg.CreatedAt,
			UpdatedAt:       msg.UpdatedAt,
		})
	}
	return out, nil
}

func (s *ChatStore) touchSession(ctx context.Context, sessionID string) error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.WithContext(ctx).Model(&model.AIChatSession{}).
		Where("id = ?", sessionID).
		Update("updated_at", time.Now().UTC()).Error
}

func parseChatMessageMetadata(raw string) (ChatMessageMetadata, error) {
	if strings.TrimSpace(raw) == "" {
		return ChatMessageMetadata{}, nil
	}
	var meta ChatMessageMetadata
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return ChatMessageMetadata{}, fmt.Errorf("unmarshal assistant metadata: %w", err)
	}
	return meta, nil
}

func normalizeScene(scene string) string {
	scene = strings.TrimSpace(scene)
	if scene == "" {
		return defaultScene
	}
	return scene
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
