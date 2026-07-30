package initialize

import (
	"context"
	"fmt"
	"strings"
	"testing"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMigrateVerificationPermissionsInheritsConfirmAuthorities(t *testing.T) {
	previous := global.GVA_DB
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(&gormadapter.CasbinRule{}); err != nil {
		t.Fatal(err)
	}
	global.GVA_DB = db
	t.Cleanup(func() {
		global.GVA_DB = previous
		if sqlDB, dbErr := db.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	})

	seed := []gormadapter.CasbinRule{
		{Ptype: "p", V0: "100", V1: "/invoice/confirm", V2: "PUT"},
		{Ptype: "p", V0: "200", V1: "/invoice/confirm", V2: "PUT"},
		{Ptype: "p", V0: "100", V1: "/invoice/verify", V2: "POST"},
		{Ptype: "p", V0: "300", V1: "/invoice/list", V2: "GET"},
	}
	if err = db.Create(&seed).Error; err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err = migrateVerificationPermissions(context.Background()); err != nil {
			t.Fatal(err)
		}
	}

	for _, authority := range []string{"100", "200"} {
		for _, rule := range inheritedVerificationRules {
			var count int64
			if err = db.Model(&gormadapter.CasbinRule{}).Where(
				"ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", authority, rule.path, rule.method,
			).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatalf("authority %s rule %s %s count = %d", authority, rule.method, rule.path, count)
			}
		}
	}
	var unrelatedCount int64
	if err = db.Model(&gormadapter.CasbinRule{}).Where(
		"ptype = ? AND v0 = ? AND v1 IN ?", "p", "300", []string{"/invoice/verify", "/invoice/verificationHistory"},
	).Count(&unrelatedCount).Error; err != nil {
		t.Fatal(err)
	}
	if unrelatedCount != 0 {
		t.Fatalf("authority without confirm permission inherited %d verification rules", unrelatedCount)
	}
}
