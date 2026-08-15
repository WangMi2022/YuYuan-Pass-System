package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

var apiRules = []system.SysApi{
	{Path: "/smart/copilot/query", Description: "业务助手只读查询", ApiGroup: "智能中心", Method: "POST"},
	{Path: "/smart/copilot/queryStream", Description: "业务助手流式只读查询", ApiGroup: "智能中心", Method: "POST"},
	{Path: "/smart/copilot/sessions", Description: "业务助手会话列表", ApiGroup: "智能中心", Method: "GET"},
	{Path: "/smart/copilot/session", Description: "业务助手会话详情", ApiGroup: "智能中心", Method: "GET"},
	{Path: "/smart/copilot/session", Description: "删除业务助手会话", ApiGroup: "智能中心", Method: "DELETE"},
	{Path: "/smart/copilot/tools", Description: "业务助手可用 Tool", ApiGroup: "智能中心", Method: "GET"},
	{Path: "/smartReport/today", Description: "今日智能日报", ApiGroup: "智能日报", Method: "GET"},
	{Path: "/smartReport/list", Description: "智能日报历史列表", ApiGroup: "智能日报", Method: "GET"},
	{Path: "/smartReport/detail", Description: "智能日报详情", ApiGroup: "智能日报", Method: "GET"},
	{Path: "/smartReport/generate", Description: "生成智能日报", ApiGroup: "智能日报", Method: "POST"},
	{Path: "/smartReport/subscription", Description: "查看智能日报订阅", ApiGroup: "智能日报", Method: "GET"},
	{Path: "/smartReport/subscription", Description: "更新智能日报订阅", ApiGroup: "智能日报", Method: "PUT"},
	{Path: "/smartReport/deliveries", Description: "智能日报发送记录", ApiGroup: "智能日报", Method: "GET"},
	{Path: "/smart/announcement/extract", Description: "公告提取日程草稿", ApiGroup: "智能草稿", Method: "POST"},
	{Path: "/smart/announcement/draft-schedule", Description: "公告生成日程草稿", ApiGroup: "智能草稿", Method: "POST"},
	{Path: "/smart/operation/draft", Description: "生成资产运营业务草稿", ApiGroup: "智能草稿", Method: "POST"},
	{Path: "/smart/operation/assets", Description: "查询业务草稿资产候选", ApiGroup: "智能草稿", Method: "GET"},
	{Path: "/smart/drafts", Description: "智能草稿列表", ApiGroup: "智能草稿", Method: "GET"},
	{Path: "/smart/draft/accept", Description: "确认智能草稿", ApiGroup: "智能草稿", Method: "POST"},
}

func Api(_ context.Context) { utils.RegisterApis(apiRules...) }
