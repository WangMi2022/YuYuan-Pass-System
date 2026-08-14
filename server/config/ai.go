package config

import (
	"errors"
	"net/url"
	"strings"
)

const (
	AIProviderOpenAICompatible = "openai-compatible"
	AIProviderAnthropic        = "anthropic"

	defaultAIProviderTimeoutSeconds = 60
)

// AI contains the centrally managed model-provider configuration. Credentials
// are write-only in browser responses and never included in normal JSON output.
type AI struct {
	Enabled               bool       `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	AllowPrivateEndpoints bool       `mapstructure:"allow-private-endpoints" json:"allow-private-endpoints" yaml:"allow-private-endpoints"`
	SensitiveWords        []string   `mapstructure:"sensitive-words" json:"sensitive-words" yaml:"sensitive-words"`
	AllowVisionModules    []string   `mapstructure:"allow-vision-modules" json:"allow-vision-modules" yaml:"allow-vision-modules"`
	OpenAICompatible      AIProvider `mapstructure:"openai-compatible" json:"openai-compatible" yaml:"openai-compatible"`
	Anthropic             AIProvider `mapstructure:"anthropic" json:"anthropic" yaml:"anthropic"`
}

type AIProvider struct {
	Enabled                    bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	BaseURL                    string `mapstructure:"base-url" json:"base-url" yaml:"base-url"`
	APIKey                     string `mapstructure:"api-key" json:"-" yaml:"api-key"`
	Model                      string `mapstructure:"model" json:"model" yaml:"model"`
	TimeoutSeconds             int    `mapstructure:"timeout-seconds" json:"timeout-seconds" yaml:"timeout-seconds"`
	InputCostMicrosPerMillion  int64  `mapstructure:"input-cost-micros-per-million" json:"input-cost-micros-per-million" yaml:"input-cost-micros-per-million"`
	OutputCostMicrosPerMillion int64  `mapstructure:"output-cost-micros-per-million" json:"output-cost-micros-per-million" yaml:"output-cost-micros-per-million"`
	APIKeyInput                string `mapstructure:"-" json:"api-key,omitempty" yaml:"-"`
	APIKeyConfigured           bool   `mapstructure:"-" json:"api-key-configured" yaml:"-"`
	ClearAPIKey                bool   `mapstructure:"-" json:"clear-api-key,omitempty" yaml:"-"`
}

func (c *AI) Normalize() {
	c.SensitiveWords = normalizeAIList(c.SensitiveWords)
	c.AllowVisionModules = normalizeAIList(c.AllowVisionModules)
	c.OpenAICompatible.normalize()
	c.Anthropic.normalize()
}

func (p *AIProvider) normalize() {
	p.BaseURL = strings.TrimSpace(p.BaseURL)
	p.Model = strings.TrimSpace(p.Model)
	if p.TimeoutSeconds <= 0 {
		p.TimeoutSeconds = defaultAIProviderTimeoutSeconds
	}
}

func (c AI) Validate() error {
	for _, word := range c.SensitiveWords {
		if len([]rune(word)) > 100 {
			return errors.New("AI 业务敏感词不能超过 100 个字符")
		}
	}
	for _, module := range c.AllowVisionModules {
		if len(module) > 80 {
			return errors.New("AI 图片发送模块标识不能超过 80 个字符")
		}
	}
	if err := c.OpenAICompatible.validate("OpenAI Compatible"); err != nil {
		return err
	}
	return c.Anthropic.validate("Anthropic")
}

func normalizeAIList(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func (p AIProvider) validate(providerName string) error {
	if len(p.BaseURL) > 2048 {
		return errors.New(providerName + " Base URL 长度超出限制")
	}
	if len(p.Model) > 200 {
		return errors.New(providerName + " 模型名称不能超过 200 个字符")
	}
	if len(p.APIKey) > 8192 {
		return errors.New(providerName + " API Key 长度超出限制")
	}
	if p.TimeoutSeconds < 1 || p.TimeoutSeconds > 120 {
		return errors.New(providerName + " 超时时间必须在 1 到 120 秒之间")
	}
	if p.InputCostMicrosPerMillion < 0 || p.OutputCostMicrosPerMillion < 0 {
		return errors.New(providerName + " 费用不能为负数")
	}
	if !p.Enabled {
		return nil
	}
	if err := validateAIProviderURL(p.BaseURL); err != nil {
		return errors.New(providerName + " Base URL 不正确")
	}
	if p.Model == "" {
		return errors.New("请配置 " + providerName + " 模型名称")
	}
	return nil
}

func (c AI) Redacted() AI {
	c.OpenAICompatible = c.OpenAICompatible.redacted()
	c.Anthropic = c.Anthropic.redacted()
	return c
}

func (p AIProvider) redacted() AIProvider {
	p.APIKeyConfigured = strings.TrimSpace(p.APIKey) != ""
	p.APIKeyInput = ""
	p.ClearAPIKey = false
	return p
}

// MergeSecrets applies write-only keys while preserving keys that are omitted
// by browser clients.
func (c AI) MergeSecrets(current AI, allow bool) AI {
	if !allow {
		return current
	}
	c.OpenAICompatible = c.OpenAICompatible.mergeSecrets(current.OpenAICompatible)
	c.Anthropic = c.Anthropic.mergeSecrets(current.Anthropic)
	return c
}

func (p AIProvider) mergeSecrets(current AIProvider) AIProvider {
	if p.ClearAPIKey {
		p.APIKey = ""
	} else if strings.TrimSpace(p.APIKeyInput) != "" {
		p.APIKey = strings.TrimSpace(p.APIKeyInput)
	} else if strings.TrimSpace(p.APIKey) == "" {
		p.APIKey = current.APIKey
	}
	p.APIKeyInput = ""
	p.APIKeyConfigured = false
	p.ClearAPIKey = false
	return p
}

func validateAIProviderURL(value string) error {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" || parsed.RawQuery != "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("invalid provider URL")
	}
	return nil
}
