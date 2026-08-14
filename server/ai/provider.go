package ai

import (
	"context"
	"net/http"
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
