package model

import "github.com/WangMi2022/mit-assets-admin/server/global"

const (
	DefaultSystemName     = "mit-assets-admin"
	DefaultSystemSubtitle = "ASSET CONTROL"
)

// LoginLogo 保存全局品牌名称和当前使用的品牌图标。
type LoginLogo struct {
	global.GVA_MODEL
	Name       string  `json:"name" form:"name" gorm:"size:180;comment:图标名称"`
	URL        string  `json:"url" form:"url" gorm:"size:1000;comment:图标地址"`
	SystemName string  `json:"systemName" form:"systemName" gorm:"size:80;comment:系统名称"`
	Subtitle   *string `json:"subtitle" form:"subtitle" gorm:"size:120;comment:品牌副标题"`
	UpdatedBy  uint    `json:"updatedBy" form:"updatedBy" gorm:"index;comment:配置用户"`
}

func (LoginLogo) TableName() string { return "system_login_logos" }
