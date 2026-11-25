package config

import (
	"errors"
	"net/url"
	"strings"
)

const (
	defaultOCRTimeoutSeconds          = 30
	defaultBaiduTimeoutSeconds        = 30
	defaultVerificationTimeoutSeconds = 30
	defaultMultimodalTimeoutSeconds   = 45
	defaultFallbackThreshold          = 0.82

	OCRProviderHTTPCompatible  = "http-compatible"
	OCRProviderBaidu           = "baidu"
	OCRProtocolBaiduVATV1      = "baidu-vat-ocr-v1"
	OCRProtocolMultipartJSONV1 = "multipart-json-v1"

	VerificationProviderBaidu          = "baidu"
	VerificationProviderHTTPCompatible = "http-compatible"
	VerificationProtocolBaiduVATV1     = "baidu-vat-invoice-v1"
	VerificationProtocolHTTPJSONV1     = "invoice-verification-json-v1"

	MultimodalProtocolOpenAICompatible = "openai-compatible"
	MultimodalProtocolAnthropic        = "anthropic"
)

// InvoiceRecognition contains the runtime providers used by the invoice worker.
// API keys are persisted in YAML but never serialized back to the browser.
type InvoiceRecognition struct {
	FallbackThreshold     float64                     `mapstructure:"fallback-threshold" json:"fallback-threshold" yaml:"fallback-threshold"`
	AllowPrivateEndpoints bool                        `mapstructure:"allow-private-endpoints" json:"allow-private-endpoints" yaml:"allow-private-endpoints"`
	Baidu                 InvoiceBaiduProvider        `mapstructure:"baidu" json:"baidu" yaml:"baidu"`
	PublicOCR             InvoicePublicOCR            `mapstructure:"public-ocr" json:"public-ocr" yaml:"public-ocr"`
	Verification          InvoiceVerificationProvider `mapstructure:"verification" json:"verification" yaml:"verification"`
	Multimodal            InvoiceMultimodalProvider   `mapstructure:"multimodal" json:"multimodal" yaml:"multimodal"`
}

type InvoiceBaiduProvider struct {
	Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	APIKey              string `mapstructure:"api-key" json:"-" yaml:"api-key"`
	SecretKey           string `mapstructure:"secret-key" json:"-" yaml:"secret-key"`
	TimeoutSeconds      int    `mapstructure:"timeout-seconds" json:"timeout-seconds" yaml:"timeout-seconds"`
	APIKeyInput         string `mapstructure:"-" json:"api-key,omitempty" yaml:"-"`
	SecretKeyInput      string `mapstructure:"-" json:"secret-key,omitempty" yaml:"-"`
	APIKeyConfigured    bool   `mapstructure:"-" json:"api-key-configured" yaml:"-"`
	SecretKeyConfigured bool   `mapstructure:"-" json:"secret-key-configured" yaml:"-"`
	ClearAPIKey         bool   `mapstructure:"-" json:"clear-api-key,omitempty" yaml:"-"`
	ClearSecretKey      bool   `mapstructure:"-" json:"clear-secret-key,omitempty" yaml:"-"`
}

type InvoicePublicOCR struct {
	Enabled          bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Provider         string `mapstructure:"provider" json:"provider" yaml:"provider"`
	Protocol         string `mapstructure:"protocol" json:"protocol" yaml:"protocol"`
	Endpoint         string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`
	APIKey           string `mapstructure:"api-key" json:"-" yaml:"api-key"`
	TimeoutSeconds   int    `mapstructure:"timeout-seconds" json:"timeout-seconds" yaml:"timeout-seconds"`
	APIKeyInput      string `mapstructure:"-" json:"api-key,omitempty" yaml:"-"`
	APIKeyConfigured bool   `mapstructure:"-" json:"api-key-configured" yaml:"-"`
	ClearAPIKey      bool   `mapstructure:"-" json:"clear-api-key,omitempty" yaml:"-"`
}

// InvoiceVerificationProvider is independent from OCR. Provider and Protocol
// are detected by the server and persisted; browser clients never choose them.
type InvoiceVerificationProvider struct {
	Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Provider            string `mapstructure:"provider" json:"provider" yaml:"provider"`
	Protocol            string `mapstructure:"protocol" json:"protocol" yaml:"protocol"`
	Endpoint            string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`
	APIKey              string `mapstructure:"api-key" json:"-" yaml:"api-key"`
	SecretKey           string `mapstructure:"secret-key" json:"-" yaml:"secret-key"`
	TimeoutSeconds      int    `mapstructure:"timeout-seconds" json:"timeout-seconds" yaml:"timeout-seconds"`
	APIKeyInput         string `mapstructure:"-" json:"api-key,omitempty" yaml:"-"`
	SecretKeyInput      string `mapstructure:"-" json:"secret-key,omitempty" yaml:"-"`
	APIKeyConfigured    bool   `mapstructure:"-" json:"api-key-configured" yaml:"-"`
	SecretKeyConfigured bool   `mapstructure:"-" json:"secret-key-configured" yaml:"-"`
	ClearAPIKey         bool   `mapstructure:"-" json:"clear-api-key,omitempty" yaml:"-"`
	ClearSecretKey      bool   `mapstructure:"-" json:"clear-secret-key,omitempty" yaml:"-"`
}

type InvoiceMultimodalProvider struct {
	Enabled          bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	BaseURL          string `mapstructure:"base-url" json:"base-url" yaml:"base-url"`
	APIKey           string `mapstructure:"api-key" json:"-" yaml:"api-key"`
	Model            string `mapstructure:"model" json:"model" yaml:"model"`
	Protocol         string `mapstructure:"protocol" json:"protocol" yaml:"protocol"`
	TimeoutSeconds   int    `mapstructure:"timeout-seconds" json:"timeout-seconds" yaml:"timeout-seconds"`
	APIKeyInput      string `mapstructure:"-" json:"api-key,omitempty" yaml:"-"`
	APIKeyConfigured bool   `mapstructure:"-" json:"api-key-configured" yaml:"-"`
	ClearAPIKey      bool   `mapstructure:"-" json:"clear-api-key,omitempty" yaml:"-"`
}

func (c *InvoiceRecognition) Normalize() {
	c.PublicOCR.Provider = strings.TrimSpace(c.PublicOCR.Provider)
	c.PublicOCR.Protocol = strings.ToLower(strings.TrimSpace(c.PublicOCR.Protocol))
	c.PublicOCR.Endpoint = strings.TrimSpace(c.PublicOCR.Endpoint)
	c.Verification.Provider = strings.ToLower(strings.TrimSpace(c.Verification.Provider))
	c.Verification.Protocol = strings.ToLower(strings.TrimSpace(c.Verification.Protocol))
	c.Verification.Endpoint = strings.TrimSpace(c.Verification.Endpoint)
	c.Multimodal.BaseURL = strings.TrimSpace(c.Multimodal.BaseURL)
	c.Multimodal.Model = strings.TrimSpace(c.Multimodal.Model)
	c.Multimodal.Protocol = strings.ToLower(strings.TrimSpace(c.Multimodal.Protocol))
	if c.PublicOCR.Protocol == "" && c.PublicOCR.Provider == OCRProviderHTTPCompatible {
		c.PublicOCR.Protocol = OCRProtocolMultipartJSONV1
	}
	if c.PublicOCR.Provider == "" && c.PublicOCR.Protocol == OCRProtocolMultipartJSONV1 {
		c.PublicOCR.Provider = OCRProviderHTTPCompatible
	}
	if c.Verification.Protocol == "" {
		switch c.Verification.Provider {
		case VerificationProviderBaidu:
			c.Verification.Protocol = VerificationProtocolBaiduVATV1
		case VerificationProviderHTTPCompatible:
			c.Verification.Protocol = VerificationProtocolHTTPJSONV1
		}
	}
	if c.Verification.Provider == "" {
		switch c.Verification.Protocol {
		case VerificationProtocolBaiduVATV1:
			c.Verification.Provider = VerificationProviderBaidu
		case VerificationProtocolHTTPJSONV1:
			c.Verification.Provider = VerificationProviderHTTPCompatible
		}
	}
	if c.PublicOCR.TimeoutSeconds <= 0 {
		c.PublicOCR.TimeoutSeconds = defaultOCRTimeoutSeconds
	}
	if c.Baidu.TimeoutSeconds <= 0 {
		c.Baidu.TimeoutSeconds = defaultBaiduTimeoutSeconds
	}
	if c.Verification.TimeoutSeconds <= 0 {
		c.Verification.TimeoutSeconds = defaultVerificationTimeoutSeconds
	}
	if c.Multimodal.TimeoutSeconds <= 0 {
		c.Multimodal.TimeoutSeconds = defaultMultimodalTimeoutSeconds
	}
	if c.FallbackThreshold <= 0 {
		c.FallbackThreshold = defaultFallbackThreshold
	}
}

func (c InvoiceRecognition) Validate() error {
	if len(c.PublicOCR.Provider) > 50 {
		return errors.New("公网 OCR 服务标识不能超过 50 个字符")
	}
	if len(c.Verification.Provider) > 50 {
		return errors.New("权威验真服务标识不能超过 50 个字符")
	}
	if len(c.PublicOCR.Endpoint) > 2048 || len(c.Verification.Endpoint) > 2048 || len(c.Multimodal.BaseURL) > 2048 {
		return errors.New("识别服务地址长度超出限制")
	}
	if c.PublicOCR.Protocol != "" && c.PublicOCR.Protocol != OCRProtocolMultipartJSONV1 {
		return errors.New("公网 OCR 协议不正确，请重新测试连接")
	}
	if c.Verification.Protocol != "" &&
		c.Verification.Protocol != VerificationProtocolBaiduVATV1 &&
		c.Verification.Protocol != VerificationProtocolHTTPJSONV1 {
		return errors.New("权威验真协议不正确，请重新测试连接")
	}
	if c.Verification.Provider != "" &&
		c.Verification.Provider != VerificationProviderBaidu &&
		c.Verification.Provider != VerificationProviderHTTPCompatible {
		return errors.New("权威验真供应商不正确，请重新测试连接")
	}
	if len(c.Multimodal.Model) > 200 {
		return errors.New("多模态模型名称不能超过 200 个字符")
	}
	if c.Multimodal.Protocol != "" &&
		c.Multimodal.Protocol != MultimodalProtocolOpenAICompatible &&
		c.Multimodal.Protocol != MultimodalProtocolAnthropic {
		return errors.New("多模态模型协议不正确，请重新测试连接")
	}
	if len(c.PublicOCR.APIKey) > 8192 || len(c.Verification.APIKey) > 8192 ||
		len(c.Verification.SecretKey) > 8192 || len(c.Multimodal.APIKey) > 8192 {
		return errors.New("识别服务 API Key 长度超出限制")
	}
	if len(c.Baidu.APIKey) > 8192 || len(c.Baidu.SecretKey) > 8192 {
		return errors.New("百度智能云凭据长度超出限制")
	}
	if c.FallbackThreshold < 0.1 || c.FallbackThreshold > 1 {
		return errors.New("多模态兜底阈值必须在 0.1 到 1 之间")
	}
	if c.PublicOCR.TimeoutSeconds < 1 || c.PublicOCR.TimeoutSeconds > 120 {
		return errors.New("公网 OCR 超时时间必须在 1 到 120 秒之间")
	}
	if c.Baidu.TimeoutSeconds < 1 || c.Baidu.TimeoutSeconds > 120 {
		return errors.New("百度智能云超时时间必须在 1 到 120 秒之间")
	}
	if c.Baidu.Enabled && (strings.TrimSpace(c.Baidu.APIKey) == "" || strings.TrimSpace(c.Baidu.SecretKey) == "") {
		return errors.New("请完整配置百度智能云 API Key 和 Secret Key")
	}
	if c.Verification.TimeoutSeconds < 1 || c.Verification.TimeoutSeconds > 120 {
		return errors.New("权威验真超时时间必须在 1 到 120 秒之间")
	}
	if c.Multimodal.TimeoutSeconds < 1 || c.Multimodal.TimeoutSeconds > 120 {
		return errors.New("多模态模型超时时间必须在 1 到 120 秒之间")
	}
	if c.PublicOCR.Enabled {
		if err := validateProviderURL(c.PublicOCR.Endpoint); err != nil {
			return errors.New("公网 OCR 接口地址不正确")
		}
		if c.PublicOCR.Provider == "" || c.PublicOCR.Protocol == "" {
			return errors.New("公网 OCR 尚未完成协议探测，请重新测试连接")
		}
	}
	if c.Verification.Enabled {
		if c.Verification.Provider == "" || c.Verification.Protocol == "" {
			return errors.New("权威验真尚未完成供应商探测，请重新测试连接")
		}
		switch c.Verification.Provider {
		case VerificationProviderBaidu:
			if c.Verification.Protocol != VerificationProtocolBaiduVATV1 ||
				strings.TrimSpace(c.Verification.Endpoint) != "" ||
				strings.TrimSpace(c.Verification.APIKey) == "" || strings.TrimSpace(c.Verification.SecretKey) == "" {
				return errors.New("百度权威验真请留空接口地址，并完整配置 API Key 和 Secret Key")
			}
		case VerificationProviderHTTPCompatible:
			if c.Verification.Protocol != VerificationProtocolHTTPJSONV1 || validateProviderURL(c.Verification.Endpoint) != nil {
				return errors.New("权威验真接口地址或协议不正确")
			}
		}
	}
	if c.Multimodal.Enabled {
		if err := validateProviderURL(c.Multimodal.BaseURL); err != nil {
			return errors.New("多模态模型 Base URL 不正确")
		}
		if c.Multimodal.Model == "" {
			return errors.New("请配置多模态模型名称")
		}
	}
	return nil
}

func (c InvoiceRecognition) Redacted() InvoiceRecognition {
	c.Baidu.APIKeyConfigured = strings.TrimSpace(c.Baidu.APIKey) != ""
	c.Baidu.SecretKeyConfigured = strings.TrimSpace(c.Baidu.SecretKey) != ""
	c.Baidu.APIKeyInput = ""
	c.Baidu.SecretKeyInput = ""
	c.Baidu.ClearAPIKey = false
	c.Baidu.ClearSecretKey = false
	c.PublicOCR.APIKeyConfigured = strings.TrimSpace(c.PublicOCR.APIKey) != ""
	c.PublicOCR.APIKeyInput = ""
	c.PublicOCR.ClearAPIKey = false
	c.Verification.APIKeyConfigured = strings.TrimSpace(c.Verification.APIKey) != ""
	c.Verification.SecretKeyConfigured = strings.TrimSpace(c.Verification.SecretKey) != ""
	c.Verification.APIKeyInput = ""
	c.Verification.SecretKeyInput = ""
	c.Verification.ClearAPIKey = false
	c.Verification.ClearSecretKey = false
	c.Multimodal.APIKeyConfigured = strings.TrimSpace(c.Multimodal.APIKey) != ""
	c.Multimodal.APIKeyInput = ""
	c.Multimodal.ClearAPIKey = false
	return c
}

// MergeSecrets applies write-only browser fields while preserving stored keys.
// Non-admin callers cannot change any invoice recognition setting.
func (c InvoiceRecognition) MergeSecrets(current InvoiceRecognition, allow bool) InvoiceRecognition {
	if !allow {
		return current
	}
	if c.PublicOCR.ClearAPIKey {
		c.PublicOCR.APIKey = ""
	} else if strings.TrimSpace(c.PublicOCR.APIKeyInput) != "" {
		c.PublicOCR.APIKey = strings.TrimSpace(c.PublicOCR.APIKeyInput)
	} else {
		c.PublicOCR.APIKey = current.PublicOCR.APIKey
	}
	if c.Baidu.ClearAPIKey {
		c.Baidu.APIKey = ""
	} else if strings.TrimSpace(c.Baidu.APIKeyInput) != "" {
		c.Baidu.APIKey = strings.TrimSpace(c.Baidu.APIKeyInput)
	} else {
		c.Baidu.APIKey = current.Baidu.APIKey
	}
	if c.Baidu.ClearSecretKey {
		c.Baidu.SecretKey = ""
	} else if strings.TrimSpace(c.Baidu.SecretKeyInput) != "" {
		c.Baidu.SecretKey = strings.TrimSpace(c.Baidu.SecretKeyInput)
	} else {
		c.Baidu.SecretKey = current.Baidu.SecretKey
	}
	if c.Verification.ClearAPIKey {
		c.Verification.APIKey = ""
	} else if strings.TrimSpace(c.Verification.APIKeyInput) != "" {
		c.Verification.APIKey = strings.TrimSpace(c.Verification.APIKeyInput)
	} else {
		c.Verification.APIKey = current.Verification.APIKey
	}
	if c.Verification.ClearSecretKey {
		c.Verification.SecretKey = ""
	} else if strings.TrimSpace(c.Verification.SecretKeyInput) != "" {
		c.Verification.SecretKey = strings.TrimSpace(c.Verification.SecretKeyInput)
	} else {
		c.Verification.SecretKey = current.Verification.SecretKey
	}
	if c.Multimodal.ClearAPIKey {
		c.Multimodal.APIKey = ""
	} else if strings.TrimSpace(c.Multimodal.APIKeyInput) != "" {
		c.Multimodal.APIKey = strings.TrimSpace(c.Multimodal.APIKeyInput)
	} else {
		c.Multimodal.APIKey = current.Multimodal.APIKey
	}
	c.PublicOCR.APIKeyInput = ""
	c.PublicOCR.APIKeyConfigured = false
	c.PublicOCR.ClearAPIKey = false
	c.Baidu.APIKeyInput = ""
	c.Baidu.SecretKeyInput = ""
	c.Baidu.APIKeyConfigured = false
	c.Baidu.SecretKeyConfigured = false
	c.Baidu.ClearAPIKey = false
	c.Baidu.ClearSecretKey = false
	c.Verification.APIKeyInput = ""
	c.Verification.SecretKeyInput = ""
	c.Verification.APIKeyConfigured = false
	c.Verification.SecretKeyConfigured = false
	c.Verification.ClearAPIKey = false
	c.Verification.ClearSecretKey = false
	c.Multimodal.APIKeyInput = ""
	c.Multimodal.APIKeyConfigured = false
	c.Multimodal.ClearAPIKey = false
	return c
}

func validateProviderURL(value string) error {
	// Provider endpoints are intentionally configurable for enterprise/private
	// gateways. Only super administrators can write this setting, requests do
	// not follow redirects, and credentials are never forwarded to a new host.
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("invalid provider URL")
	}
	return nil
}
