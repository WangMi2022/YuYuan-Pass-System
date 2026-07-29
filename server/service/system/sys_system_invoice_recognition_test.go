package system

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
)

func TestPrepareInvoiceRecognitionConfigDetectsAndPreservesProtocol(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestCount++
		if request.URL.Path != "/v1/chat/completions" {
			response.WriteHeader(http.StatusNotFound)
			return
		}
		if request.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected authorization header: %q", request.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(response).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]any{"content": "OK"}}},
		})
	}))
	defer server.Close()

	incoming := config.InvoiceRecognition{
		FallbackThreshold: 0.82,
		Multimodal: config.InvoiceMultimodalProvider{
			Enabled: true, BaseURL: server.URL, Model: "vision-model",
			APIKeyInput: "test-key", Protocol: config.MultimodalProtocolAnthropic,
			TimeoutSeconds: 5,
		},
	}
	prepared, err := prepareInvoiceRecognitionConfig(context.Background(), incoming, config.InvoiceRecognition{}, true)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Multimodal.Protocol != config.MultimodalProtocolOpenAICompatible {
		t.Fatalf("protocol = %q, want OpenAI Compatible", prepared.Multimodal.Protocol)
	}
	if requestCount != 1 {
		t.Fatalf("request count = %d, want 1", requestCount)
	}

	unchanged := prepared.Redacted()
	preparedAgain, err := prepareInvoiceRecognitionConfig(context.Background(), unchanged, prepared, true)
	if err != nil {
		t.Fatal(err)
	}
	if preparedAgain.Multimodal.Protocol != config.MultimodalProtocolOpenAICompatible || requestCount != 1 {
		t.Fatalf("unchanged configuration was probed again: protocol=%q requests=%d", preparedAgain.Multimodal.Protocol, requestCount)
	}

	spoofed := prepared.Redacted()
	spoofed.Multimodal.Protocol = config.MultimodalProtocolAnthropic
	preparedAgain, err = prepareInvoiceRecognitionConfig(context.Background(), spoofed, prepared, true)
	if err != nil {
		t.Fatal(err)
	}
	if preparedAgain.Multimodal.Protocol != config.MultimodalProtocolOpenAICompatible || requestCount != 2 {
		t.Fatalf("client protocol override was trusted: protocol=%q requests=%d", preparedAgain.Multimodal.Protocol, requestCount)
	}
}

func TestPrepareInvoiceRecognitionConfigDoesNotPersistFailedDetection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	incoming := config.InvoiceRecognition{
		FallbackThreshold: 0.82,
		Multimodal: config.InvoiceMultimodalProvider{
			Enabled: true, BaseURL: server.URL, Model: "vision-model", APIKeyInput: "test-key", TimeoutSeconds: 5,
		},
	}
	prepared, err := prepareInvoiceRecognitionConfig(context.Background(), incoming, config.InvoiceRecognition{}, true)
	if err == nil {
		t.Fatal("failed protocol detection was accepted")
	}
	if prepared.Multimodal.Protocol != "" {
		t.Fatalf("failed detection returned a protocol: %q", prepared.Multimodal.Protocol)
	}
}
