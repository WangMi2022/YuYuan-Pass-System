package ai

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	maxCompletionInputBytes  = 2 << 20
	maxCompletionOutputBytes = 4 << 20
	maxVisionInputBytes      = 10 << 20
)

var Default Gateway = gateway{}

type gateway struct{}

func (gateway) Complete(ctx context.Context, request CompletionRequest) (CompletionResult, error) {
	state, prepared, err := beginInvocation(ctx, request, nil)
	if err != nil {
		return CompletionResult{}, err
	}
	defer state.finish()

	result, err := state.provider.Complete(ctx, providerCall{
		Prompt: prepared.Prompt, Payload: prepared.Payload, MaxOutputTokens: prepared.MaxOutputTokens,
	})
	if err != nil {
		state.fail(err)
		return CompletionResult{}, err
	}
	if err := validateOutput(prepared.OutputSchema, result.Content); err != nil {
		state.fail(err)
		return CompletionResult{}, err
	}
	state.success(result.InputTokens, result.OutputTokens, resultHash(result))
	return completionResult(state, result), nil
}

func (gateway) Vision(ctx context.Context, request VisionRequest) (VisionResult, error) {
	if len(request.Image) == 0 || len(request.Image) > maxVisionInputBytes {
		err := &Error{Type: ErrorTypeValidation, Message: "AI 图片必须在 1 字节到 10MB 之间"}
		auditRejectedRequest(ctx, request.CompletionRequest, err)
		return VisionResult{}, err
	}
	if !supportedImageMIME(request.MIMEType) {
		err := &Error{Type: ErrorTypeValidation, Message: "AI 图片仅支持 JPEG、PNG、WebP 或 GIF"}
		auditRejectedRequest(ctx, request.CompletionRequest, err)
		return VisionResult{}, err
	}
	request.MIMEType = normalizedImageMIME(request.MIMEType)
	if !visionModuleAllowed(request.Module) {
		err := &Error{Type: ErrorTypePolicy, Message: "当前业务模块不允许向第三方模型发送图片"}
		auditRejectedRequest(ctx, request.CompletionRequest, err)
		return VisionResult{}, err
	}
	state, prepared, err := beginInvocation(ctx, request.CompletionRequest, request.Image)
	if err != nil {
		return VisionResult{}, err
	}
	defer state.finish()

	result, err := state.provider.Vision(ctx, providerCall{
		Prompt: prepared.Prompt, Payload: prepared.Payload, MaxOutputTokens: prepared.MaxOutputTokens,
		Image: request.Image, MIMEType: request.MIMEType,
	})
	if err != nil {
		state.fail(err)
		return VisionResult{}, err
	}
	if err := validateOutput(prepared.OutputSchema, result.Content); err != nil {
		state.fail(err)
		return VisionResult{}, err
	}
	state.success(result.InputTokens, result.OutputTokens, resultHash(result))
	return completionResult(state, result), nil
}

func (gateway) Stream(ctx context.Context, request CompletionRequest) (StreamResult, error) {
	state, prepared, err := beginInvocation(ctx, request, nil)
	if err != nil {
		return StreamResult{}, err
	}
	response, err := state.provider.Stream(ctx, providerCall{
		Prompt: prepared.Prompt, Payload: prepared.Payload, MaxOutputTokens: prepared.MaxOutputTokens,
	})
	if err != nil {
		state.fail(err)
		state.finish()
		return StreamResult{}, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_ = response.Body.Close()
		err = &Error{Type: ErrorTypeProvider, Message: fmt.Sprintf("模型流式服务返回非 2xx 状态：%d", response.StatusCode)}
		state.fail(err)
		state.finish()
		return StreamResult{}, err
	}
	response.Body = &streamAuditBody{
		ReadCloser: response.Body, state: state, provider: state.provider.Name(),
		inputEstimate: estimateRequestTokens(prepared), hasher: sha256.New(),
	}
	return StreamResult{Response: response, Provider: state.provider.Name(), Model: state.provider.Model()}, nil
}

type invocationState struct {
	invocation ModelInvocation
	provider   provider
	startedAt  time.Time
	quota      quotaLease
	finishOnce sync.Once
}

func beginInvocation(ctx context.Context, request CompletionRequest, image []byte) (*invocationState, CompletionRequest, error) {
	request = applyActor(ctx, request)
	state := &invocationState{
		startedAt: time.Now(),
		invocation: ModelInvocation{
			RequestID: uuid.NewString(), UserID: request.UserID, AuthorityID: request.AuthorityID,
			Module: strings.TrimSpace(request.Module), Operation: strings.TrimSpace(request.Operation),
			Provider: strings.TrimSpace(request.Provider), Model: strings.TrimSpace(request.Model),
			PromptKey: strings.TrimSpace(request.PromptKey), PromptVersion: request.PromptVersion,
			ObjectType: strings.TrimSpace(request.ObjectType), ObjectID: strings.TrimSpace(request.ObjectID),
			Status: InvocationStatusFailed,
		},
	}
	if err := authorizeRequest(request); err != nil {
		return rejectedInvocation(state, request, err)
	}
	prepared, redactionCount, inputHash, err := prepareRequest(ctx, request, image)
	if err != nil {
		return rejectedInvocation(state, request, err)
	}
	state.invocation.PromptKey = prepared.PromptKey
	state.invocation.PromptVersion = prepared.PromptVersion
	state.invocation.RedactionCount = redactionCount
	state.invocation.InputHash = inputHash
	selected, err := selectProvider(prepared)
	if err != nil {
		return rejectedInvocation(state, prepared, err)
	}
	state.provider = selected
	state.invocation.Provider = selected.Name()
	state.invocation.Model = selected.Model()
	inputEstimate := estimateRequestTokens(prepared)
	outputEstimate := int64(prepared.MaxOutputTokens)
	if outputEstimate <= 0 {
		outputEstimate = 1024
	}
	lease, err := acquireQuota(ctx, prepared, quotaEstimate{
		Requests: 1, Tokens: inputEstimate + outputEstimate,
		CostMicros: estimateCost(selected, inputEstimate, outputEstimate),
	})
	if err != nil {
		return rejectedInvocation(state, prepared, err)
	}
	state.quota = lease
	return state, prepared, nil
}

func rejectedInvocation(state *invocationState, request CompletionRequest, err error) (*invocationState, CompletionRequest, error) {
	state.recordError(err)
	state.finish()
	return nil, CompletionRequest{}, err
}

func auditRejectedRequest(ctx context.Context, request CompletionRequest, err error) {
	request = applyActor(ctx, request)
	state := &invocationState{
		startedAt: time.Now(),
		invocation: ModelInvocation{
			RequestID: uuid.NewString(), UserID: request.UserID, AuthorityID: request.AuthorityID,
			Module: strings.TrimSpace(request.Module), Operation: strings.TrimSpace(request.Operation),
			Provider: strings.TrimSpace(request.Provider), Model: strings.TrimSpace(request.Model),
			Status: InvocationStatusBlocked,
		},
	}
	state.recordError(err)
	state.finish()
}

func applyActor(ctx context.Context, request CompletionRequest) CompletionRequest {
	actor := actorFromContext(ctx)
	request.trustedInternal = actor.TrustedInternal
	if actor.UserID != 0 {
		request.UserID = actor.UserID
		request.AuthorityID = actor.AuthorityID
		request.PermissionPath = actor.PermissionPath
		request.PermissionMethod = actor.PermissionMethod
		request.trustedInternal = false
	}
	return request
}

func authorizeRequest(request CompletionRequest) error {
	if request.UserID == 0 {
		if request.trustedInternal {
			return nil
		}
		return &Error{Type: ErrorTypePolicy, Message: "AI 调用缺少认证身份或可信内部任务上下文"}
	}
	path := strings.TrimSpace(request.PermissionPath)
	method := strings.ToUpper(strings.TrimSpace(request.PermissionMethod))
	if path == "" || method == "" {
		return &Error{Type: ErrorTypePolicy, Message: "AI 调用缺少可审计的权限上下文"}
	}
	enforcer := utils.GetCasbin()
	if enforcer == nil {
		return &Error{Type: ErrorTypePolicy, Message: "AI 权限校验服务不可用"}
	}
	allowed, err := enforcer.Enforce(fmt.Sprintf("%d", request.AuthorityID), path, method)
	if err != nil {
		return &Error{Type: ErrorTypePolicy, Message: "AI 权限校验失败", Cause: err}
	}
	if !allowed {
		return &Error{Type: ErrorTypePolicy, Message: "当前角色无权执行该 AI 操作"}
	}
	return nil
}

func prepareRequest(ctx context.Context, request CompletionRequest, image []byte) (CompletionRequest, int, string, error) {
	request.Module = strings.TrimSpace(request.Module)
	request.Operation = strings.TrimSpace(request.Operation)
	if request.Module == "" || request.Operation == "" {
		return CompletionRequest{}, 0, "", &Error{Type: ErrorTypeValidation, Message: "AI 调用缺少模块或操作标识"}
	}
	if len(request.Module) > 80 || len(request.Operation) > 100 || len(request.PromptKey) > 120 || len(request.ObjectID) > 120 {
		return CompletionRequest{}, 0, "", &Error{Type: ErrorTypeValidation, Message: "AI 调用标识长度超出限制"}
	}
	resolved, err := resolvePrompt(ctx, request)
	if err != nil {
		return CompletionRequest{}, 0, "", err
	}
	payloadJSON := DecodePayload(resolved.Payload)
	if len(resolved.Prompt)+len(payloadJSON) > maxCompletionInputBytes {
		return CompletionRequest{}, 0, "", &Error{Type: ErrorTypeValidation, Message: "AI 文本输入超过 2MB 限制"}
	}
	hashMaterial := resolved.Prompt + payloadJSON
	if len(image) > 0 {
		imageHash := sha256.Sum256(image)
		hashMaterial += fmt.Sprintf("%x", imageHash[:])
	}
	inputHash := Hash(hashMaterial)
	configuration := global.GVA_CONFIG.AI
	redactedPrompt, redactionCount := RedactWithWords(resolved.Prompt, configuration.SensitiveWords)
	resolved.Prompt = redactedPrompt
	redactedPayload, payloadRedactionCount := RedactPayload(resolved.Payload, configuration.SensitiveWords)
	resolved.Payload = redactedPayload
	redactionCount += payloadRedactionCount
	return resolved, redactionCount, inputHash, nil
}

func selectProvider(request CompletionRequest) (provider, error) {
	configuration := global.GVA_CONFIG.AI
	configuration.Normalize()
	if request.Provider == "autocode-compat" {
		path := strings.TrimSpace(global.GVA_CONFIG.AutoCode.AiPath)
		if path == "" {
			return nil, &Error{Type: ErrorTypeDisabled, Message: "请先在 config.yaml 的 autocode.ai-path 中配置自有 AI 服务地址"}
		}
		return autoCodeProvider{path: path, allowPrivateEndpoints: true}, nil
	}
	if !configuration.Enabled {
		return nil, &Error{Type: ErrorTypeDisabled, Message: "统一 AI Gateway 尚未启用"}
	}
	providerName := strings.TrimSpace(request.Provider)
	if providerName == "" {
		if configuration.OpenAICompatible.Enabled {
			providerName = config.AIProviderOpenAICompatible
		} else if configuration.Anthropic.Enabled {
			providerName = config.AIProviderAnthropic
		}
	}
	switch providerName {
	case config.AIProviderOpenAICompatible:
		if !configuration.OpenAICompatible.Enabled {
			return nil, &Error{Type: ErrorTypeDisabled, Message: "OpenAI Compatible Provider 未启用"}
		}
		return openAICompatibleProvider{configuration: configuration.OpenAICompatible, allowPrivate: configuration.AllowPrivateEndpoints}, nil
	case config.AIProviderAnthropic:
		if !configuration.Anthropic.Enabled {
			return nil, &Error{Type: ErrorTypeDisabled, Message: "Anthropic Provider 未启用"}
		}
		return anthropicProvider{configuration: configuration.Anthropic, allowPrivate: configuration.AllowPrivateEndpoints}, nil
	default:
		return nil, &Error{Type: ErrorTypeValidation, Message: "不支持的 AI Provider：" + providerName}
	}
}

func visionModuleAllowed(module string) bool {
	module = strings.TrimSpace(module)
	for _, allowed := range global.GVA_CONFIG.AI.AllowVisionModules {
		if strings.TrimSpace(allowed) == module {
			return true
		}
	}
	return false
}

func completionResult(state *invocationState, result providerResult) CompletionResult {
	return CompletionResult{
		Data: result.Data, Content: result.Content, Provider: state.provider.Name(), Model: state.provider.Model(),
		InputTokens: result.InputTokens, OutputTokens: result.OutputTokens, DurationMS: state.durationMS(),
	}
}

func resultHash(result providerResult) string {
	if result.Content != "" {
		return Hash(result.Content)
	}
	encoded, err := json.Marshal(result.Data)
	if err != nil {
		return ""
	}
	return Hash(string(encoded))
}

func estimateRequestTokens(request CompletionRequest) int64 {
	return estimateTokens(int64(len(request.Prompt) + len(DecodePayload(request.Payload))))
}

func estimateCost(selected provider, inputTokens, outputTokens int64) int64 {
	configuration := global.GVA_CONFIG.AI
	var costs config.AIProvider
	switch selected.Name() {
	case config.AIProviderOpenAICompatible:
		costs = configuration.OpenAICompatible
	case config.AIProviderAnthropic:
		costs = configuration.Anthropic
	default:
		return 0
	}
	return inputTokens*costs.InputCostMicrosPerMillion/1_000_000 + outputTokens*costs.OutputCostMicrosPerMillion/1_000_000
}

func (s *invocationState) success(inputTokens, outputTokens int64, outputHash string) {
	s.invocation.Status = InvocationStatusSuccess
	s.invocation.ErrorType = ""
	s.invocation.InputTokens = inputTokens
	s.invocation.OutputTokens = outputTokens
	if s.provider != nil {
		s.invocation.EstimatedCostMicros = estimateCost(s.provider, inputTokens, outputTokens)
	}
	s.invocation.OutputHash = outputHash
}

func (s *invocationState) fail(err error) { s.recordError(err) }

func (s *invocationState) recordError(err error) {
	kind := ErrorKind(err)
	s.invocation.Status = InvocationStatusFailed
	if kind == ErrorTypeDisabled || kind == ErrorTypeValidation || kind == ErrorTypePolicy || kind == ErrorTypeQuota {
		s.invocation.Status = InvocationStatusBlocked
	}
	s.invocation.ErrorType = string(kind)
}

func (s *invocationState) finish() {
	s.finishOnce.Do(func() {
		s.invocation.DurationMS = s.durationMS()
		quotaState.Lock()
		defer quotaState.Unlock()
		persistInvocation(s.invocation)
		s.quota.releaseLocked()
	})
}

func (s *invocationState) durationMS() int64 { return time.Since(s.startedAt).Milliseconds() }

func persistInvocation(invocation ModelInvocation) {
	if global.GVA_DB == nil {
		return
	}
	if err := global.GVA_DB.Create(&invocation).Error; err != nil && global.GVA_LOG != nil {
		global.GVA_LOG.Error("写入 AI 模型调用审计失败", zap.String("requestId", invocation.RequestID), zap.Error(err))
	}
}

func DecodePayload(payload map[string]any) string {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(encoded)
}
