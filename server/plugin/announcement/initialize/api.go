package initialize

import (
	"context"
	model "github.com/WangMi2022/mit-assets-admin/server/model/system"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/plugin-tool/utils"
)

var apiRules = []model.SysApi{
	{
		Path:        "/info/createInfo",
		Description: "新建公告",
		ApiGroup:    "公告",
		Method:      "POST",
	},
	{
		Path:        "/info/deleteInfo",
		Description: "删除公告",
		ApiGroup:    "公告",
		Method:      "DELETE",
	},
	{
		Path:        "/info/deleteInfoByIds",
		Description: "批量删除公告",
		ApiGroup:    "公告",
		Method:      "DELETE",
	},
	{
		Path:        "/info/updateInfo",
		Description: "更新公告",
		ApiGroup:    "公告",
		Method:      "PUT",
	},
	{
		Path:        "/info/findInfo",
		Description: "根据ID获取公告",
		ApiGroup:    "公告",
		Method:      "GET",
	},
	{
		Path:        "/info/getInfoList",
		Description: "获取公告列表",
		ApiGroup:    "公告",
		Method:      "GET",
	},
	{
		Path:        "/info/notifications",
		Description: "获取公告通知",
		ApiGroup:    "公告通知",
		Method:      "GET",
	},
}

func Api(_ context.Context) {
	utils.RegisterApis(apiRules...)
}
