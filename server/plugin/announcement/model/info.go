package model

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
	"time"
)

// Info 公告 结构体
type Info struct {
	global.GVA_MODEL
	Title       string         `json:"title" form:"title" gorm:"column:title;comment:公告标题;"`                                              //标题
	Content     string         `json:"content" form:"content" gorm:"column:content;comment:公告内容;type:text;"`                              //内容
	UserID      *int           `json:"userID" form:"userID" gorm:"column:user_id;comment:发布者;"`                                           //作者
	Attachments datatypes.JSON `json:"attachments" form:"attachments" gorm:"column:attachments;comment:相关附件;" swaggertype:"array,object"` //附件
	Status      string         `json:"status" form:"status" gorm:"column:status;size:20;not null;default:published;index;comment:发布状态"`
	PublishedAt *time.Time     `json:"publishedAt" form:"publishedAt" gorm:"column:published_at;index;comment:发布时间"`
}

// TableName 公告 Info自定义表名 gva_announcements_info
func (Info) TableName() string {
	return "gva_announcements_info"
}

// Read 记录每个用户已阅读的公告，保证未读状态可跨设备同步。
type Read struct {
	global.GVA_MODEL
	UserID         uint      `json:"userID" gorm:"column:user_id;not null;uniqueIndex:idx_announcement_user_read;index"`
	AnnouncementID uint      `json:"announcementID" gorm:"column:announcement_id;not null;uniqueIndex:idx_announcement_user_read;index"`
	ReadAt         time.Time `json:"readAt" gorm:"column:read_at;not null"`
}

func (Read) TableName() string { return "gva_announcement_reads" }
