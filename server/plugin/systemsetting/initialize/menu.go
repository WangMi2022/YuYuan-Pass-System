package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
	"go.uber.org/zap"
)

var menuNames = []string{
	"systemSettings", collaborationMenuName, monitorMenuName, permissionMenuName, "state",
	"authority", "menu", "api", "user", "dictionary", "operation", "sysParams",
	"system", "apiToken", "loginLog", "sysVersion", "sysError",
}

func Menu(ctx context.Context) {
	utils.RegisterMenus(
		system.SysBaseMenu{
			ParentId: 0, Path: "admin", Name: "superAdmin", Hidden: false,
			Component: "view/superAdmin/index.vue", Sort: 5,
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
