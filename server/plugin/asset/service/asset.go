package service

import (
	"errors"
	"math"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
	assetRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model/request"
	assetResponse "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model/response"
	"gorm.io/gorm"
)

var Asset = new(assetService)

type assetService struct{}

var allowedStatuses = map[string]struct{}{
	"in_use": {}, "idle": {}, "maintenance": {}, "retired": {},
}

func money(value float64) float64 { return math.Round(value*100) / 100 }

func prepareAsset(asset *model.Asset, creating bool) error {
	asset.AssetCode = strings.ToUpper(strings.TrimSpace(asset.AssetCode))
	asset.Name = strings.TrimSpace(asset.Name)
	asset.Unit = strings.TrimSpace(asset.Unit)
	if asset.AssetCode == "" || asset.Name == "" {
		return errors.New("资产编号和资产名称不能为空")
	}
	if asset.CategoryID == 0 {
		return errors.New("请选择资产分类")
	}
	if asset.Quantity <= 0 {
		return errors.New("资产数量必须大于 0")
	}
	if asset.UnitPrice < 0 || asset.CurrentValue < 0 {
		return errors.New("金额不能为负数")
	}
	if asset.Unit == "" {
		asset.Unit = "件"
	}
	if asset.Status == "" {
		asset.Status = "in_use"
	}
	if _, ok := allowedStatuses[asset.Status]; !ok {
		return errors.New("资产状态不合法")
	}
	asset.OriginalValue = money(float64(asset.Quantity) * asset.UnitPrice)
	asset.UnitPrice = money(asset.UnitPrice)
	asset.CurrentValue = money(asset.CurrentValue)
	if creating && asset.CurrentValue == 0 && asset.OriginalValue > 0 && asset.Status != "retired" {
		asset.CurrentValue = asset.OriginalValue
	}
	if asset.CurrentValue > asset.OriginalValue && asset.OriginalValue > 0 {
		return errors.New("当前估值不能高于资产原值")
	}
	return nil
}

func (s *assetService) Create(asset *model.Asset) error {
	if err := prepareAsset(asset, true); err != nil {
		return err
	}
	var count int64
	if err := global.GVA_DB.Model(&model.Category{}).Where("id = ? AND enabled = ?", asset.CategoryID, true).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("资产分类不存在或已停用")
	}
	if err := global.GVA_DB.Create(asset).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return errors.New("资产编号已存在")
		}
		return err
	}
	return nil
}

func (s *assetService) Update(asset *model.Asset) error {
	if asset.ID == 0 {
		return errors.New("缺少资产 ID")
	}
	if err := prepareAsset(asset, false); err != nil {
		return err
	}
	fields := []string{
		"AssetCode", "Name", "CategoryID", "Brand", "Model", "SerialNumber",
		"Quantity", "Unit", "UnitPrice", "OriginalValue", "CurrentValue", "Status",
		"Location", "Custodian", "Supplier", "PurchaseDate", "WarrantyEndDate", "Photos", "Remarks",
	}
	result := global.GVA_DB.Model(&model.Asset{}).Where("id = ?", asset.ID).Select(fields).Updates(asset)
	if result.Error != nil {
		if strings.Contains(strings.ToLower(result.Error.Error()), "duplicate") {
			return errors.New("资产编号已存在")
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *assetService) Delete(id uint) error {
	if id == 0 {
		return errors.New("缺少资产 ID")
	}
	result := global.GVA_DB.Delete(&model.Asset{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *assetService) Get(id uint) (model.Asset, error) {
	var asset model.Asset
	err := global.GVA_DB.Preload("Category").First(&asset, id).Error
	return asset, err
}

func (s *assetService) List(search assetRequest.AssetSearch) ([]model.Asset, int64, error) {
	var list []model.Asset
	var total int64
	db := global.GVA_DB.Model(&model.Asset{})
	if keyword := strings.TrimSpace(search.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("asset_code ILIKE ? OR name ILIKE ? OR brand ILIKE ? OR model ILIKE ? OR serial_number ILIKE ? OR custodian ILIKE ?", like, like, like, like, like, like)
	}
	if search.CategoryID > 0 {
		db = db.Where("category_id = ?", search.CategoryID)
	}
	if search.Status != "" {
		db = db.Where("status = ?", search.Status)
	}
	if location := strings.TrimSpace(search.Location); location != "" {
		db = db.Where("location ILIKE ?", "%"+location+"%")
	}
	if search.MinValue > 0 {
		db = db.Where("current_value >= ?", search.MinValue)
	}
	if search.MaxValue > 0 {
		db = db.Where("current_value <= ?", search.MaxValue)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Preload("Category").Order("created_at DESC").Scopes(search.Paginate()).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *assetService) Dashboard() (assetResponse.Dashboard, error) {
	var result assetResponse.Dashboard
	var totals struct {
		AssetKinds    int64
		TotalQuantity int64
		OriginalValue float64
		CurrentValue  float64
	}
	if err := global.GVA_DB.Model(&model.Asset{}).
		Select("COUNT(*) AS asset_kinds, COALESCE(SUM(quantity), 0) AS total_quantity, COALESCE(SUM(original_value), 0) AS original_value, COALESCE(SUM(current_value), 0) AS current_value").
		Scan(&totals).Error; err != nil {
		return result, err
	}
	result.AssetKinds = totals.AssetKinds
	result.TotalQuantity = totals.TotalQuantity
	result.OriginalValue = money(totals.OriginalValue)
	result.CurrentValue = money(totals.CurrentValue)
	result.Depreciation = money(result.OriginalValue - result.CurrentValue)
	if err := global.GVA_DB.Model(&model.Category{}).Count(&result.CategoryCount).Error; err != nil {
		return result, err
	}

	categorySQL := `
		SELECT c.id AS category_id, c.name AS category_name, c.color,
		       COUNT(a.id) AS asset_kinds,
		       COALESCE(SUM(a.quantity), 0) AS quantity,
		       COALESCE(SUM(a.original_value), 0) AS original,
		       COALESCE(SUM(a.current_value), 0) AS current
		FROM asset_categories c
		LEFT JOIN assets a ON a.category_id = c.id AND a.deleted_at IS NULL
		WHERE c.deleted_at IS NULL AND c.enabled = true
		GROUP BY c.id, c.name, c.color, c.sort
		ORDER BY c.sort ASC, quantity DESC, c.id ASC`
	if err := global.GVA_DB.Raw(categorySQL).Scan(&result.CategorySummary).Error; err != nil {
		return result, err
	}

	if err := global.GVA_DB.Model(&model.Asset{}).
		Select("status, COALESCE(SUM(quantity), 0) AS quantity").
		Group("status").Order("quantity DESC").Scan(&result.StatusSummary).Error; err != nil {
		return result, err
	}
	if err := global.GVA_DB.Model(&model.Asset{}).
		Select("location, COALESCE(SUM(quantity), 0) AS quantity, COALESCE(SUM(current_value), 0) AS value").
		Where("location IS NOT NULL AND TRIM(location) <> ''").Group("location").Order("quantity DESC").Limit(8).
		Scan(&result.LocationSummary).Error; err != nil {
		return result, err
	}
	if err := global.GVA_DB.Preload("Category").Order("created_at DESC").Limit(8).Find(&result.RecentAssets).Error; err != nil {
		return result, err
	}
	return result, nil
}
