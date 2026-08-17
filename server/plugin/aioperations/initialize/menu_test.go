package initialize

import (
	"context"
	"strconv"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestSyncAIServiceMenuMovesMenuAndPreservesAuthorities(t *testing.T) {
	previousDB := global.GVA_DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err = db.AutoMigrate(&system.SysBaseMenu{}, &system.SysAuthorityMenu{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = previousDB })

	systemAdmin := system.SysBaseMenu{
		ParentId: 0, Path: "admin", Name: systemAdminMenuName,
		Component: "view/superAdmin/index.vue", Meta: system.Meta{Title: "系统管理"},
	}
	legacyAI := system.SysBaseMenu{
		ParentId: 0, MenuLevel: 0, Path: aiOperationsMenuName, Name: aiOperationsMenuName,
		Component: "plugin/aioperations/view/operations.vue", Sort: 98,
		Meta: system.Meta{Title: "AI 运营", Icon: "cpu", KeepAlive: true},
	}
	if err = db.Create(&systemAdmin).Error; err != nil {
		t.Fatalf("create system admin menu: %v", err)
	}
	if err = db.Create(&legacyAI).Error; err != nil {
		t.Fatalf("create legacy AI menu: %v", err)
	}
	relations := []system.SysAuthorityMenu{
		{MenuId: strconv.Itoa(int(legacyAI.ID)), AuthorityId: "100"},
		{MenuId: strconv.Itoa(int(legacyAI.ID)), AuthorityId: "200"},
		{MenuId: strconv.Itoa(int(systemAdmin.ID)), AuthorityId: "200"},
	}
	if err = db.Create(&relations).Error; err != nil {
		t.Fatalf("create legacy menu authorities: %v", err)
	}

	for run := 0; run < 2; run++ {
		if err = syncAIServiceMenu(context.Background()); err != nil {
			t.Fatalf("sync AI service menu on run %d: %v", run+1, err)
		}
	}

	var aiMenu system.SysBaseMenu
	if err = db.Where("name = ?", aiOperationsMenuName).First(&aiMenu).Error; err != nil {
		t.Fatalf("find AI service menu: %v", err)
	}
	var intelligenceServices system.SysBaseMenu
	if err = db.Where("name = ?", systemIntelligenceMenuName).First(&intelligenceServices).Error; err != nil {
		t.Fatalf("find intelligent services group: %v", err)
	}
	if intelligenceServices.ParentId != systemAdmin.ID || intelligenceServices.MenuLevel != 1 || intelligenceServices.Sort != 4 || intelligenceServices.Path != systemIntelligenceMenuName || intelligenceServices.Component != "view/routerHolder.vue" || intelligenceServices.Meta.Title != intelligenceServicesMenuTitle {
		t.Fatalf("unexpected intelligent services group: %#v", intelligenceServices)
	}
	if aiMenu.ParentId != intelligenceServices.ID || aiMenu.MenuLevel != 2 || aiMenu.Sort != 1 {
		t.Fatalf("unexpected AI service menu hierarchy: %#v", aiMenu)
	}
	if aiMenu.Meta.Title != aiServicesMenuTitle || aiMenu.Path != aiOperationsMenuName || aiMenu.Component != "plugin/aioperations/view/operations.vue" || !aiMenu.Meta.KeepAlive {
		t.Fatalf("unexpected AI service menu configuration: %#v", aiMenu)
	}

	assertAIServiceAuthorityMenu(t, db, legacyAI.ID, "100", 1)
	assertAIServiceAuthorityMenu(t, db, legacyAI.ID, "200", 1)
	assertAIServiceAuthorityMenu(t, db, intelligenceServices.ID, "100", 1)
	assertAIServiceAuthorityMenu(t, db, intelligenceServices.ID, "200", 1)
	assertAIServiceAuthorityMenu(t, db, systemAdmin.ID, "100", 1)
	assertAIServiceAuthorityMenu(t, db, systemAdmin.ID, "200", 1)
}

func TestSyncAIServiceMenuCreatesMissingSystemAdminParent(t *testing.T) {
	previousDB := global.GVA_DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err = db.AutoMigrate(&system.SysBaseMenu{}, &system.SysAuthorityMenu{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = previousDB })

	aiMenu := system.SysBaseMenu{
		ParentId: 0, Path: aiOperationsMenuName, Name: aiOperationsMenuName,
		Component: "plugin/aioperations/view/operations.vue", Meta: system.Meta{Title: "AI 运营"},
	}
	if err = db.Create(&aiMenu).Error; err != nil {
		t.Fatalf("create legacy AI menu: %v", err)
	}
	if err = db.Create(&system.SysAuthorityMenu{MenuId: strconv.Itoa(int(aiMenu.ID)), AuthorityId: "100"}).Error; err != nil {
		t.Fatalf("create AI menu authority: %v", err)
	}

	if err = syncAIServiceMenu(context.Background()); err != nil {
		t.Fatalf("sync AI service menu: %v", err)
	}

	var systemAdmin system.SysBaseMenu
	if err = db.Where("name = ?", systemAdminMenuName).First(&systemAdmin).Error; err != nil {
		t.Fatalf("find created system admin menu: %v", err)
	}
	if systemAdmin.ParentId != 0 || systemAdmin.Path != "admin" {
		t.Fatalf("unexpected created system admin menu: %#v", systemAdmin)
	}
	var migratedAI system.SysBaseMenu
	if err = db.Where("name = ?", aiOperationsMenuName).First(&migratedAI).Error; err != nil {
		t.Fatalf("find migrated AI menu: %v", err)
	}
	var intelligenceServices system.SysBaseMenu
	if err = db.Where("name = ?", systemIntelligenceMenuName).First(&intelligenceServices).Error; err != nil {
		t.Fatalf("find created intelligent services group: %v", err)
	}
	if intelligenceServices.ParentId != systemAdmin.ID || intelligenceServices.MenuLevel != 1 {
		t.Fatalf("unexpected created intelligent services group: %#v", intelligenceServices)
	}
	if migratedAI.ParentId != intelligenceServices.ID || migratedAI.MenuLevel != 2 {
		t.Fatalf("AI menu did not attach to created system admin: %#v", migratedAI)
	}
	assertAIServiceAuthorityMenu(t, db, intelligenceServices.ID, "100", 1)
	assertAIServiceAuthorityMenu(t, db, systemAdmin.ID, "100", 1)
}

func assertAIServiceAuthorityMenu(t *testing.T, db *gorm.DB, menuID uint, authorityID string, expected int64) {
	t.Helper()
	var count int64
	if err := db.Model(&system.SysAuthorityMenu{}).
		Where("sys_base_menu_id = ? AND sys_authority_authority_id = ?", strconv.Itoa(int(menuID)), authorityID).
		Count(&count).Error; err != nil {
		t.Fatalf("count menu authority relation: %v", err)
	}
	if count != expected {
		t.Fatalf("menu %d authority %s: expected %d relation(s), got %d", menuID, authorityID, expected, count)
	}
}
