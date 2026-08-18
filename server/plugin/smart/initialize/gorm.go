package initialize

import (
	"context"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/smart/model"
	"go.uber.org/zap"
)

func Gorm(ctx context.Context) {
	db := global.GVA_DB.WithContext(ctx)
	if err := db.AutoMigrate(
		&model.CopilotSession{}, &model.CopilotMessage{}, &model.CopilotRun{}, &model.SmartDailyReport{},
		&model.SmartReportSubscription{}, &model.SmartReportDelivery{}, &model.SmartDraft{}, &model.KnowledgeChunk{},
	); err != nil {
		global.GVA_LOG.Error("智能中心数据表迁移失败", zap.Error(err))
		return
	}
	if db.Dialector.Name() == "postgres" {
		if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_smart_knowledge_fts
			ON smart_knowledge_chunks USING GIN (
				to_tsvector('simple', COALESCE(title, '') || ' ' || COALESCE(content, ''))
			)`).Error; err != nil {
			global.GVA_LOG.Error("智能知识库全文索引创建失败", zap.Error(err))
		}
	}
}
