package initialize

import (
	"context"
	"strconv"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemService "github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"go.uber.org/zap"
)

const defaultAdminAuthorityID = 888

func Permission(ctx context.Context) {
	authorityID := strconv.Itoa(defaultAdminAuthorityID)
	var menus []system.SysBaseMenu
	if err := global.GVA_DB.WithContext(ctx).Where("name IN ?", menuNames).Find(&menus).Error; err == nil {
		for _, menu := range menus {
			relation := system.SysAuthorityMenu{MenuId: strconv.Itoa(int(menu.ID)), AuthorityId: authorityID}
			_ = global.GVA_DB.WithContext(ctx).Where(
				"sys_base_menu_id = ? AND sys_authority_authority_id = ?", relation.MenuId, relation.AuthorityId,
			).FirstOrCreate(&relation).Error
		}
	}
	for _, item := range apiRules {
		rule := gormadapter.CasbinRule{Ptype: "p", V0: authorityID, V1: item.Path, V2: item.Method}
		if err := global.GVA_DB.WithContext(ctx).Where(
			"ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", rule.Ptype, rule.V0, rule.V1, rule.V2,
		).FirstOrCreate(&rule).Error; err != nil {
			global.GVA_LOG.Warn("系统设置接口授权失败", zap.Error(err))
		}
	}
	if err := systemService.CasbinServiceApp.FreshCasbin(); err != nil {
		global.GVA_LOG.Warn("刷新系统设置权限失败", zap.Error(err))
	}
}
