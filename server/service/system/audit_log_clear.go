package system

import (
	"context"
	"errors"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/gorm"
)

var errInvalidAuditLogTimeRange = errors.New("日志开始时间必须早于结束时间")

func applyAuditLogTimeRange(db *gorm.DB, startTime, endTime *time.Time) *gorm.DB {
	if startTime != nil {
		db = db.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		db = db.Where("created_at <= ?", *endTime)
	}
	return db
}

func clearAuditLogs(ctx context.Context, model any, startTime, endTime *time.Time) (deleted int64, err error) {
	if startTime != nil && endTime != nil && !startTime.Before(*endTime) {
		return 0, errInvalidAuditLogTimeRange
	}
	db := global.GVA_DB.WithContext(ctx).Unscoped()
	result := applyAuditLogTimeRange(db, startTime, endTime).Where("1 = 1").Delete(model)
	return result.RowsAffected, result.Error
}
