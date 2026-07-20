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

func TestSyncBusinessNavigationGroupsPermissionMenusAndMigratesAuthorities(t *testing.T) {
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

	systemParent := system.SysBaseMenu{
		ParentId: 0, Path: "admin", Name: "superAdmin", Component: "view/superAdmin/index.vue",
		Meta: system.Meta{Title: "系统管理"},
	}
	if err = db.Create(&systemParent).Error; err != nil {
		t.Fatalf("create system parent: %v", err)
	}

	legacyMenus := []system.SysBaseMenu{
		{ParentId: systemParent.ID, MenuLevel: 1, Path: "user", Name: "user"},
		{ParentId: systemParent.ID, MenuLevel: 1, Path: "authority", Name: "authority"},
		{ParentId: systemParent.ID, MenuLevel: 1, Path: "api", Name: "api"},
		{ParentId: systemParent.ID, MenuLevel: 1, Path: "menu", Name: "menu"},
		{ParentId: systemParent.ID, MenuLevel: 1, Path: "dictionary", Name: "dictionary"},
	}
	if err = db.Create(&legacyMenus).Error; err != nil {
		t.Fatalf("create legacy menus: %v", err)
	}

	legacyRelations := []system.SysAuthorityMenu{
		{MenuId: strconv.Itoa(int(legacyMenus[0].ID)), AuthorityId: "100"},
		{MenuId: strconv.Itoa(int(legacyMenus[1].ID)), AuthorityId: "200"},
		{MenuId: strconv.Itoa(int(legacyMenus[2].ID)), AuthorityId: "300"},
		{MenuId: strconv.Itoa(int(legacyMenus[3].ID)), AuthorityId: "400"},
		{MenuId: strconv.Itoa(int(legacyMenus[4].ID)), AuthorityId: "999"},
	}
	if err = db.Create(&legacyRelations).Error; err != nil {
		t.Fatalf("create legacy authority relations: %v", err)
	}

	for run := 0; run < 2; run++ {
		if err = syncBusinessNavigation(context.Background()); err != nil {
			t.Fatalf("sync business navigation on run %d: %v", run+1, err)
		}
	}

	var permissionParent system.SysBaseMenu
	if err = db.Where("name = ?", permissionMenuName).First(&permissionParent).Error; err != nil {
		t.Fatalf("find permission parent: %v", err)
	}
	if permissionParent.ParentId != systemParent.ID || permissionParent.MenuLevel != 1 || permissionParent.Path != permissionMenuName || permissionParent.Component != "view/routerHolder.vue" || permissionParent.Meta.Title != "权限管理" {
		t.Fatalf("unexpected permission parent: %#v", permissionParent)
	}

	expectedMenus := []navigationItem{
		{name: "user", sort: 1},
		{name: "authority", sort: 2},
		{name: "api", sort: 3},
		{name: "menu", sort: 4},
	}
	for _, expected := range expectedMenus {
		var menu system.SysBaseMenu
		if err = db.Where("name = ?", expected.name).First(&menu).Error; err != nil {
			t.Fatalf("find nested menu %q: %v", expected.name, err)
		}
		if menu.ParentId != permissionParent.ID || menu.MenuLevel != 2 || menu.Sort != expected.sort {
			t.Errorf("unexpected nested menu %q: %#v", expected.name, menu)
		}
	}

	for _, authorityID := range []string{"100", "200", "300", "400"} {
		assertAuthorityMenuRelation(t, db, permissionParent.ID, authorityID, 1)
		assertAuthorityMenuRelation(t, db, systemParent.ID, authorityID, 1)
	}
	assertAuthorityMenuRelation(t, db, permissionParent.ID, "999", 0)
	assertAuthorityMenuRelation(t, db, systemParent.ID, "999", 0)
	assertAuthorityMenuRelation(t, db, legacyMenus[0].ID, "100", 1)
}

func assertAuthorityMenuRelation(t *testing.T, db *gorm.DB, menuID uint, authorityID string, expected int64) {
	t.Helper()
	var count int64
	if err := db.Model(&system.SysAuthorityMenu{}).
		Where("sys_base_menu_id = ? AND sys_authority_authority_id = ?", strconv.Itoa(int(menuID)), authorityID).
		Count(&count).Error; err != nil {
		t.Fatalf("count authority menu relation: %v", err)
	}
	if count != expected {
		t.Fatalf("menu %d authority %s: expected %d relation(s), got %d", menuID, authorityID, expected, count)
	}
}
