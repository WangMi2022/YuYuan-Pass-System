package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	announcementService "github.com/flipped-aurora/gin-vue-admin/server/plugin/announcement/service"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/schedule/model"
	scheduleRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/schedule/model/request"
	scheduleResponse "github.com/flipped-aurora/gin-vue-admin/server/plugin/schedule/model/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	maxImportedSchedules = 200
	reminderLookback     = 2 * time.Minute
)

var WorkSchedule = new(workScheduleService)

type workScheduleService struct{}

var reminderWorkerOnce sync.Once

func (s *workScheduleService) List(ctx context.Context, userID uint) ([]scheduleResponse.WorkSchedule, error) {
	if userID == 0 {
		return nil, errors.New("用户未登录")
	}

	var schedules []model.WorkSchedule
	err := global.GVA_DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("schedule_date ASC, schedule_time ASC, id ASC").
		Find(&schedules).Error
	if err != nil {
		return nil, err
	}

	result := make([]scheduleResponse.WorkSchedule, 0, len(schedules))
	for _, item := range schedules {
		result = append(result, toScheduleResponse(item))
	}
	return result, nil
}

func (s *workScheduleService) Create(ctx context.Context, userID uint, payload scheduleRequest.WorkScheduleUpsert) (scheduleResponse.WorkSchedule, error) {
	if userID == 0 {
		return scheduleResponse.WorkSchedule{}, errors.New("用户未登录")
	}
	schedule, err := scheduleFromPayload(userID, payload)
	if err != nil {
		return scheduleResponse.WorkSchedule{}, err
	}
	if schedule.ClientKey == "" {
		schedule.ClientKey = "server-" + uuid.NewString()
	}
	if err = global.GVA_DB.WithContext(ctx).Create(&schedule).Error; err != nil {
		return scheduleResponse.WorkSchedule{}, err
	}
	return toScheduleResponse(schedule), nil
}

func (s *workScheduleService) Update(ctx context.Context, userID uint, payload scheduleRequest.WorkScheduleUpsert) (scheduleResponse.WorkSchedule, error) {
	if userID == 0 || payload.ID == 0 {
		return scheduleResponse.WorkSchedule{}, errors.New("日程参数不正确")
	}

	updated, err := scheduleFromPayload(userID, payload)
	if err != nil {
		return scheduleResponse.WorkSchedule{}, err
	}

	db := global.GVA_DB.WithContext(ctx)
	var schedule model.WorkSchedule
	if err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND user_id = ?", payload.ID, userID).First(&schedule).Error; err != nil {
			return err
		}
		updated.ID = schedule.ID
		updated.ClientKey = schedule.ClientKey
		result := tx.Model(&model.WorkSchedule{}).
			Where("id = ? AND user_id = ?", schedule.ID, userID).
			Updates(scheduleUpdates(updated))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return clearScheduleNotifications(tx, userID, schedule.ID)
	}); err != nil {
		return scheduleResponse.WorkSchedule{}, err
	}
	if err = db.Where("id = ? AND user_id = ?", schedule.ID, userID).First(&schedule).Error; err != nil {
		return scheduleResponse.WorkSchedule{}, err
	}
	return toScheduleResponse(schedule), nil
}

func (s *workScheduleService) Delete(ctx context.Context, userID, id uint) error {
	if userID == 0 || id == 0 {
		return errors.New("日程参数不正确")
	}
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ? AND user_id = ?", id, userID).Delete(&model.WorkSchedule{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return clearScheduleNotifications(tx, userID, id)
	})
}

func clearScheduleNotifications(db *gorm.DB, userID, scheduleID uint) error {
	// Notifications use a unique occurrence key. They must be physically removed
	// so a revised schedule can create a fresh reminder for the same occurrence.
	return db.Unscoped().Where("user_id = ? AND schedule_id = ?", userID, scheduleID).
		Delete(&model.WorkScheduleNotification{}).Error
}

// ImportLegacy creates entries from the browser-only calendar exactly once per
// client key. Existing server entries are deliberately never overwritten.
func (s *workScheduleService) ImportLegacy(ctx context.Context, userID uint, payload scheduleRequest.LegacyScheduleImport) (int, error) {
	if userID == 0 {
		return 0, errors.New("用户未登录")
	}
	if len(payload.Schedules) == 0 {
		return 0, nil
	}
	if len(payload.Schedules) > maxImportedSchedules {
		return 0, fmt.Errorf("一次最多导入 %d 条日程", maxImportedSchedules)
	}

	prepared := make([]model.WorkSchedule, 0, len(payload.Schedules))
	seenKeys := make(map[string]struct{}, len(payload.Schedules))
	for index, item := range payload.Schedules {
		item.ClientKey = strings.TrimSpace(item.ClientKey)
		if item.ClientKey == "" {
			return 0, fmt.Errorf("第 %d 条日程缺少迁移标识", index+1)
		}
		if _, exists := seenKeys[item.ClientKey]; exists {
			continue
		}
		seenKeys[item.ClientKey] = struct{}{}
		schedule, err := scheduleFromPayload(userID, item)
		if err != nil {
			return 0, fmt.Errorf("第 %d 条日程无效：%w", index+1, err)
		}
		prepared = append(prepared, schedule)
	}

	created := 0
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, schedule := range prepared {
			result := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "client_key"}},
				DoNothing: true,
			}).Create(&schedule)
			if result.Error != nil {
				return result.Error
			}
			created += int(result.RowsAffected)
		}
		return nil
	})
	return created, err
}

func (s *workScheduleService) Notifications(ctx context.Context, userID uint, limit int) (scheduleResponse.NotificationResult, error) {
	result := scheduleResponse.NotificationResult{List: make([]scheduleResponse.NotificationItem, 0)}
	if userID == 0 {
		return result, errors.New("用户未登录")
	}
	if limit <= 0 || limit > 50 {
		limit = 12
	}

	db := global.GVA_DB.WithContext(ctx).
		Model(&model.WorkScheduleNotification{}).
		Joins("JOIN work_schedules ON work_schedules.id = work_schedule_notifications.schedule_id AND work_schedules.deleted_at IS NULL").
		Where("work_schedule_notifications.user_id = ?", userID)
	if err := db.Where("read_at IS NULL").Count(&result.UnreadCount).Error; err != nil {
		return result, err
	}

	var notifications []model.WorkScheduleNotification
	if err := db.Order("occurrence_at DESC, id DESC").Limit(limit).Find(&notifications).Error; err != nil {
		return result, err
	}
	result.List = make([]scheduleResponse.NotificationItem, 0, len(notifications))
	for _, item := range notifications {
		result.List = append(result.List, scheduleResponse.NotificationItem{
			ID:           item.ID,
			ScheduleID:   item.ScheduleID,
			OccurrenceAt: item.OccurrenceAt,
			Title:        item.Title,
			Content:      item.Content,
			IsRead:       item.ReadAt != nil,
		})
	}
	return result, nil
}

func (s *workScheduleService) MarkNotificationRead(ctx context.Context, userID, id uint) error {
	if userID == 0 || id == 0 {
		return errors.New("提醒参数不正确")
	}
	now := time.Now()
	result := global.GVA_DB.WithContext(ctx).Model(&model.WorkScheduleNotification{}).
		Where("id = ? AND user_id = ? AND read_at IS NULL", id, userID).
		Update("read_at", now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := global.GVA_DB.WithContext(ctx).Model(&model.WorkScheduleNotification{}).
			Where("id = ? AND user_id = ?", id, userID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}
	}
	return nil
}

func (s *workScheduleService) MarkAllNotificationsRead(ctx context.Context, userID uint) error {
	if userID == 0 {
		return errors.New("用户未登录")
	}
	now := time.Now()
	return global.GVA_DB.WithContext(ctx).Model(&model.WorkScheduleNotification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", now).Error
}

// ScanDueNotifications records all due occurrences in the small lookback
// window. The database unique index is the final guard against duplicate cron
// runs, restarts, and overlapping workers.
func (s *workScheduleService) ScanDueNotifications(ctx context.Context, now time.Time) ([]model.WorkScheduleNotification, error) {
	now = now.In(time.Local)
	windowStart := now.Add(-reminderLookback)
	endDate := now.Format("2006-01-02")

	var schedules []model.WorkSchedule
	if err := global.GVA_DB.WithContext(ctx).
		Where("schedule_date <= ?", endDate).
		Find(&schedules).Error; err != nil {
		return nil, err
	}

	created := make([]model.WorkScheduleNotification, 0)
	for day := startOfDay(windowStart); !day.After(startOfDay(now)); day = day.AddDate(0, 0, 1) {
		for _, schedule := range schedules {
			occurrenceAt, ok := occurrenceAt(schedule, day)
			if !ok || occurrenceAt.Before(windowStart) || occurrenceAt.After(now) {
				continue
			}
			notification := model.WorkScheduleNotification{
				UserID:       schedule.UserID,
				ScheduleID:   schedule.ID,
				OccurrenceAt: occurrenceAt,
				Title:        schedule.Title,
				Content:      reminderContent(schedule),
			}
			result := global.GVA_DB.WithContext(ctx).Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "schedule_id"}, {Name: "occurrence_at"}},
				DoNothing: true,
			}).Create(&notification)
			if result.Error != nil {
				return created, result.Error
			}
			if result.RowsAffected > 0 {
				created = append(created, notification)
			}
		}
	}
	return created, nil
}

func (s *workScheduleService) StartReminderWorker() {
	reminderWorkerOnce.Do(func() {
		run := func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			notifications, err := s.ScanDueNotifications(ctx, time.Now())
			if err != nil {
				global.GVA_LOG.Error("扫描日程提醒失败", zap.Error(err))
				return
			}
			for _, item := range notifications {
				announcementService.NotificationHub.PublishToUser(item.UserID, announcementService.NotificationEvent{
					ID:    item.ID,
					Title: item.Title,
					Kind:  "schedule",
				})
			}
		}

		if _, err := global.GVA_Timer.AddTaskByFuncWithSecond(
			"WorkScheduleReminder",
			"0 * * * * *",
			run,
			"个人日程到点提醒",
		); err != nil {
			global.GVA_LOG.Error("注册日程提醒定时任务失败", zap.Error(err))
			return
		}
		go run()
	})
}

func scheduleFromPayload(userID uint, payload scheduleRequest.WorkScheduleUpsert) (model.WorkSchedule, error) {
	date, err := parseScheduleDate(payload.Date)
	if err != nil {
		return model.WorkSchedule{}, err
	}
	timeValue, err := parseScheduleTime(payload.Time)
	if err != nil {
		return model.WorkSchedule{}, err
	}

	title := strings.TrimSpace(payload.Title)
	if title == "" {
		return model.WorkSchedule{}, errors.New("请填写日程名称")
	}
	if utf8.RuneCountInString(title) > 120 {
		return model.WorkSchedule{}, errors.New("日程名称不能超过 120 个字符")
	}
	typeValue := strings.TrimSpace(payload.Type)
	if typeValue == "" {
		return model.WorkSchedule{}, errors.New("请选择日程类型")
	}
	if utf8.RuneCountInString(typeValue) > 64 {
		return model.WorkSchedule{}, errors.New("日程类型不能超过 64 个字符")
	}
	note := strings.TrimSpace(payload.Note)
	if utf8.RuneCountInString(note) > 500 {
		return model.WorkSchedule{}, errors.New("备注不能超过 500 个字符")
	}
	clientKey := strings.TrimSpace(payload.ClientKey)
	if utf8.RuneCountInString(clientKey) > 128 {
		return model.WorkSchedule{}, errors.New("客户端标识不能超过 128 个字符")
	}

	recurrence := payload.Recurrence
	if recurrence.Mode != model.RecurrenceMonthly {
		recurrence.Mode = model.RecurrenceWeekly
	}
	if recurrence.Weekday < 1 || recurrence.Weekday > 7 {
		recurrence.Weekday = weekdayFromDate(date)
	}
	if recurrence.MonthDay < 1 || recurrence.MonthDay > 31 {
		recurrence.MonthDay = date.Day()
	}

	return model.WorkSchedule{
		UserID:             userID,
		ClientKey:          clientKey,
		Title:              title,
		ScheduleDate:       date.Format("2006-01-02"),
		ScheduleTime:       timeValue.Format("15:04"),
		Type:               typeValue,
		Note:               note,
		RecurrenceEnabled:  recurrence.Enabled,
		RecurrenceMode:     recurrence.Mode,
		RecurrenceWeekday:  recurrence.Weekday,
		RecurrenceMonthDay: recurrence.MonthDay,
	}, nil
}

func scheduleUpdates(schedule model.WorkSchedule) map[string]any {
	return map[string]any{
		"title":                schedule.Title,
		"schedule_date":        schedule.ScheduleDate,
		"schedule_time":        schedule.ScheduleTime,
		"type":                 schedule.Type,
		"note":                 schedule.Note,
		"recurrence_enabled":   schedule.RecurrenceEnabled,
		"recurrence_mode":      schedule.RecurrenceMode,
		"recurrence_weekday":   schedule.RecurrenceWeekday,
		"recurrence_month_day": schedule.RecurrenceMonthDay,
	}
}

func toScheduleResponse(item model.WorkSchedule) scheduleResponse.WorkSchedule {
	return scheduleResponse.WorkSchedule{
		ID:    item.ID,
		Title: item.Title,
		Date:  item.ScheduleDate,
		Time:  item.ScheduleTime,
		Type:  item.Type,
		Note:  item.Note,
		Recurrence: scheduleResponse.Recurrence{
			Enabled:  item.RecurrenceEnabled,
			Mode:     item.RecurrenceMode,
			Weekday:  item.RecurrenceWeekday,
			MonthDay: item.RecurrenceMonthDay,
		},
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func parseScheduleDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	date, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil || date.Format("2006-01-02") != value {
		return time.Time{}, errors.New("日期格式不正确")
	}
	return date, nil
}

func parseScheduleTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	parsed, err := time.ParseInLocation("15:04", value, time.Local)
	if err != nil || parsed.Format("15:04") != value {
		return time.Time{}, errors.New("时间格式不正确")
	}
	return parsed, nil
}

func startOfDay(value time.Time) time.Time {
	local := value.In(time.Local)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.Local)
}

func weekdayFromDate(value time.Time) int {
	weekday := int(value.Weekday())
	if weekday == 0 {
		return 7
	}
	return weekday
}

func occurrenceAt(schedule model.WorkSchedule, day time.Time) (time.Time, bool) {
	startDate, err := parseScheduleDate(schedule.ScheduleDate)
	if err != nil || day.Before(startOfDay(startDate)) {
		return time.Time{}, false
	}
	if !schedule.RecurrenceEnabled && !day.Equal(startOfDay(startDate)) {
		return time.Time{}, false
	}
	if schedule.RecurrenceEnabled {
		switch schedule.RecurrenceMode {
		case model.RecurrenceMonthly:
			if day.Day() != schedule.RecurrenceMonthDay {
				return time.Time{}, false
			}
		default:
			if weekdayFromDate(day) != schedule.RecurrenceWeekday {
				return time.Time{}, false
			}
		}
	}

	parsedTime, err := parseScheduleTime(schedule.ScheduleTime)
	if err != nil {
		return time.Time{}, false
	}
	return time.Date(day.Year(), day.Month(), day.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, time.Local), true
}

func reminderContent(schedule model.WorkSchedule) string {
	if schedule.Note != "" {
		return schedule.Note
	}
	return fmt.Sprintf("日程已到设定时间（%s），请及时处理。", schedule.ScheduleTime)
}
