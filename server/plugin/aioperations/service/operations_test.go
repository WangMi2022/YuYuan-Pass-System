package service

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/ai"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestCreatePromptValidatesSchema(t *testing.T) {
	_, err := (operationsService{}).CreatePrompt(context.Background(), PromptInput{
		PromptKey: "asset-draft", Content: "生成资产草稿", OutputSchema: `{"type":`,
	}, 1)
	if ai.ErrorKind(err) != ai.ErrorTypeSchema {
		t.Fatalf("error kind = %q, want %q", ai.ErrorKind(err), ai.ErrorTypeSchema)
	}
}

func TestCreatePromptRetriesVersionConflict(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:ai-prompt-create?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&ai.PromptTemplate{}); err != nil {
		t.Fatal(err)
	}
	previous := global.GVA_DB
	global.GVA_DB = database
	t.Cleanup(func() { global.GVA_DB = previous })

	var injected atomic.Bool
	if err := database.Callback().Create().Before("gorm:create").Register("test:prompt-version-conflict", func(tx *gorm.DB) {
		prompt, ok := tx.Statement.Dest.(*ai.PromptTemplate)
		if !ok || injected.Swap(true) {
			return
		}
		tx.Exec("INSERT INTO ai_prompt_templates (created_at, updated_at, prompt_key, version, content, status, created_by) VALUES (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?, ?, ?)", prompt.PromptKey, prompt.Version, "并发版本", ai.PromptStatusDraft, 99)
	}); err != nil {
		t.Fatal(err)
	}

	created, err := (operationsService{}).CreatePrompt(context.Background(), PromptInput{
		PromptKey: "asset-draft", Content: "生成资产草稿", OutputSchema: `{"type":"object"}`,
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if created.Version != 2 {
		t.Fatalf("version = %d, want 2", created.Version)
	}
}
