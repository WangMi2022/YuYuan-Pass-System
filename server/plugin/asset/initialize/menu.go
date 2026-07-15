package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

var menuNames = []string{"assetCenter", "assetDashboard", "assetInventory", "assetCategories"}

func Menu(_ context.Context) {
	utils.RegisterMenus(
		system.SysBaseMenu{
			ParentId: 0, Path: "assetCenter", Name: "assetCenter", Hidden: false,
			Component: "view/routerHolder.vue", Sort: 2,
			Meta: system.Meta{Title: "资产中心", Icon: "box"},
		},
		system.SysBaseMenu{
			Path: "assetDashboard", Name: "assetDashboard", Hidden: false,
			Component: "plugin/asset/view/dashboard.vue", Sort: 1,
			Meta: system.Meta{Title: "资产大屏", Icon: "data-analysis", KeepAlive: true},
		},
		system.SysBaseMenu{
			Path: "assetInventory", Name: "assetInventory", Hidden: false,
			Component: "plugin/asset/view/assets.vue", Sort: 2,
			Meta: system.Meta{Title: "资产档案", Icon: "list", KeepAlive: true},
		},
		system.SysBaseMenu{
			Path: "assetCategories", Name: "assetCategories", Hidden: false,
			Component: "plugin/asset/view/categories.vue", Sort: 3,
			Meta: system.Meta{Title: "资产分类", Icon: "collection-tag", KeepAlive: true},
		},
	)
}
