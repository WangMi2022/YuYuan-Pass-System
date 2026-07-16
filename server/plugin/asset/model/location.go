package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

const (
	LocationTypeInbound     = "inbound"
	LocationTypeUsage       = "usage"
	LocationTypeTransfer    = "transfer"
	LocationTypeReturn      = "return"
	LocationTypeMaintenance = "maintenance"
	LocationTypeDisposal    = "disposal"
)

// Location 是按资产业务环节维护的可选位置字典。
type Location struct {
	global.GVA_MODEL
	Name        string `json:"name" form:"name" gorm:"size:150;not null;uniqueIndex:idx_asset_location_type_name;comment:位置名称"`
	Type        string `json:"type" form:"type" gorm:"size:30;not null;uniqueIndex:idx_asset_location_type_name;index;comment:位置类型"`
	Code        string `json:"code" form:"code" gorm:"size:50;comment:位置编码"`
	Description string `json:"description" form:"description" gorm:"size:500;comment:位置说明"`
	Sort        int    `json:"sort" form:"sort" gorm:"default:0;comment:排序"`
	Enabled     bool   `json:"enabled" form:"enabled" gorm:"not null;comment:是否启用"`
}

func (Location) TableName() string { return "asset_locations" }
