package request

import (
	"github.com/WangMi2022/mit-assets-admin/server/model/common/request"
	"time"
)

type InfoSearch struct {
	StartCreatedAt *time.Time `json:"startCreatedAt" form:"startCreatedAt"`
	EndCreatedAt   *time.Time `json:"endCreatedAt" form:"endCreatedAt"`
	Status         string     `json:"status" form:"status"`
	request.PageInfo
}

type ReadInfo struct {
	ID uint `json:"id" binding:"required"`
}
