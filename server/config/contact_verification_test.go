package config

import "testing"

func TestContactVerificationRequiresCompleteConfigurationBeforeEnable(t *testing.T) {
	configuration := ContactVerification{
		SMS: ContactVerificationSMS{Enabled: true},
	}
	if err := configuration.Validate(Email{}); err == nil {
		t.Fatal("enabled SMS verification accepted incomplete configuration")
	}

	configuration.SMS = ContactVerificationSMS{
		Enabled:     true,
		Provider:    ContactVerificationProviderWebhook,
		Endpoint:    "https://sms.example.com/send",
		AccessToken: "token",
		SignName:    "资产平台",
		TemplateID:  "SMS_001",
	}
	if err := configuration.Validate(Email{}); err != nil {
		t.Fatalf("complete SMS verification rejected: %v", err)
	}

	configuration.Email.Enabled = true
	if err := configuration.Validate(Email{}); err == nil {
		t.Fatal("enabled email verification accepted incomplete SMTP configuration")
	}
	if err := configuration.Validate(Email{
		From: "no-reply@example.com", Host: "smtp.example.com", Secret: "secret", Port: 465,
	}); err != nil {
		t.Fatalf("complete email verification rejected: %v", err)
	}
}

func TestContactVerificationDisabledChannelsMayRemainUnconfigured(t *testing.T) {
	configuration := ContactVerification{}
	if err := configuration.Validate(Email{}); err != nil {
		t.Fatalf("disabled verification rejected empty configuration: %v", err)
	}
	configuration.Normalize()
	if configuration.Email.Subject != defaultVerificationEmailSubject {
		t.Fatalf("default email subject = %q", configuration.Email.Subject)
	}
}
