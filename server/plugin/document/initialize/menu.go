package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

var menuNames = []string{"documentManagement", "documentViewer"}

func Menu(_ context.Context) {
	utils.RegisterMenus(
		system.SysBaseMenu{
			ParentId: 0, Path: "documentManagement", Name: "documentManagement", Hidden: true,
			Component: "view/routerHolder.vue", Sort: 3,
			Meta: system.Meta{Title: "文档管理", Icon: "document"},
		},
		system.SysBaseMenu{
			Path: "documentViewer", Name: "documentViewer", Hidden: false,
			Component: "plugin/document/view/viewer.vue", Sort: 1,
			Meta: system.Meta{Title: "文档管理", Icon: "document-copy", KeepAlive: false},
		},
	)
}
