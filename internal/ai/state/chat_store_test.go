package state

import (
	"context"
	"testing"

	"github.com/cy77cc/OpsPilot/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestChatStoreFiltersBySceneAndPersistsAssistantMetadata(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:chat-store?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.AIChatSession{}, &model.AIChatMessage{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	store := NewChatStore(db)
	ctx := context.Background()

	if err := store.AppendUserMessage(ctx, "session-global", 1, "global", "全局对话", "你好"); err != nil {
		t.Fatalf("AppendUserMessage(global) error = %v", err)
	}
	assistantID, err := store.CreateAssistantMessage(ctx, "session-global", 1, "global", "全局对话")
	if err != nil {
		t.Fatalf("CreateAssistantMessage(global) error = %v", err)
	}
	if err := store.UpdateAssistantMessage(ctx, "session-global", assistantID, ChatMessageRecord{
		Content:  "最终回答",
		Thinking: "分析中",
		Status:   "completed",
		TraceID:  "trace-1",
		ThoughtChain: []map[string]any{
			{"key": "rewrite", "title": "理解你的问题", "status": "success", "content": "已提取目标"},
		},
		Recommendations: []map[string]any{
			{"id": "r-1", "title": "下一步"},
		},
		SummaryOutput: map[string]any{
			"headline": "服务当前运行正常",
		},
		RawEvidence: []string{"命令执行成功"},
	}); err != nil {
		t.Fatalf("UpdateAssistantMessage(global) error = %v", err)
	}

	if err := store.AppendUserMessage(ctx, "session-k8s", 1, "scene:k8s", "K8s 对话", "看 pod 日志"); err != nil {
		t.Fatalf("AppendUserMessage(k8s) error = %v", err)
	}

	rows, err := store.ListSessions(ctx, 1, "scene:k8s")
	if err != nil {
		t.Fatalf("ListSessions(scene:k8s) error = %v", err)
	}
	if len(rows) != 1 || rows[0].ID != "session-k8s" {
		t.Fatalf("ListSessions(scene:k8s) = %#v, want only session-k8s", rows)
	}

	row, err := store.GetSession(ctx, 1, "global", "session-global", true)
	if err != nil {
		t.Fatalf("GetSession(global) error = %v", err)
	}
	if row == nil || len(row.Messages) != 2 {
		t.Fatalf("GetSession(global) messages = %#v", row)
	}
	msg := row.Messages[1]
	if msg.TraceID != "trace-1" {
		t.Fatalf("TraceID = %q, want trace-1", msg.TraceID)
	}
	if len(msg.ThoughtChain) != 1 {
		t.Fatalf("ThoughtChain = %#v, want persisted stage", msg.ThoughtChain)
	}
	if len(msg.Recommendations) != 1 {
		t.Fatalf("Recommendations = %#v, want 1 item", msg.Recommendations)
	}
	if msg.SummaryOutput["headline"] != "服务当前运行正常" {
		t.Fatalf("SummaryOutput = %#v", msg.SummaryOutput)
	}
	if len(msg.RawEvidence) != 1 || msg.RawEvidence[0] != "命令执行成功" {
		t.Fatalf("RawEvidence = %#v", msg.RawEvidence)
	}
}
