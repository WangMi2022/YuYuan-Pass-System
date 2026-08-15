package initialize

import (
	"context"
	"errors"

	"github.com/casbin/gorm-adapter/v3"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"gorm.io/gorm"
)

const legacyAIServiceAPIGroup = "AI 运营"
const previousAIServiceAPIGroup = "AI 服务管理"

type legacyAPI struct {
	path   string
	method string
}

var legacyInvoiceProviderTestAPI = legacyAPI{path: "/invoice/provider/test", method: "POST"}

func migrateAIServiceAPIRegistry(ctx context.Context) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := migrateLegacyAIServiceGroup(tx); err != nil {
			return err
		}
		return removeLegacyInvoiceProviderTestAPI(tx)
	})
}

func migrateLegacyAIServiceGroup(tx *gorm.DB) error {
	rulesByKey := make(map[string]system.SysApi, len(apiRules))
	for _, rule := range apiRules {
		rulesByKey[apiRegistryKey(rule.Path, rule.Method)] = rule
	}

	var legacyRows []system.SysApi
	if err := tx.Where("api_group IN ? AND path LIKE ?", []string{legacyAIServiceAPIGroup, previousAIServiceAPIGroup}, "/ai/%").Find(&legacyRows).Error; err != nil {
		return err
	}
	for _, legacyRow := range legacyRows {
		key := apiRegistryKey(legacyRow.Path, legacyRow.Method)
		var current system.SysApi
		findErr := tx.Where("path = ? AND method = ? AND api_group = ?", legacyRow.Path, legacyRow.Method, aiServicesMenuTitle).First(&current).Error
		if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return findErr
		}
		if findErr == nil && current.ID != legacyRow.ID {
			if err := tx.Delete(&legacyRow).Error; err != nil {
				return err
			}
			continue
		}

		updates := map[string]any{"api_group": aiServicesMenuTitle}
		if rule, exists := rulesByKey[key]; exists {
			updates["description"] = rule.Description
		}
		if err := tx.Model(&system.SysApi{}).Where("id = ?", legacyRow.ID).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

func removeLegacyInvoiceProviderTestAPI(tx *gorm.DB) error {
	if err := tx.Where("path = ? AND method = ?", legacyInvoiceProviderTestAPI.path, legacyInvoiceProviderTestAPI.method).
		Delete(&system.SysApi{}).Error; err != nil {
		return err
	}
	return tx.Where("ptype = ? AND v1 = ? AND v2 = ?", "p", legacyInvoiceProviderTestAPI.path, legacyInvoiceProviderTestAPI.method).
		Delete(&gormadapter.CasbinRule{}).Error
}

func apiRegistryKey(path, method string) string { return method + "\x00" + path }
