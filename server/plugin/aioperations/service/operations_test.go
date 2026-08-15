package service

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/ai"
	"github.com/flipped-aurora/gin-vue-admin/server/config"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	systemService "github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/glebarez/sqlite"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func TestProviderAndInvoiceUpdatesPreserveBothSections(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte("ai: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	configurationStore := viper.New()
	configurationStore.SetConfigFile(configPath)
	if err := configurationStore.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	previousConfig, previousStore := global.GVA_CONFIG, global.GVA_VP
	global.GVA_CONFIG = config.Server{}
	global.GVA_VP = configurationStore
	t.Cleanup(func() {
		global.GVA_CONFIG = previousConfig
		global.GVA_VP = previousStore
	})

	start := make(chan struct{})
	errors := make(chan error, 2)
	go func() {
		<-start
		_, err := (operationsService{}).UpdateProviders(t.Context(), config.AI{Enabled: true})
		errors <- err
	}()
	go func() {
		<-start
		_, err := systemService.SystemConfigServiceApp.SetInvoiceRecognitionConfig(t.Context(), config.InvoiceRecognition{FallbackThreshold: 0.91})
		errors <- err
	}()
	close(start)
	for range 2 {
		if err := <-errors; err != nil {
			t.Fatal(err)
		}
	}

	if !global.GVA_CONFIG.AI.Enabled || global.GVA_CONFIG.AI.Invoice.FallbackThreshold != 0.91 {
		t.Fatalf("concurrent updates lost a configuration section: %#v", global.GVA_CONFIG.AI)
	}
	persistedStore := viper.New()
	persistedStore.SetConfigFile(configPath)
	if err := persistedStore.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	var persisted config.Server
	if err := persistedStore.Unmarshal(&persisted); err != nil {
		t.Fatal(err)
	}
	if !persisted.AI.Enabled || persisted.AI.Invoice.FallbackThreshold != 0.91 {
		t.Fatalf("persisted configuration lost a section: %#v", persisted.AI)
	}
}

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
