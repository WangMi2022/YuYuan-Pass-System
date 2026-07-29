package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

var apiRules = []system.SysApi{
	{Path: "/invoice/upload", Description: "上传发票图片", ApiGroup: "流水管理", Method: "POST"},
	{Path: "/invoice/update", Description: "核对发票信息", ApiGroup: "流水管理", Method: "PUT"},
	{Path: "/invoice/confirm", Description: "确认发票", ApiGroup: "流水管理", Method: "PUT"},
	{Path: "/invoice/reopen", Description: "重新打开发票", ApiGroup: "流水管理", Method: "PUT"},
	{Path: "/invoice/retry", Description: "重新识别发票", ApiGroup: "流水管理", Method: "PUT"},
	{Path: "/invoice/recheck", Description: "调用多模态模型重新核对发票", ApiGroup: "流水管理", Method: "POST"},
	{Path: "/invoice/provider/test", Description: "测试发票识别服务连接", ApiGroup: "运行配置", Method: "POST"},
	{Path: "/invoice/delete", Description: "删除发票", ApiGroup: "流水管理", Method: "DELETE"},
	{Path: "/invoice/list", Description: "发票台账", ApiGroup: "流水管理", Method: "GET"},
	{Path: "/invoice/detail", Description: "发票详情", ApiGroup: "流水管理", Method: "GET"},
	{Path: "/invoice/file", Description: "读取发票原图", ApiGroup: "流水管理", Method: "GET"},
	{Path: "/invoice/dashboard", Description: "流水统计", ApiGroup: "流水管理", Method: "GET"},
	{Path: "/invoice/categoryOptions", Description: "发票分类选项", ApiGroup: "流水管理", Method: "GET"},
	{Path: "/invoiceCategory/create", Description: "新增发票分类", ApiGroup: "流水分类", Method: "POST"},
	{Path: "/invoiceCategory/update", Description: "更新发票分类", ApiGroup: "流水分类", Method: "PUT"},
	{Path: "/invoiceCategory/delete", Description: "删除发票分类", ApiGroup: "流水分类", Method: "DELETE"},
	{Path: "/invoiceCategory/list", Description: "发票分类列表", ApiGroup: "流水分类", Method: "GET"},
	{Path: "/invoiceRule/create", Description: "新增分类规则", ApiGroup: "流水分类", Method: "POST"},
	{Path: "/invoiceRule/update", Description: "更新分类规则", ApiGroup: "流水分类", Method: "PUT"},
	{Path: "/invoiceRule/delete", Description: "删除分类规则", ApiGroup: "流水分类", Method: "DELETE"},
	{Path: "/invoiceRule/list", Description: "分类规则列表", ApiGroup: "流水分类", Method: "GET"},
}

func Api(_ context.Context) { utils.RegisterApis(apiRules...) }
