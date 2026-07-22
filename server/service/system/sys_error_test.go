package system

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	systemModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestClearSysErrorsPermanentlyDeletesAllRows(t *testing.T) {
	db, err := gorm.Open(
		sqlite.Open("file:sys-error-clear?mode=memory&cache=shared"),
		&gorm.Config{TranslateError: true},
	)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err = db.AutoMigrate(&systemModel.SysError{}); err != nil {
		t.Fatalf("migrate sys_error: %v", err)
	}

	previousDB := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = previousDB })

	form := "后端"
	info := "测试错误"
	logs := []systemModel.SysError{
		{Form: &form, Info: &info, Level: "fatal"},
		{Form: &form, Info: &info, Level: "error"},
		{Form: &form, Info: &info, Level: "error"},
	}
	if err = db.Create(&logs).Error; err != nil {
		t.Fatalf("seed sys_error: %v", err)
	}
	if err = db.Delete(&logs[0]).Error; err != nil {
		t.Fatalf("soft delete sys_error: %v", err)
	}

	deleted, err := (&SysErrorService{}).ClearSysErrors(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("clear sys_error: %v", err)
	}
	if deleted != int64(len(logs)) {
		t.Fatalf("deleted rows = %d, want %d", deleted, len(logs))
	}

	var remaining int64
	if err = db.Unscoped().Model(&systemModel.SysError{}).Count(&remaining).Error; err != nil {
		t.Fatalf("count remaining sys_error: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("remaining rows = %d, want 0", remaining)
	}
}
