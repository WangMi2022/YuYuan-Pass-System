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
