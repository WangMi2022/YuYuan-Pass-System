package ai

import (
	"context"
	"testing"
)

func TestPrepareRequestRedactsPrompt(t *testing.T) {
	request, count, hash, err := prepareRequest(context.Background(), CompletionRequest{
		Module: "asset", Operation: "draft", Prompt: "请联系 13800138000，token=secret-value",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if count < 2 || hash == "" || contains(request.Prompt, "13800138000") || contains(request.Prompt, "secret-value") {
		t.Fatalf("unexpected redaction result: count=%d hash=%q prompt=%q", count, hash, request.Prompt)
	}
}

func TestValidateAutoCodeMode(t *testing.T) {
	if err := ValidateAutoCodeMode("analysisChat"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAutoCodeMode("../admin"); err == nil {
		t.Fatal("unsafe mode should be rejected")
	}
}

func TestAuthorizeRequestRequiresExplicitInternalActor(t *testing.T) {
	if ErrorKind(authorizeRequest(CompletionRequest{})) != ErrorTypePolicy {
		t.Fatal("missing actor must be rejected")
	}
	request := applyActor(WithInternalActor(context.Background()), CompletionRequest{})
	if err := authorizeRequest(request); err != nil {
		t.Fatal(err)
	}
}

func TestVisionRejectsUnsupportedMIME(t *testing.T) {
	_, err := Default.Vision(context.Background(), VisionRequest{
		CompletionRequest: CompletionRequest{Module: "asset", Operation: "draft"},
		Image:             []byte("not-an-image"),
		MIMEType:          "application/pdf",
	})
	if ErrorKind(err) != ErrorTypeValidation {
		t.Fatalf("error kind = %q, want %q", ErrorKind(err), ErrorTypeValidation)
	}
}
