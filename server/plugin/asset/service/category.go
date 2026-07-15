package service

import (
	"errors"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
	assetRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model/request"
	assetResponse "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model/response"
	"gorm.io/gorm"
)

var Category = new(categoryService)

type categoryService struct{}

func normalizeCategory(category *model.Category) error {
	category.Name = strings.TrimSpace(category.Name)
	category.Code = strings.ToUpper(strings.TrimSpace(category.Code))
	if category.Name == "" || category.Code == "" {
		return errors.New("分类名称和分类编码不能为空")
	}
	if category.Color == "" {
		category.Color = "#334155"
	}
	return nil
}

func (s *categoryService) Create(category *model.Category) error {
	if err := normalizeCategory(category); err != nil {
		return err
	}
	if err := global.GVA_DB.Create(category).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return errors.New("分类名称或编码已存在")
		}
		return err
	}
	return nil
}

func (s *categoryService) Update(category *model.Category) error {
	if category.ID == 0 {
		return errors.New("缺少分类 ID")
	}
	if err := normalizeCategory(category); err != nil {
		return err
	}
	result := global.GVA_DB.Model(&model.Category{}).Where("id = ?", category.ID).
		Select("Name", "Code", "Description", "Color", "Sort", "Enabled").Updates(category)
	if result.Error != nil {
		if strings.Contains(strings.ToLower(result.Error.Error()), "duplicate") {
			return errors.New("分类名称或编码已存在")
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *categoryService) Delete(id uint) error {
	if id == 0 {
		return errors.New("缺少分类 ID")
	}
	var count int64
	if err := global.GVA_DB.Model(&model.Asset{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该分类下仍有资产，不能删除")
	}
	return global.GVA_DB.Delete(&model.Category{}, id).Error
}

func (s *categoryService) List(search assetRequest.CategorySearch) ([]assetResponse.CategoryItem, int64, error) {
	var list []assetResponse.CategoryItem
	var total int64
	db := global.GVA_DB.Model(&model.Category{})
	if keyword := strings.TrimSpace(search.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("name ILIKE ? OR code ILIKE ?", like, like)
	}
	if search.Enabled != nil {
		db = db.Where("enabled = ?", *search.Enabled)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query := db.Select(`asset_categories.*,
		COUNT(assets.id) AS asset_kinds,
		COALESCE(SUM(assets.quantity), 0) AS quantity`).
		Joins("LEFT JOIN assets ON assets.category_id = asset_categories.id AND assets.deleted_at IS NULL").
		Group("asset_categories.id").Order("asset_categories.sort ASC, asset_categories.id ASC")
	if search.PageSize > 0 {
		query = query.Scopes(search.Paginate())
	}
	if err := query.Scan(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *categoryService) Options() ([]model.Category, error) {
	var list []model.Category
	err := global.GVA_DB.Where("enabled = ?", true).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}
