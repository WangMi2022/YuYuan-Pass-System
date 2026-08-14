package initialize

import (
	"context"
	"testing"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestAssetRecognitionPermissionMigrationIsIdempotent(t *testing.T) {
	previous := global.GVA_DB
	database, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open permission test database: %v", err)
	}
	if err = database.AutoMigrate(&gormadapter.CasbinRule{}); err != nil {
		t.Fatalf("migrate casbin rules: %v", err)
	}
	global.GVA_DB = database
	t.Cleanup(func() {
		global.GVA_DB = previous
		if sqlDB, dbErr := database.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	})
	sources := []gormadapter.CasbinRule{
		{Ptype: "p", V0: "100", V1: "/asset/list", V2: "GET"},
		{Ptype: "p", V0: "100", V1: "/asset/create", V2: "POST"},
		{Ptype: "p", V0: "200", V1: "/asset/list", V2: "GET"},
		{Ptype: "p", V0: "100", V1: "/assetRecognition/list", V2: "GET"},
	}
	if err = database.Create(&sources).Error; err != nil {
		t.Fatalf("create source permission rules: %v", err)
	}
	for run := 0; run < 2; run++ {
		if err = migrateAssetRecognitionPermissions(context.Background()); err != nil {
			t.Fatalf("run permission migration %d: %v", run+1, err)
		}
	}
	var recognitionRules []gormadapter.CasbinRule
	if err = database.Where("ptype = ? AND v1 LIKE ?", "p", "/assetRecognition/%").Order("v0, v1, v2").Find(&recognitionRules).Error; err != nil {
		t.Fatalf("load migrated permission rules: %v", err)
	}
	if len(recognitionRules) != 9 {
		t.Fatalf("unexpected migrated rule count: %d (%#v)", len(recognitionRules), recognitionRules)
	}
	seen := make(map[string]int, len(recognitionRules))
	for _, rule := range recognitionRules {
		seen[rule.V0+"\x00"+rule.V1+"\x00"+rule.V2]++
	}
	for _, rule := range recognitionRules {
		if seen[rule.V0+"\x00"+rule.V1+"\x00"+rule.V2] != 1 {
			t.Fatalf("permission migration created duplicate rule: %#v", rule)
		}
	}
	for _, rule := range assetRecognitionWriteRules {
		if seen["200\x00"+rule.path+"\x00"+rule.method] != 0 {
			t.Fatalf("read-only authority inherited write rule: %#v", rule)
		}
	}
}
