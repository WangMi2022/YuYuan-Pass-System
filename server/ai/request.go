package ai

import (
	"context"
	"net/http"

	"github.com/WangMi2022/mit-assets-admin/server/model/common"
)

type actorContextKey struct{}

type Actor struct {
	UserID           uint
	AuthorityID      uint
	PermissionPath   string
	PermissionMethod string
	TrustedInternal  bool
}

func WithActor(ctx context.Context, userID, authorityID uint) context.Context {
	return context.WithValue(ctx, actorContextKey{}, Actor{UserID: userID, AuthorityID: authorityID})
}

func WithActorPermission(ctx context.Context, userID, authorityID uint, path, method string) context.Context {
	return context.WithValue(ctx, actorContextKey{}, Actor{
		UserID: userID, AuthorityID: authorityID, PermissionPath: path, PermissionMethod: method,
	})
}

func WithInternalActor(ctx context.Context) context.Context {
	return context.WithValue(ctx, actorContextKey{}, Actor{TrustedInternal: true})
}

func actorFromContext(ctx context.Context) Actor {
	actor, _ := ctx.Value(actorContextKey{}).(Actor)
	return actor
}

// CompletionRequest is intentionally business-oriented. Callers need not know
// a provider protocol, credential, or provider endpoint.
type CompletionRequest struct {
	UserID           uint
	AuthorityID      uint
	Module           string
	Operation        string
	Provider         string
	Model            string
	PromptKey        string
	PromptVersion    int
	ObjectType       string
	ObjectID         string
	Prompt           string
	Payload          common.JSONMap
	MaxOutputTokens  int
	PermissionPath   string
	PermissionMethod string
	OutputSchema     string
	trustedInternal  bool
}

type VisionRequest struct {
	CompletionRequest
	Image    []byte
	MIMEType string
}

type VisionResult = CompletionResult

type CompletionResult struct {
	Data         any
	Content      string
	Provider     string
	Model        string
	InputTokens  int64
	OutputTokens int64
	DurationMS   int64
}

type StreamResult struct {
	Response *http.Response
	Provider string
	Model    string
}

// Gateway is the only interface required by AI-enabled business services.
type Gateway interface {
	Complete(ctx context.Context, request CompletionRequest) (CompletionResult, error)
	Vision(ctx context.Context, request VisionRequest) (VisionResult, error)
	Stream(ctx context.Context, request CompletionRequest) (StreamResult, error)
}
