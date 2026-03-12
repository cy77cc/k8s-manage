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
	RawEvidence     []string         `json:"rawEvidence,omitempty"`
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
	RawEvidence     []string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ChatBlockRecord struct {
	ID          string
	BlockType   string
	Position    int
	Status      string
	Title       string
	ContentText string
	ContentJSON map[string]any
	Streaming   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ChatTurnRecord struct {
	ID           string
	Role         string
	Status       string
	Phase        string
	TraceID      string
	ParentTurnID string
	Blocks       []ChatBlockRecord
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  *time.Time
}

type ChatSessionRecord struct {
	ID        string
	Scene     string
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
	Messages  []ChatMessageRecord
	Turns     []ChatTurnRecord
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
	if err := s.upsertTurnWithBlocks(ctx, model.AIChatTurn{
		ID:        msg.ID,
		SessionID: sessionID,
		Role:      "user",
		Status:    "completed",
		Phase:     "user_message",
		CreatedAt: msg.CreatedAt,
		UpdatedAt: msg.UpdatedAt,
		CompletedAt: func() *time.Time {
			ts := msg.UpdatedAt
			return &ts
		}(),
	}, []model.AIChatBlock{
		{
			ID:          uuid.NewString(),
			TurnID:      msg.ID,
			BlockType:   "text",
			Position:    1,
			Status:      "completed",
			ContentText: msg.Content,
			CreatedAt:   msg.CreatedAt,
			UpdatedAt:   msg.UpdatedAt,
		},
	}); err != nil {
		return err
	}
	return s.touchSession(ctx, sessionID)
}

func (s *ChatStore) CreateAssistantMessage(ctx context.Context, sessionID string, userID uint64, scene, title, turnID string) (string, error) {
	if err := s.EnsureSession(ctx, sessionID, userID, scene, title); err != nil {
		return "", err
	}
	id := uuid.NewString()
	turnID = firstNonEmpty(strings.TrimSpace(turnID), id)
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
	if err := s.upsertTurnWithBlocks(ctx, model.AIChatTurn{
		ID:        turnID,
		SessionID: sessionID,
		Role:      "assistant",
		Status:    "streaming",
		Phase:     "streaming",
		CreatedAt: msg.CreatedAt,
		UpdatedAt: msg.UpdatedAt,
	}, nil); err != nil {
		return "", err
	}
	return id, s.touchSession(ctx, sessionID)
}

func (s *ChatStore) UpdateAssistantMessage(ctx context.Context, sessionID, messageID, turnID string, patch ChatMessageRecord) error {
	if s == nil || s.db == nil || strings.TrimSpace(messageID) == "" {
		return nil
	}
	metaBytes, err := json.Marshal(ChatMessageMetadata{
		ThoughtChain:    patch.ThoughtChain,
		TraceID:         strings.TrimSpace(patch.TraceID),
		Recommendations: patch.Recommendations,
		RawEvidence:     patch.RawEvidence,
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
	if err := s.syncAssistantTurn(ctx, sessionID, firstNonEmpty(strings.TrimSpace(turnID), messageID), patch); err != nil {
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
		if err := tx.Where("turn_id IN (?)", tx.Model(&model.AIChatTurn{}).Select("id").Where("session_id = ?", sessionID)).Delete(&model.AIChatBlock{}).Error; err != nil {
			return err
		}
		if err := tx.Where("session_id = ?", sessionID).Delete(&model.AIChatTurn{}).Error; err != nil {
			return err
		}
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
				RawEvidence:     msg.RawEvidence,
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
		for _, turn := range row.Turns {
			newTurnID := uuid.NewString()
			copyTurn := model.AIChatTurn{
				ID:           newTurnID,
				SessionID:    toID,
				Role:         turn.Role,
				Status:       turn.Status,
				Phase:        turn.Phase,
				TraceID:      turn.TraceID,
				ParentTurnID: turn.ParentTurnID,
				CreatedAt:    turn.CreatedAt,
				UpdatedAt:    time.Now().UTC(),
				CompletedAt:  turn.CompletedAt,
			}
			if err := tx.Create(&copyTurn).Error; err != nil {
				return err
			}
			for _, block := range turn.Blocks {
				rawJSON := ""
				if len(block.ContentJSON) > 0 {
					data, err := json.Marshal(block.ContentJSON)
					if err != nil {
						return err
					}
					rawJSON = string(data)
				}
				copyBlock := model.AIChatBlock{
					ID:          uuid.NewString(),
					TurnID:      newTurnID,
					BlockType:   block.BlockType,
					Position:    block.Position,
					Status:      block.Status,
					Title:       block.Title,
					ContentText: block.ContentText,
					ContentJSON: rawJSON,
					Streaming:   block.Streaming,
					CreatedAt:   block.CreatedAt,
					UpdatedAt:   time.Now().UTC(),
				}
				if err := tx.Create(&copyBlock).Error; err != nil {
					return err
				}
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
			RawEvidence:     meta.RawEvidence,
			CreatedAt:       msg.CreatedAt,
			UpdatedAt:       msg.UpdatedAt,
		})
	}
	turns, err := s.loadTurns(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	out.Turns = turns
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

func assistantPhaseFromPatch(patch ChatMessageRecord) string {
	switch strings.TrimSpace(patch.Status) {
	case "completed":
		return "done"
	case "error":
		return "error"
	default:
		if strings.TrimSpace(patch.Content) != "" {
			return "summary"
		}
		return "streaming"
	}
}

func stableBlockID(turnID, suffix string) string {
	return fmt.Sprintf("%s:%s", strings.TrimSpace(turnID), strings.TrimSpace(suffix))
}

func collectTurnIDs(turns []model.AIChatTurn) []string {
	out := make([]string, 0, len(turns))
	for _, turn := range turns {
		if strings.TrimSpace(turn.ID) != "" {
			out = append(out, turn.ID)
		}
	}
	return out
}

func parseBlockJSON(raw string) (map[string]any, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err == nil {
		return out, nil
	}
	var array []any
	if err := json.Unmarshal([]byte(raw), &array); err != nil {
		return nil, fmt.Errorf("unmarshal block payload: %w", err)
	}
	return map[string]any{"items": array}, nil
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

func (s *ChatStore) upsertTurnWithBlocks(ctx context.Context, turn model.AIChatTurn, blocks []model.AIChatBlock) error {
	if s == nil || s.db == nil || strings.TrimSpace(turn.ID) == "" {
		return nil
	}
	now := time.Now().UTC()
	if turn.CreatedAt.IsZero() {
		turn.CreatedAt = now
	}
	if turn.UpdatedAt.IsZero() {
		turn.UpdatedAt = now
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.AIChatTurn
		err := tx.Where("id = ?", turn.ID).Take(&existing).Error
		switch {
		case err == nil:
			updates := map[string]any{
				"session_id":     strings.TrimSpace(turn.SessionID),
				"role":           strings.TrimSpace(turn.Role),
				"status":         strings.TrimSpace(turn.Status),
				"phase":          strings.TrimSpace(turn.Phase),
				"trace_id":       strings.TrimSpace(turn.TraceID),
				"parent_turn_id": strings.TrimSpace(turn.ParentTurnID),
				"updated_at":     turn.UpdatedAt,
				"completed_at":   turn.CompletedAt,
			}
			if err := tx.Model(&model.AIChatTurn{}).Where("id = ?", turn.ID).Updates(updates).Error; err != nil {
				return err
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			if err := tx.Create(&turn).Error; err != nil {
				return err
			}
		default:
			return err
		}

		if blocks == nil {
			return nil
		}
		if err := tx.Where("turn_id = ?", turn.ID).Delete(&model.AIChatBlock{}).Error; err != nil {
			return err
		}
		if len(blocks) == 0 {
			return nil
		}
		for i := range blocks {
			if strings.TrimSpace(blocks[i].ID) == "" {
				blocks[i].ID = uuid.NewString()
			}
			blocks[i].TurnID = turn.ID
			if blocks[i].CreatedAt.IsZero() {
				blocks[i].CreatedAt = turn.CreatedAt
			}
			if blocks[i].UpdatedAt.IsZero() {
				blocks[i].UpdatedAt = turn.UpdatedAt
			}
		}
		return tx.Create(&blocks).Error
	})
}

func (s *ChatStore) syncAssistantTurn(ctx context.Context, sessionID, turnID string, patch ChatMessageRecord) error {
	turnID = strings.TrimSpace(turnID)
	if s == nil || s.db == nil || turnID == "" {
		return nil
	}
	now := time.Now().UTC()
	status := firstNonEmpty(strings.TrimSpace(patch.Status), "streaming")
	var completedAt *time.Time
	if status == "completed" || status == "error" {
		completedAt = &now
	}
	return s.upsertTurnWithBlocks(ctx, model.AIChatTurn{
		ID:          turnID,
		SessionID:   strings.TrimSpace(sessionID),
		Role:        "assistant",
		Status:      status,
		Phase:       assistantPhaseFromPatch(patch),
		TraceID:     strings.TrimSpace(patch.TraceID),
		UpdatedAt:   now,
		CompletedAt: completedAt,
	}, assistantBlocksFromPatch(turnID, patch, now))
}

func assistantBlocksFromPatch(turnID string, patch ChatMessageRecord, now time.Time) []model.AIChatBlock {
	blocks := make([]model.AIChatBlock, 0, len(patch.ThoughtChain)+4)
	position := 1
	addBlock := func(idSuffix, blockType, status, title, contentText string, contentJSON any, streaming bool) {
		rawJSON := ""
		if contentJSON != nil {
			if data, err := json.Marshal(contentJSON); err == nil {
				rawJSON = string(data)
			}
		}
		blocks = append(blocks, model.AIChatBlock{
			ID:          fmt.Sprintf("%s:%s", strings.TrimSpace(turnID), idSuffix),
			BlockType:   blockType,
			Position:    position,
			Status:      strings.TrimSpace(status),
			Title:       strings.TrimSpace(title),
			ContentText: contentText,
			ContentJSON: rawJSON,
			Streaming:   streaming,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
		position++
	}
	for idx, stage := range patch.ThoughtChain {
		title := firstNonEmpty(asString(stage["title"]), asString(stage["key"]), "执行阶段")
		contentText := firstNonEmpty(asString(stage["content"]), asString(stage["description"]), asString(stage["footer"]))
		addBlock(fmt.Sprintf("status-%d", idx+1), "status", firstNonEmpty(asString(stage["status"]), patch.Status), title, contentText, stage, patch.Status == "streaming")
	}
	if strings.TrimSpace(patch.Thinking) != "" {
		addBlock("thinking", "thinking", firstNonEmpty(patch.Status, "streaming"), "思考过程", patch.Thinking, map[string]any{"traceId": patch.TraceID}, patch.Status == "streaming")
	}
	if strings.TrimSpace(patch.Content) != "" {
		addBlock("text", "text", firstNonEmpty(patch.Status, "streaming"), "最终回答", patch.Content, nil, patch.Status == "streaming")
	}
	if len(patch.Recommendations) > 0 {
		addBlock("recommendations", "recommendations", firstNonEmpty(patch.Status, "completed"), "推荐下一步", "", patch.Recommendations, false)
	}
	if len(patch.RawEvidence) > 0 {
		addBlock("evidence", "evidence", firstNonEmpty(patch.Status, "completed"), "原始执行证据", "", map[string]any{"items": patch.RawEvidence}, false)
	}
	return blocks
}

func (s *ChatStore) loadTurns(ctx context.Context, sessionID string) ([]ChatTurnRecord, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	var turns []model.AIChatTurn
	if err := s.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&turns).Error; err != nil {
		return nil, err
	}
	if len(turns) == 0 {
		return nil, nil
	}
	var blocks []model.AIChatBlock
	if err := s.db.WithContext(ctx).
		Where("turn_id IN ?", collectTurnIDs(turns)).
		Order("position ASC, created_at ASC").
		Find(&blocks).Error; err != nil {
		return nil, err
	}
	blockMap := make(map[string][]ChatBlockRecord, len(turns))
	for _, block := range blocks {
		payload, err := parseBlockJSON(block.ContentJSON)
		if err != nil {
			return nil, err
		}
		blockMap[block.TurnID] = append(blockMap[block.TurnID], ChatBlockRecord{
			ID:          block.ID,
			BlockType:   block.BlockType,
			Position:    block.Position,
			Status:      block.Status,
			Title:       block.Title,
			ContentText: block.ContentText,
			ContentJSON: payload,
			Streaming:   block.Streaming,
			CreatedAt:   block.CreatedAt,
			UpdatedAt:   block.UpdatedAt,
		})
	}
	out := make([]ChatTurnRecord, 0, len(turns))
	for _, turn := range turns {
		out = append(out, ChatTurnRecord{
			ID:           turn.ID,
			Role:         turn.Role,
			Status:       turn.Status,
			Phase:        turn.Phase,
			TraceID:      turn.TraceID,
			ParentTurnID: turn.ParentTurnID,
			Blocks:       blockMap[turn.ID],
			CreatedAt:    turn.CreatedAt,
			UpdatedAt:    turn.UpdatedAt,
			CompletedAt:  turn.CompletedAt,
		})
	}
	return out, nil
}

func asString(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		return ""
	}
}
