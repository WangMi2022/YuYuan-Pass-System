package request

type Recurrence struct {
	Enabled   bool   `json:"enabled"`
	Mode      string `json:"mode"`
	Weekdays  []int  `json:"weekdays"`
	MonthDays []int  `json:"monthDays"`
	Weekday   int    `json:"weekday"`
	MonthDay  int    `json:"monthDay"`
}

type WorkScheduleUpsert struct {
	ID         uint       `json:"id"`
	ClientKey  string     `json:"clientKey"`
	Title      string     `json:"title"`
	Date       string     `json:"date"`
	Time       string     `json:"time"`
	Type       string     `json:"type"`
	Note       string     `json:"note"`
	Recurrence Recurrence `json:"recurrence"`
}

type LegacyScheduleImport struct {
	Schedules []WorkScheduleUpsert `json:"schedules"`
}

type NotificationRead struct {
	ID uint `json:"id" binding:"required"`
}
