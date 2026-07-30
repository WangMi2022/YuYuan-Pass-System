package config

import (
	"errors"
	"net/url"
	"strings"
)

const (
	defaultOCRTimeoutSeconds        = 30
	defaultBaiduTimeoutSeconds      = 30
	defaultMultimodalTimeoutSeconds = 45
	defaultFallbackThreshold        = 0.82

	MultimodalProtocolOpenAICompatible = "openai-compatible"
	MultimodalProtocolAnthropic        = "anthropic"
)

// InvoiceRecognition contains the runtime providers used by the invoice worker.
// API keys are persisted in YAML but never serialized back to the browser.
type InvoiceRecognition struct {
	FallbackThreshold float64                   `mapstructure:"fallback-threshold" json:"fallback-threshold" yaml:"fallback-threshold"`
	Baidu             InvoiceBaiduProvider      `mapstructure:"baidu" json:"baidu" yaml:"baidu"`
	PublicOCR         InvoicePublicOCR          `mapstructure:"public-ocr" json:"public-ocr" yaml:"public-ocr"`
	Multimodal        InvoiceMultimodalProvider `mapstructure:"multimodal" json:"multimodal" yaml:"multimodal"`
}

type InvoiceBaiduProvider struct {
	Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	VerificationEnabled bool   `mapstructure:"verification-enabled" json:"verification-enabled" yaml:"verification-enabled"`
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
	Endpoint         string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`
	APIKey           string `mapstructure:"api-key" json:"-" yaml:"api-key"`
	TimeoutSeconds   int    `mapstructure:"timeout-seconds" json:"timeout-seconds" yaml:"timeout-seconds"`
	APIKeyInput      string `mapstructure:"-" json:"api-key,omitempty" yaml:"-"`
	APIKeyConfigured bool   `mapstructure:"-" json:"api-key-configured" yaml:"-"`
	ClearAPIKey      bool   `mapstructure:"-" json:"clear-api-key,omitempty" yaml:"-"`
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
	c.PublicOCR.Endpoint = strings.TrimSpace(c.PublicOCR.Endpoint)
	c.Multimodal.BaseURL = strings.TrimSpace(c.Multimodal.BaseURL)
	c.Multimodal.Model = strings.TrimSpace(c.Multimodal.Model)
	c.Multimodal.Protocol = strings.ToLower(strings.TrimSpace(c.Multimodal.Protocol))
	if c.PublicOCR.Provider == "" {
		c.PublicOCR.Provider = "http-compatible"
	}
	if c.PublicOCR.TimeoutSeconds <= 0 {
		c.PublicOCR.TimeoutSeconds = defaultOCRTimeoutSeconds
	}
	if c.Baidu.TimeoutSeconds <= 0 {
		c.Baidu.TimeoutSeconds = defaultBaiduTimeoutSeconds
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
	if len(c.PublicOCR.Endpoint) > 2048 || len(c.Multimodal.BaseURL) > 2048 {
		return errors.New("识别服务地址长度超出限制")
	}
	if len(c.Multimodal.Model) > 200 {
		return errors.New("多模态模型名称不能超过 200 个字符")
	}
	if c.Multimodal.Protocol != "" &&
		c.Multimodal.Protocol != MultimodalProtocolOpenAICompatible &&
		c.Multimodal.Protocol != MultimodalProtocolAnthropic {
		return errors.New("多模态模型协议不正确，请重新测试连接")
	}
	if len(c.PublicOCR.APIKey) > 8192 || len(c.Multimodal.APIKey) > 8192 {
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
	if c.Baidu.VerificationEnabled && !c.Baidu.Enabled {
		return errors.New("启用百度发票验真前请先启用百度发票 OCR")
	}
	if c.Baidu.Enabled && (strings.TrimSpace(c.Baidu.APIKey) == "" || strings.TrimSpace(c.Baidu.SecretKey) == "") {
		return errors.New("请完整配置百度智能云 API Key 和 Secret Key")
	}
	if c.Multimodal.TimeoutSeconds < 1 || c.Multimodal.TimeoutSeconds > 120 {
		return errors.New("多模态模型超时时间必须在 1 到 120 秒之间")
	}
	if c.PublicOCR.Enabled {
		if err := validateProviderURL(c.PublicOCR.Endpoint); err != nil {
			return errors.New("公网 OCR 接口地址不正确")
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
	multimodalConnectionUnchanged := strings.TrimSpace(c.Multimodal.BaseURL) == strings.TrimSpace(current.Multimodal.BaseURL) &&
		strings.TrimSpace(c.Multimodal.Model) == strings.TrimSpace(current.Multimodal.Model) &&
		strings.TrimSpace(c.Multimodal.APIKeyInput) == "" && !c.Multimodal.ClearAPIKey
	if strings.TrimSpace(c.Multimodal.Protocol) == "" && multimodalConnectionUnchanged {
		// Older clients do not submit the hidden protocol field. Preserve a
		// previously detected value only while the connection identity is unchanged.
		c.Multimodal.Protocol = current.Multimodal.Protocol
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
	c.Multimodal.APIKeyInput = ""
	c.Multimodal.APIKeyConfigured = false
	c.Multimodal.ClearAPIKey = false
	return c
}

func validateProviderURL(value string) error {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("invalid provider URL")
	}
	return nil
}
