package request

import (
	"strings"

	commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
)

type AssetRecognitionSearch struct {
	commonRequest.PageInfo
	Status string `json:"status" form:"status"`
}

func (search *AssetRecognitionSearch) Normalize() {
	search.Status = strings.TrimSpace(search.Status)
	if search.Page <= 0 {
		search.Page = 1
	}
	if search.PageSize <= 0 || search.PageSize > 100 {
		search.PageSize = 20
	}
}

type AssetRecognitionID struct {
	ID uint `json:"id" form:"id"`
}

type AssetRecognitionDraftUpdate struct {
	ID    uint                        `json:"id"`
	Draft model.AssetRecognitionDraft `json:"draft"`
}
