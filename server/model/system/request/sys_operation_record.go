package request

import (
	"github.com/WangMi2022/mit-assets-admin/server/model/common/request"
	"github.com/WangMi2022/mit-assets-admin/server/model/system"
)

type SysOperationRecordSearch struct {
	system.SysOperationRecord
	request.LogTimeRange
	request.PageInfo
}
