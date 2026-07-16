package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

var apiRules = []system.SysApi{
	{Path: "/appearance/login-backgrounds", Description: "登录背景图库", ApiGroup: "系统设置", Method: "GET"},
	{Path: "/appearance/login-background", Description: "新增登录背景", ApiGroup: "系统设置", Method: "POST"},
	{Path: "/appearance/login-background/activate", Description: "切换登录背景", ApiGroup: "系统设置", Method: "PUT"},
	{Path: "/appearance/login-background", Description: "删除登录背景", ApiGroup: "系统设置", Method: "DELETE"},
	{Path: "/appearance/login-logo", Description: "设置登录图标", ApiGroup: "系统设置", Method: "PUT"},
	{Path: "/appearance/login-logo", Description: "恢复默认登录图标", ApiGroup: "系统设置", Method: "DELETE"},
}

func Api(_ context.Context) { utils.RegisterApis(apiRules...) }
