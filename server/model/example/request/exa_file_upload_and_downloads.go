package request

import (
	"github.com/WangMi2022/mit-assets-admin/server/model/common/request"
)

type ExaAttachmentCategorySearch struct {
	ClassId  int    `json:"classId" form:"classId"`
	FileType string `json:"fileType" form:"fileType"`
	request.PageInfo
}
