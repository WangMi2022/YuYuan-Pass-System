package initialize

import (
	"context"
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const defaultAdminRoleID = 888

func Permission(ctx context.Context) {
	if err := utils.RegisterPermissions(ctx, defaultAdminRoleID, menuNames, apiRules); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			global.GVA_LOG.Warn("未找到默认管理员角色，跳过流水管理自动授权", zap.Error(err))
		} else {
			global.GVA_LOG.Error("流水管理自动授权失败", zap.Error(err))
		}
	}
}
