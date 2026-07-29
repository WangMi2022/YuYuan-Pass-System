package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

var menuNames = []string{"invoiceCenter", "invoiceDashboard", "invoiceRecognition", "invoiceLedger", "invoiceCategories"}

func Menu(_ context.Context) {
	menus := []system.SysBaseMenu{
		{ParentId: 0, Path: "invoiceCenter", Name: "invoiceCenter", Component: "view/routerHolder.vue", Sort: 3, Meta: system.Meta{Title: "流水管理", Icon: "wallet"}},
		{Path: "invoiceDashboard", Name: "invoiceDashboard", Component: "plugin/invoice/view/dashboard.vue", Sort: 1, Meta: system.Meta{Title: "流水总览", Icon: "data-analysis", KeepAlive: true}},
		{Path: "invoiceRecognition", Name: "invoiceRecognition", Component: "plugin/invoice/view/recognition.vue", Sort: 2, Meta: system.Meta{Title: "发票识别", Icon: "camera", KeepAlive: true}},
		{Path: "invoiceLedger", Name: "invoiceLedger", Component: "plugin/invoice/view/ledger.vue", Sort: 3, Meta: system.Meta{Title: "发票台账", Icon: "tickets", KeepAlive: true}},
		{Path: "invoiceCategories", Name: "invoiceCategories", Component: "plugin/invoice/view/categories.vue", Sort: 4, Meta: system.Meta{Title: "分类规则", Icon: "collection-tag", KeepAlive: true}},
	}
	utils.RegisterMenus(menus...)
	var parent system.SysBaseMenu
	if err := global.GVA_DB.Where("name = ?", "invoiceCenter").First(&parent).Error; err != nil {
		return
	}
	_ = global.GVA_DB.Model(&parent).Updates(map[string]any{
		"parent_id": 0, "menu_level": 0, "title": "流水管理", "icon": "wallet", "sort": 3, "hidden": false,
	}).Error
	for _, menu := range menus[1:] {
		_ = global.GVA_DB.Model(&system.SysBaseMenu{}).Where("name = ?", menu.Name).Updates(map[string]any{
			"parent_id": parent.ID, "menu_level": 1, "path": menu.Path, "component": menu.Component,
			"title": menu.Meta.Title, "icon": menu.Meta.Icon, "sort": menu.Sort,
			"hidden": false, "keep_alive": menu.Meta.KeepAlive,
		}).Error
	}
}
