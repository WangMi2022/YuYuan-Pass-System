package system

import (
	"context"
	"errors"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

const (
	coreAdminAuthorityID = "888"
	reloadSystemPath     = "/system/reloadSystem"
	reloadSystemMethod   = "POST"
)

var requiredCoreAdminRules = []gormadapter.CasbinRule{
	{Ptype: "p", V0: coreAdminAuthorityID, V1: reloadSystemPath, V2: reloadSystemMethod},
}

// EnsureCoreAdminPermissions repairs mandatory administrator-only permissions
// that must also be present in databases created before the route was added.
func (casbinService *CasbinService) EnsureCoreAdminPermissions(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return errors.New("database is not initialized")
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
