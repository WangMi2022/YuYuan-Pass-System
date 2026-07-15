package request

import commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type DocumentSearch struct {
	commonRequest.PageInfo
	FileExt string `json:"fileExt" form:"fileExt"`
}

type UpdateContent struct {
	ID      uint   `json:"ID" form:"ID"`
	Title   string `json:"title" form:"title"`
	Content string `json:"content" form:"content"`
	Remarks string `json:"remarks" form:"remarks"`
}
