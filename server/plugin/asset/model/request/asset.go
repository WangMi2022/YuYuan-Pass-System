package request

import commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type AssetSearch struct {
	commonRequest.PageInfo
	CategoryID uint    `json:"categoryId" form:"categoryId"`
	Status     string  `json:"status" form:"status"`
	Location   string  `json:"location" form:"location"`
	MinValue   float64 `json:"minValue" form:"minValue"`
	MaxValue   float64 `json:"maxValue" form:"maxValue"`
}

type CategorySearch struct {
	commonRequest.PageInfo
	Enabled *bool `json:"enabled" form:"enabled"`
}
