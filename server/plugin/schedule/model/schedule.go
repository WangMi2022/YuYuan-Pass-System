package model

import (
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/global"
)

const (
	RecurrenceDaily   = "daily"
	RecurrenceWeekly  = "weekly"
	RecurrenceMonthly = "monthly"
)

// WorkSchedule is a personal calendar entry. Ownership is always resolved from
// the authenticated user and is never accepted from a client request.
type WorkSchedule struct {
	global.GVA_MODEL
	UserID              uint   `json:"-" gorm:"column:user_id;not null;index;uniqueIndex:idx_work_schedule_user_client;comment:所属用户"`
	ClientKey           string `json:"-" gorm:"column:client_key;size:128;not null;uniqueIndex:idx_work_schedule_user_client;comment:客户端迁移标识"`
	Title               string `json:"title" gorm:"column:title;size:120;not null;comment:日程标题"`
	ScheduleDate        string `json:"date" gorm:"column:schedule_date;size:10;not null;index;comment:开始日期"`
	ScheduleTime        string `json:"time" gorm:"column:schedule_time;size:5;not null;comment:触发时间"`
	Type                string `json:"type" gorm:"column:type;size:64;not null;index;comment:日程类型"`
	Note                string `json:"note" gorm:"column:note;size:500;comment:备注"`
	RecurrenceEnabled   bool   `json:"-" gorm:"column:recurrence_enabled;not null;default:false;comment:是否重复"`
	RecurrenceMode      string `json:"-" gorm:"column:recurrence_mode;size:16;not null;default:weekly;comment:重复周期"`
	RecurrenceWeekday   int    `json:"-" gorm:"column:recurrence_weekday;not null;default:1;comment:每周星期"`
	RecurrenceMonthDay  int    `json:"-" gorm:"column:recurrence_month_day;not null;default:1;comment:每月日期"`
	RecurrenceWeekdays  string `json:"-" gorm:"column:recurrence_weekdays;type:text;comment:每周多选日期"`
	RecurrenceMonthDays string `json:"-" gorm:"column:recurrence_month_days;type:text;comment:每月多选日期"`
}

func (WorkSchedule) TableName() string { return "work_schedules" }

// WorkScheduleNotification is a durable inbox entry for one due occurrence.
// The compound unique index makes reminder generation idempotent across cron
// retries and process restarts.
type WorkScheduleNotification struct {
	global.GVA_MODEL
	UserID       uint       `json:"-" gorm:"column:user_id;not null;index;uniqueIndex:idx_work_schedule_notification_occurrence;comment:接收用户"`
	ScheduleID   uint       `json:"scheduleId" gorm:"column:schedule_id;not null;index;uniqueIndex:idx_work_schedule_notification_occurrence;comment:日程ID"`
	OccurrenceAt time.Time  `json:"occurrenceAt" gorm:"column:occurrence_at;not null;index;uniqueIndex:idx_work_schedule_notification_occurrence;comment:发生时间"`
	Title        string     `json:"title" gorm:"column:title;size:120;not null;comment:提醒标题"`
	Content      string     `json:"content" gorm:"column:content;size:500;comment:提醒内容"`
	ReadAt       *time.Time `json:"readAt" gorm:"column:read_at;index;comment:已读时间"`
}

func (WorkScheduleNotification) TableName() string { return "work_schedule_notifications" }
