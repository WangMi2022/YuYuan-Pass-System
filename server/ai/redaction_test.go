package ai

import "testing"

func TestRedactMasksSensitiveValues(t *testing.T) {
	input := "phone=13800138000 email=test@example.com token=abc123 ip=10.0.0.1 id=110101199001011234"
	output, count := Redact(input)
	if count < 5 {
		t.Fatalf("redaction count = %d, want at least 5", count)
	}
	for _, raw := range []string{"13800138000", "test@example.com", "abc123", "10.0.0.1", "110101199001011234"} {
		if contains(output, raw) {
			t.Fatalf("sensitive value %q remains in %q", raw, output)
		}
	}
}

func TestHashIsStable(t *testing.T) {
	if Hash("same") != Hash("same") || Hash("same") == Hash("different") {
		t.Fatal("hash must be stable and content-sensitive")
	}
}

func TestRedactPayloadMasksNestedSecretsAndBusinessWords(t *testing.T) {
	payload := map[string]any{
		"taxNo":  "91350211M000100Y43",
		"nested": []any{map[string]any{"access_token": "secret-token", "note": "天枢计划"}},
	}
	redacted, count := RedactPayload(payload, []string{"天枢计划"})
	encoded := DecodePayload(redacted)
	if count < 3 || contains(encoded, "91350211M000100Y43") || contains(encoded, "secret-token") || contains(encoded, "天枢计划") {
		t.Fatalf("unexpected redaction: count=%d payload=%s", count, encoded)
	}
}

func contains(value, fragment string) bool {
	for index := 0; index+len(fragment) <= len(value); index++ {
		if value[index:index+len(fragment)] == fragment {
			return true
		}
	}
	return false
}
