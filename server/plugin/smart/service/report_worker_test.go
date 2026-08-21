package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/service/reportemail"
)

func TestSendReportEmailDispatchesSmartDailyAndDoesNotRepeatSuccess(t *testing.T) {
	setupSmartTestDB(t)

	previousSender := sendReportToSystemInbox
	t.Cleanup(func() {
		sendReportToSystemInbox = previousSender
	})

	var sendCount int
	sendReportToSystemInbox = func(message reportemail.Message) error {
		sendCount++
		if !strings.Contains(message.Subject, "智能日报") || message.Title != "智能日报" {
			t.Fatalf("unexpected report message: %#v", message)
		}
		if !strings.Contains(message.Subtitle, "生成时间") {
			t.Fatalf("report message is missing generated-at metadata: %#v", message)
		}
		return nil
	}

	input := ReportEmailInput{ReportType: ReportEmailTypeSmartDaily}
	result, err := Smart.SendReportEmail(context.Background(), 1, 888, input)
	if err != nil {
		t.Fatal(err)
	}
	if result.ReportType != ReportEmailTypeSmartDaily || result.ReportID == 0 || result.DeliveryID == 0 || result.Status != "sent" || result.AlreadySent {
		t.Fatalf("unexpected first report delivery: %#v", result)
	}
	if sendCount != 1 {
		t.Fatalf("expected one send, got %d", sendCount)
	}

	second, err := Smart.SendReportEmail(context.Background(), 1, 888, input)
	if err != nil {
		t.Fatal(err)
	}
	if !second.AlreadySent || second.ReportID != result.ReportID || second.DeliveryID != result.DeliveryID || sendCount != 1 {
		t.Fatalf("successful delivery should be idempotent: first=%#v second=%#v sendCount=%d", result, second, sendCount)
	}
}

func TestSendReportEmailRejectsUnknownReportType(t *testing.T) {
	result, err := Smart.SendReportEmail(context.Background(), 1, 888, ReportEmailInput{ReportType: "arbitrary_html"})
	if err == nil || !strings.Contains(err.Error(), "不支持的报告类型") {
		t.Fatalf("expected controlled report-type error, got result=%#v err=%v", result, err)
	}
}

func TestSendReportEmailScopesExplicitReportIDByAuthority(t *testing.T) {
	setupSmartTestDB(t)
	report, err := Smart.GenerateReport(context.Background(), 1, 999)
	if err != nil {
		t.Fatal(err)
	}

	previousSender := sendReportToSystemInbox
	t.Cleanup(func() { sendReportToSystemInbox = previousSender })
	sendReportToSystemInbox = func(reportemail.Message) error {
		t.Fatal("out-of-scope report must not reach the email sender")
		return nil
	}

	_, err = Smart.SendReportEmail(context.Background(), 1, 888, ReportEmailInput{
		ReportType: ReportEmailTypeSmartDaily,
		ReportID:   report.ID,
	})
	if err == nil || !strings.Contains(err.Error(), "无权访问") {
		t.Fatalf("expected authority-scoped report error, got %v", err)
	}
}

func TestSendReportEmailAllowsManualRetryAfterFailure(t *testing.T) {
	setupSmartTestDB(t)

	previousSender := sendReportToSystemInbox
	t.Cleanup(func() { sendReportToSystemInbox = previousSender })
	var attempts int
	sendReportToSystemInbox = func(reportemail.Message) error {
		attempts++
		if attempts == 1 {
			return errors.New("temporary smtp failure")
		}
		return nil
	}

	input := ReportEmailInput{ReportType: ReportEmailTypeSmartDaily}
	if _, err := Smart.SendReportEmail(context.Background(), 1, 888, input); err == nil {
		t.Fatal("expected the first SMTP attempt to fail")
	}
	result, err := Smart.SendReportEmail(context.Background(), 1, 888, input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "sent" || result.AlreadySent || attempts != 2 {
		t.Fatalf("unexpected retry result=%#v attempts=%d", result, attempts)
	}
}
