package initialize

import (
	"context"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/site/model"
	"go.uber.org/zap"
)

func Gorm(ctx context.Context) {
	if err := global.GVA_DB.WithContext(ctx).AutoMigrate(&model.SiteBookmark{}); err != nil {
		global.GVA_LOG.Error("站点管理数据表迁移失败", zap.Error(err))
	}
}
