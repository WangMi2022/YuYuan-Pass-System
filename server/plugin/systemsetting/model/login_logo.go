package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// LoginLogo 保存登录页当前使用的品牌图标。
type LoginLogo struct {
	global.GVA_MODEL
	Name      string `json:"name" form:"name" gorm:"size:180;not null;comment:图标名称"`
	URL       string `json:"url" form:"url" gorm:"size:1000;not null;comment:图标地址"`
	UpdatedBy uint   `json:"updatedBy" form:"updatedBy" gorm:"index;comment:配置用户"`
}

func (LoginLogo) TableName() string { return "system_login_logos" }
