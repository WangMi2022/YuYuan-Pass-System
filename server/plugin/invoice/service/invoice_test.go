package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model"
	invoiceRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/provider"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type recognizerFunc func(context.Context, provider.Input) (provider.Result, error)

func (fn recognizerFunc) Recognize(ctx context.Context, input provider.Input) (provider.Result, error) {
	return fn(ctx, input)
}

type verifierFunc func(context.Context, provider.VerificationRequest) (provider.VerificationResult, error)

func (fn verifierFunc) Verify(ctx context.Context, request provider.VerificationRequest) (provider.VerificationResult, error) {
	return fn(ctx, request)
}

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
		&model.InvoiceVerification{}, &model.ClassificationRule{}, &model.RecognitionJob{}, &model.InvoiceFileCleanupJob{},
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
	verifiedAt := time.Now()
	invoice := model.Invoice{
		Direction: "expense", InvoiceType: "增值税专用发票", VerificationType: "special_vat_invoice",
		InvoiceCode: "044001", InvoiceNumber: number,
		IssueDate: &issueDate, SellerName: "测试销售方", SellerTaxNo: "91440000TEST",
		AmountCents: 10000, TaxCents: 600, TotalCents: 10600, Currency: "CNY",
		CategoryID: &categoryID, Status: model.InvoiceStatusPendingReview,
		VerificationStatus:   model.InvoiceVerificationVerifiedValid,
		VerificationProvider: "test-verifier", VerificationCheckedAt: &verifiedAt,
		FileName: number + ".png", FileKey: number + ".png", FileHash: hashInvoiceKey(number),
		MimeType: "image/png", FileSize: 128, StorageType: "local",
		CreatedBy: userID, AuthorityID: authorityID,
	}
	invoice.VerificationFingerprint = invoiceVerificationFingerprint(invoice)
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

func TestInvoiceListIgnoresZeroValueOptionalDates(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	createReviewableInvoice(t, 14, 100, category.ID, "EMPTY-DATE-FILTER")
	zeroDate := time.Time{}
	search := invoiceRequest.InvoiceSearch{
		PageInfo:  commonRequest.PageInfo{Page: 1, PageSize: 20},
		StartDate: &zeroDate,
		EndDate:   &zeroDate,
	}

	list, total, err := (InvoiceService{}).List(search, AccessScope{UserID: 14, AuthorityID: 100})
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("zero-value optional dates should not filter invoices: total=%d list=%#v err=%v", total, list, err)
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

func TestConfirmRequiresVerificationForCurrentFields(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 211, 100, category.ID, "VERIFY-GATE-001")
	if err := global.GVA_DB.Model(&invoice).Updates(map[string]any{
		"verification_status":      model.InvoiceVerificationUnverified,
		"verification_fingerprint": "", "verification_checked_at": nil,
	}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := (InvoiceService{}).Confirm(invoice.ID, AccessScope{UserID: 211, AuthorityID: 100}); err == nil || !strings.Contains(err.Error(), "权威查验") {
		t.Fatalf("unverified invoice was confirmed: %v", err)
	}

	verifiedAt := time.Now()
	staleFingerprint := hashInvoiceKey("stale")
	if err := global.GVA_DB.Model(&invoice).Updates(map[string]any{
		"verification_status":      model.InvoiceVerificationVerifiedValid,
		"verification_fingerprint": staleFingerprint, "verification_checked_at": &verifiedAt,
	}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := (InvoiceService{}).Confirm(invoice.ID, AccessScope{UserID: 211, AuthorityID: 100}); err == nil || !strings.Contains(err.Error(), "字段已在查验后变更") {
		t.Fatalf("stale verification was accepted: %v", err)
	}
}

func TestBaiduVerificationPersistsHistoryAndAllowsConfirmation(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 212, 100, category.ID, "VERIFY-OK-001")
	if err := global.GVA_DB.Model(&invoice).Updates(map[string]any{
		"verification_status":      model.InvoiceVerificationUnverified,
		"verification_fingerprint": "", "verification_checked_at": nil,
	}).Error; err != nil {
		t.Fatal(err)
	}

	previousFactory := newInvoiceVerifier
	newInvoiceVerifier = func(config.InvoiceRecognition) (provider.Verifier, error) {
		return verifierFunc(func(_ context.Context, request provider.VerificationRequest) (provider.VerificationResult, error) {
			if request.InvoiceType != "special_vat_invoice" || request.TotalAmount != "100.00" {
				t.Fatalf("unexpected verification request: %#v", request)
			}
			return provider.VerificationResult{
				VerifyResult: "0001", VerifyMessage: "查验成功发票一致", InvalidSign: "N",
				VerifyFrequency: "1", ProviderLogID: "log-1",
				Official: map[string]string{
					"invoiceType": "增值税专用发票", "invoiceCode": invoice.InvoiceCode,
					"invoiceNumber": invoice.InvoiceNumber, "issueDate": "20260701",
					"sellerName": invoice.SellerName, "sellerTaxNo": invoice.SellerTaxNo,
					"amountCents": "10000", "taxCents": "600", "totalCents": "10600",
				}, RawPayload: `{"words_result":{"VerifyResult":"0001"}}`,
			}, nil
		}), nil
	}
	t.Cleanup(func() { newInvoiceVerifier = previousFactory })

	outcome, err := (VerificationService{}).Verify(t.Context(), invoice.ID, AccessScope{UserID: 212, AuthorityID: 100})
	if err != nil {
		t.Fatal(err)
	}
	if outcome.Invoice.VerificationStatus != model.InvoiceVerificationVerifiedValid || outcome.Attempt.ID == 0 || len(outcome.Attempt.Differences) != 0 {
		t.Fatalf("unexpected verification outcome: %#v", outcome)
	}
	history, err := (VerificationService{}).History(invoice.ID, AccessScope{UserID: 212, AuthorityID: 100})
	if err != nil || len(history) != 1 || history[0].ProviderLogID != "log-1" {
		t.Fatalf("unexpected public history: %#v err=%v", history, err)
	}
	serializedHistory, err := json.Marshal(history)
	if err != nil || strings.Contains(string(serializedHistory), "words_result") {
		t.Fatalf("raw verification payload leaked in JSON: %s err=%v", serializedHistory, err)
	}
	if _, err = (InvoiceService{}).Confirm(invoice.ID, AccessScope{UserID: 212, AuthorityID: 100}); err != nil {
		t.Fatalf("verified invoice could not be confirmed: %v", err)
	}
}

func TestVerificationMismatchCannotConfirm(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 213, 100, category.ID, "VERIFY-DIFF-001")
	if err := global.GVA_DB.Model(&invoice).Updates(map[string]any{
		"verification_status":      model.InvoiceVerificationUnverified,
		"verification_fingerprint": "", "verification_checked_at": nil,
	}).Error; err != nil {
		t.Fatal(err)
	}
	previousFactory := newInvoiceVerifier
	newInvoiceVerifier = func(config.InvoiceRecognition) (provider.Verifier, error) {
		return verifierFunc(func(context.Context, provider.VerificationRequest) (provider.VerificationResult, error) {
			return provider.VerificationResult{
				VerifyResult: "0001", VerifyMessage: "查验成功发票一致", InvalidSign: "N",
				Official: map[string]string{"sellerName": "权威销售方"},
			}, nil
		}), nil
	}
	t.Cleanup(func() { newInvoiceVerifier = previousFactory })
	outcome, err := (VerificationService{}).Verify(t.Context(), invoice.ID, AccessScope{UserID: 213, AuthorityID: 100})
	if err != nil {
		t.Fatal(err)
	}
	if outcome.Invoice.VerificationStatus != model.InvoiceVerificationInconsistent || len(outcome.Attempt.Differences) != 1 {
		t.Fatalf("mismatch was not preserved: %#v", outcome)
	}
	if _, err = (InvoiceService{}).Confirm(invoice.ID, AccessScope{UserID: 213, AuthorityID: 100}); err == nil || !strings.Contains(err.Error(), "权威查验") {
		t.Fatalf("mismatched invoice was confirmed: %v", err)
	}
}

func TestReopenConfirmedInvoiceReleasesConfirmation(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	first := createReviewableInvoice(t, 23, 100, category.ID, "REOPEN-001")
	second := createReviewableInvoice(t, 24, 200, category.ID, "REOPEN-001")
	admin := AccessScope{UserID: 1, AuthorityID: defaultAdminRoleID, All: true}

	confirmed, err := (InvoiceService{}).Confirm(first.ID, admin)
	if err != nil || confirmed.DuplicateKey == nil || confirmed.ConfirmedAt == nil {
		t.Fatalf("confirm invoice before reopening: invoice=%#v err=%v", confirmed, err)
	}
	reopened, err := (InvoiceService{}).Reopen(first.ID, admin)
	if err != nil {
		t.Fatalf("reopen confirmed invoice: %v", err)
	}
	if reopened.Status != model.InvoiceStatusPendingReview || reopened.DuplicateKey != nil || reopened.ConfirmedBy != 0 || reopened.ConfirmedAt != nil {
		t.Fatalf("reopened invoice kept confirmation state: %#v", reopened)
	}
	if _, err = (InvoiceService{}).Confirm(second.ID, admin); err != nil {
		t.Fatalf("released duplicate key should allow another invoice to be confirmed: %v", err)
	}
	dashboard, err := (InvoiceService{}).Dashboard(admin)
	if err != nil || dashboard.ConfirmedCount != 1 || dashboard.TotalCents != second.TotalCents {
		t.Fatalf("dashboard should exclude reopened invoice: %#v err=%v", dashboard, err)
	}
}

func TestReopenRequiresAdminAndConfirmedStatus(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 25, 100, category.ID, "REOPEN-AUTH-001")
	admin := AccessScope{UserID: 1, AuthorityID: defaultAdminRoleID, All: true}

	if _, err := (InvoiceService{}).Reopen(invoice.ID, admin); err == nil || !strings.Contains(err.Error(), "尚未确认") {
		t.Fatalf("expected pending invoice to reject reopen, got %v", err)
	}
	if _, err := (InvoiceService{}).Confirm(invoice.ID, admin); err != nil {
		t.Fatalf("confirm invoice before permission check: %v", err)
	}
	if _, err := (InvoiceService{}).Reopen(invoice.ID, AccessScope{UserID: 25, AuthorityID: 100}); err == nil || !strings.Contains(err.Error(), "管理员") {
		t.Fatalf("expected non-admin reopen to be rejected, got %v", err)
	}
	var persisted model.Invoice
	if err := global.GVA_DB.First(&persisted, invoice.ID).Error; err != nil || persisted.Status != model.InvoiceStatusConfirmed {
		t.Fatalf("rejected reopen changed invoice state: invoice=%#v err=%v", persisted, err)
	}
}

func TestUpdateConfirmedInvoiceRequiresExplicitReopen(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 26, 100, category.ID, "REOPEN-BEFORE-EDIT")
	admin := AccessScope{UserID: 1, AuthorityID: defaultAdminRoleID, All: true}
	confirmed, err := (InvoiceService{}).Confirm(invoice.ID, admin)
	if err != nil {
		t.Fatalf("confirm invoice before update guard: %v", err)
	}

	confirmed.SellerName = "不应写入的销售方"
	_, err = (InvoiceService{}).Update(invoiceRequest.InvoiceUpdate{
		ID: confirmed.ID, Direction: confirmed.Direction, InvoiceType: confirmed.InvoiceType,
		InvoiceCode: confirmed.InvoiceCode, InvoiceNumber: confirmed.InvoiceNumber, IssueDate: confirmed.IssueDate,
		BuyerName: confirmed.BuyerName, BuyerTaxNo: confirmed.BuyerTaxNo,
		SellerName: confirmed.SellerName, SellerTaxNo: confirmed.SellerTaxNo,
		AmountCents: confirmed.AmountCents, TaxCents: confirmed.TaxCents, TotalCents: confirmed.TotalCents,
		CategoryID: confirmed.CategoryID, ReviewNotes: confirmed.ReviewNotes,
	}, admin)
	if err == nil || !strings.Contains(err.Error(), "先重新打开") {
		t.Fatalf("expected confirmed update to require reopen, got %v", err)
	}
	var persisted model.Invoice
	if err = global.GVA_DB.First(&persisted, invoice.ID).Error; err != nil || persisted.Status != model.InvoiceStatusConfirmed || persisted.SellerName == confirmed.SellerName {
		t.Fatalf("rejected confirmed update changed invoice: invoice=%#v err=%v", persisted, err)
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

func TestRecheckReturnsModelCandidateWithoutPersistingIt(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 52, 500, category.ID, "RECHECK-ORIGINAL")
	storageRoot := t.TempDir()
	fileKey := "recheck.png"
	if err := os.WriteFile(filepath.Join(storageRoot, fileKey), []byte("private invoice image"), 0o600); err != nil {
		t.Fatalf("write invoice fixture: %v", err)
	}
	if err := global.GVA_DB.Model(&invoice).Updates(map[string]any{
		"storage_type": "local", "storage_root": storageRoot, "file_key": fileKey,
		"mime_type": "image/png", "file_size": 21,
	}).Error; err != nil {
		t.Fatalf("configure invoice storage: %v", err)
	}

	previousFactory := newRecheckRecognizer
	newRecheckRecognizer = func(config.InvoiceRecognition) (provider.Recognizer, error) {
		return recognizerFunc(func(_ context.Context, input provider.Input) (provider.Result, error) {
			if string(input.Data) != "private invoice image" {
				t.Fatalf("unexpected recheck input: %q", string(input.Data))
			}
			return provider.Result{
				Provider: "multimodal-ai", InvoiceNumber: "MODEL-CANDIDATE",
				AmountCents: 10000, TaxCents: 600, TotalCents: 10600,
				Confidence: 0.94, RawText: "sensitive raw text", RawPayload: "large provider payload",
				Items: []model.InvoiceItem{{InvoiceID: 999, Name: "服务费", AmountCents: 10000}},
			}, nil
		}), nil
	}
	t.Cleanup(func() { newRecheckRecognizer = previousFactory })

	result, err := (RecognitionService{}).Recheck(
		t.Context(), invoice.ID, AccessScope{UserID: 52, AuthorityID: 500},
	)
	if err != nil {
		t.Fatalf("recheck invoice: %v", err)
	}
	if result.InvoiceNumber != "MODEL-CANDIDATE" || result.RawText != "" || result.RawPayload != "" ||
		len(result.Items) != 1 || result.Items[0].InvoiceID != 0 {
		t.Fatalf("unexpected recheck candidate: %#v", result)
	}
	var persisted model.Invoice
	if err = global.GVA_DB.First(&persisted, invoice.ID).Error; err != nil {
		t.Fatalf("reload invoice after recheck: %v", err)
	}
	if persisted.InvoiceNumber != "RECHECK-ORIGINAL" {
		t.Fatalf("recheck persisted an unconfirmed model result: %#v", persisted)
	}
}

func TestRecheckRejectsUnauthorizedAndConfirmedInvoicesBeforeCallingModel(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 53, 500, category.ID, "RECHECK-GUARD")
	previousFactory := newRecheckRecognizer
	var factoryCalls atomic.Int32
	newRecheckRecognizer = func(config.InvoiceRecognition) (provider.Recognizer, error) {
		factoryCalls.Add(1)
		return recognizerFunc(func(context.Context, provider.Input) (provider.Result, error) {
			return provider.Result{}, nil
		}), nil
	}
	t.Cleanup(func() { newRecheckRecognizer = previousFactory })

	if _, err := (RecognitionService{}).Recheck(
		t.Context(), invoice.ID, AccessScope{UserID: 999, AuthorityID: 999},
	); err == nil {
		t.Fatal("unauthorized recheck unexpectedly succeeded")
	}
	if err := global.GVA_DB.Model(&invoice).Update("status", model.InvoiceStatusConfirmed).Error; err != nil {
		t.Fatalf("confirm invoice fixture: %v", err)
	}
	if _, err := (RecognitionService{}).Recheck(
		t.Context(), invoice.ID, AccessScope{UserID: 53, AuthorityID: 500},
	); err == nil || !strings.Contains(err.Error(), "已确认") {
		t.Fatalf("confirmed invoice recheck error = %v", err)
	}
	if factoryCalls.Load() != 0 {
		t.Fatalf("model factory called %d times for rejected requests", factoryCalls.Load())
	}
}

func TestRecheckSingleflightSurvivesLeadingCallerCancellation(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 54, 500, category.ID, "RECHECK-SHARED")
	storageRoot := t.TempDir()
	fileKey := "recheck-shared.png"
	if err := os.WriteFile(filepath.Join(storageRoot, fileKey), []byte("shared invoice image"), 0o600); err != nil {
		t.Fatalf("write invoice fixture: %v", err)
	}
	if err := global.GVA_DB.Model(&invoice).Updates(map[string]any{
		"storage_type": "local", "storage_root": storageRoot, "file_key": fileKey,
		"mime_type": "image/png", "file_size": 20,
	}).Error; err != nil {
		t.Fatalf("configure invoice storage: %v", err)
	}

	previousFactory := newRecheckRecognizer
	started := make(chan struct{})
	release := make(chan struct{})
	var factoryCalls atomic.Int32
	newRecheckRecognizer = func(config.InvoiceRecognition) (provider.Recognizer, error) {
		factoryCalls.Add(1)
		return recognizerFunc(func(ctx context.Context, _ provider.Input) (provider.Result, error) {
			close(started)
			select {
			case <-release:
				return provider.Result{Provider: "multimodal-ai", InvoiceNumber: "SHARED-RESULT", Confidence: 0.9}, nil
			case <-ctx.Done():
				return provider.Result{}, ctx.Err()
			}
		}), nil
	}
	t.Cleanup(func() { newRecheckRecognizer = previousFactory })

	firstCtx, cancelFirst := context.WithCancel(t.Context())
	firstDone := make(chan error, 1)
	go func() {
		_, err := (RecognitionService{}).Recheck(
			firstCtx, invoice.ID, AccessScope{UserID: 54, AuthorityID: 500},
		)
		firstDone <- err
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("shared recheck did not reach the model")
	}
	cancelFirst()
	select {
	case err := <-firstDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("leading caller error = %v, want context canceled", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("canceled leading caller did not return promptly")
	}

	secondDone := make(chan struct {
		result provider.Result
		err    error
	}, 1)
	go func() {
		result, err := (RecognitionService{}).Recheck(
			t.Context(), invoice.ID, AccessScope{UserID: 54, AuthorityID: 500},
		)
		secondDone <- struct {
			result provider.Result
			err    error
		}{result: result, err: err}
	}()
	time.Sleep(20 * time.Millisecond)
	close(release)
	select {
	case response := <-secondDone:
		if response.err != nil || response.result.InvoiceNumber != "SHARED-RESULT" {
			t.Fatalf("joined caller result = %#v, err = %v", response.result, response.err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("joined caller did not receive the shared result")
	}
	if factoryCalls.Load() != 1 {
		t.Fatalf("model factory called %d times, want 1", factoryCalls.Load())
	}
}

func TestProcessClaimedRecognitionJobsRunsConcurrently(t *testing.T) {
	jobs := make([]model.RecognitionJob, recognitionBatchWorkers)
	started := make(chan struct{}, recognitionBatchWorkers)
	release := make(chan struct{})
	done := make(chan struct{})
	var active atomic.Int32
	var peak atomic.Int32

	go func() {
		processClaimedRecognitionJobs(jobs, func(model.RecognitionJob) {
			current := active.Add(1)
			for {
				previous := peak.Load()
				if current <= previous || peak.CompareAndSwap(previous, current) {
					break
				}
			}
			started <- struct{}{}
			<-release
			active.Add(-1)
		})
		close(done)
	}()

	for range recognitionBatchWorkers {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("recognition batch did not start all workers concurrently")
		}
	}
	if peak.Load() != recognitionBatchWorkers {
		t.Fatalf("peak recognition concurrency = %d, want %d", peak.Load(), recognitionBatchWorkers)
	}
	close(release)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("recognition batch workers did not finish")
	}
}

func TestRecognizeWithSlotCapsTotalModelConcurrency(t *testing.T) {
	requestCount := recognitionBatchWorkers + 2
	started := make(chan struct{}, requestCount)
	release := make(chan struct{})
	var active atomic.Int32
	var peak atomic.Int32
	recognizer := recognizerFunc(func(context.Context, provider.Input) (provider.Result, error) {
		current := active.Add(1)
		for {
			previous := peak.Load()
			if current <= previous || peak.CompareAndSwap(previous, current) {
				break
			}
		}
		started <- struct{}{}
		<-release
		active.Add(-1)
		return provider.Result{Provider: "test"}, nil
	})

	var workers sync.WaitGroup
	workers.Add(requestCount)
	for range requestCount {
		go func() {
			defer workers.Done()
			_, _ = recognizeWithSlot(t.Context(), recognizer, provider.Input{})
		}()
	}
	for range recognitionBatchWorkers {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("shared recognition slots did not admit the expected workers")
		}
	}
	select {
	case <-started:
		t.Fatal("shared recognition slots admitted more than four model calls")
	case <-time.After(120 * time.Millisecond):
	}
	close(release)
	workers.Wait()
	if peak.Load() != recognitionBatchWorkers {
		t.Fatalf("peak model concurrency = %d, want %d", peak.Load(), recognitionBatchWorkers)
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
