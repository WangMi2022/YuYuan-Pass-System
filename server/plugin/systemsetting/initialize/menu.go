package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

var menuNames = []string{"systemSettings"}

func Menu(_ context.Context) {
	utils.RegisterMenus(
		system.SysBaseMenu{
			ParentId: 0, Path: "admin", Name: "superAdmin", Hidden: false,
			Component: "view/superAdmin/index.vue", Sort: 3,
			Meta: system.Meta{Title: "系统管理", Icon: "setting"},
		},
		system.SysBaseMenu{
			Path: "systemSettings", Name: "systemSettings", Hidden: false,
			Component: "view/superAdmin/systemSettings/index.vue", Sort: 8,
			Meta: system.Meta{Title: "系统设置", Icon: "picture"},
		},
	)
	// 兼容已初始化数据库：注册逻辑不会覆盖既有菜单标题，因此显式统一父菜单名称。
	_ = global.GVA_DB.Model(&system.SysBaseMenu{}).Where("name = ?", "superAdmin").
		Updates(map[string]any{"title": "系统管理", "icon": "setting"}).Error
}
