package system

import (
	"context"
	"errors"
	"strconv"

	sysModel "github.com/WangMi2022/mit-assets-admin/server/model/system"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

const (
	coreAdminAuthorityID = "888"
	reloadSystemPath     = "/system/reloadSystem"
	reloadSystemMethod   = "POST"
	swaggerPath          = "/swagger/*any"
	freshCasbinPath      = "/api/freshCasbin"
)

var requiredCoreAdminRules = []gormadapter.CasbinRule{
	{Ptype: "p", V0: coreAdminAuthorityID, V1: reloadSystemPath, V2: reloadSystemMethod},
	{Ptype: "p", V0: coreAdminAuthorityID, V1: swaggerPath, V2: "GET"},
	{Ptype: "p", V0: coreAdminAuthorityID, V1: freshCasbinPath, V2: "GET"},
	{Ptype: "p", V0: coreAdminAuthorityID, V1: "/autoCode/llmAuto", V2: "POST"},
	{Ptype: "p", V0: coreAdminAuthorityID, V1: "/autoCode/llmAutoSSE", V2: "POST"},
	{Ptype: "p", V0: coreAdminAuthorityID, V1: "/autoCode/initMenu", V2: "POST"},
	{Ptype: "p", V0: coreAdminAuthorityID, V1: "/autoCode/initAPI", V2: "POST"},
	{Ptype: "p", V0: coreAdminAuthorityID, V1: "/autoCode/initDictionary", V2: "POST"},
}

type requiredContactVerificationPermission struct {
	path        string
	method      string
	description string
}

var requiredContactVerificationPermissions = []requiredContactVerificationPermission{
	{path: "/user/contactVerificationCapabilities", method: "GET", description: "获取联系方式验证码能力"},
	{path: "/user/sendContactVerificationCode", method: "POST", description: "发送联系方式验证码"},
	{path: "/user/updateSelfContact", method: "PUT", description: "验证并更新自身联系方式"},
}

// EnsureCoreAdminPermissions repairs mandatory administrator-only permissions
// that must also be present in databases created before the route was added.
func (casbinService *CasbinService) EnsureCoreAdminPermissions(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return errors.New("database is not initialized")
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where(
			"(path = ? AND method = ?) OR (path = ? AND method = ?) OR (path = ? AND method = ?) OR (path = ? AND method = ?)",
			"/autoCode/llmAuto", "POST",
			"/autoCode/llmAutoSSE", "POST",
			swaggerPath, "GET",
			freshCasbinPath, "GET",
		).Delete(&sysModel.SysIgnoreApi{}).Error; err != nil {
			return err
		}
		for i := range requiredCoreAdminRules {
			rule := requiredCoreAdminRules[i]
			if err := tx.Where(
				"ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?",
				rule.Ptype, rule.V0, rule.V1, rule.V2,
			).FirstOrCreate(&rule).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// EnsureContactVerificationPermissions repairs the self-service API resources
// and grants them to every active authority without removing existing custom
// permissions. These routes are authenticated user capabilities, not admin
// management APIs, so old databases need an idempotent data upgrade.
func (casbinService *CasbinService) EnsureContactVerificationPermissions(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return errors.New("database is not initialized")
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, permission := range requiredContactVerificationPermissions {
			var api sysModel.SysApi
			result := tx.Where("path = ? AND method = ?", permission.path, permission.method).Limit(1).Find(&api)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected > 0 {
				continue
			}

			api = sysModel.SysApi{}
			result = tx.Unscoped().Where("path = ? AND method = ?", permission.path, permission.method).Limit(1).Find(&api)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected > 0 {
				if err := tx.Unscoped().Model(&api).UpdateColumn("deleted_at", nil).Error; err != nil {
					return err
				}
			} else {
				api = sysModel.SysApi{
					ApiGroup:    "系统用户",
					Method:      permission.method,
					Path:        permission.path,
					Description: permission.description,
				}
				if err := tx.Create(&api).Error; err != nil {
					return err
				}
			}
		}

		var authorities []sysModel.SysAuthority
		if err := tx.Find(&authorities).Error; err != nil {
			return err
		}
		for _, authority := range authorities {
			for _, permission := range requiredContactVerificationPermissions {
				rule := gormadapter.CasbinRule{
					Ptype: "p",
					V0:    strconv.Itoa(int(authority.AuthorityId)),
					V1:    permission.path,
					V2:    permission.method,
				}
				if err := tx.Where(
					"ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?",
					rule.Ptype, rule.V0, rule.V1, rule.V2,
				).FirstOrCreate(&rule).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
