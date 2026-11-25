package initialize

import (
	"context"

	"github.com/WangMi2022/mit-assets-admin/server/model/system"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/plugin-tool/utils"
)

var apiRules = []system.SysApi{
	{Path: "/workSchedule/list", Description: "日程列表", ApiGroup: "工作日程", Method: "GET"},
	{Path: "/workSchedule/notifications", Description: "日程提醒", ApiGroup: "工作日程", Method: "GET"},
	{Path: "/workSchedule/create", Description: "新建日程", ApiGroup: "工作日程", Method: "POST"},
	{Path: "/workSchedule/update", Description: "更新日程", ApiGroup: "工作日程", Method: "PUT"},
	{Path: "/workSchedule/delete", Description: "删除日程", ApiGroup: "工作日程", Method: "DELETE"},
	{Path: "/workSchedule/import", Description: "导入日程", ApiGroup: "工作日程", Method: "POST"},
	{Path: "/workSchedule/notifications/read", Description: "标记日程提醒已读", ApiGroup: "工作日程", Method: "POST"},
	{Path: "/workSchedule/notifications/readAll", Description: "全部日程提醒已读", ApiGroup: "工作日程", Method: "POST"},
}

func Api(_ context.Context) { utils.RegisterApis(apiRules...) }
