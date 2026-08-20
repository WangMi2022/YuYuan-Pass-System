package service

import (
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/systemsetting/model"
	exampleService "github.com/WangMi2022/mit-assets-admin/server/service/example"
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
		subtitle := model.DefaultSystemSubtitle
		return model.LoginLogo{
			SystemName: model.DefaultSystemName,
			Subtitle:   &subtitle,
		}, nil
	}
	if err == nil {
		item.SystemName = strings.TrimSpace(item.SystemName)
		if item.SystemName == "" {
			item.SystemName = model.DefaultSystemName
		}
		if item.Subtitle == nil {
			subtitle := model.DefaultSystemSubtitle
			item.Subtitle = &subtitle
		} else {
			subtitle := strings.TrimSpace(*item.Subtitle)
			item.Subtitle = &subtitle
		}
		if item.URL != "" {
			previewURL, previewErr := exampleService.ResolveConfiguredMediaPreviewURL(item.URL)
			if previewErr != nil {
				return model.LoginLogo{}, previewErr
			}
			item.URL = previewURL
		}
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

func (s *loginLogoService) SaveBranding(systemName, subtitle string, updatedBy uint) error {
	systemName = strings.TrimSpace(systemName)
	subtitle = strings.TrimSpace(subtitle)
	if systemName == "" {
		return errors.New("系统名称不能为空")
	}
	if utf8.RuneCountInString(systemName) > 80 {
		return errors.New("系统名称不能超过80个字符")
	}
	if utf8.RuneCountInString(subtitle) > 120 {
		return errors.New("品牌副标题不能超过120个字符")
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var current model.LoginLogo
		err := tx.Order("updated_at DESC").First(&current).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&model.LoginLogo{
				SystemName: systemName,
				Subtitle:   &subtitle,
				UpdatedBy:  updatedBy,
			}).Error
		}
		if err != nil {
			return err
		}
		return tx.Model(&current).Updates(map[string]any{
			"system_name": systemName,
			"subtitle":    subtitle,
			"updated_by":  updatedBy,
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
	return global.GVA_DB.Model(&item).Updates(map[string]any{
		"name": "",
		"url":  "",
	}).Error
}
