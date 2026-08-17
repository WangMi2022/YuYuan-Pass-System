package ai

import (
	"context"
	"errors"
	"strings"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"gorm.io/gorm"
)

func resolvePrompt(ctx context.Context, request CompletionRequest) (CompletionRequest, error) {
	request.PromptKey = strings.TrimSpace(request.PromptKey)
	if request.PromptKey == "" {
		return request, nil
	}
	if global.GVA_DB == nil {
		return CompletionRequest{}, &Error{Type: ErrorTypeProvider, Message: "Prompt 模板存储尚未初始化"}
	}
	query := global.GVA_DB.WithContext(ctx).Where("prompt_key = ?", request.PromptKey)
	if request.PromptVersion > 0 {
		query = query.Where("version = ?", request.PromptVersion)
	} else {
		query = query.Where("status = ?", PromptStatusActive)
	}
	var template PromptTemplate
	if err := query.Order("version DESC").First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return CompletionRequest{}, &Error{Type: ErrorTypeValidation, Message: "指定的 Prompt 模板不存在或未激活"}
		}
		return CompletionRequest{}, &Error{Type: ErrorTypeProvider, Message: "读取 Prompt 模板失败", Cause: err}
	}
	if request.Prompt == "" {
		request.Prompt = template.Content
	} else {
		request.Prompt = template.Content + "\n\n" + request.Prompt
	}
	request.PromptVersion = template.Version
	request.OutputSchema = template.OutputSchema
	return request, nil
}
