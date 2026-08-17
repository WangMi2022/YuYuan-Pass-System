package initialize

import (
	"context"
	"testing"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/system"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestMigrateAIServiceAPIRegistry(t *testing.T) {
	previousDB := global.GVA_DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err = db.AutoMigrate(&system.SysApi{}, &gormadapter.CasbinRule{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = previousDB })

	legacyRows := []system.SysApi{
		{Path: "/ai/providers", Method: "GET", ApiGroup: legacyAIServiceAPIGroup, Description: "旧描述"},
		{Path: "/ai/providers", Method: "PUT", ApiGroup: legacyAIServiceAPIGroup, Description: "重复旧描述"},
		{Path: "/ai/providers", Method: "PUT", ApiGroup: previousAIServiceAPIGroup, Description: "新描述"},
		{Path: legacyInvoiceProviderTestAPI.path, Method: legacyInvoiceProviderTestAPI.method, ApiGroup: "运行配置"},
	}
	if err = db.Create(&legacyRows).Error; err != nil {
		t.Fatalf("seed API metadata: %v", err)
	}
	if err = db.Create(&[]gormadapter.CasbinRule{
		{Ptype: "p", V0: "888", V1: legacyInvoiceProviderTestAPI.path, V2: legacyInvoiceProviderTestAPI.method},
		{Ptype: "p", V0: "888", V1: "/ai/providers", V2: "GET"},
	}).Error; err != nil {
		t.Fatalf("seed Casbin metadata: %v", err)
	}

	if err = migrateAIServiceAPIRegistry(context.Background()); err != nil {
		t.Fatalf("migrate API metadata: %v", err)
	}

	var providerGET system.SysApi
	if err = db.Where("path = ? AND method = ?", "/ai/providers", "GET").First(&providerGET).Error; err != nil {
		t.Fatalf("load migrated provider API: %v", err)
	}
	if providerGET.ApiGroup != aiServicesMenuTitle || providerGET.Description != "查看模型接入状态" {
		t.Fatalf("unexpected migrated provider API: %#v", providerGET)
	}

	var providerPUTs []system.SysApi
	if err = db.Where("path = ? AND method = ?", "/ai/providers", "PUT").Find(&providerPUTs).Error; err != nil {
		t.Fatalf("load migrated duplicate APIs: %v", err)
	}
	if len(providerPUTs) != 1 || providerPUTs[0].ApiGroup != aiServicesMenuTitle {
		t.Fatalf("unexpected duplicate provider APIs: %#v", providerPUTs)
	}

	var obsoleteCount int64
	if err = db.Model(&system.SysApi{}).Where("path = ? AND method = ?", legacyInvoiceProviderTestAPI.path, legacyInvoiceProviderTestAPI.method).Count(&obsoleteCount).Error; err != nil {
		t.Fatalf("count obsolete API metadata: %v", err)
	}
	if obsoleteCount != 0 {
		t.Fatalf("obsolete API metadata remains: %d", obsoleteCount)
	}
	var obsoleteRuleCount int64
	if err = db.Model(&gormadapter.CasbinRule{}).Where("ptype = ? AND v1 = ? AND v2 = ?", "p", legacyInvoiceProviderTestAPI.path, legacyInvoiceProviderTestAPI.method).Count(&obsoleteRuleCount).Error; err != nil {
		t.Fatalf("count obsolete Casbin rules: %v", err)
	}
	if obsoleteRuleCount != 0 {
		t.Fatalf("obsolete Casbin rules remain: %d", obsoleteRuleCount)
	}
}
