package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model"
	invoiceRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/provider"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupInvoiceServiceTestDB(t *testing.T) {
	t.Helper()
	previous := global.GVA_DB
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open invoice test database: %v", err)
	}
	if err = db.AutoMigrate(
		&model.InvoiceCategory{}, &model.Invoice{}, &model.InvoiceItem{},
		&model.ClassificationRule{}, &model.RecognitionJob{}, &model.InvoiceFileCleanupJob{},
	); err != nil {
		t.Fatalf("migrate invoice tables: %v", err)
	}
	global.GVA_DB = db
	t.Cleanup(func() {
		global.GVA_DB = previous
		if sqlDB, dbErr := db.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	})
}

func createInvoiceTestCategory(t *testing.T) model.InvoiceCategory {
	t.Helper()
	category := model.InvoiceCategory{
		Name: "测试分类-" + t.Name(), Code: "TEST-" + fmt.Sprint(time.Now().UnixNano()),
		Color: "#2563EB", Enabled: true,
	}
	if err := global.GVA_DB.Create(&category).Error; err != nil {
		t.Fatalf("create invoice category: %v", err)
	}
	return category
}

func createReviewableInvoice(t *testing.T, userID, authorityID uint, categoryID uint, number string) model.Invoice {
	t.Helper()
	issueDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	invoice := model.Invoice{
		Direction: "expense", InvoiceCode: "044001", InvoiceNumber: number,
		IssueDate: &issueDate, SellerName: "测试销售方", SellerTaxNo: "91440000TEST",
		AmountCents: 10000, TaxCents: 600, TotalCents: 10600, Currency: "CNY",
		CategoryID: &categoryID, Status: model.InvoiceStatusPendingReview,
		FileName: number + ".png", FileKey: number + ".png", FileHash: hashInvoiceKey(number),
		MimeType: "image/png", FileSize: 128, StorageType: "local",
		CreatedBy: userID, AuthorityID: authorityID,
	}
	if err := global.GVA_DB.Create(&invoice).Error; err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	return invoice
}

func TestInvoiceScopeKeepsPersonalAndRoleWideDataSeparated(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	createReviewableInvoice(t, 11, 100, category.ID, "PERSONAL-1")
	createReviewableInvoice(t, 12, 100, category.ID, "ROLE-2")
	createReviewableInvoice(t, 13, 200, category.ID, "OTHER-3")
	search := invoiceRequest.InvoiceSearch{PageInfo: commonRequest.PageInfo{Page: 1, PageSize: 20}}

	personal, total, err := (InvoiceService{}).List(search, AccessScope{UserID: 11, AuthorityID: 100})
	if err != nil || total != 1 || len(personal) != 1 || personal[0].CreatedBy != 11 {
		t.Fatalf("unexpected personal scope result: total=%d list=%#v err=%v", total, personal, err)
	}
	roleWide, total, err := (InvoiceService{}).List(search, AccessScope{
		UserID: 11, AuthorityID: 100, RoleWide: true, AuthorityIDs: []uint{100},
	})
	if err != nil || total != 2 || len(roleWide) != 2 {
		t.Fatalf("unexpected role-wide result: total=%d list=%#v err=%v", total, roleWide, err)
	}
}

func TestConfirmPersistsUniqueDuplicateKeyAndFeedsDashboard(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	first := createReviewableInvoice(t, 21, 100, category.ID, "DUP-001")
	second := createReviewableInvoice(t, 22, 200, category.ID, "DUP-001")
	admin := AccessScope{UserID: 1, AuthorityID: defaultAdminRoleID, All: true}

	confirmed, err := (InvoiceService{}).Confirm(first.ID, admin)
	if err != nil || confirmed.Status != model.InvoiceStatusConfirmed || confirmed.DuplicateKey == nil {
		t.Fatalf("confirm first invoice: invoice=%#v err=%v", confirmed, err)
	}
	if _, err = (InvoiceService{}).Confirm(second.ID, admin); err == nil || !strings.Contains(err.Error(), "相同号码") {
		t.Fatalf("expected duplicate confirmation error, got %v", err)
	}

	duplicateKey := *confirmed.DuplicateKey
	direct := global.GVA_DB.Model(&model.Invoice{}).Where("id = ?", second.ID).Updates(map[string]any{
		"status": model.InvoiceStatusConfirmed, "duplicate_key": &duplicateKey,
	})
	if direct.Error == nil || !isUniqueConstraintError(direct.Error) {
		t.Fatalf("database unique key did not reject duplicate: %v", direct.Error)
	}

	dashboard, err := (InvoiceService{}).Dashboard(admin)
	if err != nil || dashboard.ConfirmedCount != 1 || dashboard.TotalCents != first.TotalCents {
		t.Fatalf("unexpected confirmed dashboard: %#v err=%v", dashboard, err)
	}
}

func TestConfirmRejectsInconsistentAmounts(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 31, 300, category.ID, "AMOUNT-001")
	if err := global.GVA_DB.Model(&invoice).Updates(map[string]any{
		"amount_cents": 100, "tax_cents": 20, "total_cents": 999,
	}).Error; err != nil {
		t.Fatalf("prepare invalid amount: %v", err)
	}
	_, err := (InvoiceService{}).Confirm(invoice.ID, AccessScope{UserID: 31, AuthorityID: 300})
	if err == nil || !strings.Contains(err.Error(), "价税合计") {
		t.Fatalf("expected amount validation error, got %v", err)
	}
}

func TestRecognitionResultDoesNotOverwriteManualReview(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 41, 400, category.ID, "ORIGINAL")
	if err := global.GVA_DB.Model(&invoice).Updates(map[string]any{
		"status": model.InvoiceStatusRecognizing,
	}).Error; err != nil {
		t.Fatalf("mark invoice recognizing: %v", err)
	}
	if err := global.GVA_DB.First(&invoice, invoice.ID).Error; err != nil {
		t.Fatalf("reload invoice snapshot: %v", err)
	}
	now := time.Now()
	job := model.RecognitionJob{
		InvoiceID: invoice.ID, Status: model.RecognitionJobProcessing, Attempts: 1,
		MaxAttempts: 3, LockedAt: &now, LockToken: "active-token",
	}
	if err := global.GVA_DB.Create(&job).Error; err != nil {
		t.Fatalf("create recognition job: %v", err)
	}
	if err := global.GVA_DB.Model(&model.Invoice{}).Where("id = ?", invoice.ID).Updates(map[string]any{
		"invoice_number": "MANUAL", "status": model.InvoiceStatusPendingReview,
		"classification_source": model.ClassificationManual,
	}).Error; err != nil {
		t.Fatalf("apply manual review: %v", err)
	}

	err := saveRecognitionResult(invoice, job, provider.Result{
		Provider: "test", InvoiceNumber: "OCR-OVERWRITE", AmountCents: 100,
		TotalCents: 100, Confidence: 0.9,
	})
	if err != nil {
		t.Fatalf("finish obsolete recognition result: %v", err)
	}
	var persisted model.Invoice
	if err = global.GVA_DB.First(&persisted, invoice.ID).Error; err != nil {
		t.Fatalf("reload reviewed invoice: %v", err)
	}
	if persisted.InvoiceNumber != "MANUAL" || persisted.Status != model.InvoiceStatusPendingReview {
		t.Fatalf("recognition overwrote manual review: %#v", persisted)
	}
	var persistedJob model.RecognitionJob
	if err = global.GVA_DB.First(&persistedJob, job.ID).Error; err != nil {
		t.Fatalf("reload recognition job: %v", err)
	}
	if persistedJob.Status != model.RecognitionJobCompleted || persistedJob.LockToken != "" {
		t.Fatalf("obsolete recognition job was not completed safely: %#v", persistedJob)
	}
}

func TestRecognitionResultRequiresCurrentLease(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 51, 500, category.ID, "LEASE-001")
	if err := global.GVA_DB.Model(&invoice).Update("status", model.InvoiceStatusRecognizing).Error; err != nil {
		t.Fatalf("mark invoice recognizing: %v", err)
	}
	now := time.Now()
	persistedJob := model.RecognitionJob{
		InvoiceID: invoice.ID, Status: model.RecognitionJobProcessing, Attempts: 1,
		MaxAttempts: 3, LockedAt: &now, LockToken: "new-token",
	}
	if err := global.GVA_DB.Create(&persistedJob).Error; err != nil {
		t.Fatalf("create recognition job: %v", err)
	}
	staleJob := persistedJob
	staleJob.LockToken = "old-token"
	err := saveRecognitionResult(invoice, staleJob, provider.Result{
		Provider: "test", InvoiceNumber: "STALE", AmountCents: 100,
		TotalCents: 100, Confidence: 0.8,
	})
	if !errors.Is(err, errRecognitionLeaseLost) {
		t.Fatalf("expected lease loss, got %v", err)
	}
	var reloaded model.Invoice
	if err = global.GVA_DB.First(&reloaded, invoice.ID).Error; err != nil {
		t.Fatalf("reload invoice: %v", err)
	}
	if reloaded.InvoiceNumber != "LEASE-001" || reloaded.Status != model.InvoiceStatusRecognizing {
		t.Fatalf("stale worker changed invoice: %#v", reloaded)
	}
}

func TestRecoverStaleRecognitionJobsClearsExpiredAndMissingLeases(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 61, 600, category.ID, "STALE-001")
	missingLeaseInvoice := createReviewableInvoice(t, 61, 600, category.ID, "STALE-NULL-001")
	lockedAt := time.Now().Add(-recognitionLeaseTimeout - time.Minute)
	jobs := []model.RecognitionJob{{
		InvoiceID: invoice.ID, Status: model.RecognitionJobProcessing, Attempts: 1,
		MaxAttempts: 3, LockedAt: &lockedAt, LockToken: "expired-token",
	}, {
		InvoiceID: missingLeaseInvoice.ID, Status: model.RecognitionJobProcessing, Attempts: 1,
		MaxAttempts: 3, LockToken: "missing-lease-token",
	}}
	if err := global.GVA_DB.Create(&jobs).Error; err != nil {
		t.Fatalf("create stale jobs: %v", err)
	}
	recoverStaleRecognitionJobs()
	for _, job := range jobs {
		var recovered model.RecognitionJob
		if err := global.GVA_DB.First(&recovered, job.ID).Error; err != nil {
			t.Fatalf("reload stale job: %v", err)
		}
		if recovered.Status != model.RecognitionJobPending || recovered.LockedAt != nil || recovered.LockToken != "" {
			t.Fatalf("stale lease was not recovered: %#v", recovered)
		}
	}
}

func TestRecoverCleanupJobWithMissingLease(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	job := model.InvoiceFileCleanupJob{
		FileKey: "missing-lease.png", StorageType: "local",
		Status: model.FileCleanupJobProcessing, Attempts: 1, LockToken: "missing-lease-token",
	}
	if err := global.GVA_DB.Create(&job).Error; err != nil {
		t.Fatalf("create cleanup job: %v", err)
	}

	recoverStaleInvoiceFileCleanupJobs()

	var recovered model.InvoiceFileCleanupJob
	if err := global.GVA_DB.First(&recovered, job.ID).Error; err != nil {
		t.Fatalf("reload cleanup job: %v", err)
	}
	if recovered.Status != model.FileCleanupJobPending || recovered.LockedAt != nil || recovered.LockToken != "" {
		t.Fatalf("cleanup lease was not recovered: %#v", recovered)
	}
}

func TestDeleteInvoiceObjectRejectsMissingPersistedLocation(t *testing.T) {
	currentRoot := t.TempDir()
	filePath := filepath.Join(currentRoot, "same-key.png")
	if err := os.WriteFile(filePath, []byte("must stay"), 0o600); err != nil {
		t.Fatalf("write current object: %v", err)
	}
	previousRoot := global.GVA_CONFIG.Local.StorePath
	global.GVA_CONFIG.Local.StorePath = currentRoot
	t.Cleanup(func() { global.GVA_CONFIG.Local.StorePath = previousRoot })

	err := deleteInvoiceObject(t.Context(), model.Invoice{
		FileKey: "same-key.png", StorageType: "local",
	})
	if err == nil || !strings.Contains(err.Error(), "存储位置缺失") {
		t.Fatalf("expected missing persisted location error, got %v", err)
	}
	if _, err = os.Stat(filePath); err != nil {
		t.Fatalf("current object was touched: %v", err)
	}
}

func TestDeleteUsesTransactionalCleanupOutbox(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 71, 700, category.ID, "DELETE-001")
	storageRoot := t.TempDir()
	fileKey := "delete-001.png"
	filePath := filepath.Join(storageRoot, fileKey)
	if err := os.WriteFile(filePath, []byte("private invoice"), 0o600); err != nil {
		t.Fatalf("write invoice object: %v", err)
	}
	if err := global.GVA_DB.Model(&invoice).Updates(map[string]any{
		"file_key": fileKey, "storage_type": "local", "storage_root": storageRoot,
	}).Error; err != nil {
		t.Fatalf("configure invoice storage: %v", err)
	}
	job := model.RecognitionJob{InvoiceID: invoice.ID, Status: model.RecognitionJobPending, MaxAttempts: 3}
	if err := global.GVA_DB.Create(&job).Error; err != nil {
		t.Fatalf("create recognition job: %v", err)
	}

	if err := (InvoiceService{}).Delete(invoice.ID, AccessScope{UserID: 71, AuthorityID: 700}); err != nil {
		t.Fatalf("delete invoice: %v", err)
	}
	if _, err := os.Stat(filePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("invoice object was not removed: %v", err)
	}
	if err := global.GVA_DB.First(&model.Invoice{}, invoice.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("invoice record still visible after delete: %v", err)
	}
	var cleanupCount int64
	if err := global.GVA_DB.Model(&model.InvoiceFileCleanupJob{}).Count(&cleanupCount).Error; err != nil || cleanupCount != 0 {
		t.Fatalf("completed cleanup job was not removed: count=%d err=%v", cleanupCount, err)
	}
}

func TestDeleteKeepsRetryableCleanupJobWhenStorageFails(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 81, 800, category.ID, "DELETE-RETRY-001")
	if err := global.GVA_DB.Model(&invoice).Updates(map[string]any{
		"storage_type": "unavailable-storage", "file_key": "retry-object.png",
	}).Error; err != nil {
		t.Fatalf("configure unavailable storage: %v", err)
	}

	if err := (InvoiceService{}).Delete(invoice.ID, AccessScope{UserID: 81, AuthorityID: 800}); err != nil {
		t.Fatalf("delete should commit with retryable cleanup: %v", err)
	}
	if err := global.GVA_DB.First(&model.Invoice{}, invoice.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("invoice record still visible after outbox commit: %v", err)
	}
	var cleanupJob model.InvoiceFileCleanupJob
	if err := global.GVA_DB.First(&cleanupJob).Error; err != nil {
		t.Fatalf("load retryable cleanup job: %v", err)
	}
	if cleanupJob.Status != model.FileCleanupJobPending || cleanupJob.Attempts != 1 || cleanupJob.LastError == "" || cleanupJob.NextRunAt == nil {
		t.Fatalf("cleanup failure was not preserved for retry: %#v", cleanupJob)
	}
}
