package request

import (
	"github.com/WangMi2022/mit-assets-admin/server/model/common/request"
	"github.com/WangMi2022/mit-assets-admin/server/model/system"
)

type SysLoginLogSearch struct {
	system.SysLoginLog
	request.LogTimeRange
	request.PageInfo
}
