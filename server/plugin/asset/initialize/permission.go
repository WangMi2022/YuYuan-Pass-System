package initialize

import (
	"context"
	"errors"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/plugin-tool/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const defaultAdminAuthorityID = 888

// Permission 首次安装时仅把资产模块授权给默认管理员，其他角色可在角色管理中按需分配。
func Permission(ctx context.Context) {
	if err := utils.RegisterPermissions(ctx, defaultAdminAuthorityID, menuNames, apiRules); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			global.GVA_LOG.Warn("未找到默认管理员角色，跳过资产模块自动授权", zap.Error(err))
		} else {
			global.GVA_LOG.Error("资产模块自动授权失败", zap.Error(err))
		}
	}
	if err := migrateAssetRecognitionPermissions(ctx); err != nil {
		global.GVA_LOG.Error("资产智能建档权限迁移失败", zap.Error(err))
	}
}

type assetRecognitionPermissionRule struct {
	path   string
	method string
}

var assetRecognitionReadRules = []assetRecognitionPermissionRule{
	{path: "/assetRecognition/list", method: "GET"},
	{path: "/assetRecognition/detail", method: "GET"},
}

var assetRecognitionWriteRules = []assetRecognitionPermissionRule{
	{path: "/assetRecognition/create", method: "POST"},
	{path: "/assetRecognition/retry", method: "POST"},
	{path: "/assetRecognition/draft", method: "PUT"},
	{path: "/assetRecognition/confirm", method: "POST"},
	{path: "/assetRecognition/delete", method: "DELETE"},
}

func migrateAssetRecognitionPermissions(ctx context.Context) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var readSources []gormadapter.CasbinRule
		if err := tx.Where("ptype = ? AND v1 = ? AND v2 = ?", "p", "/asset/list", "GET").Find(&readSources).Error; err != nil {
			return err
		}
		if err := inheritAssetRecognitionRules(tx, assetRecognitionAuthorityIDs(readSources), assetRecognitionReadRules); err != nil {
			return err
		}
		var writeSources []gormadapter.CasbinRule
		if err := tx.Where("ptype = ? AND v1 = ? AND v2 = ?", "p", "/asset/create", "POST").Find(&writeSources).Error; err != nil {
			return err
		}
		return inheritAssetRecognitionRules(tx, assetRecognitionAuthorityIDs(writeSources), assetRecognitionWriteRules)
	})
}

func assetRecognitionAuthorityIDs(rules []gormadapter.CasbinRule) []string {
	seen := map[string]struct{}{}
	authorities := make([]string, 0, len(rules))
	for _, rule := range rules {
		if rule.V0 == "" {
			continue
		}
		if _, exists := seen[rule.V0]; exists {
			continue
		}
		seen[rule.V0] = struct{}{}
		authorities = append(authorities, rule.V0)
	}
	return authorities
}

func inheritAssetRecognitionRules(tx *gorm.DB, authorities []string, rules []assetRecognitionPermissionRule) error {
	if len(authorities) == 0 || len(rules) == 0 {
		return nil
	}
	paths := make([]string, 0, len(rules))
	for _, rule := range rules {
		paths = append(paths, rule.path)
	}
	var existing []gormadapter.CasbinRule
	if err := tx.Where("ptype = ? AND v0 IN ? AND v1 IN ?", "p", authorities, paths).Find(&existing).Error; err != nil {
		return err
	}
	existingSet := make(map[string]struct{}, len(existing))
	for _, rule := range existing {
		existingSet[rule.V0+"\x00"+rule.V1+"\x00"+rule.V2] = struct{}{}
	}
	missing := make([]gormadapter.CasbinRule, 0, len(authorities)*len(rules))
	for _, authority := range authorities {
		for _, rule := range rules {
			key := authority + "\x00" + rule.path + "\x00" + rule.method
			if _, exists := existingSet[key]; exists {
				continue
			}
			existingSet[key] = struct{}{}
			missing = append(missing, gormadapter.CasbinRule{Ptype: "p", V0: authority, V1: rule.path, V2: rule.method})
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return tx.Create(&missing).Error
}
