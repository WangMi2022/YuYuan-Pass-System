package ai

import (
	"context"
	"net/http"
	"strings"
)

type providerCall struct {
	Prompt          string
	Payload         map[string]any
	MaxOutputTokens int
	Image           []byte
	MIMEType        string
}

type providerResult struct {
	Data         any
	Content      string
	InputTokens  int64
	OutputTokens int64
}

type provider interface {
	Name() string
	Model() string
	Complete(context.Context, providerCall) (providerResult, error)
	Vision(context.Context, providerCall) (providerResult, error)
	Stream(context.Context, providerCall) (*http.Response, error)
}

func providerMessageText(call providerCall) string {
	prompt := strings.TrimSpace(call.Prompt)
	payload := strings.TrimSpace(DecodePayload(call.Payload))
	if payload == "" || payload == "{}" || payload == "null" {
		return prompt
	}
	if prompt == "" {
		return "以下是已脱敏的业务上下文 JSON：\n" + payload
	}
	return prompt + "\n\n以下是已脱敏的业务上下文 JSON，请只基于它回答：\n" + payload
}
