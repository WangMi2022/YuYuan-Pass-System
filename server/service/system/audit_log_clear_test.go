package system

import (
	"context"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	systemModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemRequest "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
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

	deleted, err := (&LoginLogService{}).ClearLoginLogs(context.Background(), nil, nil)
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

	deleted, err := (&OperationRecordService{}).ClearOperationRecords(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("clear operation records: %v", err)
	}
	assertAuditLogsCleared(t, db, &systemModel.SysOperationRecord{}, deleted, len(records))
}

func TestClearLoginLogsDeletesOnlySelectedTimeRange(t *testing.T) {
	db := setupAuditLogTestDB(t, &systemModel.SysLoginLog{})
	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	logs := []systemModel.SysLoginLog{
		{GVA_MODEL: global.GVA_MODEL{CreatedAt: now.AddDate(0, 0, -60)}, Username: "old"},
		{GVA_MODEL: global.GVA_MODEL{CreatedAt: now.AddDate(0, 0, -20)}, Username: "middle"},
		{GVA_MODEL: global.GVA_MODEL{CreatedAt: now.AddDate(0, 0, -2)}, Username: "recent"},
	}
	if err := db.Create(&logs).Error; err != nil {
		t.Fatalf("seed ranged login logs: %v", err)
	}
	endTime := now.AddDate(0, 0, -30)
	_, total, err := (&LoginLogService{}).GetLoginLogInfoList(systemRequest.SysLoginLogSearch{
		LogTimeRange: commonRequest.LogTimeRange{EndTime: &endTime},
		PageInfo:     commonRequest.PageInfo{Page: 1, PageSize: 1},
	})
	if err != nil || total != 1 {
		t.Fatalf("preview ranged login logs: total=%d err=%v", total, err)
	}
	if err := db.Delete(&logs[0]).Error; err != nil {
		t.Fatalf("soft delete ranged login log: %v", err)
	}

	deleted, err := (&LoginLogService{}).ClearLoginLogs(context.Background(), nil, &endTime)
	if err != nil {
		t.Fatalf("clear ranged login logs: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted rows = %d, want 1", deleted)
	}
	assertRemainingAuditLogNames(t, db, &systemModel.SysLoginLog{}, []string{"middle", "recent"})
}

func TestClearOperationRecordsDeletesOnlySelectedTimeRange(t *testing.T) {
	db := setupAuditLogTestDB(t, &systemModel.SysOperationRecord{})
	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	records := []systemModel.SysOperationRecord{
		{GVA_MODEL: global.GVA_MODEL{CreatedAt: now.AddDate(0, 0, -60)}, Path: "/old"},
		{GVA_MODEL: global.GVA_MODEL{CreatedAt: now.AddDate(0, 0, -20)}, Path: "/middle"},
		{GVA_MODEL: global.GVA_MODEL{CreatedAt: now.AddDate(0, 0, -2)}, Path: "/recent"},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("seed ranged operation records: %v", err)
	}

	startTime := now.AddDate(0, 0, -30)
	endTime := now.AddDate(0, 0, -10)
	_, total, err := (&OperationRecordService{}).GetSysOperationRecordInfoList(systemRequest.SysOperationRecordSearch{
		LogTimeRange: commonRequest.LogTimeRange{StartTime: &startTime, EndTime: &endTime},
		PageInfo:     commonRequest.PageInfo{Page: 1, PageSize: 1},
	})
	if err != nil || total != 1 {
		t.Fatalf("preview ranged operation records: total=%d err=%v", total, err)
	}
	deleted, err := (&OperationRecordService{}).ClearOperationRecords(context.Background(), &startTime, &endTime)
	if err != nil {
		t.Fatalf("clear ranged operation records: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted rows = %d, want 1", deleted)
	}

	var remaining []systemModel.SysOperationRecord
	if err := db.Order("id").Find(&remaining).Error; err != nil {
		t.Fatalf("find remaining operation records: %v", err)
	}
	if len(remaining) != 2 || remaining[0].Path != "/old" || remaining[1].Path != "/recent" {
		t.Fatalf("unexpected remaining operation records: %#v", remaining)
	}
}

func TestClearSysErrorsDeletesOnlySelectedTimeRange(t *testing.T) {
	db := setupAuditLogTestDB(t, &systemModel.SysError{})
	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	oldInfo, middleInfo, recentInfo := "old", "middle", "recent"
	errors := []systemModel.SysError{
		{GVA_MODEL: global.GVA_MODEL{CreatedAt: now.AddDate(0, 0, -60)}, Info: &oldInfo},
		{GVA_MODEL: global.GVA_MODEL{CreatedAt: now.AddDate(0, 0, -20)}, Info: &middleInfo},
		{GVA_MODEL: global.GVA_MODEL{CreatedAt: now.AddDate(0, 0, -2)}, Info: &recentInfo},
	}
	if err := db.Create(&errors).Error; err != nil {
		t.Fatalf("seed ranged error logs: %v", err)
	}

	endTime := now.AddDate(0, 0, -30)
	_, total, err := (&SysErrorService{}).GetSysErrorInfoList(context.Background(), systemRequest.SysErrorSearch{
		LogTimeRange: commonRequest.LogTimeRange{EndTime: &endTime},
		PageInfo:     commonRequest.PageInfo{Page: 1, PageSize: 1},
	})
	if err != nil || total != 1 {
		t.Fatalf("preview ranged error logs: total=%d err=%v", total, err)
	}
	deleted, err := (&SysErrorService{}).ClearSysErrors(context.Background(), nil, &endTime)
	if err != nil {
		t.Fatalf("clear ranged error logs: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted rows = %d, want 1", deleted)
	}

	var remaining []systemModel.SysError
	if err := db.Order("id").Find(&remaining).Error; err != nil {
		t.Fatalf("find remaining error logs: %v", err)
	}
	if len(remaining) != 2 || remaining[0].Info == nil || *remaining[0].Info != "middle" ||
		remaining[1].Info == nil || *remaining[1].Info != "recent" {
		t.Fatalf("unexpected remaining error logs: %#v", remaining)
	}
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

func assertRemainingAuditLogNames(t *testing.T, db *gorm.DB, model any, expected []string) {
	t.Helper()
	var remaining []systemModel.SysLoginLog
	if err := db.Model(model).Order("id").Find(&remaining).Error; err != nil {
		t.Fatalf("find remaining audit logs: %v", err)
	}
	if len(remaining) != len(expected) {
		t.Fatalf("remaining rows = %d, want %d", len(remaining), len(expected))
	}
	for index, name := range expected {
		if remaining[index].Username != name {
			t.Fatalf("remaining row %d username = %q, want %q", index, remaining[index].Username, name)
		}
	}
}
