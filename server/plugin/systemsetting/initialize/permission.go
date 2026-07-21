package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
	"go.uber.org/zap"
)

const defaultAdminAuthorityID = 888

func Permission(ctx context.Context) {
	if err := utils.RegisterPermissions(ctx, defaultAdminAuthorityID, menuNames, apiRules); err != nil {
		global.GVA_LOG.Error("系统设置自动授权失败", zap.Error(err))
		return
	}
}
