package system

import (
	"context"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
)

type LoginLogService struct{}

var LoginLogServiceApp = new(LoginLogService)

func (loginLogService *LoginLogService) CreateLoginLog(loginLog system.SysLoginLog) (err error) {
	err = global.GVA_DB.Create(&loginLog).Error
	return err
}

func (loginLogService *LoginLogService) DeleteLoginLogByIds(ids request.IdsReq) (err error) {
	err = global.GVA_DB.Delete(&[]system.SysLoginLog{}, "id in (?)", ids.Ids).Error
	return err
}

// ClearLoginLogs 永久清理指定时间范围内的登录日志；边界均为空时清空全部。
func (loginLogService *LoginLogService) ClearLoginLogs(ctx context.Context, startTime, endTime *time.Time) (deleted int64, err error) {
	return clearAuditLogs(ctx, &system.SysLoginLog{}, startTime, endTime)
}

func (loginLogService *LoginLogService) DeleteLoginLog(loginLog system.SysLoginLog) (err error) {
	err = global.GVA_DB.Delete(&loginLog).Error
	return err
}

func (loginLogService *LoginLogService) GetLoginLog(id uint) (loginLog system.SysLoginLog, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&loginLog).Error
	return
}

func (loginLogService *LoginLogService) GetLoginLogInfoList(info systemReq.SysLoginLogSearch) (list interface{}, total int64, err error) {
	// 创建db
	db := global.GVA_DB.Model(&system.SysLoginLog{})
	var loginLogs []system.SysLoginLog
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.Username != "" {
		db = db.Where("username LIKE ?", "%"+info.Username+"%")
	}
	if info.Status != false {
		db = db.Where("status = ?", info.Status)
	}
	db = applyAuditLogTimeRange(db, info.StartTime, info.EndTime)
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Scopes(info.Paginate()).Order("id desc").Preload("User").Find(&loginLogs).Error
	return loginLogs, total, err
}
