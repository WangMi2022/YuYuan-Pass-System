package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
	"go.uber.org/zap"
)

func Gorm(ctx context.Context) {
	if err := global.GVA_DB.WithContext(ctx).AutoMigrate(&model.Category{}, &model.Asset{}); err != nil {
		global.GVA_LOG.Error("资产模块数据表迁移失败", zap.Error(err))
		return
	}
	seedCategories(ctx)
}

func seedCategories(ctx context.Context) {
	var count int64
	if err := global.GVA_DB.WithContext(ctx).Model(&model.Category{}).Count(&count).Error; err != nil || count > 0 {
		return
	}
	categories := []model.Category{
		{Name: "座椅/板凳", Code: "FURN-CHAIR", Color: "#0F766E", Sort: 10, Enabled: true, Description: "办公椅、会议椅、板凳等座具"},
		{Name: "桌类家具", Code: "FURN-DESK", Color: "#7C3AED", Sort: 20, Enabled: true, Description: "办公桌、会议桌、工作台等"},
		{Name: "电脑整机", Code: "IT-COMPUTER", Color: "#2563EB", Sort: 30, Enabled: true, Description: "台式机、笔记本、工作站等"},
		{Name: "显示设备", Code: "IT-DISPLAY", Color: "#0891B2", Sort: 40, Enabled: true, Description: "显示器、电视、投影仪及大屏"},
		{Name: "网络设备", Code: "IT-NETWORK", Color: "#4F46E5", Sort: 50, Enabled: true, Description: "交换机、路由器、防火墙及无线设备"},
		{Name: "办公设备", Code: "OFFICE-EQUIP", Color: "#D97706", Sort: 60, Enabled: true, Description: "打印机、扫描仪、碎纸机等"},
		{Name: "生产设备", Code: "PROD-EQUIP", Color: "#DC2626", Sort: 70, Enabled: true, Description: "生产、检测、维修相关设备"},
		{Name: "其他资产", Code: "OTHER", Color: "#475569", Sort: 99, Enabled: true, Description: "暂未归入其他分类的资产"},
	}
	if err := global.GVA_DB.WithContext(ctx).Create(&categories).Error; err != nil {
		global.GVA_LOG.Warn("预置资产分类失败", zap.Error(err))
	}
}
