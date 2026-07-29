package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

	recognizer := MultimodalRecognizer{BaseURL: server.URL, APIKey: "test-key", Model: "vision-model", Timeout: time.Second}
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
