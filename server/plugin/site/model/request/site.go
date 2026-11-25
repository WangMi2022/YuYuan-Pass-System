package request

import commonRequest "github.com/WangMi2022/mit-assets-admin/server/model/common/request"

type SiteSearch struct {
	commonRequest.PageInfo
	Category string `json:"category" form:"category"`
	Enabled  *bool  `json:"enabled" form:"enabled"`
}
