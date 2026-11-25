package provider

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRestrictedProviderIPRanges(t *testing.T) {
	tests := map[string]bool{
		"0.0.0.1":              true,
		"127.0.0.1":            true,
		"10.0.0.8":             true,
		"172.16.0.1":           true,
		"192.168.1.1":          true,
		"169.254.169.254":      true,
		"100.64.0.1":           true,
		"::1":                  true,
		"fd00::1":              true,
		"8.8.8.8":              false,
		"2001:4860:4860::8888": false,
	}
	for value, want := range tests {
		t.Run(value, func(t *testing.T) {
			if got := restrictedProviderIP(net.ParseIP(value)); got != want {
				t.Fatalf("restrictedProviderIP(%q) = %v, want %v", value, got, want)
			}
		})
	}
}

func TestProviderHTTPClientRejectsPrivateEndpointByDefault(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		requestCount++
		_ = response
	}))
	defer server.Close()

	_, err := (&HTTPVerifier{Endpoint: server.URL, Timeout: time.Second}).Detect(t.Context())
	if err == nil || !strings.Contains(err.Error(), "受限内网地址") {
		t.Fatalf("private endpoint was not rejected: %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("private endpoint received %d requests", requestCount)
	}
}
