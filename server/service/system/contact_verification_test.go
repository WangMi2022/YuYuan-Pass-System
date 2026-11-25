package system

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/config"
)

func TestSendSMSVerificationCodeUsesConfiguredWebhookContract(t *testing.T) {
	var received smsVerificationPayload
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		authorization = request.Header.Get("Authorization")
		if err := json.NewDecoder(request.Body).Decode(&received); err != nil {
			t.Errorf("decode webhook payload: %v", err)
		}
		writer.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	configuration := config.ContactVerificationSMS{
		Provider:    config.ContactVerificationProviderWebhook,
		Endpoint:    server.URL,
		AccessToken: "test-token",
		SignName:    "资产平台",
		TemplateID:  "SMS_001",
	}
	if err := sendSMSVerificationCode(t.Context(), configuration, "17612345678", "123456"); err != nil {
		t.Fatalf("sendSMSVerificationCode() error = %v", err)
	}
	if authorization != "Bearer test-token" {
		t.Fatalf("authorization = %q", authorization)
	}
	if received.Phone != "17612345678" || received.Code != "123456" {
		t.Fatalf("payload target/code = %q/%q", received.Phone, received.Code)
	}
	if received.SignName != "资产平台" || received.TemplateID != "SMS_001" || received.Purpose != "update_contact" {
		t.Fatalf("payload metadata = %+v", received)
	}
}

func TestNormalizeContactVerificationTarget(t *testing.T) {
	if channel, target, err := normalizeContactVerificationTarget("email", "User@Example.com"); err != nil || channel != "email" || target != "user@example.com" {
		t.Fatalf("email normalization = %q/%q/%v", channel, target, err)
	}
	if _, _, err := normalizeContactVerificationTarget("phone", "123"); err == nil {
		t.Fatal("invalid phone was accepted")
	}
	if _, _, err := normalizeContactVerificationTarget("fax", "123"); err == nil {
		t.Fatal("unsupported channel was accepted")
	}
}
