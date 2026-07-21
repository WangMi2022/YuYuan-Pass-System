package utils

import (
	"context"
	"strconv"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"gorm.io/gorm"
)

type permissionRuleIdentity struct {
	path, method string
}

// RegisterPermissions idempotently grants a plugin's menus and APIs to one
// authority using a fixed number of database round trips.
func RegisterPermissions(ctx context.Context, authorityID uint, menuNames []string, apis []system.SysApi) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return registerPermissions(tx, authorityID, menuNames, apis)
	})
}

func registerPermissions(db *gorm.DB, authorityID uint, menuNames []string, apis []system.SysApi) error {
	if err := db.Select("authority_id").First(&system.SysAuthority{}, "authority_id = ?", authorityID).Error; err != nil {
		return err
	}
	authorityIDString := strconv.Itoa(int(authorityID))
	if err := registerMenuPermissions(db, authorityIDString, menuNames); err != nil {
		return err
	}
	return registerAPIPermissions(db, authorityIDString, apis)
}

func registerMenuPermissions(db *gorm.DB, authorityID string, menuNames []string) error {
	if len(menuNames) == 0 {
		return nil
	}
	uniqueNames := make([]string, 0, len(menuNames))
	seenNames := make(map[string]struct{}, len(menuNames))
	for _, name := range menuNames {
		if _, exists := seenNames[name]; exists {
			continue
		}
		seenNames[name] = struct{}{}
		uniqueNames = append(uniqueNames, name)
	}
	var menus []system.SysBaseMenu
	if err := db.Select("id").Where("name IN ?", uniqueNames).Find(&menus).Error; err != nil {
		return err
	}
	if len(menus) == 0 {
		return nil
	}
	menuIDs := make([]string, 0, len(menus))
	for _, menu := range menus {
		menuIDs = append(menuIDs, strconv.Itoa(int(menu.ID)))
	}
	var existing []system.SysAuthorityMenu
	if err := db.Where(
		"sys_authority_authority_id = ? AND sys_base_menu_id IN ?", authorityID, menuIDs,
	).Find(&existing).Error; err != nil {
		return err
	}
	existingSet := make(map[string]struct{}, len(existing))
	for _, relation := range existing {
		existingSet[relation.MenuId] = struct{}{}
	}
	missing := make([]system.SysAuthorityMenu, 0, len(menuIDs))
	for _, menuID := range menuIDs {
		if _, exists := existingSet[menuID]; exists {
			continue
		}
		existingSet[menuID] = struct{}{}
		missing = append(missing, system.SysAuthorityMenu{MenuId: menuID, AuthorityId: authorityID})
	}
	if len(missing) == 0 {
		return nil
	}
	return db.Create(&missing).Error
}

func registerAPIPermissions(db *gorm.DB, authorityID string, apis []system.SysApi) error {
	if len(apis) == 0 {
		return nil
	}
	paths := make([]string, 0, len(apis))
	pathSet := make(map[string]struct{}, len(apis))
	for _, api := range apis {
		if _, exists := pathSet[api.Path]; exists {
			continue
		}
		pathSet[api.Path] = struct{}{}
		paths = append(paths, api.Path)
	}
	var existing []gormadapter.CasbinRule
	if err := db.Where("ptype = ? AND v0 = ? AND v1 IN ?", "p", authorityID, paths).Find(&existing).Error; err != nil {
		return err
	}
	existingSet := make(map[permissionRuleIdentity]struct{}, len(existing))
	for _, rule := range existing {
		existingSet[permissionRuleIdentity{path: rule.V1, method: rule.V2}] = struct{}{}
	}
	missing := make([]gormadapter.CasbinRule, 0, len(apis))
	for _, api := range apis {
		identity := permissionRuleIdentity{path: api.Path, method: api.Method}
		if _, exists := existingSet[identity]; exists {
			continue
		}
		existingSet[identity] = struct{}{}
		missing = append(missing, gormadapter.CasbinRule{Ptype: "p", V0: authorityID, V1: api.Path, V2: api.Method})
	}
	if len(missing) == 0 {
		return nil
	}
	return db.Create(&missing).Error
}
