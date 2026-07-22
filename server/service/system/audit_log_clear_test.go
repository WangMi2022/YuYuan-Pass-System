package system

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	systemModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestClearLoginLogsPermanentlyDeletesAllRows(t *testing.T) {
	db := setupAuditLogTestDB(t, &systemModel.SysLoginLog{})
	logs := []systemModel.SysLoginLog{
		{Username: "admin", Status: true},
		{Username: "auditor", Status: false},
	}
	if err := db.Create(&logs).Error; err != nil {
		t.Fatalf("seed login logs: %v", err)
	}
	if err := db.Delete(&logs[0]).Error; err != nil {
		t.Fatalf("soft delete login log: %v", err)
	}

	deleted, err := (&LoginLogService{}).ClearLoginLogs(context.Background())
	if err != nil {
		t.Fatalf("clear login logs: %v", err)
	}
	assertAuditLogsCleared(t, db, &systemModel.SysLoginLog{}, deleted, len(logs))
}

func TestClearOperationRecordsPermanentlyDeletesAllRows(t *testing.T) {
	db := setupAuditLogTestDB(t, &systemModel.SysOperationRecord{})
	records := []systemModel.SysOperationRecord{
		{Method: "GET", Path: "/audit", Status: 200},
		{Method: "DELETE", Path: "/audit", Status: 200},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("seed operation records: %v", err)
	}
	if err := db.Delete(&records[0]).Error; err != nil {
		t.Fatalf("soft delete operation record: %v", err)
	}

	deleted, err := (&OperationRecordService{}).ClearOperationRecords(context.Background())
	if err != nil {
		t.Fatalf("clear operation records: %v", err)
	}
	assertAuditLogsCleared(t, db, &systemModel.SysOperationRecord{}, deleted, len(records))
}

func setupAuditLogTestDB(t *testing.T, model any) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err = db.AutoMigrate(model); err != nil {
		t.Fatalf("migrate audit log table: %v", err)
	}

	previousDB := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = previousDB })
	return db
}

func assertAuditLogsCleared(t *testing.T, db *gorm.DB, model any, deleted int64, expected int) {
	t.Helper()
	if deleted != int64(expected) {
		t.Fatalf("deleted rows = %d, want %d", deleted, expected)
	}
	var remaining int64
	if err := db.Unscoped().Model(model).Count(&remaining).Error; err != nil {
		t.Fatalf("count remaining audit logs: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("remaining rows = %d, want 0", remaining)
	}
}
