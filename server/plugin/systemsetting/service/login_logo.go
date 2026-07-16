package service

import (
	"errors"
	"net/url"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/systemsetting/model"
	"gorm.io/gorm"
)

var LoginLogo = new(loginLogoService)

type loginLogoService struct{}

func normalizeLogo(item *model.LoginLogo) error {
	if item == nil {
		return errors.New("登录图标数据为空")
	}
	item.Name = strings.TrimSpace(item.Name)
	item.URL = strings.TrimSpace(item.URL)
	if item.Name == "" || item.URL == "" {
		return errors.New("图标名称和地址不能为空")
	}
	parsed, err := url.ParseRequestURI(item.URL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("登录图标地址不正确")
	}
	return nil
}

func (s *loginLogoService) Current() (model.LoginLogo, error) {
	var item model.LoginLogo
	err := global.GVA_DB.Order("updated_at DESC").First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.LoginLogo{}, nil
	}
	return item, err
}

func (s *loginLogoService) Save(item *model.LoginLogo) error {
	if err := normalizeLogo(item); err != nil {
		return err
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var current model.LoginLogo
		err := tx.Order("updated_at DESC").First(&current).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(item).Error
		}
		if err != nil {
			return err
		}
		return tx.Model(&current).Updates(map[string]any{
			"name":       item.Name,
			"url":        item.URL,
			"updated_by": item.UpdatedBy,
		}).Error
	})
}

func (s *loginLogoService) Reset() error {
	var item model.LoginLogo
	err := global.GVA_DB.Order("updated_at DESC").First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return global.GVA_DB.Delete(&item).Error
}
