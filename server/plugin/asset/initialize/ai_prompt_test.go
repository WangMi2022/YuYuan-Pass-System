package initialize

import (
	"encoding/json"
	"strings"
	"testing"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
)

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
