package request

import (
	"errors"
	"strings"
	"time"

	commonRequest "github.com/WangMi2022/mit-assets-admin/server/model/common/request"
)

type QualitySearch struct {
	commonRequest.PageInfo
	StartDate *time.Time `json:"startDate" form:"startDate" time_format:"2006-01-02"`
	EndDate   *time.Time `json:"endDate" form:"endDate" time_format:"2006-01-02"`
	Provider  string     `json:"provider" form:"provider"`
	Model     string     `json:"model" form:"model"`
	FileType  string     `json:"fileType" form:"fileType"`
}

func (search *QualitySearch) Normalize() error {
	now := time.Now()
	if search.StartDate == nil || search.StartDate.IsZero() {
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -29)
		search.StartDate = &start
	}
	if search.EndDate == nil || search.EndDate.IsZero() {
		end := now
		search.EndDate = &end
	} else {
		end := time.Date(search.EndDate.Year(), search.EndDate.Month(), search.EndDate.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), search.EndDate.Location())
		search.EndDate = &end
	}
	if search.EndDate.Before(*search.StartDate) {
		return errors.New("结束日期不能早于开始日期")
	}
	if search.EndDate.Sub(*search.StartDate) > 366*24*time.Hour {
		return errors.New("质量统计时间范围不能超过 366 天")
	}
	search.Provider = strings.TrimSpace(search.Provider)
	search.Model = strings.TrimSpace(search.Model)
	search.FileType = strings.TrimSpace(search.FileType)
	if search.Page <= 0 {
		search.Page = 1
	}
	if search.PageSize <= 0 || search.PageSize > 100 {
		search.PageSize = 20
	}
	return nil
}
