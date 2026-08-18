package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/goccy/go-json"
)

const maxProviderResponseSize = 4 << 20

type openAICompatibleProvider struct {
	configuration config.AIProvider
	allowPrivate  bool
}

func (p openAICompatibleProvider) Name() string  { return config.AIProviderOpenAICompatible }
func (p openAICompatibleProvider) Model() string { return p.configuration.Model }

func (p openAICompatibleProvider) Complete(ctx context.Context, call providerCall) (providerResult, error) {
	response, err := p.do(ctx, call, false)
	if err != nil {
		return providerResult{}, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, maxProviderResponseSize+1))
	if err != nil {
		return providerResult{}, &Error{Type: ErrorTypeProvider, Message: "读取模型响应失败", Cause: err}
	}
	if len(body) > maxProviderResponseSize {
		return providerResult{}, &Error{Type: ErrorTypeProvider, Message: "模型响应超过 4MB 限制"}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return providerResult{}, providerStatusError(response.StatusCode)
	}
	var payload struct {
		Choices []struct {
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int64 `json:"prompt_tokens"`
			CompletionTokens int64 `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return providerResult{}, &Error{Type: ErrorTypeProvider, Message: "解析模型响应失败", Cause: err}
	}
	if len(payload.Choices) == 0 {
		return providerResult{}, &Error{Type: ErrorTypeProvider, Message: "模型未返回结果"}
	}
	content, err := openAIContent(payload.Choices[0].Message.Content)
	if err != nil {
		return providerResult{}, err
	}
	return providerResult{Content: content, Data: content, InputTokens: payload.Usage.PromptTokens, OutputTokens: payload.Usage.CompletionTokens}, nil
}

func (p openAICompatibleProvider) Vision(ctx context.Context, call providerCall) (providerResult, error) {
	return p.Complete(ctx, call)
}

func (p openAICompatibleProvider) Stream(ctx context.Context, call providerCall) (*http.Response, error) {
	return p.do(ctx, call, true)
}

func (p openAICompatibleProvider) do(ctx context.Context, call providerCall, streaming bool) (*http.Response, error) {
	messageText := providerMessageText(call)
	messages := any([]map[string]string{{"role": "user", "content": messageText}})
	if len(call.Image) > 0 {
		mimeType := normalizedImageMIME(call.MIMEType)
		messages = []map[string]any{{
			"role": "user",
			"content": []map[string]any{
				{"type": "text", "text": messageText},
				{"type": "image_url", "image_url": map[string]any{
					"url": "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(call.Image), "detail": "high",
				}},
			},
		}}
	}
	requestBody := map[string]any{
		"model":       p.configuration.Model,
		"messages":    messages,
		"temperature": 0,
	}
	if call.MaxOutputTokens > 0 {
		requestBody["max_tokens"] = call.MaxOutputTokens
	}
	if streaming {
		requestBody["stream"] = true
	}
	encoded, err := json.Marshal(requestBody)
	if err != nil {
		return nil, &Error{Type: ErrorTypeValidation, Message: "模型请求序列化失败", Cause: err}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIChatURL(p.configuration.BaseURL), bytes.NewReader(encoded))
	if err != nil {
		return nil, &Error{Type: ErrorTypeValidation, Message: "模型服务地址不正确", Cause: err}
	}
	req.Header.Set("Content-Type", "application/json")
	if streaming {
		req.Header.Set("Accept", "text/event-stream")
	}
	if key := strings.TrimSpace(p.configuration.APIKey); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	client := providerHTTPClient(time.Duration(p.configuration.TimeoutSeconds)*time.Second, p.allowPrivate)
	if streaming {
		client.Timeout = 0
	}
	response, err := client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, &Error{Type: ErrorTypeTimeout, Message: "模型请求超时", Cause: err}
		}
		return nil, &Error{Type: ErrorTypeProvider, Message: "调用模型失败", Cause: err}
	}
	return response, nil
}

func normalizedImageMIME(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return value
	default:
		return "image/png"
	}
}

func supportedImageMIME(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "image/jpeg" || value == "image/png" || value == "image/webp" || value == "image/gif"
}

func openAIChatURL(baseURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return strings.TrimSpace(baseURL)
	}
	path := strings.TrimRight(parsed.Path, "/")
	if !strings.HasSuffix(path, "/chat/completions") {
		if strings.HasSuffix(path, "/v1") {
			path += "/chat/completions"
		} else {
			path += "/v1/chat/completions"
		}
	}
	parsed.Path = path
	parsed.RawPath = ""
	parsed.Fragment = ""
	return parsed.String()
}

func openAIContent(raw json.RawMessage) (string, error) {
	var content string
	if err := json.Unmarshal(raw, &content); err == nil && strings.TrimSpace(content) != "" {
		return content, nil
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err != nil {
		return "", &Error{Type: ErrorTypeProvider, Message: "模型返回内容格式不兼容"}
	}
	var builder strings.Builder
	for _, part := range parts {
		if part.Type == "text" || part.Type == "output_text" {
			builder.WriteString(part.Text)
		}
	}
	if strings.TrimSpace(builder.String()) == "" {
		return "", &Error{Type: ErrorTypeProvider, Message: "模型返回内容为空"}
	}
	return builder.String(), nil
}

func providerStatusError(statusCode int) error {
	message := fmt.Sprintf("模型服务返回状态 %d", statusCode)
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		message = fmt.Sprintf("模型服务鉴权失败（HTTP %d）", statusCode)
	}
	return &Error{Type: ErrorTypeProvider, Message: message}
}
