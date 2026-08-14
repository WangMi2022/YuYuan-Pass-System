package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

var menuNames = []string{"aiOperations"}

func Menu(_ context.Context) {
	menus := []system.SysBaseMenu{{
		ParentId: 0, Path: "aiOperations", Name: "aiOperations", Component: "plugin/aioperations/view/operations.vue", Sort: 98,
		Meta: system.Meta{Title: "AI 运营", Icon: "cpu", KeepAlive: true},
	}}
	utils.RegisterMenus(menus...)
	_ = global.GVA_DB.Model(&system.SysBaseMenu{}).Where("name = ?", "aiOperations").Updates(map[string]any{
		"parent_id": 0, "menu_level": 0, "path": "aiOperations", "component": "plugin/aioperations/view/operations.vue",
		"title": "AI 运营", "icon": "cpu", "sort": 98, "hidden": false, "keep_alive": true,
	}).Error
}
