package initialize

import (
	"context"
	"errors"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const defaultAdminRoleID = 888

var inheritedVerificationRules = []struct {
	path, method string
}{
	{path: "/invoice/verify", method: "POST"},
	{path: "/invoice/verificationHistory", method: "GET"},
}

func Permission(ctx context.Context) {
	if err := utils.RegisterPermissions(ctx, defaultAdminRoleID, menuNames, apiRules); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			global.GVA_LOG.Warn("未找到默认管理员角色，跳过流水管理自动授权", zap.Error(err))
		} else {
			global.GVA_LOG.Error("流水管理自动授权失败", zap.Error(err))
		}
	}
	if err := migrateVerificationPermissions(ctx); err != nil {
		global.GVA_LOG.Error("发票验真权限迁移失败", zap.Error(err))
	}
}

func migrateVerificationPermissions(ctx context.Context) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var confirmRules []gormadapter.CasbinRule
		if err := tx.Where(
			"ptype = ? AND v1 = ? AND v2 = ?", "p", "/invoice/confirm", "PUT",
		).Find(&confirmRules).Error; err != nil {
			return err
		}
		authorities := make([]string, 0, len(confirmRules))
		authoritySet := make(map[string]struct{}, len(confirmRules))
		for _, rule := range confirmRules {
			if rule.V0 == "" {
				continue
			}
			if _, exists := authoritySet[rule.V0]; exists {
				continue
			}
			authoritySet[rule.V0] = struct{}{}
			authorities = append(authorities, rule.V0)
		}
		if len(authorities) == 0 {
			return nil
		}

		paths := make([]string, 0, len(inheritedVerificationRules))
		for _, rule := range inheritedVerificationRules {
			paths = append(paths, rule.path)
		}
		var existing []gormadapter.CasbinRule
		if err := tx.Where(
			"ptype = ? AND v0 IN ? AND v1 IN ?", "p", authorities, paths,
		).Find(&existing).Error; err != nil {
			return err
		}
		existingSet := make(map[string]struct{}, len(existing))
		for _, rule := range existing {
			existingSet[rule.V0+"\x00"+rule.V1+"\x00"+rule.V2] = struct{}{}
		}
		missing := make([]gormadapter.CasbinRule, 0, len(authorities)*len(inheritedVerificationRules))
		for _, authority := range authorities {
			for _, rule := range inheritedVerificationRules {
				key := authority + "\x00" + rule.path + "\x00" + rule.method
				if _, exists := existingSet[key]; exists {
					continue
				}
				existingSet[key] = struct{}{}
				missing = append(missing, gormadapter.CasbinRule{
					Ptype: "p", V0: authority, V1: rule.path, V2: rule.method,
				})
			}
		}
		if len(missing) == 0 {
			return nil
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&missing).Error
	})
}
