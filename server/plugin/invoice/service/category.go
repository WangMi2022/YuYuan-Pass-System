package service

import (
	"errors"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model"
	invoiceRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model/request"
	"gorm.io/gorm"
)

type CategoryService struct{}

func normalizeCategory(category *model.InvoiceCategory) error {
	category.Name = strings.TrimSpace(category.Name)
	category.Code = strings.ToUpper(strings.TrimSpace(category.Code))
	category.Description = strings.TrimSpace(category.Description)
	if category.Name == "" || category.Code == "" {
		return errors.New("分类名称和编码不能为空")
	}
	if category.Color == "" {
		category.Color = "#6d5dfb"
	}
	return nil
}

func (CategoryService) Create(category *model.InvoiceCategory) error {
	if err := normalizeCategory(category); err != nil {
		return err
	}
	if err := global.GVA_DB.Create(category).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return errors.New("分类名称或编码已存在")
		}
		return err
	}
	return nil
}

func (CategoryService) Update(category *model.InvoiceCategory) error {
	if category.ID == 0 {
		return errors.New("缺少分类 ID")
	}
	if err := normalizeCategory(category); err != nil {
		return err
	}
	result := global.GVA_DB.Model(&model.InvoiceCategory{}).Where("id = ?", category.ID).
		Select("Name", "Code", "Description", "Color", "Sort", "Enabled").Updates(category)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (CategoryService) Delete(id uint) error {
	if id == 0 {
		return errors.New("缺少分类 ID")
	}
	var count int64
	if err := global.GVA_DB.Model(&model.Invoice{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该分类已有发票，不能删除")
	}
	if err := global.GVA_DB.Model(&model.ClassificationRule{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该分类已有匹配规则，不能删除")
	}
	return global.GVA_DB.Delete(&model.InvoiceCategory{}, id).Error
}

func (CategoryService) List(search invoiceRequest.CategorySearch) ([]model.InvoiceCategory, int64, error) {
	var list []model.InvoiceCategory
	var total int64
	db := global.GVA_DB.Model(&model.InvoiceCategory{})
	if keyword := strings.TrimSpace(search.Keyword); keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		db = db.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", like, like)
	}
	if search.Enabled != nil {
		db = db.Where("enabled = ?", *search.Enabled)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("sort ASC, id ASC").Scopes(search.Paginate()).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (CategoryService) Options() ([]model.InvoiceCategory, error) {
	var list []model.InvoiceCategory
	err := global.GVA_DB.Where("enabled = ?", true).Order("sort ASC, id ASC").Limit(100).Find(&list).Error
	return list, err
}
