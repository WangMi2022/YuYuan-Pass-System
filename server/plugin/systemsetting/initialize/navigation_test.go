package initialize

import (
	"context"
	"strconv"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/system"
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
		{ParentId: systemParent.ID, MenuLevel: 1, Path: "operation", Name: "operation"},
		{ParentId: systemParent.ID, MenuLevel: 1, Path: "loginLog", Name: "loginLog"},
		{ParentId: systemParent.ID, MenuLevel: 1, Path: "sysError", Name: "sysError"},
		{ParentId: 0, MenuLevel: 0, Path: "https://www.gin-vue-admin.com", Name: "https://www.gin-vue-admin.com"},
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
		{MenuId: strconv.Itoa(int(legacyMenus[4].ID)), AuthorityId: "200"},
		{MenuId: strconv.Itoa(int(legacyMenus[5].ID)), AuthorityId: "500"},
		{MenuId: strconv.Itoa(int(legacyMenus[6].ID)), AuthorityId: "600"},
		{MenuId: strconv.Itoa(int(legacyMenus[7].ID)), AuthorityId: "700"},
		{MenuId: strconv.Itoa(int(legacyMenus[7].ID)), AuthorityId: "200"},
		// 旧结构下，勾选权限子菜单时会同时保存系统管理父菜单。
		{MenuId: strconv.Itoa(int(systemParent.ID)), AuthorityId: "100"},
		{MenuId: strconv.Itoa(int(systemParent.ID)), AuthorityId: "200"},
		{MenuId: strconv.Itoa(int(systemParent.ID)), AuthorityId: "400"},
		{MenuId: strconv.Itoa(int(systemParent.ID)), AuthorityId: "999"},
		{MenuId: strconv.Itoa(int(systemParent.ID)), AuthorityId: "500"},
		{MenuId: strconv.Itoa(int(systemParent.ID)), AuthorityId: "600"},
		{MenuId: strconv.Itoa(int(systemParent.ID)), AuthorityId: "700"},
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
	if permissionParent.ParentId != 0 || permissionParent.MenuLevel != 0 || permissionParent.Sort != 7 || permissionParent.Path != permissionMenuName || permissionParent.Component != "view/routerHolder.vue" || permissionParent.Meta.Title != "权限管理" {
		t.Fatalf("unexpected permission parent: %#v", permissionParent)
	}
	var auditParent system.SysBaseMenu
	if err = db.Where("name = ?", auditMenuName).First(&auditParent).Error; err != nil {
		t.Fatalf("find audit parent: %v", err)
	}
	if auditParent.ParentId != 0 || auditParent.MenuLevel != 0 || auditParent.Sort != 8 || auditParent.Path != auditMenuName || auditParent.Component != "view/routerHolder.vue" || auditParent.Meta.Title != "审计平台" {
		t.Fatalf("unexpected audit parent: %#v", auditParent)
	}
	if err = db.Where("name = ?", "superAdmin").First(&systemParent).Error; err != nil {
		t.Fatalf("reload system parent: %v", err)
	}
	if systemParent.ParentId != 0 || systemParent.MenuLevel != 0 || systemParent.Sort != 9 {
		t.Fatalf("unexpected system parent: %#v", systemParent)
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
		if menu.ParentId != permissionParent.ID || menu.MenuLevel != 1 || menu.Sort != expected.sort {
			t.Errorf("unexpected nested menu %q: %#v", expected.name, menu)
		}
	}

	auditMenus := []navigationItem{
		{name: "operation", sort: 1},
		{name: "loginLog", sort: 2},
		{name: "sysError", sort: 3},
	}
	for _, expected := range auditMenus {
		var menu system.SysBaseMenu
		if err = db.Where("name = ?", expected.name).First(&menu).Error; err != nil {
			t.Fatalf("find audit menu %q: %v", expected.name, err)
		}
		if menu.ParentId != auditParent.ID || menu.MenuLevel != 1 || menu.Sort != expected.sort {
			t.Errorf("unexpected audit menu %q: %#v", expected.name, menu)
		}
	}

	var workCalendarParent system.SysBaseMenu
	if err = db.Where("name = ?", workCalendarMenuName).First(&workCalendarParent).Error; err != nil {
		t.Fatalf("find work calendar parent: %v", err)
	}
	if workCalendarParent.ParentId != 0 || workCalendarParent.MenuLevel != 0 || workCalendarParent.Sort != 4 || workCalendarParent.Path != workCalendarMenuName || workCalendarParent.Component != "view/routerHolder.vue" || workCalendarParent.Meta.Title != "工作日历" {
		t.Fatalf("unexpected work calendar parent: %#v", workCalendarParent)
	}
	var workSchedule system.SysBaseMenu
	if err = db.Where("name = ?", "workSchedule").First(&workSchedule).Error; err != nil {
		t.Fatalf("find work schedule child: %v", err)
	}
	if workSchedule.ParentId != workCalendarParent.ID || workSchedule.MenuLevel != 1 || workSchedule.Sort != 1 || workSchedule.Component != "view/workCalendar/index.vue" {
		t.Fatalf("unexpected work schedule child: %#v", workSchedule)
	}

	for _, authorityID := range []string{"100", "200", "300", "400"} {
		assertAuthorityMenuRelation(t, db, permissionParent.ID, authorityID, 1)
	}
	for _, authorityID := range []string{"200", "500", "600", "700"} {
		assertAuthorityMenuRelation(t, db, auditParent.ID, authorityID, 1)
	}
	// 仅有权限管理子菜单的角色不应残留空的系统管理入口。
	assertAuthorityMenuRelation(t, db, systemParent.ID, "100", 0)
	assertAuthorityMenuRelation(t, db, systemParent.ID, "300", 0)
	assertAuthorityMenuRelation(t, db, systemParent.ID, "400", 0)
	assertAuthorityMenuRelation(t, db, systemParent.ID, "500", 0)
	assertAuthorityMenuRelation(t, db, systemParent.ID, "600", 0)
	assertAuthorityMenuRelation(t, db, systemParent.ID, "700", 0)
	// 同时拥有系统管理子菜单的角色必须保留系统管理父菜单。
	assertAuthorityMenuRelation(t, db, systemParent.ID, "200", 1)
	assertAuthorityMenuRelation(t, db, permissionParent.ID, "999", 0)
	assertAuthorityMenuRelation(t, db, systemParent.ID, "999", 1)
	assertAuthorityMenuRelation(t, db, legacyMenus[0].ID, "100", 1)

	var projectHome system.SysBaseMenu
	if err = db.First(&projectHome, legacyMenus[8].ID).Error; err != nil {
		t.Fatalf("reload migrated project home: %v", err)
	}
	if projectHome.Name != "https://github.com/WangMi2022/YuYuan-Pass-System" || projectHome.Path != projectHome.Name || !projectHome.Hidden {
		t.Fatalf("unexpected migrated project home: %#v", projectHome)
	}
}

func TestSyncBusinessNavigationGroupsSystemManagementMenusAndPreservesAuthorities(t *testing.T) {
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
		{ParentId: systemParent.ID, MenuLevel: 1, Path: "dictionary", Name: "dictionary"},
		{ParentId: systemParent.ID, MenuLevel: 1, Path: "sysParams", Name: "sysParams"},
		{ParentId: systemParent.ID, MenuLevel: 1, Path: "systemSettings", Name: "systemSettings"},
		{ParentId: systemParent.ID, MenuLevel: 1, Path: "system", Name: "system"},
		{ParentId: systemParent.ID, MenuLevel: 1, Path: "apiToken", Name: "apiToken"},
		{ParentId: systemParent.ID, MenuLevel: 1, Path: "sysVersion", Name: "sysVersion"},
		{ParentId: systemParent.ID, MenuLevel: 1, Path: "aiOperations", Name: "aiOperations"},
	}
	if err = db.Create(&legacyMenus).Error; err != nil {
		t.Fatalf("create legacy system menus: %v", err)
	}

	legacyRelations := []system.SysAuthorityMenu{
		{MenuId: strconv.Itoa(int(legacyMenus[0].ID)), AuthorityId: "100"},
		{MenuId: strconv.Itoa(int(legacyMenus[0].ID)), AuthorityId: "200"},
		{MenuId: strconv.Itoa(int(legacyMenus[1].ID)), AuthorityId: "200"},
		{MenuId: strconv.Itoa(int(legacyMenus[2].ID)), AuthorityId: "300"},
		{MenuId: strconv.Itoa(int(legacyMenus[3].ID)), AuthorityId: "200"},
		{MenuId: strconv.Itoa(int(legacyMenus[4].ID)), AuthorityId: "400"},
		{MenuId: strconv.Itoa(int(legacyMenus[5].ID)), AuthorityId: "500"},
		{MenuId: strconv.Itoa(int(legacyMenus[6].ID)), AuthorityId: "600"},
		{MenuId: strconv.Itoa(int(systemParent.ID)), AuthorityId: "100"},
		{MenuId: strconv.Itoa(int(systemParent.ID)), AuthorityId: "200"},
		{MenuId: strconv.Itoa(int(systemParent.ID)), AuthorityId: "300"},
		{MenuId: strconv.Itoa(int(systemParent.ID)), AuthorityId: "400"},
		{MenuId: strconv.Itoa(int(systemParent.ID)), AuthorityId: "500"},
		{MenuId: strconv.Itoa(int(systemParent.ID)), AuthorityId: "600"},
		{MenuId: strconv.Itoa(int(systemParent.ID)), AuthorityId: "999"},
	}
	if err = db.Create(&legacyRelations).Error; err != nil {
		t.Fatalf("create legacy authority relations: %v", err)
	}

	for run := 0; run < 2; run++ {
		if err = syncBusinessNavigation(context.Background()); err != nil {
			t.Fatalf("sync business navigation on run %d: %v", run+1, err)
		}
	}

	expectedGroups := []struct {
		name     string
		title    string
		icon     string
		sort     int
		children []navigationItem
		authors  []string
	}{
		{
			name: systemDataMenuName, title: "基础数据", icon: "collection-tag", sort: 1,
			children: []navigationItem{{name: "dictionary", title: "数据字典", icon: "notebook", sort: 1}, {name: "sysParams", title: "系统参数", icon: "compass", sort: 2}},
			authors:  []string{"100", "200"},
		},
		{
			name: systemConfigurationMenuName, title: "平台设置", icon: "setting", sort: 2,
			children: []navigationItem{{name: "systemSettings", title: "登录外观", icon: "picture", sort: 1}, {name: "system", title: "运行配置", icon: "operation", sort: 2}},
			authors:  []string{"200", "300"},
		},
		{
			name: systemIntegrationMenuName, title: "开放与运维", icon: "connection", sort: 3,
			children: []navigationItem{{name: "apiToken", title: "接口令牌", icon: "key", sort: 1}, {name: "sysVersion", title: "配置版本", icon: "server", sort: 2}},
			authors:  []string{"400", "500"},
		},
		{
			name: systemIntelligenceMenuName, title: "智能服务", icon: "cpu", sort: 4,
			children: []navigationItem{{name: "aiOperations", title: "智能能力配置", icon: "cpu", sort: 1}},
			authors:  []string{"600"},
		},
	}

	for _, expectedGroup := range expectedGroups {
		var group system.SysBaseMenu
		if err = db.Where("name = ?", expectedGroup.name).First(&group).Error; err != nil {
			t.Fatalf("find system management group %q: %v", expectedGroup.name, err)
		}
		if group.ParentId != systemParent.ID || group.MenuLevel != 1 || group.Path != expectedGroup.name || group.Component != "view/routerHolder.vue" || group.Sort != expectedGroup.sort || group.Meta.Title != expectedGroup.title || group.Meta.Icon != expectedGroup.icon {
			t.Errorf("unexpected system management group %q: %#v", expectedGroup.name, group)
		}
		var count int64
		if err = db.Model(&system.SysBaseMenu{}).Where("name = ?", expectedGroup.name).Count(&count).Error; err != nil {
			t.Fatalf("count system management group %q: %v", expectedGroup.name, err)
		}
		if count != 1 {
			t.Errorf("expected one system management group %q, got %d", expectedGroup.name, count)
		}
		for _, expectedChild := range expectedGroup.children {
			var child system.SysBaseMenu
			if err = db.Where("name = ?", expectedChild.name).First(&child).Error; err != nil {
				t.Fatalf("find system management child %q: %v", expectedChild.name, err)
			}
			if child.ParentId != group.ID || child.MenuLevel != 2 || child.Sort != expectedChild.sort || child.Meta.Title != expectedChild.title || child.Meta.Icon != expectedChild.icon {
				t.Errorf("unexpected system management child %q: %#v", expectedChild.name, child)
			}
		}
		for _, authorityID := range expectedGroup.authors {
			assertAuthorityMenuRelation(t, db, group.ID, authorityID, 1)
		}
	}

	for _, authorityID := range []string{"100", "200", "300", "400", "500", "600"} {
		assertAuthorityMenuRelation(t, db, systemParent.ID, authorityID, 1)
	}
	// 显式授予的父级权限未关联已迁出的旧菜单，迁移不能静默删除。
	assertAuthorityMenuRelation(t, db, systemParent.ID, "999", 1)
	assertAuthorityMenuRelation(t, db, legacyMenus[0].ID, "100", 1)
	assertAuthorityMenuRelation(t, db, legacyMenus[6].ID, "600", 1)
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
