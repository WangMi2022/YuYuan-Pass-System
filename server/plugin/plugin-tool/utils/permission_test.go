package utils

import (
	"errors"
	"testing"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"gorm.io/gorm"
)

func TestRegisterPermissionsBatchesAndDeduplicatesRelations(t *testing.T) {
	db, counter := setupMetadataTestDB(t)
	if err := db.AutoMigrate(
		&system.SysAuthority{}, &system.SysAuthorityMenu{}, &gormadapter.CasbinRule{},
	); err != nil {
		t.Fatalf("migrate permission tables: %v", err)
	}
	parentID := uint(0)
	authority := system.SysAuthority{AuthorityId: 888, AuthorityName: "管理员", ParentId: &parentID}
	if err := db.Create(&authority).Error; err != nil {
		t.Fatalf("create authority: %v", err)
	}
	menus := []system.SysBaseMenu{{Name: "asset"}, {Name: "assetList"}}
	if err := db.Create(&menus).Error; err != nil {
		t.Fatalf("create menus: %v", err)
	}
	apis := []system.SysApi{
		{Path: "/asset/list", Method: "GET"},
		{Path: "/asset/create", Method: "POST"},
		{Path: "/asset/list", Method: "GET"},
	}

	counter.reset()
	if err := registerPermissions(db, authority.AuthorityId, []string{"asset", "assetList", "missing", "asset"}, apis); err != nil {
		t.Fatalf("register permissions: %v", err)
	}
	if len(counter.statements) > 6 {
		t.Fatalf("initial permission registration executed %d SQL statements, want at most 6: %v", len(counter.statements), counter.statements)
	}
	var menuRelationCount int64
	if err := db.Model(&system.SysAuthorityMenu{}).Count(&menuRelationCount).Error; err != nil {
		t.Fatalf("count menu permissions: %v", err)
	}
	if menuRelationCount != 2 {
		t.Fatalf("registered %d menu permissions, want 2", menuRelationCount)
	}
	var apiRuleCount int64
	if err := db.Model(&gormadapter.CasbinRule{}).Where("ptype = ? AND v0 = ?", "p", "888").Count(&apiRuleCount).Error; err != nil {
		t.Fatalf("count API permissions: %v", err)
	}
	if apiRuleCount != 2 {
		t.Fatalf("registered %d API permissions, want 2", apiRuleCount)
	}

	counter.reset()
	if err := registerPermissions(db, authority.AuthorityId, []string{"asset", "assetList"}, apis); err != nil {
		t.Fatalf("repeat permission registration: %v", err)
	}
	if len(counter.statements) != 4 {
		t.Fatalf("repeat permission registration executed %d SQL statements, want 4: %v", len(counter.statements), counter.statements)
	}
}

func TestRegisterPermissionsRequiresExistingAuthority(t *testing.T) {
	db, _ := setupMetadataTestDB(t)
	if err := db.AutoMigrate(&system.SysAuthority{}, &system.SysAuthorityMenu{}, &gormadapter.CasbinRule{}); err != nil {
		t.Fatalf("migrate permission tables: %v", err)
	}
	err := registerPermissions(db, 999, nil, nil)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected missing authority error, got %v", err)
	}
}
