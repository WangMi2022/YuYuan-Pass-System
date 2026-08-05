package service

import (
	"context"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/schedule/model"
	scheduleRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/schedule/model/request"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestScheduleFromPayloadNormalizesRecurrence(t *testing.T) {
	payload := scheduleRequest.WorkScheduleUpsert{
		Title: "  月度盘点  ",
		Date:  "2026-08-05",
		Time:  "09:30",
		Type:  "asset",
		Note:  "  核对资产状态  ",
		Recurrence: scheduleRequest.Recurrence{
			Enabled:  true,
			Mode:     model.RecurrenceMonthly,
			MonthDay: 5,
		},
	}

	schedule, err := scheduleFromPayload(7, payload)
	if err != nil {
		t.Fatalf("scheduleFromPayload() error = %v", err)
	}
	if schedule.UserID != 7 || schedule.Title != "月度盘点" || schedule.Note != "核对资产状态" {
		t.Fatalf("unexpected normalized schedule: %#v", schedule)
	}
	if !schedule.RecurrenceEnabled || schedule.RecurrenceMode != model.RecurrenceMonthly || schedule.RecurrenceMonthDay != 5 {
		t.Fatalf("unexpected recurrence: %#v", schedule)
	}
}

func TestOccurrenceAtRespectsStartDateAndRule(t *testing.T) {
	schedule := model.WorkSchedule{
		ScheduleDate:       "2026-08-03",
		ScheduleTime:       "09:15",
		RecurrenceEnabled:  true,
		RecurrenceMode:     model.RecurrenceWeekly,
		RecurrenceWeekday:  1,
		RecurrenceMonthDay: 3,
	}
	beforeStart := time.Date(2026, 7, 27, 0, 0, 0, 0, time.Local)
	if _, ok := occurrenceAt(schedule, beforeStart); ok {
		t.Fatal("recurring schedule matched before its start date")
	}
	matchingDay := time.Date(2026, 8, 10, 0, 0, 0, 0, time.Local)
	occurrence, ok := occurrenceAt(schedule, matchingDay)
	if !ok || occurrence.Format("2006-01-02 15:04") != "2026-08-10 09:15" {
		t.Fatalf("unexpected weekly occurrence: %v, %t", occurrence, ok)
	}
}

func TestScanDueNotificationsIsIdempotent(t *testing.T) {
	originalDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = originalDB })

	db, err := gorm.Open(sqlite.Open("file:work-schedule-reminder?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err = db.AutoMigrate(&model.WorkSchedule{}, &model.WorkScheduleNotification{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	global.GVA_DB = db

	now := time.Date(2026, 8, 5, 9, 16, 0, 0, time.Local)
	schedule := model.WorkSchedule{
		UserID:             42,
		ClientKey:          "test-schedule",
		Title:              "晨会",
		ScheduleDate:       "2026-08-05",
		ScheduleTime:       "09:15",
		Type:               "meeting",
		RecurrenceMode:     model.RecurrenceWeekly,
		RecurrenceWeekday:  3,
		RecurrenceMonthDay: 5,
	}
	if err = db.Create(&schedule).Error; err != nil {
		t.Fatalf("create schedule: %v", err)
	}

	first, err := WorkSchedule.ScanDueNotifications(context.Background(), now)
	if err != nil {
		t.Fatalf("first scan: %v", err)
	}
	if len(first) != 1 {
		t.Fatalf("first scan created %d notifications, want 1", len(first))
	}
	if first[0].Title != "晨会" || first[0].UserID != 42 {
		t.Fatalf("unexpected notification: %#v", first[0])
	}

	second, err := WorkSchedule.ScanDueNotifications(context.Background(), now)
	if err != nil {
		t.Fatalf("second scan: %v", err)
	}
	if len(second) != 0 {
		t.Fatalf("second scan created %d duplicate notifications", len(second))
	}
}

func TestScanDueNotificationsBackfillsOnlyTodaysOverdueSchedules(t *testing.T) {
	originalDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = originalDB })

	db, err := gorm.Open(sqlite.Open("file:work-schedule-today-backfill?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err = db.AutoMigrate(&model.WorkSchedule{}, &model.WorkScheduleNotification{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	global.GVA_DB = db

	schedules := []model.WorkSchedule{
		{
			UserID:             42,
			ClientKey:          "today-overdue",
			Title:              "今天已过点",
			ScheduleDate:       "2026-08-05",
			ScheduleTime:       "09:15",
			Type:               "meeting",
			RecurrenceMode:     model.RecurrenceWeekly,
			RecurrenceWeekday:  3,
			RecurrenceMonthDay: 5,
		},
		{
			UserID:             42,
			ClientKey:          "today-future",
			Title:              "今天尚未到点",
			ScheduleDate:       "2026-08-05",
			ScheduleTime:       "15:00",
			Type:               "meeting",
			RecurrenceMode:     model.RecurrenceWeekly,
			RecurrenceWeekday:  3,
			RecurrenceMonthDay: 5,
		},
		{
			UserID:             42,
			ClientKey:          "yesterday-overdue",
			Title:              "昨天已过点",
			ScheduleDate:       "2026-08-04",
			ScheduleTime:       "09:15",
			Type:               "meeting",
			RecurrenceMode:     model.RecurrenceWeekly,
			RecurrenceWeekday:  2,
			RecurrenceMonthDay: 4,
		},
	}
	if err = db.Create(&schedules).Error; err != nil {
		t.Fatalf("create schedules: %v", err)
	}

	now := time.Date(2026, 8, 5, 14, 19, 0, 0, time.Local)
	created, err := WorkSchedule.ScanDueNotifications(context.Background(), now)
	if err != nil {
		t.Fatalf("scan reminders: %v", err)
	}
	if len(created) != 1 || created[0].Title != "今天已过点" {
		t.Fatalf("created reminders = %#v, want only today's overdue schedule", created)
	}
}

func TestUpdateClearsPreviousNotifications(t *testing.T) {
	originalDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = originalDB })

	db, err := gorm.Open(sqlite.Open("file:work-schedule-update?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err = db.AutoMigrate(&model.WorkSchedule{}, &model.WorkScheduleNotification{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	global.GVA_DB = db

	schedule := model.WorkSchedule{
		UserID:             42,
		ClientKey:          "schedule-to-update",
		Title:              "原日程",
		ScheduleDate:       "2026-08-05",
		ScheduleTime:       "09:15",
		Type:               "meeting",
		RecurrenceMode:     model.RecurrenceWeekly,
		RecurrenceWeekday:  3,
		RecurrenceMonthDay: 5,
	}
	if err = db.Create(&schedule).Error; err != nil {
		t.Fatalf("create schedule: %v", err)
	}
	if err = db.Create(&model.WorkScheduleNotification{
		UserID:       schedule.UserID,
		ScheduleID:   schedule.ID,
		OccurrenceAt: time.Date(2026, 8, 5, 9, 15, 0, 0, time.Local),
		Title:        schedule.Title,
	}).Error; err != nil {
		t.Fatalf("create notification: %v", err)
	}

	_, err = WorkSchedule.Update(context.Background(), schedule.UserID, scheduleRequest.WorkScheduleUpsert{
		ID:    schedule.ID,
		Title: "已修改日程",
		Date:  "2026-08-05",
		Time:  "10:00",
		Type:  "meeting",
	})
	if err != nil {
		t.Fatalf("update schedule: %v", err)
	}

	var count int64
	if err = db.Unscoped().Model(&model.WorkScheduleNotification{}).
		Where("user_id = ? AND schedule_id = ?", schedule.UserID, schedule.ID).
		Count(&count).Error; err != nil {
		t.Fatalf("count notifications: %v", err)
	}
	if count != 0 {
		t.Fatalf("notifications after schedule update = %d, want 0", count)
	}
}

func TestDeleteClearsPreviousNotifications(t *testing.T) {
	originalDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = originalDB })

	db, err := gorm.Open(sqlite.Open("file:work-schedule-delete?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err = db.AutoMigrate(&model.WorkSchedule{}, &model.WorkScheduleNotification{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	global.GVA_DB = db

	schedule := model.WorkSchedule{
		UserID:             42,
		ClientKey:          "schedule-to-delete",
		Title:              "待删除日程",
		ScheduleDate:       "2026-08-05",
		ScheduleTime:       "09:15",
		Type:               "meeting",
		RecurrenceMode:     model.RecurrenceWeekly,
		RecurrenceWeekday:  3,
		RecurrenceMonthDay: 5,
	}
	if err = db.Create(&schedule).Error; err != nil {
		t.Fatalf("create schedule: %v", err)
	}
	if err = db.Create(&model.WorkScheduleNotification{
		UserID:       schedule.UserID,
		ScheduleID:   schedule.ID,
		OccurrenceAt: time.Date(2026, 8, 5, 9, 15, 0, 0, time.Local),
		Title:        schedule.Title,
	}).Error; err != nil {
		t.Fatalf("create notification: %v", err)
	}

	if err = WorkSchedule.Delete(context.Background(), schedule.UserID, schedule.ID); err != nil {
		t.Fatalf("delete schedule: %v", err)
	}

	var count int64
	if err = db.Unscoped().Model(&model.WorkScheduleNotification{}).
		Where("user_id = ? AND schedule_id = ?", schedule.UserID, schedule.ID).
		Count(&count).Error; err != nil {
		t.Fatalf("count notifications: %v", err)
	}
	if count != 0 {
		t.Fatalf("notifications after schedule delete = %d, want 0", count)
	}
}
