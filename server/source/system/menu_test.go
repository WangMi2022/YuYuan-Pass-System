package system

import (
	"context"
	"testing"

	systemModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
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
	if permissionParent.ParentId != 0 || permissionParent.MenuLevel != 0 || permissionParent.Sort != 5 ||
		permissionParent.Path != "permissionManagement" || permissionParent.Component != "view/routerHolder.vue" ||
		permissionParent.Meta.Title != "权限管理" {
		t.Fatalf("unexpected permission parent: %#v", permissionParent)
	}

	var systemParent systemModel.SysBaseMenu
	if err = db.Where("name = ?", "superAdmin").First(&systemParent).Error; err != nil {
		t.Fatalf("find system parent: %v", err)
	}
	if systemParent.ParentId != 0 || systemParent.MenuLevel != 0 || systemParent.Sort != 7 {
		t.Fatalf("unexpected system parent: %#v", systemParent)
	}

	var auditParent systemModel.SysBaseMenu
	if err = db.Where("name = ?", "auditPlatform").First(&auditParent).Error; err != nil {
		t.Fatalf("find audit parent: %v", err)
	}
	if auditParent.ParentId != 0 || auditParent.MenuLevel != 0 || auditParent.Sort != 6 ||
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
}
