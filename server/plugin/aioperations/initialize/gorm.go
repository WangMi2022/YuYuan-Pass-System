package initialize

import (
	"context"

	"github.com/WangMi2022/mit-assets-admin/server/ai"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"go.uber.org/zap"
)

func Gorm(ctx context.Context) {
	if err := global.GVA_DB.WithContext(ctx).AutoMigrate(&ai.ModelInvocation{}, &ai.UsageQuota{}, &ai.PromptTemplate{}); err != nil {
		global.GVA_LOG.Error("AI Gateway 数据表迁移失败", zap.Error(err))
	}
}
