package system

import (
	"context"
	"testing"

	sysModel "github.com/WangMi2022/mit-assets-admin/server/model/system"
	gormadapter "github.com/casbin/gorm-adapter/v3"
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

func TestEnsureContactVerificationPermissionsRepairsExistingDatabase(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err = db.AutoMigrate(&gormadapter.CasbinRule{}, &sysModel.SysApi{}, &sysModel.SysAuthority{}); err != nil {
		t.Fatalf("migrate permission tables: %v", err)
	}
	authorities := []sysModel.SysAuthority{
		{AuthorityId: 888, AuthorityName: "平台管理员"},
		{AuthorityId: 9528, AuthorityName: "普通用户"},
		{AuthorityId: 10001, AuthorityName: "资产管理员"},
	}
	if err = db.Create(&authorities).Error; err != nil {
		t.Fatalf("seed authorities: %v", err)
	}
	if err = db.Create(&sysModel.SysApi{
		Path:        "/user/contactVerificationCapabilities",
		Method:      "GET",
		ApiGroup:    "系统用户",
		Description: "保留管理员自定义描述",
	}).Error; err != nil {
		t.Fatalf("seed existing contact API: %v", err)
	}
	if err = db.Create(&gormadapter.CasbinRule{
		Ptype: "p",
		V0:    "10001",
		V1:    "/asset/read",
		V2:    "GET",
	}).Error; err != nil {
		t.Fatalf("seed existing custom permission: %v", err)
	}

	service := new(CasbinService)
	if err = service.EnsureContactVerificationPermissions(context.Background(), db); err != nil {
		t.Fatalf("repair contact verification permissions: %v", err)
	}
	if err = service.EnsureContactVerificationPermissions(context.Background(), db); err != nil {
		t.Fatalf("repeat contact verification permission repair: %v", err)
	}

	routes := []struct {
		path   string
		method string
	}{
		{path: "/user/contactVerificationCapabilities", method: "GET"},
		{path: "/user/sendContactVerificationCode", method: "POST"},
		{path: "/user/updateSelfContact", method: "PUT"},
	}
	for _, route := range routes {
		var apiCount int64
		if err = db.Model(&sysModel.SysApi{}).
			Where("path = ? AND method = ?", route.path, route.method).
			Count(&apiCount).Error; err != nil {
			t.Fatalf("count API %s %s: %v", route.path, route.method, err)
		}
		if apiCount != 1 {
			t.Fatalf("API %s %s count = %d, want 1", route.path, route.method, apiCount)
		}
		for _, authority := range authorities {
			var policyCount int64
			if err = db.Model(&gormadapter.CasbinRule{}).
				Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", authority.AuthorityId, route.path, route.method).
				Count(&policyCount).Error; err != nil {
				t.Fatalf("count policy %s %s authority %d: %v", route.path, route.method, authority.AuthorityId, err)
			}
			if policyCount != 1 {
				t.Fatalf("policy %s %s authority %d count = %d, want 1", route.path, route.method, authority.AuthorityId, policyCount)
			}
		}
	}

	var description string
	if err = db.Model(&sysModel.SysApi{}).
		Where("path = ? AND method = ?", "/user/contactVerificationCapabilities", "GET").
		Pluck("description", &description).Error; err != nil {
		t.Fatalf("read existing API description: %v", err)
	}
	if description != "保留管理员自定义描述" {
		t.Fatalf("existing API description = %q, want preserved metadata", description)
	}

	var customPolicyCount int64
	if err = db.Model(&gormadapter.CasbinRule{}).
		Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", "10001", "/asset/read", "GET").
		Count(&customPolicyCount).Error; err != nil {
		t.Fatalf("count existing custom policy: %v", err)
	}
	if customPolicyCount != 1 {
		t.Fatalf("existing custom policy count = %d, want 1", customPolicyCount)
	}
}

func TestEnsureContactVerificationPermissionsRejectsMissingDatabase(t *testing.T) {
	err := new(CasbinService).EnsureContactVerificationPermissions(context.Background(), nil)
	if err == nil {
		t.Fatal("expected missing database error")
	}
}

func TestEnsureContactVerificationPermissionsRestoresSoftDeletedAPI(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err = db.AutoMigrate(&gormadapter.CasbinRule{}, &sysModel.SysApi{}, &sysModel.SysAuthority{}); err != nil {
		t.Fatalf("migrate permission tables: %v", err)
	}
	if err = db.Create(&sysModel.SysAuthority{AuthorityId: 9528, AuthorityName: "普通用户"}).Error; err != nil {
		t.Fatalf("seed authority: %v", err)
	}
	api := sysModel.SysApi{
		Path:        "/user/contactVerificationCapabilities",
		Method:      "GET",
		ApiGroup:    "自定义分组",
		Description: "保留软删除前的元数据",
	}
	if err = db.Create(&api).Error; err != nil {
		t.Fatalf("seed contact API: %v", err)
	}
	if err = db.Delete(&api).Error; err != nil {
		t.Fatalf("soft delete contact API: %v", err)
	}

	if err = new(CasbinService).EnsureContactVerificationPermissions(context.Background(), db); err != nil {
		t.Fatalf("restore contact verification API: %v", err)
	}

	var activeAPI sysModel.SysApi
	if err = db.Where("path = ? AND method = ?", api.Path, api.Method).First(&activeAPI).Error; err != nil {
		t.Fatalf("read restored contact API: %v", err)
	}
	if activeAPI.ID != api.ID {
		t.Fatalf("restored API ID = %d, want original ID %d", activeAPI.ID, api.ID)
	}
	if activeAPI.ApiGroup != api.ApiGroup || activeAPI.Description != api.Description {
		t.Fatalf("restored API metadata changed: group=%q description=%q", activeAPI.ApiGroup, activeAPI.Description)
	}
	var unscopedCount int64
	if err = db.Unscoped().Model(&sysModel.SysApi{}).
		Where("path = ? AND method = ?", api.Path, api.Method).
		Count(&unscopedCount).Error; err != nil {
		t.Fatalf("count contact APIs: %v", err)
	}
	if unscopedCount != 1 {
		t.Fatalf("unscoped contact API count = %d, want 1", unscopedCount)
	}
}

func TestEnsureContactVerificationPermissionsPrefersExistingActiveAPI(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err = db.AutoMigrate(&gormadapter.CasbinRule{}, &sysModel.SysApi{}, &sysModel.SysAuthority{}); err != nil {
		t.Fatalf("migrate permission tables: %v", err)
	}
	if err = db.Create(&sysModel.SysAuthority{AuthorityId: 9528, AuthorityName: "普通用户"}).Error; err != nil {
		t.Fatalf("seed authority: %v", err)
	}
	deletedAPI := sysModel.SysApi{
		Path:        "/user/contactVerificationCapabilities",
		Method:      "GET",
		ApiGroup:    "历史分组",
		Description: "历史软删除记录",
	}
	if err = db.Create(&deletedAPI).Error; err != nil {
		t.Fatalf("seed deleted contact API: %v", err)
	}
	if err = db.Delete(&deletedAPI).Error; err != nil {
		t.Fatalf("soft delete contact API: %v", err)
	}
	activeAPI := sysModel.SysApi{
		Path:        deletedAPI.Path,
		Method:      deletedAPI.Method,
		ApiGroup:    "当前分组",
		Description: "当前活动记录",
	}
	if err = db.Create(&activeAPI).Error; err != nil {
		t.Fatalf("seed active contact API: %v", err)
	}

	if err = new(CasbinService).EnsureContactVerificationPermissions(context.Background(), db); err != nil {
		t.Fatalf("repair contact verification permissions: %v", err)
	}

	var activeAPIs []sysModel.SysApi
	if err = db.Where("path = ? AND method = ?", activeAPI.Path, activeAPI.Method).Find(&activeAPIs).Error; err != nil {
		t.Fatalf("read active contact APIs: %v", err)
	}
	if len(activeAPIs) != 1 || activeAPIs[0].ID != activeAPI.ID {
		t.Fatalf("active contact APIs = %#v, want only ID %d", activeAPIs, activeAPI.ID)
	}
	var deletedStillPresent int64
	if err = db.Unscoped().Model(&sysModel.SysApi{}).
		Where("id = ? AND deleted_at IS NOT NULL", deletedAPI.ID).
		Count(&deletedStillPresent).Error; err != nil {
		t.Fatalf("count historical deleted API: %v", err)
	}
	if deletedStillPresent != 1 {
		t.Fatalf("historical deleted API count = %d, want 1", deletedStillPresent)
	}
}
