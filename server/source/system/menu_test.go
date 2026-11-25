package system

import (
	"context"
	"testing"

	systemModel "github.com/WangMi2022/mit-assets-admin/server/model/system"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestInitMenuCreatesTopLevelPermissionManagement(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err = db.AutoMigrate(
		&systemModel.SysBaseMenu{},
		&systemModel.SysBaseMenuParameter{},
		&systemModel.SysBaseMenuBtn{},
	); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	ctx := context.WithValue(context.Background(), "db", db)
	if _, err = (&initMenu{}).InitializeData(ctx); err != nil {
		t.Fatalf("initialize menus: %v", err)
	}

	var permissionParent systemModel.SysBaseMenu
	if err = db.Where("name = ?", "permissionManagement").First(&permissionParent).Error; err != nil {
		t.Fatalf("find permission parent: %v", err)
	}
	if permissionParent.ParentId != 0 || permissionParent.MenuLevel != 0 || permissionParent.Sort != 7 ||
		permissionParent.Path != "permissionManagement" || permissionParent.Component != "view/routerHolder.vue" ||
		permissionParent.Meta.Title != "权限管理" {
		t.Fatalf("unexpected permission parent: %#v", permissionParent)
	}

	var systemParent systemModel.SysBaseMenu
	if err = db.Where("name = ?", "superAdmin").First(&systemParent).Error; err != nil {
		t.Fatalf("find system parent: %v", err)
	}
	if systemParent.ParentId != 0 || systemParent.MenuLevel != 0 || systemParent.Sort != 9 {
		t.Fatalf("unexpected system parent: %#v", systemParent)
	}

	expectedSystemGroups := map[string]struct {
		title    string
		sort     int
		children map[string]int
	}{
		"systemData": {
			title: "基础数据", sort: 1,
			children: map[string]int{"dictionary": 1, "sysParams": 2},
		},
		"systemConfiguration": {
			title: "平台设置", sort: 2,
			children: map[string]int{"system": 2},
		},
		"systemIntegration": {
			title: "开放与运维", sort: 3,
			children: map[string]int{"apiToken": 1, "sysVersion": 2},
		},
		"systemIntelligence": {
			title: "智能服务", sort: 4,
			children: map[string]int{},
		},
	}
	for name, expected := range expectedSystemGroups {
		var group systemModel.SysBaseMenu
		if err = db.Where("name = ?", name).First(&group).Error; err != nil {
			t.Fatalf("find system management group %q: %v", name, err)
		}
		if group.ParentId != systemParent.ID || group.MenuLevel != 1 || group.Path != name || group.Component != "view/routerHolder.vue" || group.Sort != expected.sort || group.Meta.Title != expected.title {
			t.Errorf("unexpected system management group %q: %#v", name, group)
		}
		for childName, childSort := range expected.children {
			var child systemModel.SysBaseMenu
			if err = db.Where("name = ?", childName).First(&child).Error; err != nil {
				t.Fatalf("find system management child %q: %v", childName, err)
			}
			if child.ParentId != group.ID || child.MenuLevel != 2 || child.Sort != childSort {
				t.Errorf("unexpected system management child %q: %#v", childName, child)
			}
		}
	}

	var auditParent systemModel.SysBaseMenu
	if err = db.Where("name = ?", "auditPlatform").First(&auditParent).Error; err != nil {
		t.Fatalf("find audit parent: %v", err)
	}
	if auditParent.ParentId != 0 || auditParent.MenuLevel != 0 || auditParent.Sort != 8 ||
		auditParent.Path != "auditPlatform" || auditParent.Component != "view/routerHolder.vue" ||
		auditParent.Meta.Title != "审计平台" {
		t.Fatalf("unexpected audit parent: %#v", auditParent)
	}

	expectedChildren := map[string]int{
		"user":      1,
		"authority": 2,
		"api":       3,
		"menu":      4,
	}
	for name, sort := range expectedChildren {
		var child systemModel.SysBaseMenu
		if err = db.Where("name = ?", name).First(&child).Error; err != nil {
			t.Fatalf("find permission child %q: %v", name, err)
		}
		if child.ParentId != permissionParent.ID || child.MenuLevel != 1 || child.Sort != sort {
			t.Errorf("unexpected permission child %q: %#v", name, child)
		}
	}

	expectedAuditChildren := map[string]int{
		"operation": 1,
		"loginLog":  2,
		"sysError":  3,
	}
	for name, sort := range expectedAuditChildren {
		var child systemModel.SysBaseMenu
		if err = db.Where("name = ?", name).First(&child).Error; err != nil {
			t.Fatalf("find audit child %q: %v", name, err)
		}
		if child.ParentId != auditParent.ID || child.MenuLevel != 1 || child.Sort != sort {
			t.Errorf("unexpected audit child %q: %#v", name, child)
		}
	}

	var workCalendarParent systemModel.SysBaseMenu
	if err = db.Where("name = ?", "workCalendar").First(&workCalendarParent).Error; err != nil {
		t.Fatalf("find work calendar parent: %v", err)
	}
	if workCalendarParent.ParentId != 0 || workCalendarParent.MenuLevel != 0 || workCalendarParent.Sort != 4 ||
		workCalendarParent.Path != "workCalendar" || workCalendarParent.Component != "view/routerHolder.vue" || workCalendarParent.Meta.Title != "工作日历" {
		t.Fatalf("unexpected work calendar parent: %#v", workCalendarParent)
	}
	var workSchedule systemModel.SysBaseMenu
	if err = db.Where("name = ?", "workSchedule").First(&workSchedule).Error; err != nil {
		t.Fatalf("find work schedule child: %v", err)
	}
	if workSchedule.ParentId != workCalendarParent.ID || workSchedule.MenuLevel != 1 || workSchedule.Sort != 1 ||
		workSchedule.Component != "view/workCalendar/index.vue" {
		t.Fatalf("unexpected work schedule child: %#v", workSchedule)
	}
}
