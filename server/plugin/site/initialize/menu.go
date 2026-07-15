package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

var menuNames = []string{"siteManagement", "siteBookmarks"}

func Menu(_ context.Context) {
	utils.RegisterMenus(
		system.SysBaseMenu{
			ParentId: 0, Path: "siteManagement", Name: "siteManagement", Hidden: false,
			Component: "view/routerHolder.vue", Sort: 4,
			Meta: system.Meta{Title: "站点管理", Icon: "link"},
		},
		system.SysBaseMenu{
			Path: "siteBookmarks", Name: "siteBookmarks", Hidden: false,
			Component: "plugin/site/view/bookmarks.vue", Sort: 1,
			Meta: system.Meta{Title: "站点收藏", Icon: "link", KeepAlive: true},
		},
	)
}
