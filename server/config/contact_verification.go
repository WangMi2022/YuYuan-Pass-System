package config

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	ContactVerificationProviderWebhook = "webhook"
	defaultVerificationEmailSubject    = "账号安全验证码"
)

// ContactVerification controls self-service phone and email verification.
// A channel is available only when it is both enabled and fully configured.
type ContactVerification struct {
	SMS   ContactVerificationSMS   `mapstructure:"sms" json:"sms" yaml:"sms"`
	Email ContactVerificationEmail `mapstructure:"email" json:"email" yaml:"email"`
}

type ContactVerificationSMS struct {
	Enabled     bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Provider    string `mapstructure:"provider" json:"provider" yaml:"provider"`
	Endpoint    string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`
	AccessToken string `mapstructure:"access-token" json:"access-token" yaml:"access-token"`
	SignName    string `mapstructure:"sign-name" json:"sign-name" yaml:"sign-name"`
	TemplateID  string `mapstructure:"template-id" json:"template-id" yaml:"template-id"`
}

type ContactVerificationEmail struct {
	Enabled bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Subject string `mapstructure:"subject" json:"subject" yaml:"subject"`
}

func (c *ContactVerification) Normalize() {
	c.SMS.Provider = strings.ToLower(strings.TrimSpace(c.SMS.Provider))
	c.SMS.Endpoint = strings.TrimSpace(c.SMS.Endpoint)
	c.SMS.AccessToken = strings.TrimSpace(c.SMS.AccessToken)
	c.SMS.SignName = strings.TrimSpace(c.SMS.SignName)
	c.SMS.TemplateID = strings.TrimSpace(c.SMS.TemplateID)
	c.Email.Subject = strings.TrimSpace(c.Email.Subject)
	if c.Email.Subject == "" {
		c.Email.Subject = defaultVerificationEmailSubject
	}
}

func (c ContactVerification) Validate(smtp Email) error {
	c.Normalize()
	if c.SMS.Enabled && !c.SMS.Ready() {
		return fmt.Errorf("短信验证码配置不完整，无法开启")
	}
	if c.Email.Enabled && !c.Email.Ready(smtp) {
		return fmt.Errorf("邮件验证码配置不完整，无法开启")
	}
	return nil
}

func (c ContactVerificationSMS) Ready() bool {
	c.Provider = strings.ToLower(strings.TrimSpace(c.Provider))
	if c.Provider != ContactVerificationProviderWebhook ||
		strings.TrimSpace(c.Endpoint) == "" ||
		strings.TrimSpace(c.AccessToken) == "" ||
		strings.TrimSpace(c.SignName) == "" ||
		strings.TrimSpace(c.TemplateID) == "" {
		return false
	}
	endpoint, err := url.ParseRequestURI(strings.TrimSpace(c.Endpoint))
	return err == nil && (endpoint.Scheme == "http" || endpoint.Scheme == "https") && endpoint.Host != ""
}

func (c ContactVerificationEmail) Ready(smtp Email) bool {
	return strings.TrimSpace(smtp.From) != "" &&
		strings.TrimSpace(smtp.Host) != "" &&
		strings.TrimSpace(smtp.Secret) != "" &&
		smtp.Port > 0 && smtp.Port <= 65535
}
