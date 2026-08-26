package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/config"
)

func TestAnthropicProviderReturnsStopReason(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"content":[{"type":"text","text":"partial"}],"stop_reason":"max_tokens","usage":{"input_tokens":10,"output_tokens":500}}`))
	}))
	defer server.Close()

	provider := anthropicProvider{
		configuration: config.AIProvider{BaseURL: server.URL, Model: "test", TimeoutSeconds: 5},
		allowPrivate:  true,
	}
	result, err := provider.Complete(context.Background(), providerCall{Prompt: "test", MaxOutputTokens: 500})
	if err != nil {
		t.Fatal(err)
	}
	if result.Content != "partial" || result.OutputTokens != 500 || result.FinishReason != "max_tokens" {
		t.Fatalf("unexpected provider result: %#v", result)
	}
}
