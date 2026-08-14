package ai

import (
	"crypto/sha256"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
)

func TestParseStreamUsage(t *testing.T) {
	openAI := "data: {\"usage\":{\"prompt_tokens\":12,\"completion_tokens\":7}}\n\ndata: [DONE]\n"
	input, output := parseStreamUsage(config.AIProviderOpenAICompatible, []byte(openAI))
	if input != 12 || output != 7 {
		t.Fatalf("OpenAI usage = (%d, %d)", input, output)
	}
	anthropic := "data: {\"message\":{\"usage\":{\"input_tokens\":9,\"output_tokens\":1}}}\n\ndata: {\"usage\":{\"output_tokens\":6}}\n"
	input, output = parseStreamUsage(config.AIProviderAnthropic, []byte(anthropic))
	if input != 9 || output != 6 {
		t.Fatalf("Anthropic usage = (%d, %d)", input, output)
	}
}

func TestStreamAuditBodyFallsBackToEstimates(t *testing.T) {
	state := &invocationState{invocation: ModelInvocation{}, startedAt: time.Now()}
	body := &streamAuditBody{
		ReadCloser: io.NopCloser(strings.NewReader("data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\n")),
		state:      state, provider: config.AIProviderOpenAICompatible, inputEstimate: 11, hasher: sha256.New(),
	}
	if _, err := io.ReadAll(body); err != nil {
		t.Fatal(err)
	}
	if state.invocation.Status != InvocationStatusSuccess || state.invocation.InputTokens != 11 || state.invocation.OutputTokens == 0 || state.invocation.OutputHash == "" {
		t.Fatalf("unexpected invocation: %#v", state.invocation)
	}
	if err := body.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestStreamAuditBodyMarksEarlyCloseAsFailed(t *testing.T) {
	state := &invocationState{invocation: ModelInvocation{}, startedAt: time.Now()}
	body := &streamAuditBody{
		ReadCloser: io.NopCloser(strings.NewReader("data: unfinished")),
		state:      state,
		hasher:     sha256.New(),
	}
	if err := body.Close(); err != nil {
		t.Fatal(err)
	}
	if state.invocation.Status != InvocationStatusFailed || state.invocation.ErrorType != string(ErrorTypeProvider) {
		t.Fatalf("unexpected invocation: %#v", state.invocation)
	}
}

func TestStreamAuditBodyStopsAtLimit(t *testing.T) {
	state := &invocationState{invocation: ModelInvocation{}, startedAt: time.Now()}
	body := &streamAuditBody{
		ReadCloser: io.NopCloser(strings.NewReader(strings.Repeat("x", maxStreamOutputBytes+1))),
		state:      state,
		hasher:     sha256.New(),
	}
	written, err := io.Copy(io.Discard, body)
	if ErrorKind(err) != ErrorTypeSchema {
		t.Fatalf("error kind = %q, want %q", ErrorKind(err), ErrorTypeSchema)
	}
	if written != maxStreamOutputBytes || state.invocation.Status != InvocationStatusFailed {
		t.Fatalf("written = %d, invocation = %#v", written, state.invocation)
	}
}
