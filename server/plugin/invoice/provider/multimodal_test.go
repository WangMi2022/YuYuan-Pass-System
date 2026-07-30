package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
)

func TestMultimodalRecognizerUsesImageInputAndParsesResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("unexpected authorization header")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		serialized, _ := json.Marshal(body)
		if !strings.Contains(string(serialized), "data:image/png;base64,") {
			t.Fatalf("request does not contain an image data URL: %s", serialized)
		}
		content := `{"invoiceType":"电子发票","invoiceNumber":"12345678","issueDate":"2026-07-29","sellerName":"示例公司","amountCents":1000,"taxCents":60,"totalCents":1060,"rawText":"发票原文","confidence":0.91,"fieldConfidences":{"invoiceNumber":0.96},"items":[]}`
		_ = json.NewEncoder(response).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]any{"content": "```json\n" + content + "\n```"}}},
		})
	}))
	defer server.Close()

	recognizer := MultimodalRecognizer{BaseURL: server.URL, APIKey: "test-key", Model: "vision-model", Timeout: time.Second, AllowPrivateEndpoints: true}
	result, err := recognizer.Recognize(context.Background(), Input{
		FileName: "invoice.png", ContentType: "image/png", Data: probePNG,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.InvoiceNumber != "12345678" || result.TotalCents != 1060 || result.IssueDate == nil {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestChatCompletionsURLAcceptsBaseOrFullEndpoint(t *testing.T) {
	tests := map[string]string{
		"https://api.example.com":                     "https://api.example.com/v1/chat/completions",
		"https://api.example.com/v1/":                 "https://api.example.com/v1/chat/completions",
		"https://api.example.com/v1/chat/completions": "https://api.example.com/v1/chat/completions",
		"https://api.example.com/v1?tenant=asset":     "https://api.example.com/v1/chat/completions?tenant=asset",
	}
	for input, expected := range tests {
		if actual := chatCompletionsURL(input); actual != expected {
			t.Fatalf("%s: got %s want %s", input, actual, expected)
		}
	}
}

func TestMultimodalRecognizerDetectsAnthropicProtocol(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestCount++
		switch request.URL.Path {
		case "/v1/chat/completions":
			_ = json.NewEncoder(response).Encode(map[string]any{
				"choices": []map[string]any{{"message": map[string]any{
					"content": "", "reasoning_content": "thinking",
				}}},
			})
		case "/v1/messages":
			if request.Header.Get("x-api-key") != "test-key" {
				t.Fatal("missing Anthropic API key header")
			}
			if request.Header.Get("anthropic-version") != "2023-06-01" {
				t.Fatal("missing Anthropic version header")
			}
			if request.Header.Get("Authorization") != "" {
				t.Fatal("Anthropic request should not use the Bearer header")
			}
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			serialized, _ := json.Marshal(body)
			if !strings.Contains(string(serialized), `"type":"base64"`) ||
				!strings.Contains(string(serialized), `"media_type":"image/png"`) {
				t.Fatalf("request does not contain an Anthropic image source: %s", serialized)
			}
			if body["max_tokens"] != float64(probeMaxTokens) {
				t.Fatalf("probe max_tokens = %#v, want %d", body["max_tokens"], probeMaxTokens)
			}
			_ = json.NewEncoder(response).Encode(map[string]any{
				"content": []map[string]any{{"type": "thinking", "thinking": "done"}, {"type": "text", "text": "OK"}},
			})
		default:
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
	}))
	defer server.Close()

	recognizer := MultimodalRecognizer{BaseURL: server.URL, APIKey: "test-key", Model: "vision-model", Timeout: time.Second, AllowPrivateEndpoints: true}
	protocol, err := recognizer.Probe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if protocol != config.MultimodalProtocolAnthropic || requestCount != 2 {
		t.Fatalf("detected protocol = %q, requests = %d", protocol, requestCount)
	}
}

func TestMultimodalRecognizerUsesStoredAnthropicProtocol(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestCount++
		if request.URL.Path != "/v1/messages" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		content := `{"invoiceType":"电子发票","invoiceNumber":"87654321","issueDate":"2026-07-29","sellerName":"示例公司","totalCents":1060,"rawText":"发票原文","confidence":0.93,"fieldConfidences":{},"items":[]}`
		_ = json.NewEncoder(response).Encode(map[string]any{
			"content": []map[string]any{{"type": "text", "text": content}},
		})
	}))
	defer server.Close()

	recognizer := MultimodalRecognizer{
		BaseURL: server.URL, APIKey: "test-key", Model: "vision-model",
		Protocol: config.MultimodalProtocolAnthropic, Timeout: time.Second, AllowPrivateEndpoints: true,
	}
	result, err := recognizer.Recognize(context.Background(), Input{
		FileName: "invoice.png", ContentType: "image/png", Data: probePNG,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.InvoiceNumber != "87654321" || requestCount != 1 {
		t.Fatalf("unexpected result or request count: %#v, %d", result, requestCount)
	}
}

func TestAnthropicMessagesURLAcceptsBaseOrFullEndpoint(t *testing.T) {
	tests := map[string]string{
		"https://api.example.com":                       "https://api.example.com/v1/messages",
		"https://api.example.com/v1/":                   "https://api.example.com/v1/messages",
		"https://api.example.com/anthropic":             "https://api.example.com/anthropic/v1/messages",
		"https://api.example.com/anthropic/v1/messages": "https://api.example.com/anthropic/v1/messages",
		"https://api.example.com/v1?tenant=asset":       "https://api.example.com/v1/messages?tenant=asset",
	}
	for input, expected := range tests {
		if actual := anthropicMessagesURL(input); actual != expected {
			t.Fatalf("%s: got %s want %s", input, actual, expected)
		}
	}
}

func TestMultimodalRecognizerDoesNotForwardAPIKeyThroughRedirect(t *testing.T) {
	destinationRequests := 0
	destination := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		destinationRequests++
		if request.Header.Get("x-api-key") != "" || request.Header.Get("Authorization") != "" {
			t.Errorf("provider credential was forwarded through redirect")
		}
		response.WriteHeader(http.StatusOK)
	}))
	defer destination.Close()

	redirect := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/messages" {
			t.Errorf("unexpected path: %s", request.URL.Path)
		}
		http.Redirect(response, request, destination.URL, http.StatusTemporaryRedirect)
	}))
	defer redirect.Close()

	recognizer := MultimodalRecognizer{
		BaseURL: redirect.URL, APIKey: "test-key", Model: "vision-model",
		Protocol: config.MultimodalProtocolAnthropic, Timeout: time.Second, AllowPrivateEndpoints: true,
	}
	if _, err := recognizer.Probe(context.Background()); err == nil {
		t.Fatal("redirect response was accepted")
	}
	if destinationRequests != 0 {
		t.Fatalf("redirect destination received %d requests", destinationRequests)
	}
}
