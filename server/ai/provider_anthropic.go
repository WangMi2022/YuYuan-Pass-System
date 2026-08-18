package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/goccy/go-json"
)

type anthropicProvider struct {
	configuration config.AIProvider
	allowPrivate  bool
}

func (p anthropicProvider) Name() string  { return config.AIProviderAnthropic }
func (p anthropicProvider) Model() string { return p.configuration.Model }

func (p anthropicProvider) Complete(ctx context.Context, call providerCall) (providerResult, error) {
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
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int64 `json:"input_tokens"`
			OutputTokens int64 `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return providerResult{}, &Error{Type: ErrorTypeProvider, Message: "解析模型响应失败", Cause: err}
	}
	var builder strings.Builder
	for _, part := range payload.Content {
		if part.Type == "text" {
			builder.WriteString(part.Text)
		}
	}
	content := strings.TrimSpace(builder.String())
	if content == "" {
		return providerResult{}, &Error{Type: ErrorTypeProvider, Message: "模型返回内容为空"}
	}
	return providerResult{Content: content, Data: content, InputTokens: payload.Usage.InputTokens, OutputTokens: payload.Usage.OutputTokens}, nil
}

func (p anthropicProvider) Vision(ctx context.Context, call providerCall) (providerResult, error) {
	return p.Complete(ctx, call)
}

func (p anthropicProvider) Stream(ctx context.Context, call providerCall) (*http.Response, error) {
	return p.do(ctx, call, true)
}

func (p anthropicProvider) do(ctx context.Context, call providerCall, streaming bool) (*http.Response, error) {
	messageText := providerMessageText(call)
	messages := any([]map[string]string{{"role": "user", "content": messageText}})
	if len(call.Image) > 0 {
		messages = []map[string]any{{
			"role": "user",
			"content": []map[string]any{
				{"type": "image", "source": map[string]any{
					"type": "base64", "media_type": normalizedImageMIME(call.MIMEType), "data": base64.StdEncoding.EncodeToString(call.Image),
				}},
				{"type": "text", "text": messageText},
			},
		}}
	}
	requestBody := map[string]any{
		"model":      p.configuration.Model,
		"max_tokens": call.MaxOutputTokens,
		"messages":   messages,
	}
	if call.MaxOutputTokens <= 0 {
		requestBody["max_tokens"] = 1024
	}
	if streaming {
		requestBody["stream"] = true
	}
	encoded, err := json.Marshal(requestBody)
	if err != nil {
		return nil, &Error{Type: ErrorTypeValidation, Message: "模型请求序列化失败", Cause: err}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicMessagesURL(p.configuration.BaseURL), bytes.NewReader(encoded))
	if err != nil {
		return nil, &Error{Type: ErrorTypeValidation, Message: "模型服务地址不正确", Cause: err}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	if streaming {
		req.Header.Set("Accept", "text/event-stream")
	}
	if key := strings.TrimSpace(p.configuration.APIKey); key != "" {
		req.Header.Set("x-api-key", key)
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

func anthropicMessagesURL(baseURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return strings.TrimSpace(baseURL)
	}
	path := strings.TrimRight(parsed.Path, "/")
	if !strings.HasSuffix(path, "/messages") {
		if strings.HasSuffix(path, "/v1") {
			path += "/messages"
		} else {
			path += "/v1/messages"
		}
	}
	parsed.Path = path
	parsed.RawPath = ""
	parsed.Fragment = ""
	return parsed.String()
}
