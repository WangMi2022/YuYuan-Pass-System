package ai

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common"
	commonResponse "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/goccy/go-json"
)

type autoCodeProvider struct {
	path                  string
	allowPrivateEndpoints bool
}

func (p autoCodeProvider) Name() string  { return "autocode-compat" }
func (p autoCodeProvider) Model() string { return "legacy-autocode" }

func (p autoCodeProvider) Complete(ctx context.Context, call providerCall) (providerResult, error) {
	response, err := p.do(ctx, call, false)
	if err != nil {
		return providerResult{}, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 4<<20+1))
	if err != nil {
		return providerResult{}, &Error{Type: ErrorTypeProvider, Message: "读取代码生成模型响应失败", Cause: err}
	}
	if len(body) > 4<<20 {
		return providerResult{}, &Error{Type: ErrorTypeProvider, Message: "代码生成模型响应超过 4MB 限制"}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return providerResult{}, &Error{Type: ErrorTypeProvider, Message: fmt.Sprintf("代码生成模型返回非 2xx 状态：%d", response.StatusCode)}
	}
	var envelope commonResponse.Response
	if err := json.Unmarshal(body, &envelope); err != nil {
		return providerResult{}, &Error{Type: ErrorTypeProvider, Message: "解析代码生成模型响应失败", Cause: err}
	}
	if envelope.Code != commonResponse.SUCCESS {
		return providerResult{}, &Error{Type: ErrorTypeProvider, Message: "代码生成模型返回业务错误：" + strings.TrimSpace(envelope.Msg)}
	}
	return providerResult{Data: envelope.Data}, nil
}

func (p autoCodeProvider) Vision(context.Context, providerCall) (providerResult, error) {
	return providerResult{}, &Error{Type: ErrorTypeValidation, Message: "自动代码兼容 Provider 不支持图片调用"}
}

func (p autoCodeProvider) Stream(ctx context.Context, call providerCall) (*http.Response, error) {
	return p.do(ctx, call, true)
}

func (p autoCodeProvider) do(ctx context.Context, call providerCall, streaming bool) (*http.Response, error) {
	mode := strings.TrimSpace(fmt.Sprintf("%v", call.Payload["mode"]))
	if err := ValidateAutoCodeMode(mode); err != nil {
		return nil, err
	}
	endpoint := strings.ReplaceAll(p.path, "{FUNC}", mode)
	payload := make(common.JSONMap, len(call.Payload)+1)
	for key, value := range call.Payload {
		payload[key] = value
	}
	if streaming && strings.TrimSpace(fmt.Sprintf("%v", payload["response_mode"])) == "" {
		payload["response_mode"] = "streaming"
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &Error{Type: ErrorTypeValidation, Message: "代码生成模型请求序列化失败", Cause: err}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, &Error{Type: ErrorTypeValidation, Message: "代码生成模型地址不正确", Cause: err}
	}
	req.Header.Set("Content-Type", "application/json")
	if streaming {
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Accept-Encoding", "identity")
		req.Header.Set("Cache-Control", "no-cache")
	}
	client := providerHTTPClient(0, p.allowPrivateEndpoints)
	if streaming {
		client.Timeout = 0
	}
	response, err := client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, &Error{Type: ErrorTypeTimeout, Message: "代码生成模型请求超时", Cause: err}
		}
		return nil, &Error{Type: ErrorTypeProvider, Message: "调用代码生成模型失败", Cause: err}
	}
	return response, nil
}

func ValidateAutoCodeMode(mode string) error {
	if mode == "" {
		return &Error{Type: ErrorTypeValidation, Message: "llmAuto 缺少 mode 参数"}
	}
	if len(mode) > 64 {
		return &Error{Type: ErrorTypeValidation, Message: "llmAuto mode 参数过长"}
	}
	if _, ok := allowedAutoCodeModes[mode]; !ok {
		return &Error{Type: ErrorTypeValidation, Message: "llmAuto mode 不受支持：" + mode}
	}
	return nil
}

var allowedAutoCodeModes = map[string]struct{}{
	"ai": {}, "analysisChat": {}, "apiCompletion": {}, "addFunc": {}, "autoCompleteFunc": {}, "autoExportTemplate": {},
	"createWeb": {}, "dict": {}, "dictEye": {}, "exportCompletion": {}, "eye": {}, "workflowPromptChat": {},
}
