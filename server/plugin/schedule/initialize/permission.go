package initialize

import (
	"context"
	"errors"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/plugin-tool/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const defaultAdminAuthorityID = 888

func Permission(ctx context.Context) {
	if err := utils.RegisterPermissions(ctx, defaultAdminAuthorityID, nil, apiRules); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			global.GVA_LOG.Warn("未找到默认管理员角色，跳过日程权限自动授权", zap.Error(err))
			return
		}
		global.GVA_LOG.Error("日程权限自动授权失败", zap.Error(err))
	}
}
