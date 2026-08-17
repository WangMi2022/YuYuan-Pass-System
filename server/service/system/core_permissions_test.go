package system

import (
	"context"
	"testing"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	sysModel "github.com/WangMi2022/mit-assets-admin/server/model/system"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestEnsureCoreAdminPermissionsRepairsExistingDatabase(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err = db.AutoMigrate(&gormadapter.CasbinRule{}, &sysModel.SysIgnoreApi{}); err != nil {
		t.Fatalf("migrate casbin table: %v", err)
	}
	if err = db.Create(&[]sysModel.SysIgnoreApi{
		{Path: "/autoCode/llmAuto", Method: "POST"},
		{Path: "/autoCode/llmAutoSSE", Method: "POST"},
		{Path: swaggerPath, Method: "GET"},
		{Path: freshCasbinPath, Method: "GET"},
	}).Error; err != nil {
		t.Fatalf("seed obsolete ignored AI routes: %v", err)
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
	for _, rule := range requiredCoreAdminRules {
		var ruleCount int64
		if err = db.Model(&gormadapter.CasbinRule{}).
			Where(
				"ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?",
				rule.Ptype, rule.V0, rule.V1, rule.V2,
			).
			Count(&ruleCount).Error; err != nil {
			t.Fatalf("count repaired permission %s %s: %v", rule.V1, rule.V2, err)
		}
		if ruleCount != 1 {
			t.Fatalf("permission %s %s count = %d, want 1", rule.V1, rule.V2, ruleCount)
		}
	}
	var ignoredCount int64
	if err = db.Model(&sysModel.SysIgnoreApi{}).
		Where("(path IN ? AND method = ?) OR (path IN ? AND method = ?)", []string{"/autoCode/llmAuto", "/autoCode/llmAutoSSE"}, "POST", []string{swaggerPath, freshCasbinPath}, "GET").
		Count(&ignoredCount).Error; err != nil {
		t.Fatalf("count obsolete ignored AI routes: %v", err)
	}
	if ignoredCount != 0 {
		t.Fatalf("obsolete ignored AI routes count = %d, want 0", ignoredCount)
	}
}

func TestEnsureCoreAdminPermissionsRejectsMissingDatabase(t *testing.T) {
	err := new(CasbinService).EnsureCoreAdminPermissions(context.Background(), nil)
	if err == nil {
		t.Fatal("expected missing database error")
	}
}
