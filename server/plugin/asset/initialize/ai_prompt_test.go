package initialize

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/ai"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/glebarez/sqlite"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
	"gorm.io/gorm"
)

func useAssetPromptTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open asset prompt test database: %v", err)
	}
	if err = database.AutoMigrate(&ai.PromptTemplate{}); err != nil {
		t.Fatalf("migrate asset prompt table: %v", err)
	}
	previous := global.GVA_DB
	global.GVA_DB = database
	t.Cleanup(func() {
		global.GVA_DB = previous
		if sqlDB, dbErr := database.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	})
	return database
}

func compileAssetRecognitionOutputSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("asset-recognition-output.json", strings.NewReader(assetRecognitionOutputSchema)); err != nil {
		t.Fatalf("add asset recognition schema: %v", err)
	}
	compiled, err := compiler.Compile("asset-recognition-output.json")
	if err != nil {
		t.Fatalf("compile asset recognition schema: %v", err)
	}
	return compiled
}

func decodeAssetRecognitionOutput(t *testing.T, content string) any {
	t.Helper()
	var output any
	if err := json.Unmarshal([]byte(content), &output); err != nil {
		t.Fatalf("decode asset recognition output: %v", err)
	}
	return output
}

func TestAssetRecognitionOutputSchemaAcceptsPartialResult(t *testing.T) {
	output := decodeAssetRecognitionOutput(t, `{
		"brand":"Codex Lab",
		"model":"QA-2026",
		"serialNumber":"QA-M4-20260814-001",
		"productionDate":"2026-08-01",
		"rawText":"TEST ASSET NAMEPLATE",
		"fieldConfidences":{"brand":0.99,"model":0.99,"serialNumber":0.99}
	}`)
	if err := compileAssetRecognitionOutputSchema(t).Validate(output); err != nil {
		t.Fatalf("partial recognition result should remain reviewable: %v", err)
	}
}

func TestAssetRecognitionOutputSchemaRejectsUnknownField(t *testing.T) {
	output := decodeAssetRecognitionOutput(t, `{"unexpected":"value"}`)
	if err := compileAssetRecognitionOutputSchema(t).Validate(output); err == nil {
		t.Fatal("unknown recognition field should be rejected")
	}
}

func TestSeedAssetRecognitionPromptUpgradesLegacyActiveVersion(t *testing.T) {
	database := useAssetPromptTestDatabase(t)
	legacy := ai.PromptTemplate{
		PromptKey: assetRecognitionPromptKey, Version: 1, Content: "legacy prompt",
		OutputSchema: legacyAssetRecognitionOutputSchema, Status: ai.PromptStatusActive,
	}
	if err := database.Create(&legacy).Error; err != nil {
		t.Fatalf("create legacy prompt: %v", err)
	}

	seedAssetRecognitionPrompt(context.Background())

	var templates []ai.PromptTemplate
	if err := database.Where("prompt_key = ?", assetRecognitionPromptKey).Order("version ASC").Find(&templates).Error; err != nil {
		t.Fatalf("load upgraded prompts: %v", err)
	}
	if len(templates) != 2 {
		t.Fatalf("expected two prompt versions, got %d", len(templates))
	}
	if templates[0].Version != 1 || templates[0].Status != ai.PromptStatusRetired || templates[0].ActivatedAt != nil {
		t.Fatalf("legacy prompt was not retired: %#v", templates[0])
	}
	if templates[1].Version != 2 || templates[1].Status != ai.PromptStatusActive || templates[1].ActivatedAt == nil {
		t.Fatalf("upgraded prompt was not activated: %#v", templates[1])
	}
	if templates[1].Content != assetRecognitionPrompt || templates[1].OutputSchema != assetRecognitionOutputSchema {
		t.Fatalf("upgraded prompt does not match built-in version: %#v", templates[1])
	}
}

func TestSeedAssetRecognitionPromptUpgradeIsIdempotent(t *testing.T) {
	database := useAssetPromptTestDatabase(t)
	legacy := ai.PromptTemplate{
		PromptKey: assetRecognitionPromptKey, Version: 1, Content: "legacy prompt",
		OutputSchema: legacyAssetRecognitionOutputSchema, Status: ai.PromptStatusActive,
	}
	if err := database.Create(&legacy).Error; err != nil {
		t.Fatalf("create legacy prompt: %v", err)
	}

	seedAssetRecognitionPrompt(context.Background())
	seedAssetRecognitionPrompt(context.Background())

	var count int64
	if err := database.Model(&ai.PromptTemplate{}).Where("prompt_key = ?", assetRecognitionPromptKey).Count(&count).Error; err != nil {
		t.Fatalf("count prompt versions: %v", err)
	}
	if count != 2 {
		t.Fatalf("idempotent migration created unexpected versions: %d", count)
	}
	var active ai.PromptTemplate
	if err := database.Where("prompt_key = ? AND status = ?", assetRecognitionPromptKey, ai.PromptStatusActive).First(&active).Error; err != nil {
		t.Fatalf("load active prompt: %v", err)
	}
	if active.Version != 2 {
		t.Fatalf("unexpected active prompt version after repeated migration: %d", active.Version)
	}
}

func TestSeedAssetRecognitionPromptPreservesCustomizedActiveVersion(t *testing.T) {
	database := useAssetPromptTestDatabase(t)
	customSchema := `{"type":"object","additionalProperties":false,"properties":{"custom":{"type":"string"}}}`
	custom := ai.PromptTemplate{
		PromptKey: assetRecognitionPromptKey, Version: 7, Content: "administrator customized prompt",
		OutputSchema: customSchema, Status: ai.PromptStatusActive,
	}
	if err := database.Create(&custom).Error; err != nil {
		t.Fatalf("create customized prompt: %v", err)
	}

	seedAssetRecognitionPrompt(context.Background())

	var templates []ai.PromptTemplate
	if err := database.Where("prompt_key = ?", assetRecognitionPromptKey).Find(&templates).Error; err != nil {
		t.Fatalf("load customized prompt: %v", err)
	}
	if len(templates) != 1 {
		t.Fatalf("customized prompt should not be versioned automatically: %d", len(templates))
	}
	if templates[0].Version != 7 || templates[0].Content != custom.Content || templates[0].OutputSchema != customSchema || templates[0].Status != ai.PromptStatusActive {
		t.Fatalf("customized prompt was changed: %#v", templates[0])
	}
}
