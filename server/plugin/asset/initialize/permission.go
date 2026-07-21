package initialize

import (
	"context"
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const defaultAdminAuthorityID = 888

// Permission 首次安装时仅把资产模块授权给默认管理员，其他角色可在角色管理中按需分配。
func Permission(ctx context.Context) {
	if err := utils.RegisterPermissions(ctx, defaultAdminAuthorityID, menuNames, apiRules); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			global.GVA_LOG.Warn("未找到默认管理员角色，跳过资产模块自动授权", zap.Error(err))
		} else {
			global.GVA_LOG.Error("资产模块自动授权失败", zap.Error(err))
		}
		return
	}
}
