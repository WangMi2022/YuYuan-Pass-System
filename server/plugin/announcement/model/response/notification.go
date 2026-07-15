package response

import (
	"time"

	"gorm.io/datatypes"
)

type NotificationItem struct {
	ID          uint           `json:"ID"`
	CreatedAt   time.Time      `json:"CreatedAt"`
	UpdatedAt   time.Time      `json:"UpdatedAt"`
	Title       string         `json:"title"`
	Content     string         `json:"content"`
	Attachments datatypes.JSON `json:"attachments"`
	PublishedAt *time.Time     `json:"publishedAt"`
	Publisher   string         `json:"publisher"`
	IsRead      bool           `json:"isRead"`
}

type NotificationResult struct {
	List        []NotificationItem `json:"list"`
	UnreadCount int64              `json:"unreadCount"`
}
