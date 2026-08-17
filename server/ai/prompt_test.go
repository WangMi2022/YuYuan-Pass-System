package ai

import (
	"context"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestResolvePromptUsesActiveAndExplicitVersion(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:ai-prompt-resolve?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&PromptTemplate{}); err != nil {
		t.Fatal(err)
	}
	previous := global.GVA_DB
	global.GVA_DB = database
	t.Cleanup(func() { global.GVA_DB = previous })

	templates := []PromptTemplate{
		{PromptKey: "asset-draft", Version: 1, Content: "旧模板", Status: PromptStatusRetired},
		{PromptKey: "asset-draft", Version: 2, Content: "当前模板", OutputSchema: `{"type":"object"}`, Status: PromptStatusActive},
	}
	if err := database.Create(&templates).Error; err != nil {
		t.Fatal(err)
	}
	active, err := resolvePrompt(context.Background(), CompletionRequest{PromptKey: "asset-draft", Prompt: "业务输入"})
	if err != nil {
		t.Fatal(err)
	}
	if active.Prompt != "当前模板\n\n业务输入" || active.PromptVersion != 2 || active.OutputSchema == "" {
		t.Fatalf("unexpected active prompt: %#v", active)
	}
	explicit, err := resolvePrompt(context.Background(), CompletionRequest{PromptKey: "asset-draft", PromptVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	if explicit.Prompt != "旧模板" || explicit.PromptVersion != 1 {
		t.Fatalf("unexpected explicit prompt: %#v", explicit)
	}
}
