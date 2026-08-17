package system

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/WangMi2022/mit-assets-admin/server/ai"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/common"
)

func buildLLMAutoPath(llm common.JSONMap) (string, error) {
	mode := strings.TrimSpace(fmt.Sprintf("%v", llm["mode"]))
	if err := validateAutoCodeMode(mode); err != nil {
		return "", err
	}
	path := strings.TrimSpace(global.GVA_CONFIG.AutoCode.AiPath)
	if path == "" {
		return "", fmt.Errorf("请先在 config.yaml 的 autocode.ai-path 中配置自有 AI 服务地址")
	}
	return strings.ReplaceAll(path, "{FUNC}", mode), nil
}

func validateAutoCodeMode(mode string) error {
	return ai.ValidateAutoCodeMode(mode)
}

func (s *AutoCodeService) LLMAuto(ctx context.Context, llm common.JSONMap) (interface{}, error) {
	result, err := ai.Default.Complete(ctx, ai.CompletionRequest{
		Module: "autocode", Operation: "generate", Provider: "autocode-compat", Payload: llm,
	})
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (s *AutoCodeService) LLMAutoStream(ctx context.Context, llm common.JSONMap) (*http.Response, error) {
	result, err := ai.Default.Stream(ctx, ai.CompletionRequest{
		Module: "autocode", Operation: "generate_stream", Provider: "autocode-compat", Payload: llm,
	})
	if err != nil {
		return nil, err
	}
	return result.Response, nil
}
