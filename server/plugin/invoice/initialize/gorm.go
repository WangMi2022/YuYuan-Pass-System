package initialize

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model"
	"go.uber.org/zap"
)

func Gorm(ctx context.Context) {
	if err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		&model.InvoiceCategory{},
		&model.Invoice{},
		&model.InvoiceVerification{},
		&model.InvoiceItem{},
		&model.ClassificationRule{},
		&model.RecognitionJob{},
		&model.InvoiceReviewCorrection{},
		&model.InvoiceFileCleanupJob{},
	); err != nil {
		global.GVA_LOG.Error("流水管理数据表迁移失败", zap.Error(err))
		return
	}
	if err := backfillInvoiceStorageLocations(ctx); err != nil {
		global.GVA_LOG.Error("发票存储位置迁移失败", zap.Error(err))
		return
	}
	seedInvoiceCategories(ctx)
	seedClassificationRules(ctx)
}

func backfillInvoiceStorageLocations(ctx context.Context) error {
	localRoot := strings.TrimSpace(global.GVA_CONFIG.Local.StorePath)
	if localRoot != "" {
		absoluteRoot, err := filepath.Abs(filepath.Clean(localRoot))
		if err != nil {
			return err
		}
		if err = global.GVA_DB.WithContext(ctx).Unscoped().Model(&model.Invoice{}).
			Where("storage_type = ? AND COALESCE(storage_root, '') = ''", "local").
			Update("storage_root", absoluteRoot).Error; err != nil {
			return err
		}
	}

	endpoint := strings.TrimSpace(global.GVA_CONFIG.Minio.Endpoint)
	bucket := strings.TrimSpace(global.GVA_CONFIG.Minio.BucketName)
	if endpoint == "" || bucket == "" {
		return nil
	}
	return global.GVA_DB.WithContext(ctx).Unscoped().Model(&model.Invoice{}).Where(
		"storage_type = ? AND (COALESCE(storage_endpoint, '') = '' OR COALESCE(storage_bucket, '') = '')",
		"minio",
	).Updates(map[string]any{
		"storage_endpoint": endpoint,
		"storage_bucket":   bucket,
		"storage_use_ssl":  global.GVA_CONFIG.Minio.UseSSL,
	}).Error
}

func seedInvoiceCategories(ctx context.Context) {
	categories := []model.InvoiceCategory{
		{Name: "办公采购", Code: "OFFICE", Description: "办公用品、耗材及日常行政采购", Color: "#2563EB", Sort: 10, Enabled: true},
		{Name: "固定资产", Code: "FIXED-ASSET", Description: "设备、家具及其他可资本化资产", Color: "#7C3AED", Sort: 20, Enabled: true},
		{Name: "差旅交通", Code: "TRAVEL", Description: "机票、铁路、住宿、打车及差旅相关支出", Color: "#0891B2", Sort: 30, Enabled: true},
		{Name: "软件订阅", Code: "SOFTWARE", Description: "软件许可、云服务及信息化订阅", Color: "#4F46E5", Sort: 40, Enabled: true},
		{Name: "维修服务", Code: "MAINTENANCE", Description: "设备维修、保养及技术服务", Color: "#D97706", Sort: 50, Enabled: true},
		{Name: "物流运输", Code: "LOGISTICS", Description: "快递、货运及物流配送", Color: "#059669", Sort: 60, Enabled: true},
		{Name: "餐饮招待", Code: "CATERING", Description: "工作餐、会议餐及业务招待", Color: "#DC2626", Sort: 70, Enabled: true},
		{Name: "其他支出", Code: "OTHER", Description: "暂未归入其他分类的流水", Color: "#64748B", Sort: 99, Enabled: true},
	}
	for index := range categories {
		item := categories[index]
		if err := global.GVA_DB.WithContext(ctx).Where("code = ?", item.Code).FirstOrCreate(&item).Error; err != nil {
			global.GVA_LOG.Warn("预置发票分类失败", zap.Error(err), zap.String("code", item.Code))
		}
	}
}

func seedClassificationRules(ctx context.Context) {
	var categories []model.InvoiceCategory
	if err := global.GVA_DB.WithContext(ctx).Find(&categories).Error; err != nil {
		return
	}
	categoryByCode := make(map[string]uint, len(categories))
	for _, category := range categories {
		categoryByCode[category.Code] = category.ID
	}
	type seedRule struct {
		Name, Field, MatchType, Pattern, CategoryCode string
		Weight, Priority                              int
	}
	rules := []seedRule{
		{Name: "办公用品关键词", Field: "item", MatchType: "contains", Pattern: "办公", Weight: 75, Priority: 50, CategoryCode: "OFFICE"},
		{Name: "电脑设备关键词", Field: "item", MatchType: "contains", Pattern: "电脑", Weight: 85, Priority: 80, CategoryCode: "FIXED-ASSET"},
		{Name: "住宿服务关键词", Field: "item", MatchType: "contains", Pattern: "住宿", Weight: 80, Priority: 70, CategoryCode: "TRAVEL"},
		{Name: "客运服务关键词", Field: "raw", MatchType: "contains", Pattern: "旅客运输", Weight: 80, Priority: 70, CategoryCode: "TRAVEL"},
		{Name: "软件订阅关键词", Field: "all", MatchType: "contains", Pattern: "软件", Weight: 75, Priority: 60, CategoryCode: "SOFTWARE"},
		{Name: "云服务关键词", Field: "all", MatchType: "contains", Pattern: "云服务", Weight: 80, Priority: 70, CategoryCode: "SOFTWARE"},
		{Name: "维修服务关键词", Field: "all", MatchType: "contains", Pattern: "维修", Weight: 80, Priority: 70, CategoryCode: "MAINTENANCE"},
		{Name: "快递物流关键词", Field: "all", MatchType: "contains", Pattern: "快递", Weight: 75, Priority: 60, CategoryCode: "LOGISTICS"},
		{Name: "餐饮服务关键词", Field: "all", MatchType: "contains", Pattern: "餐饮", Weight: 75, Priority: 60, CategoryCode: "CATERING"},
	}
	for _, seed := range rules {
		categoryID := categoryByCode[seed.CategoryCode]
		if categoryID == 0 {
			continue
		}
		rule := model.ClassificationRule{
			Name: seed.Name, Field: seed.Field, MatchType: seed.MatchType, Pattern: seed.Pattern,
			Weight: seed.Weight, Priority: seed.Priority, Enabled: true, CategoryID: categoryID,
		}
		if err := global.GVA_DB.WithContext(ctx).Where("name = ?", rule.Name).FirstOrCreate(&rule).Error; err != nil {
			global.GVA_LOG.Warn("预置发票分类规则失败", zap.Error(err), zap.String("name", rule.Name))
		}
	}
}
