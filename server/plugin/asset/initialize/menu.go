package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

var menuNames = []string{
	"assetCenter", "assetDashboard", "assetInventory", "assetCategories",
	"assetInbound", "assetIssue", "assetTransfer", "assetReturn", "assetMaintenance", "assetScrap",
}

func Menu(_ context.Context) {
	menus := []system.SysBaseMenu{
		{
			ParentId: 0, Path: "assetCenter", Name: "assetCenter", Hidden: false,
			Component: "view/routerHolder.vue", Sort: 2,
			Meta: system.Meta{Title: "资产管理", Icon: "box"},
		},
		{
			Path: "assetDashboard", Name: "assetDashboard", Hidden: true,
			Component: "plugin/asset/view/dashboard.vue", Sort: 0,
			Meta: system.Meta{Title: "资产大屏", Icon: "data-analysis", KeepAlive: true},
		},
		{
			Path: "assetInventory", Name: "assetInventory", Hidden: false,
			Component: "plugin/asset/view/assets.vue", Sort: 1,
			Meta: system.Meta{Title: "资产档案", Icon: "list", KeepAlive: true},
		},
		{
			Path: "assetCategories", Name: "assetCategories", Hidden: false,
			Component: "plugin/asset/view/categories.vue", Sort: 2,
			Meta: system.Meta{Title: "分类管理", Icon: "collection-tag", KeepAlive: true},
		},
		{Path: "assetInbound", Name: "assetInbound", Component: "plugin/asset/view/operations.vue", Sort: 3, Meta: system.Meta{Title: "入库管理", Icon: "box"}},
		{Path: "assetIssue", Name: "assetIssue", Component: "plugin/asset/view/operations.vue", Sort: 4, Meta: system.Meta{Title: "领用管理", Icon: "user"}},
		{Path: "assetTransfer", Name: "assetTransfer", Component: "plugin/asset/view/operations.vue", Sort: 5, Meta: system.Meta{Title: "调拨管理", Icon: "switch"}},
		{Path: "assetReturn", Name: "assetReturn", Component: "plugin/asset/view/operations.vue", Sort: 6, Meta: system.Meta{Title: "归还管理", Icon: "refresh-left"}},
		{Path: "assetMaintenance", Name: "assetMaintenance", Component: "plugin/asset/view/operations.vue", Sort: 7, Meta: system.Meta{Title: "维修管理", Icon: "tools"}},
		{Path: "assetScrap", Name: "assetScrap", Component: "plugin/asset/view/operations.vue", Sort: 8, Meta: system.Meta{Title: "报废管理", Icon: "delete"}},
	}
	utils.RegisterMenus(menus...)

	// RegisterMenus 不更新已有记录，显式同步二开后的菜单层级和显示属性。
	var parent system.SysBaseMenu
	if err := global.GVA_DB.Where("name = ?", "assetCenter").First(&parent).Error; err != nil {
		return
	}
	_ = global.GVA_DB.Model(&parent).Updates(map[string]any{
		"title": "资产管理", "icon": "box", "sort": 2, "hidden": false,
	}).Error
	for _, menu := range menus[1:] {
		_ = global.GVA_DB.Model(&system.SysBaseMenu{}).Where("name = ?", menu.Name).Updates(map[string]any{
			"parent_id": parent.ID, "path": menu.Path, "component": menu.Component,
			"title": menu.Meta.Title, "icon": menu.Meta.Icon, "sort": menu.Sort,
			"hidden": menu.Hidden, "keep_alive": menu.Meta.KeepAlive,
		}).Error
	}
}
