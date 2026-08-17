package initialize

import (
	"context"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/smart/model"
	"go.uber.org/zap"
)

func Gorm(ctx context.Context) {
	if err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		&model.CopilotSession{}, &model.CopilotMessage{}, &model.SmartDailyReport{},
		&model.SmartReportSubscription{}, &model.SmartReportDelivery{}, &model.SmartDraft{},
	); err != nil {
		global.GVA_LOG.Error("智能中心数据表迁移失败", zap.Error(err))
	}
}
