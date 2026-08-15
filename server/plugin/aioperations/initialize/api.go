package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
	"go.uber.org/zap"
)

var apiRules = []system.SysApi{
	{Path: "/ai/providers", Description: "查看模型接入状态", ApiGroup: aiServicesMenuTitle, Method: "GET"},
	{Path: "/ai/providers", Description: "更新模型接入配置", ApiGroup: aiServicesMenuTitle, Method: "PUT"},
	{Path: "/ai/invoice-recognition", Description: "查看智能识别配置", ApiGroup: aiServicesMenuTitle, Method: "GET"},
	{Path: "/ai/invoice-recognition", Description: "更新智能识别配置", ApiGroup: aiServicesMenuTitle, Method: "PUT"},
	{Path: "/ai/invoice-recognition/test", Description: "测试智能识别服务", ApiGroup: aiServicesMenuTitle, Method: "POST"},
	{Path: "/ai/usage/summary", Description: "查看智能服务用量汇总", ApiGroup: aiServicesMenuTitle, Method: "GET"},
	{Path: "/ai/invocations", Description: "查看智能服务调用日志", ApiGroup: aiServicesMenuTitle, Method: "GET"},
	{Path: "/ai/quotas", Description: "查看智能服务配额", ApiGroup: aiServicesMenuTitle, Method: "GET"},
	{Path: "/ai/quotas", Description: "更新智能服务配额", ApiGroup: aiServicesMenuTitle, Method: "PUT"},
	{Path: "/ai/prompts", Description: "查看 Prompt 模板版本", ApiGroup: aiServicesMenuTitle, Method: "GET"},
	{Path: "/ai/prompts", Description: "创建 Prompt 模板版本", ApiGroup: aiServicesMenuTitle, Method: "POST"},
	{Path: "/ai/prompts/activate", Description: "激活 Prompt 模板版本", ApiGroup: aiServicesMenuTitle, Method: "PUT"},
}

func Api(ctx context.Context) {
	if err := migrateAIServiceAPIRegistry(ctx); err != nil {
		global.GVA_LOG.Error("智能能力配置 API 元数据迁移失败", zap.Error(err))
	}
	utils.RegisterApis(apiRules...)
}
