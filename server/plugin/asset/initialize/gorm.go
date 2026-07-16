package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
	"go.uber.org/zap"
)

func Gorm(ctx context.Context) {
	if err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		&model.Category{},
		&model.Location{},
		&model.Asset{},
		&model.AssetOperationOrder{},
		&model.AssetOperationItem{},
		&model.AssetOperationRecord{},
	); err != nil {
		global.GVA_LOG.Error("资产模块数据表迁移失败", zap.Error(err))
		return
	}
	seedCategories(ctx)
	seedLocations(ctx)
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

func seedLocations(ctx context.Context) {
	locations := []model.Location{
		{Name: "资产仓库 A 区", Type: model.LocationTypeInbound, Code: "RK-A", Sort: 10, Enabled: true, Description: "常规资产验收入库区"},
		{Name: "资产仓库 B 区", Type: model.LocationTypeInbound, Code: "RK-B", Sort: 20, Enabled: true, Description: "大件及批量资产入库区"},
		{Name: "信息设备备件库", Type: model.LocationTypeInbound, Code: "RK-IT", Sort: 30, Enabled: true, Description: "信息设备及配件入库区"},
		{Name: "研发中心 6F", Type: model.LocationTypeUsage, Code: "SY-RD6", Sort: 10, Enabled: true, Description: "研发部门资产使用区"},
		{Name: "一号办公楼 4F", Type: model.LocationTypeUsage, Code: "SY-OFFICE1-4", Sort: 20, Enabled: true, Description: "行政办公资产使用区"},
		{Name: "生产车间 A 区", Type: model.LocationTypeUsage, Code: "SY-PLANT-A", Sort: 30, Enabled: true, Description: "生产设备使用区"},
		{Name: "二号办公楼 3F", Type: model.LocationTypeTransfer, Code: "DR-OFFICE2-3", Sort: 10, Enabled: true, Description: "跨办公区调拨目标位置"},
		{Name: "容灾机房", Type: model.LocationTypeTransfer, Code: "DR-DR", Sort: 20, Enabled: true, Description: "信息设备调拨目标位置"},
		{Name: "设计中心", Type: model.LocationTypeTransfer, Code: "DR-DESIGN", Sort: 30, Enabled: true, Description: "设计类资产调拨目标位置"},
		{Name: "资产仓库 A 区", Type: model.LocationTypeReturn, Code: "GH-A", Sort: 10, Enabled: true, Description: "常规资产归还接收区"},
		{Name: "行政物资库", Type: model.LocationTypeReturn, Code: "GH-OFFICE", Sort: 20, Enabled: true, Description: "办公物资归还接收区"},
		{Name: "设备维修中心", Type: model.LocationTypeMaintenance, Code: "WX-CENTER", Sort: 10, Enabled: true, Description: "内部设备维修位置"},
		{Name: "外送维修待验区", Type: model.LocationTypeMaintenance, Code: "WX-OUT", Sort: 20, Enabled: true, Description: "外送维修后待验收位置"},
		{Name: "信息设备检修间", Type: model.LocationTypeMaintenance, Code: "WX-IT", Sort: 30, Enabled: true, Description: "信息设备检修位置"},
		{Name: "报废暂存区", Type: model.LocationTypeDisposal, Code: "CZ-SCRAP", Sort: 10, Enabled: true, Description: "已批准报废资产暂存位置"},
		{Name: "待拍卖区", Type: model.LocationTypeDisposal, Code: "CZ-AUCTION", Sort: 20, Enabled: true, Description: "待公开处置资产存放位置"},
		{Name: "环保回收交接区", Type: model.LocationTypeDisposal, Code: "CZ-RECYCLE", Sort: 30, Enabled: true, Description: "环保回收处置交接位置"},
	}
	for index := range locations {
		item := locations[index]
		if err := global.GVA_DB.WithContext(ctx).Where("type = ? AND name = ?", item.Type, item.Name).
			FirstOrCreate(&item).Error; err != nil {
			global.GVA_LOG.Warn("预置资产位置失败", zap.Error(err))
		}
	}
}
