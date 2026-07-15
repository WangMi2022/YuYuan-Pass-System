package model

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// SiteBookmark 保存工作中常用的 HTTP/HTTPS 站点收藏。
type SiteBookmark struct {
	global.GVA_MODEL
	Name          string     `json:"name" form:"name" gorm:"size:120;not null;index;comment:站点名称"`
	URL           string     `json:"url" form:"url" gorm:"size:800;not null;comment:站点地址"`
	Category      string     `json:"category" form:"category" gorm:"size:80;index;comment:站点分类"`
	Description   string     `json:"description" form:"description" gorm:"size:500;comment:站点说明"`
	Color         string     `json:"color" form:"color" gorm:"size:20;default:#2563eb;comment:展示颜色"`
	Sort          int        `json:"sort" form:"sort" gorm:"default:0;comment:排序"`
	Enabled       bool       `json:"enabled" form:"enabled" gorm:"default:true;comment:是否启用"`
	VisitCount    int        `json:"visitCount" form:"visitCount" gorm:"default:0;comment:访问次数"`
	LastVisitedAt *time.Time `json:"lastVisitedAt" form:"lastVisitedAt" gorm:"comment:最近访问时间"`
}

func (SiteBookmark) TableName() string { return "site_bookmarks" }
