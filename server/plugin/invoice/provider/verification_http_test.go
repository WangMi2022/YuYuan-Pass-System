package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
)

func TestHTTPVerifierDetectsCompatibleGateway(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.Header.Get("Authorization") != "Bearer test-ak" ||
			request.Header.Get("X-API-Secret") != "test-sk" {
			t.Errorf("unexpected verification probe request")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil || body["probe"] != true {
			t.Errorf("unexpected probe body: %#v err=%v", body, err)
		}
		_ = json.NewEncoder(response).Encode(map[string]any{"code": 0, "msg": "ok", "data": map[string]any{}})
	}))
	defer server.Close()

	info, err := (&HTTPVerifier{
		Endpoint: server.URL, APIKey: "test-ak", SecretKey: "test-sk", Timeout: time.Second, AllowPrivateEndpoints: true,
	}).Detect(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if info.Provider != config.VerificationProviderHTTPCompatible || info.Protocol != config.VerificationProtocolHTTPJSONV1 {
		t.Fatalf("unexpected detection: %#v", info)
	}
}

func TestHTTPVerifierMapsCanonicalRequestAndOutcome(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var body struct {
			Invoice VerificationRequest `json:"invoice"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Errorf("decode canonical request: %v", err)
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		if body.Invoice.InvoiceNum != "INV-001" || body.Invoice.InvoiceType != "special_vat_invoice" {
			t.Errorf("unexpected canonical request: %#v", body.Invoice)
		}
		_ = json.NewEncoder(response).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"outcome": "valid", "verifyResult": "vendor-ok", "verifyMessage": "查验一致",
				"invalidSign": "N", "official": map[string]string{"invoiceNumber": "INV-001"},
			},
		})
	}))
	defer server.Close()

	result, err := (&HTTPVerifier{Endpoint: server.URL, Timeout: time.Second, AllowPrivateEndpoints: true}).Verify(t.Context(), VerificationRequest{
		InvoiceNum: "INV-001", InvoiceType: "special_vat_invoice",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != VerificationOutcomeValid || result.VerifyResult != "vendor-ok" ||
		result.Official["invoiceNumber"] != "INV-001" || result.RawPayload == "" {
		t.Fatalf("unexpected verification result: %#v", result)
	}
}

func TestHTTPVerifierRejectsUnknownOutcome(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(response).Encode(map[string]any{
			"code": 0, "data": map[string]any{"outcome": "vendor-specific-success"},
		})
	}))
	defer server.Close()
	if _, err := (&HTTPVerifier{Endpoint: server.URL, Timeout: time.Second, AllowPrivateEndpoints: true}).Verify(t.Context(), VerificationRequest{}); err == nil {
		t.Fatal("unknown vendor outcome was accepted")
	}
}

func TestHTTPVerifierDoesNotForwardCredentialsThroughRedirect(t *testing.T) {
	targetCalls := 0
	target := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		targetCalls++
		if request.Header.Get("Authorization") != "" || request.Header.Get("X-API-Secret") != "" {
			t.Errorf("verification credentials were forwarded through redirect")
		}
	}))
	defer target.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Location", target.URL)
		response.WriteHeader(http.StatusTemporaryRedirect)
	}))
	defer redirect.Close()

	_, err := (&HTTPVerifier{Endpoint: redirect.URL, APIKey: "secret-ak", SecretKey: "secret-sk", Timeout: time.Second, AllowPrivateEndpoints: true}).Verify(t.Context(), VerificationRequest{})
	if err == nil || targetCalls != 0 {
		t.Fatalf("redirect was followed or not rejected: err=%v calls=%d", err, targetCalls)
	}
}

func TestBaiduVerificationOutcomeIsCanonical(t *testing.T) {
	tests := map[string]struct {
		verifyResult string
		invalidSign  string
		want         VerificationOutcome
	}{
		"valid":     {"0001", "N", VerificationOutcomeValid},
		"voided":    {"0001", "Y", VerificationOutcomeVoided},
		"red":       {"0001", "H", VerificationOutcomeRedFlushed},
		"mismatch":  {"0006", "", VerificationOutcomeMismatch},
		"not-found": {"0009", "", VerificationOutcomeNotFound},
		"deferred":  {"1014", "", VerificationOutcomeDeferred},
		"expired":   {"1015", "", VerificationOutcomeExpired},
		"unknown":   {"9999", "", VerificationOutcomeUnavailable},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := baiduVerificationOutcome(test.verifyResult, test.invalidSign); got != test.want {
				t.Fatalf("outcome = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNewVerificationAdapterSelectsDetectedProvider(t *testing.T) {
	configuration := config.InvoiceRecognition{Verification: config.InvoiceVerificationProvider{
		Enabled: true, Provider: config.VerificationProviderHTTPCompatible,
		Protocol: config.VerificationProtocolHTTPJSONV1, Endpoint: "https://verify.example",
		TimeoutSeconds: 5,
	}}
	configuration.Normalize()
	adapter, err := NewVerificationAdapter(configuration)
	if err != nil {
		t.Fatal(err)
	}
	if adapter.Provider != config.VerificationProviderHTTPCompatible || adapter.Protocol != config.VerificationProtocolHTTPJSONV1 {
		t.Fatalf("unexpected adapter metadata: %#v", adapter)
	}
	if _, ok := adapter.Verifier.(*HTTPVerifier); !ok {
		t.Fatalf("unexpected adapter implementation: %T", adapter.Verifier)
	}
}
