package initialize

import (
	"context"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const aiOperationsMenuName = "aiOperations"
const systemAdminMenuName = "superAdmin"
const aiServicesMenuTitle = "AI 服务管理"

var menuNames = []string{aiOperationsMenuName}

func Menu(ctx context.Context) {
	menus := []system.SysBaseMenu{{
		ParentId: 0, MenuLevel: 1, Path: aiOperationsMenuName, Name: aiOperationsMenuName, Component: "plugin/aioperations/view/operations.vue", Sort: 7,
		Meta: system.Meta{Title: aiServicesMenuTitle, Icon: "cpu", KeepAlive: true},
	}}
	utils.RegisterMenus(menus...)
	if err := syncAIServiceMenu(ctx); err != nil {
		global.GVA_LOG.Error("AI 服务管理菜单同步失败", zap.Error(err))
	}
}

func syncAIServiceMenu(ctx context.Context) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		systemAdmin := system.SysBaseMenu{
			ParentId: 0, Path: "admin", Name: systemAdminMenuName,
			Component: "view/superAdmin/index.vue", Sort: 9,
			Meta: system.Meta{Title: "系统管理", Icon: "setting"},
		}
		if err := tx.Where("name = ?", systemAdminMenuName).FirstOrCreate(&systemAdmin).Error; err != nil {
			return err
		}

		aiMenu := system.SysBaseMenu{
			ParentId: systemAdmin.ID, MenuLevel: 1,
			Path: aiOperationsMenuName, Name: aiOperationsMenuName, Hidden: false,
			Component: "plugin/aioperations/view/operations.vue", Sort: 7,
			Meta: system.Meta{Title: aiServicesMenuTitle, Icon: "cpu", KeepAlive: true},
		}
		if err := tx.Where("name = ?", aiOperationsMenuName).FirstOrCreate(&aiMenu).Error; err != nil {
			return err
		}
		if err := tx.Model(&system.SysBaseMenu{}).Where("name = ?", aiOperationsMenuName).Updates(map[string]any{
			"parent_id": systemAdmin.ID, "menu_level": 1, "path": aiOperationsMenuName,
			"component": "plugin/aioperations/view/operations.vue", "title": aiServicesMenuTitle, "icon": "cpu",
			"sort": 7, "hidden": false, "keep_alive": true,
		}).Error; err != nil {
			return err
		}

		var authorityIDs []string
		if err := tx.Model(&system.SysAuthorityMenu{}).
			Distinct("sys_authority_authority_id").
			Where("sys_base_menu_id = ?", strconv.Itoa(int(aiMenu.ID))).
			Pluck("sys_authority_authority_id", &authorityIDs).Error; err != nil {
			return err
		}
		for _, authorityID := range authorityIDs {
			relation := system.SysAuthorityMenu{
				MenuId: strconv.Itoa(int(systemAdmin.ID)), AuthorityId: authorityID,
			}
			if err := tx.Where(
				"sys_base_menu_id = ? AND sys_authority_authority_id = ?", relation.MenuId, relation.AuthorityId,
			).FirstOrCreate(&relation).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
