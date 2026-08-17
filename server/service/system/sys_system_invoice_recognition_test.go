package system

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/config"
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
		FallbackThreshold: 0.82, AllowPrivateEndpoints: true,
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
		FallbackThreshold: 0.82, AllowPrivateEndpoints: true,
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

func TestPrepareInvoiceRecognitionConfigDetectsPublicOCRProtocol(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestCount++
		if request.Header.Get("Authorization") != "Bearer ocr-key" {
			t.Errorf("unexpected OCR authorization header")
		}
		_ = json.NewEncoder(response).Encode(map[string]any{"code": 0, "data": map[string]any{}})
	}))
	defer server.Close()

	incoming := config.InvoiceRecognition{AllowPrivateEndpoints: true, PublicOCR: config.InvoicePublicOCR{
		Enabled: true, Endpoint: server.URL, APIKeyInput: "ocr-key", TimeoutSeconds: 5,
	}}
	prepared, err := prepareInvoiceRecognitionConfig(t.Context(), incoming, config.InvoiceRecognition{}, true)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.PublicOCR.Provider != config.OCRProviderHTTPCompatible ||
		prepared.PublicOCR.Protocol != config.OCRProtocolMultipartJSONV1 || requestCount != 1 {
		t.Fatalf("public OCR was not detected once: %#v requests=%d", prepared.PublicOCR, requestCount)
	}
	preparedAgain, err := prepareInvoiceRecognitionConfig(t.Context(), prepared.Redacted(), prepared, true)
	if err != nil {
		t.Fatal(err)
	}
	if preparedAgain.PublicOCR.Protocol != config.OCRProtocolMultipartJSONV1 || requestCount != 1 {
		t.Fatalf("unchanged public OCR was probed again: %#v requests=%d", preparedAgain.PublicOCR, requestCount)
	}
}

func TestPrepareInvoiceRecognitionConfigDetectsVerificationProvider(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestCount++
		if request.Header.Get("Authorization") != "Bearer verify-key" || request.Header.Get("X-API-Secret") != "verify-secret" {
			t.Errorf("unexpected verification credentials")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil || body["probe"] != true {
			t.Errorf("unexpected verification probe: %#v err=%v", body, err)
		}
		_ = json.NewEncoder(response).Encode(map[string]any{"code": 0, "data": map[string]any{}})
	}))
	defer server.Close()

	incoming := config.InvoiceRecognition{AllowPrivateEndpoints: true, Verification: config.InvoiceVerificationProvider{
		Enabled: true, Endpoint: server.URL, APIKeyInput: "verify-key", SecretKeyInput: "verify-secret", TimeoutSeconds: 5,
	}}
	prepared, err := prepareInvoiceRecognitionConfig(t.Context(), incoming, config.InvoiceRecognition{}, true)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Verification.Provider != config.VerificationProviderHTTPCompatible ||
		prepared.Verification.Protocol != config.VerificationProtocolHTTPJSONV1 || requestCount != 1 {
		t.Fatalf("verification provider was not detected once: %#v requests=%d", prepared.Verification, requestCount)
	}
	preparedAgain, err := prepareInvoiceRecognitionConfig(t.Context(), prepared.Redacted(), prepared, true)
	if err != nil {
		t.Fatal(err)
	}
	if preparedAgain.Verification.Protocol != config.VerificationProtocolHTTPJSONV1 || requestCount != 1 {
		t.Fatalf("unchanged verification provider was probed again: %#v requests=%d", preparedAgain.Verification, requestCount)
	}
}

func TestPrepareInvoiceRecognitionConfigRejectsFailedVerificationDetection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	incoming := config.InvoiceRecognition{AllowPrivateEndpoints: true, Verification: config.InvoiceVerificationProvider{
		Enabled: true, Endpoint: server.URL, APIKeyInput: "invalid-key", TimeoutSeconds: 5,
	}}
	prepared, err := prepareInvoiceRecognitionConfig(t.Context(), incoming, config.InvoiceRecognition{}, true)
	if err == nil {
		t.Fatal("failed verification detection was accepted")
	}
	if prepared.Verification.Provider != "" || prepared.Verification.Protocol != "" {
		t.Fatalf("failed detection returned metadata: %#v", prepared.Verification)
	}
}
