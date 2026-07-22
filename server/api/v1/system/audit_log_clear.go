package system

import (
	"errors"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

func resolveLogClearRange(req request.LogDeleteReq) (startTime, endTime *time.Time, shouldClear bool, err error) {
	hasTimeRange := req.StartTime != nil || req.EndTime != nil
	if req.ClearAll && hasTimeRange {
		return nil, nil, false, errors.New("全部日志与时间范围不能同时选择")
	}
	if req.StartTime != nil && req.EndTime != nil && !req.StartTime.Before(*req.EndTime) {
		return nil, nil, false, errors.New("日志开始时间必须早于结束时间")
	}
	if req.ClearAll {
		return nil, nil, true, nil
	}
	return req.StartTime, req.EndTime, hasTimeRange, nil
}
