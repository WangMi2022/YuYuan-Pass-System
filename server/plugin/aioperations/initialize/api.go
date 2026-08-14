package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

var apiRules = []system.SysApi{
	{Path: "/ai/providers", Description: "查看 AI Provider 状态", ApiGroup: "AI 运营", Method: "GET"},
	{Path: "/ai/providers", Description: "更新 AI Provider 配置", ApiGroup: "AI 运营", Method: "PUT"},
	{Path: "/ai/usage/summary", Description: "查看 AI 用量汇总", ApiGroup: "AI 运营", Method: "GET"},
	{Path: "/ai/invocations", Description: "查看 AI 调用审计", ApiGroup: "AI 运营", Method: "GET"},
	{Path: "/ai/quotas", Description: "查看 AI 配额", ApiGroup: "AI 运营", Method: "GET"},
	{Path: "/ai/quotas", Description: "更新 AI 配额", ApiGroup: "AI 运营", Method: "PUT"},
	{Path: "/ai/prompts", Description: "查看 AI Prompt 模板", ApiGroup: "AI 运营", Method: "GET"},
	{Path: "/ai/prompts", Description: "创建 AI Prompt 模板", ApiGroup: "AI 运营", Method: "POST"},
	{Path: "/ai/prompts/activate", Description: "激活 AI Prompt 模板", ApiGroup: "AI 运营", Method: "PUT"},
}

func Api(_ context.Context) { utils.RegisterApis(apiRules...) }
