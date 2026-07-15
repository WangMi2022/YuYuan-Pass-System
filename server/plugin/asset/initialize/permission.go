package initialize

import (
	"context"
	"strconv"

	"github.com/casbin/gorm-adapter/v3"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemService "github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"go.uber.org/zap"
)

const defaultAdminAuthorityID = 888

// Permission 首次安装时仅把资产模块授权给默认管理员，其他角色可在角色管理中按需分配。
func Permission(ctx context.Context) {
	var authority system.SysAuthority
	if err := global.GVA_DB.WithContext(ctx).Where("authority_id = ?", defaultAdminAuthorityID).First(&authority).Error; err != nil {
		global.GVA_LOG.Warn("未找到默认管理员角色，跳过资产模块自动授权", zap.Error(err))
		return
	}
	authorityID := strconv.Itoa(defaultAdminAuthorityID)
	var menus []system.SysBaseMenu
	if err := global.GVA_DB.WithContext(ctx).Where("name IN ?", menuNames).Find(&menus).Error; err != nil {
		global.GVA_LOG.Error("查询资产菜单失败", zap.Error(err))
		return
	}
	for _, menu := range menus {
		relation := system.SysAuthorityMenu{MenuId: strconv.Itoa(int(menu.ID)), AuthorityId: authorityID}
		if err := global.GVA_DB.WithContext(ctx).Where(
			"sys_base_menu_id = ? AND sys_authority_authority_id = ?", relation.MenuId, relation.AuthorityId,
		).FirstOrCreate(&relation).Error; err != nil {
			global.GVA_LOG.Warn("资产菜单授权失败", zap.Error(err))
		}
	}
	for _, item := range apiRules {
		rule := gormadapter.CasbinRule{Ptype: "p", V0: authorityID, V1: item.Path, V2: item.Method}
		if err := global.GVA_DB.WithContext(ctx).Where(
			"ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", rule.Ptype, rule.V0, rule.V1, rule.V2,
		).FirstOrCreate(&rule).Error; err != nil {
			global.GVA_LOG.Warn("资产接口授权失败", zap.Error(err))
		}
	}
	if err := systemService.CasbinServiceApp.FreshCasbin(); err != nil {
		global.GVA_LOG.Warn("刷新资产接口权限失败", zap.Error(err))
	}
}
