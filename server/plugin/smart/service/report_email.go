package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/system"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/smart/model"
	"github.com/WangMi2022/mit-assets-admin/server/service/reportemail"
	"gorm.io/gorm"
)

const ReportEmailTypeSmartDaily = "smart_daily"

type ReportEmailInput struct {
	ReportType string `json:"reportType" binding:"required,max=64"`
	ReportID   uint   `json:"reportId"`
}

type ReportEmailResult struct {
	ReportType  string `json:"reportType"`
	ReportID    uint   `json:"reportId"`
	DeliveryID  uint   `json:"deliveryId"`
	Status      string `json:"status"`
	AlreadySent bool   `json:"alreadySent"`
}

var (
	sendReportToSystemInbox = reportemail.SendToSystemInbox
	sendReportToMailbox     = reportemail.SendToMailbox
)

// SendReportEmail is the stable seam used by HTTP callers. Clients select a
// server-known report type and optional report ID; recipients, authorization,
// subject and content are always resolved by the selected provider.
func (s *smartService) SendReportEmail(ctx context.Context, userID, authorityID uint, input ReportEmailInput) (ReportEmailResult, error) {
	if userID == 0 {
		return ReportEmailResult{}, errors.New("用户未登录")
	}
	reportType := strings.ToLower(strings.TrimSpace(input.ReportType))
	if reportType == "" || len(reportType) > 64 {
		return ReportEmailResult{}, errors.New("报告类型不正确")
	}
	switch reportType {
	case ReportEmailTypeSmartDaily:
		return s.sendSmartDailyReportEmail(ctx, userID, authorityID, input.ReportID)
	default:
		return ReportEmailResult{}, fmt.Errorf("不支持的报告类型: %s", reportType)
	}
}

func (s *smartService) sendSmartDailyReportEmail(ctx context.Context, userID, authorityID, reportID uint) (ReportEmailResult, error) {
	report, exists, err := findSmartDailyReportForEmail(ctx, userID, authorityID, reportID)
	if err != nil {
		return ReportEmailResult{}, err
	}
	if exists {
		if delivery, found, findErr := findSuccessfulManualDelivery(ctx, userID, report.ID); findErr != nil {
			return ReportEmailResult{}, findErr
		} else if found {
			return smartDailyReportEmailResult(report, delivery, true), nil
		}
	}
	if reportID == 0 {
		report, err = s.GenerateReport(ctx, userID, authorityID)
		if err != nil {
			return ReportEmailResult{}, err
		}
	}
	user := system.SysUser{}
	user.ID = userID
	if err = deliverChannel(ctx, user, report, reportDeliveryChannelManualMail, time.Now()); err != nil {
		return ReportEmailResult{}, err
	}
	delivery, err := findManualDelivery(ctx, userID, report.ID)
	if err != nil {
		return ReportEmailResult{}, err
	}
	if delivery.Status != "sent" {
		if delivery.Error != "" {
			return ReportEmailResult{}, errors.New(delivery.Error)
		}
		return ReportEmailResult{}, errors.New("报告尚未完成发送，请稍后重试")
	}
	return smartDailyReportEmailResult(report, delivery, false), nil
}

func findSmartDailyReportForEmail(ctx context.Context, userID, authorityID, reportID uint) (model.SmartDailyReport, bool, error) {
	var report model.SmartDailyReport
	query := global.GVA_DB.WithContext(ctx).Where("user_id = ? AND authority_id = ?", userID, authorityID)
	if reportID != 0 {
		query = query.Where("id = ?", reportID)
	} else {
		start, _ := todayRange()
		query = query.Where("report_date = ?", start)
	}
	err := query.First(&report).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if reportID != 0 {
			return model.SmartDailyReport{}, false, errors.New("报告不存在或无权访问")
		}
		return model.SmartDailyReport{}, false, nil
	}
	return report, err == nil, err
}

func findSuccessfulManualDelivery(ctx context.Context, userID, reportID uint) (model.SmartReportDelivery, bool, error) {
	var delivery model.SmartReportDelivery
	err := global.GVA_DB.WithContext(ctx).
		Where("user_id = ? AND report_id = ? AND channel = ? AND status = ?", userID, reportID, reportDeliveryChannelManualMail, "sent").
		First(&delivery).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.SmartReportDelivery{}, false, nil
	}
	return delivery, err == nil, err
}

func findManualDelivery(ctx context.Context, userID, reportID uint) (model.SmartReportDelivery, error) {
	var delivery model.SmartReportDelivery
	err := global.GVA_DB.WithContext(ctx).
		Where("user_id = ? AND report_id = ? AND channel = ?", userID, reportID, reportDeliveryChannelManualMail).
		First(&delivery).Error
	return delivery, err
}

func smartDailyReportEmailResult(report model.SmartDailyReport, delivery model.SmartReportDelivery, alreadySent bool) ReportEmailResult {
	return ReportEmailResult{
		ReportType:  ReportEmailTypeSmartDaily,
		ReportID:    report.ID,
		DeliveryID:  delivery.ID,
		Status:      delivery.Status,
		AlreadySent: alreadySent,
	}
}

func smartDailyReportEmailMessage(report model.SmartDailyReport) reportemail.Message {
	summary := strings.TrimSpace(report.Summary)
	if summary == "" {
		summary = "日报暂无正文。"
	}
	generatedAt := report.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = time.Now()
	}
	return reportemail.Message{
		Subject:  "智能日报 - " + report.ReportDate.Format("2006-01-02"),
		Title:    "智能日报",
		Subtitle: fmt.Sprintf("报告日期：%s · 生成时间：%s", report.ReportDate.Format("2006-01-02"), generatedAt.Format("2006-01-02 15:04:05")),
		Body:     summary,
	}
}
