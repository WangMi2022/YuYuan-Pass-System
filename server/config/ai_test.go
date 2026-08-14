package config

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAIRedactedJSONDoesNotExposeKeys(t *testing.T) {
	configuration := AI{
		OpenAICompatible: AIProvider{APIKey: "openai-secret", APIKeyInput: "replacement"},
		Anthropic:        AIProvider{APIKey: "anthropic-secret"},
	}
	encoded, err := json.Marshal(configuration.Redacted())
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	for _, secret := range []string{"openai-secret", "anthropic-secret", "replacement"} {
		if strings.Contains(text, secret) {
			t.Fatalf("redacted JSON contains %q: %s", secret, text)
		}
	}
	if !strings.Contains(text, `"api-key-configured":true`) {
		t.Fatalf("redacted JSON does not report configured key: %s", text)
	}
}

func TestAIMergeSecretsPreservesOmittedKey(t *testing.T) {
	current := AI{OpenAICompatible: AIProvider{APIKey: "existing"}}
	incoming := AI{OpenAICompatible: AIProvider{Model: "new-model"}}
	merged := incoming.MergeSecrets(current, true)
	if merged.OpenAICompatible.APIKey != "existing" || merged.OpenAICompatible.Model != "new-model" {
		t.Fatalf("unexpected merged configuration: %#v", merged.OpenAICompatible)
	}
}

func TestAIProviderURLRejectsQueryCredentials(t *testing.T) {
	if err := validateAIProviderURL("https://api.example.com/v1?api-key=secret"); err == nil {
		t.Fatal("provider URL query must be rejected")
	}
}
