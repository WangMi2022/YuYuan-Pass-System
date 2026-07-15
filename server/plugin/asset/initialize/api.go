package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

var apiRules = []system.SysApi{
	{Path: "/asset/create", Description: "新增资产", ApiGroup: "资产管理", Method: "POST"},
	{Path: "/asset/update", Description: "更新资产", ApiGroup: "资产管理", Method: "PUT"},
	{Path: "/asset/delete", Description: "删除资产", ApiGroup: "资产管理", Method: "DELETE"},
	{Path: "/asset/detail", Description: "资产详情", ApiGroup: "资产管理", Method: "GET"},
	{Path: "/asset/list", Description: "资产列表", ApiGroup: "资产管理", Method: "GET"},
	{Path: "/asset/dashboard", Description: "资产统计大屏", ApiGroup: "资产管理", Method: "GET"},
	{Path: "/asset/categoryOptions", Description: "资产分类选项", ApiGroup: "资产管理", Method: "GET"},
	{Path: "/asset/uploadPhoto", Description: "上传资产图片", ApiGroup: "资产管理", Method: "POST"},
	{Path: "/asset/deletePhoto", Description: "删除资产图片", ApiGroup: "资产管理", Method: "DELETE"},
	{Path: "/assetCategory/create", Description: "新增资产分类", ApiGroup: "资产分类", Method: "POST"},
	{Path: "/assetCategory/update", Description: "更新资产分类", ApiGroup: "资产分类", Method: "PUT"},
	{Path: "/assetCategory/delete", Description: "删除资产分类", ApiGroup: "资产分类", Method: "DELETE"},
	{Path: "/assetCategory/list", Description: "资产分类列表", ApiGroup: "资产分类", Method: "GET"},
}

func Api(_ context.Context) { utils.RegisterApis(apiRules...) }
