package service

import (
	"errors"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model"
	invoiceRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model/request"
	"gorm.io/gorm"
)

type RuleService struct{}

func normalizeRule(rule *model.ClassificationRule) error {
	rule.Name = strings.TrimSpace(rule.Name)
	rule.Field = strings.ToLower(strings.TrimSpace(rule.Field))
	rule.MatchType = strings.ToLower(strings.TrimSpace(rule.MatchType))
	rule.Pattern = strings.TrimSpace(rule.Pattern)
	if rule.Name == "" || rule.Pattern == "" || rule.CategoryID == 0 {
		return errors.New("规则名称、匹配内容和目标分类不能为空")
	}
	validFields := map[string]bool{"all": true, "seller": true, "item": true, "type": true, "raw": true}
	if !validFields[rule.Field] {
		return errors.New("匹配字段不正确")
	}
	if rule.MatchType != "contains" && rule.MatchType != "exact" {
		return errors.New("匹配方式不正确")
	}
	if rule.Weight < 1 || rule.Weight > 100 {
		return errors.New("规则分值必须在 1 到 100 之间")
	}
	var count int64
	if err := global.GVA_DB.Model(&model.InvoiceCategory{}).Where("id = ? AND enabled = ?", rule.CategoryID, true).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("目标分类不存在或已停用")
	}
	return nil
}

func (RuleService) Create(rule *model.ClassificationRule) error {
	if err := normalizeRule(rule); err != nil {
		return err
	}
	return global.GVA_DB.Create(rule).Error
}

func (RuleService) Update(rule *model.ClassificationRule) error {
	if rule.ID == 0 {
		return errors.New("缺少规则 ID")
	}
	if err := normalizeRule(rule); err != nil {
		return err
	}
	result := global.GVA_DB.Model(&model.ClassificationRule{}).Where("id = ?", rule.ID).
		Select("Name", "Field", "MatchType", "Pattern", "Weight", "Priority", "Enabled", "CategoryID").Updates(rule)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (RuleService) Delete(id uint) error {
	if id == 0 {
		return errors.New("缺少规则 ID")
	}
	return global.GVA_DB.Delete(&model.ClassificationRule{}, id).Error
}

func (RuleService) List(search invoiceRequest.RuleSearch) ([]model.ClassificationRule, int64, error) {
	var list []model.ClassificationRule
	var total int64
	db := global.GVA_DB.Model(&model.ClassificationRule{})
	if keyword := strings.TrimSpace(search.Keyword); keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		db = db.Where("LOWER(name) LIKE ? OR LOWER(pattern) LIKE ?", like, like)
	}
	if search.CategoryID > 0 {
		db = db.Where("category_id = ?", search.CategoryID)
	}
	if search.Enabled != nil {
		db = db.Where("enabled = ?", *search.Enabled)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Preload("Category").Order("priority DESC, id ASC").Scopes(search.Paginate()).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
