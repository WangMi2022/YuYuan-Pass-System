package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/ai"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/common"
	systemModel "github.com/WangMi2022/mit-assets-admin/server/model/system"
	announcementModel "github.com/WangMi2022/mit-assets-admin/server/plugin/announcement/model"
	assetModel "github.com/WangMi2022/mit-assets-admin/server/plugin/asset/model"
	invoiceModel "github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model"
	scheduleModel "github.com/WangMi2022/mit-assets-admin/server/plugin/schedule/model"
	smartModel "github.com/WangMi2022/mit-assets-admin/server/plugin/smart/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupSmartTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	previousDB := global.GVA_DB
	previousAI := global.GVA_CONFIG.AI
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{TranslateError: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open smart test database: %v", err)
	}
	if err := database.AutoMigrate(
		&assetModel.Category{}, &assetModel.Asset{}, &assetModel.AssetOperationOrder{}, &assetModel.AssetRiskEvent{},
		&invoiceModel.Invoice{}, &scheduleModel.WorkSchedule{}, &announcementModel.Info{}, &announcementModel.Read{},
		&ai.ModelInvocation{}, &smartModel.CopilotSession{}, &smartModel.CopilotMessage{}, &smartModel.CopilotRun{}, &smartModel.KnowledgeChunk{}, &smartModel.SmartDailyReport{},
		&smartModel.SmartReportSubscription{}, &smartModel.SmartReportDelivery{}, &smartModel.SmartDraft{},
	); err != nil {
		t.Fatalf("migrate smart test tables: %v", err)
	}
	if err := database.Exec("CREATE TABLE IF NOT EXISTS sys_users (id integer primary key, nick_name text, username text, email text, authority_id integer)").Error; err != nil {
		t.Fatalf("create smart test users table: %v", err)
	}
	global.GVA_DB = database
	global.GVA_CONFIG.AI.Enabled = false
	t.Cleanup(func() {
		global.GVA_DB = previousDB
		global.GVA_CONFIG.AI = previousAI
		if sqlDB, dbErr := database.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	})
	return database
}

func TestClassifyReadOnlyBusinessQuestions(t *testing.T) {
	cases := []struct {
		question string
		intent   string
		tool     string
	}{
		{"查询未来质保到期的资产", "warranty", "asset.warranty.expiring"},
		{"当前有哪些高风险异常", "risk", "asset.risk.list"},
		{"我有哪些未读公告", "announcement", "announcement.unread"},
		{"查看待复核发票", "invoice_review", "invoice.pending_reviews"},
		{"查看发票识别质量", "invoice_quality", "invoice.provider_quality"},
		{"今天的日程安排", "schedule", "schedule.today"},
	}
	for _, item := range cases {
		intent, tool := classify(item.question)
		if intent != item.intent || tool != item.tool {
			t.Fatalf("classify(%q) = (%q, %q), want (%q, %q)", item.question, intent, tool, item.intent, item.tool)
		}
	}
}

func TestWriteIntentRejectsCommandsButAllowsReadOnlyQuestions(t *testing.T) {
	for _, question := range []string{"帮我报废资产 12", "创建一张领用单", "删除这张发票", "确认并提交业务单"} {
		if !isWriteIntent(question) {
			t.Fatalf("expected write intent for %q", question)
		}
	}
	for _, question := range []string{"查询报废资产", "查看维修单详情", "统计今日入库数量", "有哪些待确认发票"} {
		if isWriteIntent(question) {
			t.Fatalf("unexpected write intent for %q", question)
		}
	}
}

func TestEveryCopilotToolHasPermissionMapping(t *testing.T) {
	if len(toolDefinitions) != 13 {
		t.Fatalf("tool count = %d, want 13", len(toolDefinitions))
	}
	for _, definition := range toolDefinitions {
		path, ok := toolPermissionPath(definition.Name)
		if !ok || path == "" || !definition.ReadOnly {
			t.Fatalf("invalid tool definition: %#v, path = %q", definition, path)
		}
	}
	if _, ok := toolPermissionPath("asset.delete"); ok {
		t.Fatal("unregistered write tool must not have a permission mapping")
	}
}

func TestCopilotSessionsAndReportsAreScopedByAuthority(t *testing.T) {
	database := setupSmartTestDB(t)
	now := time.Now()
	sessions := []smartModel.CopilotSession{
		{UserID: 1, AuthorityID: 888, Title: "管理员会话", LastMessageAt: now},
		{UserID: 1, AuthorityID: 889, Title: "普通角色会话", LastMessageAt: now},
	}
	if err := database.Create(&sessions).Error; err != nil {
		t.Fatal(err)
	}
	adminSessions, err := Smart.Sessions(1, 888)
	if err != nil || len(adminSessions) != 1 || adminSessions[0].ID != sessions[0].ID {
		t.Fatalf("unexpected admin sessions: %#v, err: %v", adminSessions, err)
	}
	if _, _, err := Smart.Session(1, 889, sessions[0].ID); err == nil {
		t.Fatal("expected cross-authority session access to be rejected")
	}

	reportDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	reports := []smartModel.SmartDailyReport{
		{UserID: 1, AuthorityID: 888, ReportDate: reportDate, Metrics: common.JSONMap{}, Summary: "管理员日报", GeneratedBy: "deterministic", GeneratedAt: now},
		{UserID: 1, AuthorityID: 889, ReportDate: reportDate, Metrics: common.JSONMap{}, Summary: "普通角色日报", GeneratedBy: "deterministic", GeneratedAt: now},
	}
	if err := database.Create(&reports).Error; err != nil {
		t.Fatal(err)
	}
	adminReports, total, err := Smart.Reports(1, 888, ReportListInput{Page: 1, PageSize: 10})
	if err != nil || total != 1 || len(adminReports) != 1 || adminReports[0].ID != reports[0].ID {
		t.Fatalf("unexpected admin reports: %#v, total: %d, err: %v", adminReports, total, err)
	}
	if _, err := Smart.Report(1, 889, reports[0].ID); err == nil {
		t.Fatal("expected cross-authority report access to be rejected")
	}
}

func TestDeleteSessionAlsoDeletesScopedCopilotRuns(t *testing.T) {
	database := setupSmartTestDB(t)
	sessions := []smartModel.CopilotSession{
		{UserID: 1, AuthorityID: 888, Title: "待删除", LastMessageAt: time.Now()},
		{UserID: 1, AuthorityID: 889, Title: "保留", LastMessageAt: time.Now()},
	}
	if err := database.Create(&sessions).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Create(&[]smartModel.CopilotMessage{
		{SessionID: sessions[0].ID, UserID: 1, AuthorityID: 888, Role: smartModel.MessageRoleUser, Content: "需要删除的消息"},
		{SessionID: sessions[1].ID, UserID: 1, AuthorityID: 889, Role: smartModel.MessageRoleUser, Content: "需要保留的消息"},
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Create(&[]smartModel.CopilotRun{
		{RequestID: "deleted-run", SessionID: sessions[0].ID, UserID: 1, AuthorityID: 888, Planner: rulePlannerName, Intent: "schedule", Status: "success"},
		{RequestID: "retained-run", SessionID: sessions[1].ID, UserID: 1, AuthorityID: 889, Planner: rulePlannerName, Intent: "schedule", Status: "success"},
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := Smart.DeleteSession(1, 888, sessions[0].ID); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	var deletedMessages, deletedRuns, retainedRuns int64
	if err := database.Model(&smartModel.CopilotMessage{}).Where("session_id = ? AND user_id = ? AND authority_id = ?", sessions[0].ID, 1, 888).Count(&deletedMessages).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Model(&smartModel.CopilotRun{}).Where("session_id = ? AND user_id = ? AND authority_id = ?", sessions[0].ID, 1, 888).Count(&deletedRuns).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Model(&smartModel.CopilotRun{}).Where("session_id = ? AND user_id = ? AND authority_id = ?", sessions[1].ID, 1, 889).Count(&retainedRuns).Error; err != nil {
		t.Fatal(err)
	}
	if deletedMessages != 0 || deletedRuns != 0 || retainedRuns != 1 {
		t.Fatalf("unexpected delete counts: messages=%d deletedRuns=%d retainedRuns=%d", deletedMessages, deletedRuns, retainedRuns)
	}
}

func TestNormalizeReportChannels(t *testing.T) {
	channels, err := normalizeReportChannels("in_app, email, in_app")
	if err != nil || channels != "in_app,email" {
		t.Fatalf("normalizeReportChannels() = %q, %v", channels, err)
	}
	if _, err := normalizeReportChannels("webhook"); err == nil {
		t.Fatal("expected unsupported channel to be rejected")
	}
}

func TestAssetOperationType(t *testing.T) {
	for _, value := range []string{"inbound", "issue", "transfer", "return", "maintenance", "scrap"} {
		if got, err := assetOperationType(value); err != nil || got != value {
			t.Fatalf("assetOperationType(%q) = %q, %v", value, got, err)
		}
	}
	if _, err := assetOperationType("delete"); err == nil {
		t.Fatal("expected unsupported operation type to be rejected")
	}
}

func TestExtractScheduleKeepsEvidenceAndNormalizesDate(t *testing.T) {
	payload := extractSchedule("年度盘点通知", "请于2026年8月20日 下午 14:30 到地点：一号会议室参加盘点。请提交资产清单。")
	if payload["date"] != "2026-08-20" || payload["time"] != "14:30" || payload["location"] != "一号会议室" {
		t.Fatalf("unexpected extracted schedule: %#v", payload)
	}
	todos, ok := payload["todos"].([]string)
	if !ok || len(todos) == 0 {
		t.Fatalf("expected todo evidence, got %#v", payload["todos"])
	}
	if payload["confidence"].(float64) <= 0.5 {
		t.Fatalf("expected confident extraction, got %#v", payload["confidence"])
	}
}

func TestLoadInvoiceProviderQualityBindsFailureStatus(t *testing.T) {
	database := setupSmartTestDB(t)
	invoices := []invoiceModel.Invoice{
		{Direction: "expense", InvoiceNumber: "QUALITY-1", SellerName: "供应商甲", Currency: "CNY", Status: invoiceModel.InvoiceStatusConfirmed, RecognitionProvider: "vision", RecognitionDurationMS: 100, FileName: "a.pdf", FileKey: "a", FileHash: "quality-a", MimeType: "application/pdf", FileSize: 1, StorageType: "local", CreatedBy: 1, AuthorityID: 888},
		{Direction: "expense", InvoiceNumber: "QUALITY-2", SellerName: "供应商乙", Currency: "CNY", Status: invoiceModel.InvoiceStatusRecognitionFailed, RecognitionProvider: "vision", RecognitionDurationMS: 300, FileName: "b.pdf", FileKey: "b", FileHash: "quality-b", MimeType: "application/pdf", FileSize: 1, StorageType: "local", CreatedBy: 2, AuthorityID: 888},
	}
	if err := database.Create(&invoices).Error; err != nil {
		t.Fatal(err)
	}
	rows, err := loadInvoiceProviderQuality(1, 888)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Provider != "vision" || rows[0].Total != 2 || rows[0].Failed != 1 || rows[0].AvgDuration != 200 {
		t.Fatalf("unexpected provider quality rows: %#v", rows)
	}
}

func TestScheduleDraftNoteKeepsLocationAndTodos(t *testing.T) {
	note := scheduleDraftNote(common.JSONMap{
		"location": "一号会议室",
		"todos":    []any{"携带资产清单", "完成签字"},
		"note":     "原公告内容",
	})
	for _, expected := range []string{"地点：一号会议室", "待办：携带资产清单；完成签字", "公告原文：原公告内容"} {
		if !strings.Contains(note, expected) {
			t.Fatalf("schedule note %q missing %q", note, expected)
		}
	}
}

func TestAnnouncementDraftRefreshesExpiredDraftWithoutDuplicate(t *testing.T) {
	database := setupSmartTestDB(t)
	now := time.Now()
	announcement := announcementModel.Info{Title: "盘点安排", Content: "请于2026年8月20日 09:30 到地点：仓库参加盘点。", Status: "published", PublishedAt: &now}
	if err := database.Create(&announcement).Error; err != nil {
		t.Fatal(err)
	}
	expiredAt := now.Add(-time.Hour)
	existing := smartModel.SmartDraft{UserID: 1, DraftType: smartModel.DraftTypeSchedule, SourceID: announcement.ID, Status: smartModel.DraftStatusDraft, Payload: common.JSONMap{"title": "旧草稿"}, Confidence: 0.2, ExpiresAt: &expiredAt}
	if err := database.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}
	refreshed, err := Smart.AnnouncementDraft(1, announcement.ID)
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.ID != existing.ID || refreshed.ExpiresAt == nil || !refreshed.ExpiresAt.After(now) || refreshed.Payload["title"] != announcement.Title || refreshed.DedupKey == nil {
		t.Fatalf("unexpected refreshed draft: %#v", refreshed)
	}
	reused, err := Smart.AnnouncementDraft(1, announcement.ID)
	if err != nil || reused.ID != existing.ID {
		t.Fatalf("unexpected reused draft: %#v, err: %v", reused, err)
	}
	var count int64
	if err := database.Model(&smartModel.SmartDraft{}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("draft count = %d, err = %v", count, err)
	}
}

func TestGenerateReportIsIdempotentAndUsesBusinessMetrics(t *testing.T) {
	database := setupSmartTestDB(t)
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	category := assetModel.Category{Name: "智能日报测试分类", Code: "SMART-REPORT", Enabled: true}
	if err := database.Create(&category).Error; err != nil {
		t.Fatal(err)
	}
	warranty := start.AddDate(0, 0, 20)
	assets := []assetModel.Asset{
		{AssetCode: "SMART-PENDING", Name: "待入库设备", CategoryID: category.ID, Quantity: 1, Unit: "件", Status: assetModel.AssetStatusPendingInbound, WarrantyEndDate: &warranty},
		{AssetCode: "SMART-MAINTENANCE", Name: "维修设备", CategoryID: category.ID, Quantity: 1, Unit: "件", Status: assetModel.AssetStatusMaintenance},
	}
	assets[0].CreatedAt, assets[0].UpdatedAt = now, now
	assets[1].CreatedAt, assets[1].UpdatedAt = now, start.AddDate(0, 0, -10)
	if err := database.Create(&assets).Error; err != nil {
		t.Fatal(err)
	}
	operation := assetModel.AssetOperationOrder{OrderNo: "SMART-OP-1", Type: "inbound", Status: assetModel.OperationStatusCompleted, BusinessDate: start, CreatedBy: 1}
	if err := database.Create(&operation).Error; err != nil {
		t.Fatal(err)
	}
	risk := assetModel.AssetRiskEvent{Fingerprint: "smart-risk", AssetID: assets[0].ID, RuleCode: "SMART_RISK", RuleVersion: 1, Category: "status", Severity: assetModel.RiskSeverityHigh, Status: assetModel.RiskStatusOpen, Title: "测试风险", FirstDetectedAt: now, LastDetectedAt: now, LastScanRunID: 1}
	if err := database.Create(&risk).Error; err != nil {
		t.Fatal(err)
	}
	confirmedAt := now
	invoices := []invoiceModel.Invoice{
		{Direction: "expense", InvoiceNumber: "SMART-INV-1", SellerName: "测试供应商", TotalCents: 12345, Currency: "CNY", Status: invoiceModel.InvoiceStatusConfirmed, RecognitionProvider: "test", RecognitionConfidence: 0.95, FileName: "a.pdf", FileKey: "a", FileHash: "hash-a", MimeType: "application/pdf", FileSize: 1, StorageType: "local", CreatedBy: 1, AuthorityID: 888, ConfirmedAt: &confirmedAt},
		{Direction: "expense", InvoiceNumber: "SMART-INV-2", SellerName: "测试供应商", Currency: "CNY", Status: invoiceModel.InvoiceStatusPendingReview, RecognitionProvider: "test", RecognitionConfidence: 0.4, FileName: "b.pdf", FileKey: "b", FileHash: "hash-b", MimeType: "application/pdf", FileSize: 1, StorageType: "local", CreatedBy: 1, AuthorityID: 888},
	}
	for index := range invoices {
		invoices[index].CreatedAt, invoices[index].UpdatedAt = now, now
	}
	if err := database.Create(&invoices).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Model(&invoiceModel.Invoice{}).Where("id = ?", invoices[0].ID).Updates(map[string]any{"status": invoiceModel.InvoiceStatusConfirmed, "confirmed_at": confirmedAt}).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Model(&invoiceModel.Invoice{}).Where("id = ?", invoices[1].ID).Updates(map[string]any{"status": invoiceModel.InvoiceStatusPendingReview, "recognition_confidence": 0.4}).Error; err != nil {
		t.Fatal(err)
	}
	var persistedInvoices []invoiceModel.Invoice
	if err := database.Order("id ASC").Find(&persistedInvoices).Error; err != nil {
		t.Fatal(err)
	}
	if len(persistedInvoices) != 2 || persistedInvoices[0].Status != invoiceModel.InvoiceStatusConfirmed || persistedInvoices[1].Status != invoiceModel.InvoiceStatusPendingReview {
		t.Fatalf("unexpected persisted invoice statuses: %#v", persistedInvoices)
	}
	schedule := scheduleModel.WorkSchedule{UserID: 1, ClientKey: "smart-report-schedule", Title: "今日会议", ScheduleDate: start.Format("2006-01-02"), ScheduleTime: "09:00", Type: "meeting"}
	if err := database.Create(&schedule).Error; err != nil {
		t.Fatal(err)
	}
	announcement := announcementModel.Info{Title: "测试公告", Content: "测试内容", Status: "published", PublishedAt: &now}
	if err := database.Create(&announcement).Error; err != nil {
		t.Fatal(err)
	}
	invocation := ai.ModelInvocation{RequestID: "smart-report-ai", UserID: 1, AuthorityID: 888, Module: "smart", Operation: "test", Provider: "test", Status: ai.InvocationStatusSuccess, DurationMS: 240, EstimatedCostMicros: 900}
	invocation.CreatedAt, invocation.UpdatedAt = now, now
	if err := database.Create(&invocation).Error; err != nil {
		t.Fatal(err)
	}

	first, err := Smart.GenerateReport(context.Background(), 1, 888)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Smart.GenerateReport(context.Background(), 1, 888)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == 0 || second.ID != first.ID {
		t.Fatalf("report IDs = (%d, %d), want one idempotent report", first.ID, second.ID)
	}
	var reportCount int64
	if err := database.Model(&smartModel.SmartDailyReport{}).Count(&reportCount).Error; err != nil || reportCount != 1 {
		t.Fatalf("report count = %d, err = %v", reportCount, err)
	}
	assetsMetric := second.Metrics["assets"].(map[string]any)
	if assetsMetric["todayOperationTotal"] != float64(1) || assetsMetric["pendingInbound"] != float64(1) || assetsMetric["maintenanceOverdue"] != float64(1) {
		t.Fatalf("unexpected asset metrics: %#v", assetsMetric)
	}
	invoicesMetric := second.Metrics["invoices"].(map[string]any)
	if invoicesMetric["todayConfirmed"] != float64(1) || invoicesMetric["pendingReview"] != float64(1) || invoicesMetric["lowConfidence"] != float64(1) || invoicesMetric["confirmedTodayCents"] != float64(12345) {
		t.Fatalf("unexpected invoice metrics: %#v", invoicesMetric)
	}
	collaborationMetric := second.Metrics["collaboration"].(map[string]any)
	if collaborationMetric["todaySchedules"] != float64(1) || collaborationMetric["unreadAnnouncements"] != float64(1) {
		t.Fatalf("unexpected collaboration metrics: %#v", collaborationMetric)
	}
	systemMetric := second.Metrics["system"].(map[string]any)
	if systemMetric["aiCalls"] != float64(1) || systemMetric["aiAverageDurationMs"] != float64(240) || systemMetric["aiEstimatedCostMicros"] != float64(900) {
		t.Fatalf("unexpected system metrics: %#v", systemMetric)
	}
}

func TestDeliverChannelDoesNotRepeatSuccessfulDelivery(t *testing.T) {
	database := setupSmartTestDB(t)
	report := smartModel.SmartDailyReport{UserID: 7, AuthorityID: 888, ReportDate: time.Now(), Metrics: common.JSONMap{}, Summary: "日报", GeneratedBy: "deterministic", GeneratedAt: time.Now()}
	if err := database.Create(&report).Error; err != nil {
		t.Fatal(err)
	}
	user := systemModel.SysUser{}
	user.ID = 7
	now := time.Now()
	if err := deliverChannel(context.Background(), user, report, "in_app", now); err != nil {
		t.Fatal(err)
	}
	if err := deliverChannel(context.Background(), user, report, "in_app", now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	var deliveries []smartModel.SmartReportDelivery
	if err := database.Find(&deliveries).Error; err != nil {
		t.Fatal(err)
	}
	if len(deliveries) != 1 || deliveries[0].Status != "sent" || deliveries[0].RetryCount != 1 {
		t.Fatalf("unexpected deliveries: %#v", deliveries)
	}
}
