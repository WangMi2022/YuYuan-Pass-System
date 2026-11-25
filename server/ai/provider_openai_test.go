package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/config"
)

func TestOpenAIChatURL(t *testing.T) {
	cases := map[string]string{
		"https://example.com":                     "https://example.com/v1/chat/completions",
		"https://example.com/v1":                  "https://example.com/v1/chat/completions",
		"https://example.com/v1/chat/completions": "https://example.com/v1/chat/completions",
	}
	for input, expected := range cases {
		if actual := openAIChatURL(input); actual != expected {
			t.Fatalf("openAIChatURL(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestOpenAIContent(t *testing.T) {
	content, err := openAIContent([]byte(`"ready"`))
	if err != nil || content != "ready" {
		t.Fatalf("content = %q, err = %v", content, err)
	}
}

func TestOpenAIProviderReturnsFinishReason(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"choices":[{"message":{"content":"partial"},"finish_reason":"length"}],"usage":{"prompt_tokens":10,"completion_tokens":500}}`))
	}))
	defer server.Close()

	provider := openAICompatibleProvider{
		configuration: config.AIProvider{BaseURL: server.URL, Model: "test", TimeoutSeconds: 5},
		allowPrivate:  true,
	}
	result, err := provider.Complete(context.Background(), providerCall{Prompt: "test", MaxOutputTokens: 500})
	if err != nil {
		t.Fatal(err)
	}
	if result.Content != "partial" || result.OutputTokens != 500 || result.FinishReason != "length" {
		t.Fatalf("unexpected provider result: %#v", result)
	}
}
