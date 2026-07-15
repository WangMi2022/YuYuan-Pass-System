package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// LoginBackground 保存可供登录页切换的背景图片。
type LoginBackground struct {
	global.GVA_MODEL
	Name      string `json:"name" form:"name" gorm:"size:180;not null;comment:图片名称"`
	URL       string `json:"url" form:"url" gorm:"size:1000;not null;uniqueIndex;comment:图片地址"`
	IsActive  bool   `json:"isActive" form:"isActive" gorm:"default:false;index;comment:是否启用"`
	CreatedBy uint   `json:"createdBy" form:"createdBy" gorm:"index;comment:上传用户"`
}

func (LoginBackground) TableName() string { return "system_login_backgrounds" }
