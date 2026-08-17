package initialize

import (
	"context"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/system"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/plugin-tool/utils"
	"go.uber.org/zap"
)

var menuNames = []string{
	"systemSettings", workCalendarMenuName, "workSchedule", collaborationMenuName, monitorMenuName, permissionMenuName, auditMenuName, "state",
	"authority", "menu", "api", "user", "dictionary", "operation", "sysParams",
	"system", "apiToken", "loginLog", "sysVersion", "sysError",
	systemDataMenuName, systemConfigurationMenuName, systemIntegrationMenuName, systemIntelligenceMenuName,
}

func Menu(ctx context.Context) {
	utils.RegisterMenus(
		system.SysBaseMenu{
			ParentId: 0, Path: "workCalendar", Name: workCalendarMenuName, Hidden: false,
			Component: "view/routerHolder.vue", Sort: 4,
			Meta: system.Meta{Title: "工作日历", Icon: "calendar"},
		},
		system.SysBaseMenu{
			Path: "schedule", Name: "workSchedule", Hidden: false,
			Component: "view/workCalendar/index.vue", Sort: 1,
			Meta: system.Meta{Title: "日程总览", Icon: "calendar", KeepAlive: true},
		},
		system.SysBaseMenu{
			ParentId: 0, Path: "admin", Name: "superAdmin", Hidden: false,
			Component: "view/superAdmin/index.vue", Sort: 7,
			Meta: system.Meta{Title: "系统管理", Icon: "setting"},
		},
		system.SysBaseMenu{
			Path: "systemSettings", Name: "systemSettings", Hidden: false,
			Component: "view/superAdmin/systemSettings/index.vue", Sort: 4,
			Meta: system.Meta{Title: "系统设置", Icon: "setting"},
		},
	)
	if err := syncBusinessNavigation(ctx); err != nil {
		global.GVA_LOG.Error("业务菜单同步失败", zap.Error(err))
	}
}
