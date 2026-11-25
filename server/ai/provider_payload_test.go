package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/config"
)

func TestOpenAICompatibleProviderSendsBusinessPayload(t *testing.T) {
	assertProviderSendsBusinessPayload(t, func(baseURL string) (*http.Response, error) {
		provider := openAICompatibleProvider{
			configuration: config.AIProvider{BaseURL: baseURL, Model: "test-model", TimeoutSeconds: 5},
			allowPrivate:  true,
		}
		return provider.do(context.Background(), providerCall{
			Prompt:  "只根据业务数据回答",
			Payload: map[string]any{"question": "今天有哪些日程？", "count": 0},
		}, false)
	})
}

func TestAnthropicProviderSendsBusinessPayload(t *testing.T) {
	assertProviderSendsBusinessPayload(t, func(baseURL string) (*http.Response, error) {
		provider := anthropicProvider{
			configuration: config.AIProvider{BaseURL: baseURL, Model: "test-model", TimeoutSeconds: 5},
			allowPrivate:  true,
		}
		return provider.do(context.Background(), providerCall{
			Prompt:  "只根据业务数据回答",
			Payload: map[string]any{"question": "今天有哪些日程？", "count": 0},
		}, false)
	})
}

func assertProviderSendsBusinessPayload(t *testing.T, invoke func(baseURL string) (*http.Response, error)) {
	t.Helper()
	requestBody := make(chan []byte, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requestBody <- body
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"content":[],"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer server.Close()

	response, err := invoke(server.URL)
	if err != nil {
		t.Fatalf("provider request failed: %v", err)
	}
	_ = response.Body.Close()

	var payload map[string]any
	if err := json.Unmarshal(<-requestBody, &payload); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	encoded, _ := json.Marshal(payload["messages"])
	messageText := string(encoded)
	for _, expected := range []string{"只根据业务数据回答", "今天有哪些日程？", `\"count\":0`} {
		if !strings.Contains(messageText, expected) {
			t.Fatalf("provider messages do not contain %q: %s", expected, messageText)
		}
	}
}
