package initialize

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	model "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

func Menu(ctx context.Context) {
	entities := []model.SysBaseMenu{
		{
			ParentId:  9,
			Path:      "anInfo",
			Name:      "anInfo",
			Hidden:    false,
			Component: "plugin/announcement/view/info.vue",
			Sort:      5,
			Meta:      model.Meta{Title: "公告管理", Icon: "bell"},
		},
	}
	utils.RegisterMenus(entities...)
	var parent model.SysBaseMenu
	if err := global.GVA_DB.Where("name = ?", "superAdmin").First(&parent).Error; err == nil {
		_ = global.GVA_DB.Model(&model.SysBaseMenu{}).Where("name = ?", "anInfo").Updates(map[string]any{
			"parent_id": parent.ID,
			"title":     "公告管理",
			"icon":      "bell",
		}).Error
	}
}
