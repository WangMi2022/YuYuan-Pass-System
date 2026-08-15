package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

var menuNames = []string{"smartCenter", "smartCopilot", "smartReport", "smartDrafts"}

func Menu(_ context.Context) {
	menus := []system.SysBaseMenu{
		{ParentId: 0, Path: "smartCenter", Name: "smartCenter", Component: "view/routerHolder.vue", Sort: 3, Meta: system.Meta{Title: "智能中心", Icon: "cpu", KeepAlive: true}},
		{Path: "smartCopilot", Name: "smartCopilot", Component: "plugin/smart/view/copilot.vue", Sort: 1, Meta: system.Meta{Title: "业务助手", Icon: "chat-dot-round", KeepAlive: true}},
		{Path: "smartReport", Name: "smartReport", Component: "plugin/smart/view/report.vue", Sort: 2, Meta: system.Meta{Title: "智能日报", Icon: "data-analysis", KeepAlive: true}},
		{Path: "smartDrafts", Name: "smartDrafts", Component: "plugin/smart/view/drafts.vue", Sort: 3, Meta: system.Meta{Title: "智能草稿", Icon: "document", KeepAlive: true}},
	}
	utils.RegisterMenus(menus...)
	var parent system.SysBaseMenu
	if err := global.GVA_DB.Where("name = ?", "smartCenter").First(&parent).Error; err != nil {
		return
	}
	_ = global.GVA_DB.Model(&parent).Updates(map[string]any{"parent_id": 0, "menu_level": 0, "path": "smartCenter", "component": "view/routerHolder.vue", "title": "智能中心", "icon": "cpu", "sort": 3, "hidden": false, "keep_alive": true}).Error
	for _, item := range menus[1:] {
		_ = global.GVA_DB.Model(&system.SysBaseMenu{}).Where("name = ?", item.Name).Updates(map[string]any{"parent_id": parent.ID, "menu_level": 1, "path": item.Path, "component": item.Component, "title": item.Meta.Title, "icon": item.Meta.Icon, "sort": item.Sort, "hidden": false, "keep_alive": true}).Error
	}
}
