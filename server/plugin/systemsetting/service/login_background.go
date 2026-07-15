package service

import (
	"errors"
	"net/url"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/systemsetting/model"
	"gorm.io/gorm"
)

var LoginBackground = new(loginBackgroundService)

type loginBackgroundService struct{}

func normalizeBackground(item *model.LoginBackground) error {
	if item == nil {
		return errors.New("背景图片数据为空")
	}
	item.Name = strings.TrimSpace(item.Name)
	item.URL = strings.TrimSpace(item.URL)
	if item.Name == "" || item.URL == "" {
		return errors.New("图片名称和地址不能为空")
	}
	parsed, err := url.ParseRequestURI(item.URL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("背景图片地址不正确")
	}
	return nil
}

func (s *loginBackgroundService) Create(item *model.LoginBackground) error {
	if err := normalizeBackground(item); err != nil {
		return err
	}
	return global.GVA_DB.Create(item).Error
}

func (s *loginBackgroundService) List() ([]model.LoginBackground, error) {
	var list []model.LoginBackground
	err := global.GVA_DB.Order("is_active DESC, created_at DESC").Find(&list).Error
	return list, err
}

func (s *loginBackgroundService) Current() (model.LoginBackground, error) {
	var item model.LoginBackground
	err := global.GVA_DB.Where("is_active = ?", true).Order("updated_at DESC").First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.LoginBackground{}, nil
	}
	return item, err
}

func (s *loginBackgroundService) Activate(id uint) error {
	if id == 0 {
		return errors.New("请选择背景图片")
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var item model.LoginBackground
		if err := tx.First(&item, id).Error; err != nil {
			return errors.New("背景图片不存在")
		}
		if err := tx.Model(&model.LoginBackground{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
			return err
		}
		return tx.Model(&item).Update("is_active", true).Error
	})
}

func (s *loginBackgroundService) Delete(id uint) error {
	if id == 0 {
		return errors.New("背景图片参数不正确")
	}
	var item model.LoginBackground
	if err := global.GVA_DB.First(&item, id).Error; err != nil {
		return errors.New("背景图片不存在")
	}
	if item.IsActive {
		return errors.New("当前使用中的背景不能删除")
	}
	return global.GVA_DB.Delete(&item).Error
}
