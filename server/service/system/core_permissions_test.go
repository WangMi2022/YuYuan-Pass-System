package system

import (
	"context"
	"testing"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestEnsureCoreAdminPermissionsRepairsExistingDatabase(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err = db.AutoMigrate(&gormadapter.CasbinRule{}); err != nil {
		t.Fatalf("migrate casbin table: %v", err)
	}

	service := new(CasbinService)
	if err = service.EnsureCoreAdminPermissions(context.Background(), db); err != nil {
		t.Fatalf("repair core administrator permissions: %v", err)
	}
	if err = service.EnsureCoreAdminPermissions(context.Background(), db); err != nil {
		t.Fatalf("repeat core administrator permission repair: %v", err)
	}

	var count int64
	if err = db.Model(&gormadapter.CasbinRule{}).
		Where(
			"ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?",
			"p", coreAdminAuthorityID, reloadSystemPath, reloadSystemMethod,
		).
		Count(&count).Error; err != nil {
		t.Fatalf("count repaired reload permission: %v", err)
	}
	if count != 1 {
		t.Fatalf("reload permission count = %d, want 1", count)
	}
}

func TestEnsureCoreAdminPermissionsRejectsMissingDatabase(t *testing.T) {
	err := new(CasbinService).EnsureCoreAdminPermissions(context.Background(), nil)
	if err == nil {
		t.Fatal("expected missing database error")
	}
}
