package response

import "time"

type Recurrence struct {
	Enabled   bool   `json:"enabled"`
	Mode      string `json:"mode"`
	Weekdays  []int  `json:"weekdays"`
	MonthDays []int  `json:"monthDays"`
	Weekday   int    `json:"weekday"`
	MonthDay  int    `json:"monthDay"`
}

type WorkSchedule struct {
	ID         uint       `json:"id"`
	Title      string     `json:"title"`
	Date       string     `json:"date"`
	Time       string     `json:"time"`
	Type       string     `json:"type"`
	Note       string     `json:"note"`
	Recurrence Recurrence `json:"recurrence"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

type NotificationItem struct {
	ID           uint      `json:"id"`
	ScheduleID   uint      `json:"scheduleId"`
	OccurrenceAt time.Time `json:"occurrenceAt"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	IsRead       bool      `json:"isRead"`
}

type NotificationResult struct {
	List        []NotificationItem `json:"list"`
	UnreadCount int64              `json:"unreadCount"`
}
