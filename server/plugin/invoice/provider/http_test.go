package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPRecognizerRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"code":0,"data":{"rawPayload":"`))
		_, _ = response.Write([]byte(strings.Repeat("x", maxOCRResponseSize)))
		_, _ = response.Write([]byte(`"}}`))
	}))
	defer server.Close()

	recognizer := HTTPRecognizer{Endpoint: server.URL, Timeout: time.Second}
	_, err := recognizer.Recognize(context.Background(), Input{FileName: "invoice.png", Data: []byte("image")})
	if err == nil || !strings.Contains(err.Error(), "2MB") {
		t.Fatalf("expected oversized response error, got %v", err)
	}
}

func TestHTTPRecognizerProbeRequiresCompatibleSuccessEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("unexpected authorization header")
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"code":0,"msg":"ok","data":{}}`))
	}))
	defer server.Close()

	recognizer := HTTPRecognizer{Endpoint: server.URL, Token: "test-key", Timeout: time.Second}
	if err := recognizer.Probe(context.Background()); err != nil {
		t.Fatalf("probe failed: %v", err)
	}
}

func TestHTTPRecognizerProbeRejectsBusinessError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"code":401,"msg":"invalid token","data":{}}`))
	}))
	defer server.Close()

	recognizer := HTTPRecognizer{Endpoint: server.URL, Timeout: time.Second}
	if err := recognizer.Probe(context.Background()); err == nil || !strings.Contains(err.Error(), "invalid token") {
		t.Fatalf("expected business error, got %v", err)
	}
}
