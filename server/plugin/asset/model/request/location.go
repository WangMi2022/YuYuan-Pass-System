package request

import commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type LocationSearch struct {
	commonRequest.PageInfo
	Type    string `json:"type" form:"type"`
	Enabled *bool  `json:"enabled" form:"enabled"`
}

type LocationOptionsSearch struct {
	Type string `json:"type" form:"type"`
}
