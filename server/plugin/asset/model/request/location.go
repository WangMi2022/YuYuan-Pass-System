package request

import commonRequest "github.com/WangMi2022/mit-assets-admin/server/model/common/request"

type LocationSearch struct {
	commonRequest.PageInfo
	Type    string `json:"type" form:"type"`
	Enabled *bool  `json:"enabled" form:"enabled"`
}

type LocationOptionsSearch struct {
	Type string `json:"type" form:"type"`
}
