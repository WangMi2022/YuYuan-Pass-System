package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

var apiRules = []system.SysApi{
	{Path: "/site/create", Description: "新增收藏站点", ApiGroup: "站点管理", Method: "POST"},
	{Path: "/site/update", Description: "更新收藏站点", ApiGroup: "站点管理", Method: "PUT"},
	{Path: "/site/delete", Description: "删除收藏站点", ApiGroup: "站点管理", Method: "DELETE"},
	{Path: "/site/list", Description: "收藏站点列表", ApiGroup: "站点管理", Method: "GET"},
	{Path: "/site/detail", Description: "收藏站点详情", ApiGroup: "站点管理", Method: "GET"},
	{Path: "/site/categories", Description: "收藏站点分类", ApiGroup: "站点管理", Method: "GET"},
	{Path: "/site/visit", Description: "记录站点访问", ApiGroup: "站点管理", Method: "POST"},
}

func Api(_ context.Context) { utils.RegisterApis(apiRules...) }
