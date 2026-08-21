package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/system"
	announcementService "github.com/WangMi2022/mit-assets-admin/server/plugin/announcement/service"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/smart/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	reportDeliveryChannelInApp      = "in_app"
	reportDeliveryChannelEmail      = "email"
	reportDeliveryChannelManualMail = "email_manual"
)

var reportWorkerOnce sync.Once

// StartReportWorker sends enabled subscriptions at their configured local time.
// Delivery rows are unique per user, report and channel, so retries and process
// restarts cannot create duplicate notifications.
func StartReportWorker() {
	reportWorkerOnce.Do(func() {
		if global.GVA_Timer == nil {
			global.GVA_LOG.Warn("定时任务服务不可用，跳过智能日报发送任务")
			return
		}
		if _, err := global.GVA_Timer.AddTaskByFuncWithSecond(
			"SmartDailyReportDelivery", "0 * * * * *", dispatchReports, "智能日报订阅发送",
		); err != nil {
			global.GVA_LOG.Error("注册智能日报发送任务失败", zap.Error(err))
			return
		}
		go dispatchReports()
	})
}

func dispatchReports() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	now := time.Now()
	var subscriptions []model.SmartReportSubscription
	if err := global.GVA_DB.WithContext(ctx).
		Where("enabled = ? AND delivery_time = ?", true, now.Format("15:04")).
		Find(&subscriptions).Error; err != nil {
		global.GVA_LOG.Error("读取智能日报订阅失败", zap.Error(err))
		return
	}
	start, _ := todayRange()
	var failedUserIDs []uint
	if err := global.GVA_DB.WithContext(ctx).Table("smart_report_deliveries AS d").
		Joins("JOIN smart_daily_reports AS r ON r.id = d.report_id").
		Where("d.status = ? AND d.retry_count < ? AND r.report_date = ?", "failed", 3, start).
		Distinct("d.user_id").Pluck("d.user_id", &failedUserIDs).Error; err != nil {
		global.GVA_LOG.Warn("读取智能日报失败投递记录失败", zap.Error(err))
	}
	if len(failedUserIDs) > 0 {
		var retrySubscriptions []model.SmartReportSubscription
		if err := global.GVA_DB.WithContext(ctx).Where("enabled = ? AND user_id IN ?", true, failedUserIDs).Find(&retrySubscriptions).Error; err == nil {
			seen := make(map[uint]struct{}, len(subscriptions)+len(retrySubscriptions))
			for _, item := range subscriptions {
				seen[item.UserID] = struct{}{}
			}
			for _, item := range retrySubscriptions {
				if _, exists := seen[item.UserID]; !exists {
					subscriptions = append(subscriptions, item)
				}
			}
		}
	}
	for _, subscription := range subscriptions {
		if err := deliverSubscription(ctx, subscription, now); err != nil {
			global.GVA_LOG.Error("发送智能日报失败", zap.Uint("userID", subscription.UserID), zap.Error(err))
		}
	}
}

func deliverSubscription(ctx context.Context, subscription model.SmartReportSubscription, now time.Time) error {
	var user system.SysUser
	if err := global.GVA_DB.WithContext(ctx).First(&user, subscription.UserID).Error; err != nil {
		return err
	}
	report, err := Smart.GenerateReport(ctx, user.ID, user.AuthorityId)
	if err != nil {
		return err
	}
	for _, channel := range strings.Split(subscription.Channels, ",") {
		channel = strings.TrimSpace(channel)
		if channel == "" {
			continue
		}
		if err := deliverChannel(ctx, user, report, channel, now); err != nil {
			global.GVA_LOG.Warn("智能日报渠道发送失败", zap.Uint("userID", user.ID), zap.String("channel", channel), zap.Error(err))
		}
	}
	return nil
}

func deliverChannel(ctx context.Context, user system.SysUser, report model.SmartDailyReport, channel string, now time.Time) error {
	var delivery model.SmartReportDelivery
	query := global.GVA_DB.WithContext(ctx).Where("user_id = ? AND report_id = ? AND channel = ?", user.ID, report.ID, channel)
	err := query.First(&delivery).Error
	if err == nil {
		if delivery.Status == "sent" || (channel != reportDeliveryChannelManualMail && delivery.RetryCount >= 3) {
			return nil
		}
		if delivery.Status == "sending" && delivery.LastAttempt != nil && delivery.LastAttempt.After(now.Add(-10*time.Minute)) {
			return nil
		}
		claimQuery := global.GVA_DB.WithContext(ctx).Model(&model.SmartReportDelivery{}).
			Where("id = ? AND status IN ?", delivery.ID, []string{"failed", "sending"})
		if channel == reportDeliveryChannelManualMail {
			claimQuery = claimQuery.Where("status = ? OR last_attempt IS NULL OR last_attempt < ?", "failed", now.Add(-10*time.Minute))
		} else {
			claimQuery = claimQuery.Where("retry_count < ? AND (last_attempt IS NULL OR last_attempt < ?)", 3, now.Add(-10*time.Minute))
		}
		claim := claimQuery.Updates(map[string]any{"status": "sending", "retry_count": gorm.Expr("retry_count + 1"), "last_attempt": now, "error": ""})
		if claim.Error != nil {
			return claim.Error
		}
		if claim.RowsAffected == 0 {
			return nil
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	} else {
		delivery = model.SmartReportDelivery{UserID: user.ID, ReportID: report.ID, Channel: channel, Status: "sending", RetryCount: 1, LastAttempt: &now}
		if err := global.GVA_DB.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&delivery).Error; err != nil {
			return err
		}
		if delivery.ID == 0 {
			if err := query.First(&delivery).Error; err != nil {
				return err
			}
			return nil
		}
	}

	var sendErr error
	switch channel {
	case reportDeliveryChannelInApp:
		announcementService.NotificationHub.PublishToUser(user.ID, announcementService.NotificationEvent{ID: report.ID, Title: "智能日报已生成", Kind: "smart_daily_report"})
	case reportDeliveryChannelEmail:
		if strings.TrimSpace(user.Email) == "" {
			sendErr = fmt.Errorf("用户未配置邮箱")
		} else {
			sendErr = sendReportToMailbox(user.Email, smartDailyReportEmailMessage(report))
		}
	case reportDeliveryChannelManualMail:
		sendErr = sendReportToSystemInbox(smartDailyReportEmailMessage(report))
	default:
		sendErr = fmt.Errorf("不支持的日报渠道: %s", channel)
	}
	if sendErr != nil {
		_ = global.GVA_DB.WithContext(ctx).Model(&model.SmartReportDelivery{}).Where("id = ?", delivery.ID).Updates(map[string]any{"status": "failed", "error": sendErr.Error()}).Error
		return sendErr
	}
	return global.GVA_DB.WithContext(ctx).Model(&model.SmartReportDelivery{}).Where("id = ? AND status = ?", delivery.ID, "sending").Updates(map[string]any{"status": "sent", "error": "", "sent_at": now}).Error
}
