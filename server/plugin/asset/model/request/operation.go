package request

import (
	"time"

	commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

type OperationSearch struct {
	commonRequest.PageInfo
	Type      string `json:"type" form:"type"`
	Status    string `json:"status" form:"status"`
	StartDate string `json:"startDate" form:"startDate"`
	EndDate   string `json:"endDate" form:"endDate"`
}

type SaveOperation struct {
	ID              uint      `json:"ID"`
	Type            string    `json:"type"`
	BusinessDate    time.Time `json:"businessDate"`
	TargetLocation  string    `json:"targetLocation"`
	TargetCustodian string    `json:"targetCustodian"`
	Reason          string    `json:"reason"`
	Remarks         string    `json:"remarks"`
	AssetIDs        []uint    `json:"assetIds"`
	Submit          bool      `json:"submit"`
}

type OperationAssetSearch struct {
	Type    string `json:"type" form:"type"`
	Keyword string `json:"keyword" form:"keyword"`
}
