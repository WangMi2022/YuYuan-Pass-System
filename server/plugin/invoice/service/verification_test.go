package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model"
	invoiceRequest "github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model/request"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/provider"
	"gorm.io/gorm"
)

func successfulVerificationResult(request provider.VerificationRequest) provider.VerificationResult {
	official := map[string]string{
		"invoiceNumber": request.InvoiceNum,
		"issueDate":     request.InvoiceDate,
	}
	if request.InvoiceCode != "" {
		official["invoiceCode"] = request.InvoiceCode
	}
	if request.CheckCode != "" {
		official["checkCode"] = request.CheckCode
	}
	switch request.AmountMode {
	case model.VerificationAmountModeAmount:
		official["amountCents"] = "10000"
	case model.VerificationAmountModeTotal:
		official["totalCents"] = "10600"
	}
	return provider.VerificationResult{
		Outcome:      provider.VerificationOutcomeValid,
		VerifyResult: "0001", VerifyMessage: "查验成功发票一致", InvalidSign: "N", Official: official,
	}
}

func resetInvoiceVerification(t *testing.T, invoice model.Invoice) {
	t.Helper()
	if err := global.GVA_DB.Model(&invoice).Updates(map[string]any{
		"verification_status":      model.InvoiceVerificationUnverified,
		"verification_fingerprint": "", "verification_checked_at": nil,
	}).Error; err != nil {
		t.Fatal(err)
	}
}

func TestBuildVerificationRequestFollowsCanonicalTicketRules(t *testing.T) {
	issueDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	base := model.Invoice{
		InvoiceCode: "044001", InvoiceNumber: "12345678", CheckCode: "123456",
		IssueDate: &issueDate, AmountCents: 10000, TaxCents: 600, TotalCents: 10600,
	}
	tests := []struct {
		name        string
		invoice     model.Invoice
		amountMode  string
		totalAmount string
		wantError   string
		clearCode   bool
		clearCheck  bool
	}{
		{name: "special amount", invoice: model.Invoice{VerificationType: "special_vat_invoice"}, amountMode: "amount", totalAmount: "100.00"},
		{name: "electronic special check code", invoice: model.Invoice{VerificationType: "elec_special_vat_invoice"}, clearCheck: true, wantError: "校验码"},
		{name: "digital special total", invoice: model.Invoice{VerificationType: "elec_invoice_special"}, clearCode: true, amountMode: "total", totalAmount: "106.00"},
		{name: "paper motor requires code", invoice: model.Invoice{VerificationType: "motor_vehicle_invoice", VerificationAmountMode: "amount"}, clearCode: true, wantError: "发票代码"},
		{name: "electronic motor total", invoice: model.Invoice{VerificationType: "motor_vehicle_invoice", VerificationAmountMode: "total"}, clearCode: true, amountMode: "total", totalAmount: "106.00"},
		{name: "normal uses check code", invoice: model.Invoice{VerificationType: "normal_invoice"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			invoice := base
			invoice.VerificationType = test.invoice.VerificationType
			invoice.VerificationAmountMode = test.invoice.VerificationAmountMode
			if test.clearCode {
				invoice.InvoiceCode = ""
			}
			if test.clearCheck {
				invoice.CheckCode = ""
			}
			request, err := buildVerificationRequest(invoice)
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("error = %v, want %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if request.AmountMode != test.amountMode || request.TotalAmount != test.totalAmount {
				t.Fatalf("request = %#v", request)
			}
		})
	}
}

func TestVerificationSuccessWithoutRequiredAuthorityFieldsIsUnavailable(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 301, 100, category.ID, "VERIFY-EMPTY-OFFICIAL")
	resetInvoiceVerification(t, invoice)

	previousFactory := newInvoiceVerifier
	newInvoiceVerifier = func(config.InvoiceRecognition) (provider.VerificationAdapter, error) {
		return testVerificationAdapter(verifierFunc(func(context.Context, provider.VerificationRequest) (provider.VerificationResult, error) {
			return provider.VerificationResult{
				Outcome:      provider.VerificationOutcomeValid,
				VerifyResult: "0001", VerifyMessage: "查验成功发票一致", InvalidSign: "N",
				Official: map[string]string{},
			}, nil
		})), nil
	}
	t.Cleanup(func() { newInvoiceVerifier = previousFactory })

	outcome, err := (VerificationService{}).Verify(t.Context(), invoice.ID, AccessScope{UserID: 301, AuthorityID: 100})
	if err != nil {
		t.Fatal(err)
	}
	if outcome.Invoice.VerificationStatus != model.InvoiceVerificationUnavailable {
		t.Fatalf("missing authority fields produced status %q", outcome.Invoice.VerificationStatus)
	}
	if !strings.Contains(outcome.Attempt.VerifyMessage, "缺少权威字段") {
		t.Fatalf("missing authority fields were not explained: %#v", outcome.Attempt)
	}
	if _, err = (InvoiceService{}).Confirm(invoice.ID, AccessScope{UserID: 301, AuthorityID: 100}); err == nil {
		t.Fatal("invoice with incomplete authority response was confirmed")
	}
}

func TestVerificationDisabledDoesNotCreateProviderOrHistory(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	setInvoiceVerificationEnabledForTest(false)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 305, 100, category.ID, "VERIFY-SWITCH-OFF")
	resetInvoiceVerification(t, invoice)

	providerCreated := false
	previousFactory := newInvoiceVerifier
	newInvoiceVerifier = func(config.InvoiceRecognition) (provider.VerificationAdapter, error) {
		providerCreated = true
		return provider.VerificationAdapter{}, errors.New("provider must not be created")
	}
	t.Cleanup(func() { newInvoiceVerifier = previousFactory })

	_, err := (VerificationService{}).Verify(t.Context(), invoice.ID, AccessScope{UserID: 305, AuthorityID: 100})
	if err == nil || !strings.Contains(err.Error(), "已在运行配置中关闭") {
		t.Fatalf("unexpected disabled verification error: %v", err)
	}
	if providerCreated {
		t.Fatal("disabled verification created a provider")
	}
	var count int64
	if err = global.GVA_DB.Model(&model.InvoiceVerification{}).Where("invoice_id = ?", invoice.ID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("disabled verification created %d history records", count)
	}
}

func TestVerificationDisabledBeforeProviderCallReleasesAttempt(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 306, 100, category.ID, "VERIFY-SWITCH-DURING-START")
	resetInvoiceVerification(t, invoice)

	providerCalled := false
	previousFactory := newInvoiceVerifier
	newInvoiceVerifier = func(config.InvoiceRecognition) (provider.VerificationAdapter, error) {
		setInvoiceVerificationEnabledForTest(false)
		return testVerificationAdapter(verifierFunc(func(context.Context, provider.VerificationRequest) (provider.VerificationResult, error) {
			providerCalled = true
			return provider.VerificationResult{}, nil
		})), nil
	}
	t.Cleanup(func() { newInvoiceVerifier = previousFactory })

	_, err := (VerificationService{}).Verify(t.Context(), invoice.ID, AccessScope{UserID: 306, AuthorityID: 100})
	if err == nil || !strings.Contains(err.Error(), "已在运行配置中关闭") {
		t.Fatalf("unexpected verification error: %v", err)
	}
	if providerCalled {
		t.Fatal("provider was called after verification was disabled")
	}
	var persisted model.Invoice
	if err = global.GVA_DB.First(&persisted, invoice.ID).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.ActiveVerificationID != nil || persisted.VerificationStatus != model.InvoiceVerificationUnavailable {
		t.Fatalf("disabled attempt was not released: %#v", persisted)
	}
	var attempt model.InvoiceVerification
	if err = global.GVA_DB.Where("invoice_id = ?", invoice.ID).First(&attempt).Error; err != nil {
		t.Fatal(err)
	}
	if attempt.CompletedAt == nil || attempt.Status != model.InvoiceVerificationUnavailable || !strings.Contains(attempt.VerifyMessage, "已在运行配置中关闭") {
		t.Fatalf("disabled attempt was not completed: %#v", attempt)
	}
}

func TestEditingDuringVerificationKeepsLeaseAndMakesResultStale(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 302, 100, category.ID, "VERIFY-ACTIVE-LEASE")
	resetInvoiceVerification(t, invoice)

	started := make(chan struct{})
	release := make(chan struct{})
	previousFactory := newInvoiceVerifier
	newInvoiceVerifier = func(config.InvoiceRecognition) (provider.VerificationAdapter, error) {
		return testVerificationAdapter(verifierFunc(func(_ context.Context, request provider.VerificationRequest) (provider.VerificationResult, error) {
			close(started)
			<-release
			return successfulVerificationResult(request), nil
		})), nil
	}
	t.Cleanup(func() { newInvoiceVerifier = previousFactory })

	verifyDone := make(chan error, 1)
	go func() {
		_, err := (VerificationService{}).Verify(t.Context(), invoice.ID, AccessScope{UserID: 302, AuthorityID: 100})
		verifyDone <- err
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("verification did not start")
	}

	_, err := (InvoiceService{}).Update(invoiceRequest.InvoiceUpdate{
		ID: invoice.ID, Direction: invoice.Direction, InvoiceType: invoice.InvoiceType,
		VerificationType: invoice.VerificationType, VerificationAmountMode: invoice.VerificationAmountMode,
		InvoiceCode: invoice.InvoiceCode, InvoiceNumber: "VERIFY-EDITED", CheckCode: invoice.CheckCode,
		IssueDate: invoice.IssueDate, BuyerName: invoice.BuyerName, BuyerTaxNo: invoice.BuyerTaxNo,
		SellerName: invoice.SellerName, SellerTaxNo: invoice.SellerTaxNo,
		AmountCents: invoice.AmountCents, TaxCents: invoice.TaxCents, TotalCents: invoice.TotalCents,
		CategoryID: invoice.CategoryID,
	}, AccessScope{UserID: 302, AuthorityID: 100})
	if err != nil {
		close(release)
		t.Fatalf("edit invoice during verification: %v", err)
	}
	var edited model.Invoice
	if err = global.GVA_DB.First(&edited, invoice.ID).Error; err != nil {
		close(release)
		t.Fatal(err)
	}
	if edited.ActiveVerificationID == nil || edited.VerificationStatus != model.InvoiceVerificationVerifying {
		close(release)
		t.Fatalf("edit released active verification lease: %#v", edited)
	}
	if _, err = (VerificationService{}).Verify(t.Context(), invoice.ID, AccessScope{UserID: 302, AuthorityID: 100}); err == nil || !strings.Contains(err.Error(), "正在查验") {
		close(release)
		t.Fatalf("second verification was not blocked: %v", err)
	}
	close(release)
	select {
	case err = <-verifyDone:
		if err != nil {
			t.Fatalf("finish first verification: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("first verification did not finish")
	}
	if err = global.GVA_DB.First(&edited, invoice.ID).Error; err != nil {
		t.Fatal(err)
	}
	if edited.ActiveVerificationID != nil || edited.VerificationStatus != model.InvoiceVerificationUnverified {
		t.Fatalf("stale verification result was accepted: %#v", edited)
	}
}

func TestExpiredVerificationLeaseCanBeTakenOverWithoutLateOverwrite(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 303, 100, category.ID, "VERIFY-EXPIRED-LEASE")
	resetInvoiceVerification(t, invoice)
	oldAttempt := model.InvoiceVerification{
		InvoiceID: invoice.ID, Provider: "baidu-vat-invoice-verification",
		Status: model.InvoiceVerificationVerifying, LocalFingerprint: invoiceVerificationFingerprint(invoice),
		RequestedBy: 303,
	}
	if err := global.GVA_DB.Create(&oldAttempt).Error; err != nil {
		t.Fatal(err)
	}
	startedAt := time.Now().Add(-verificationLeaseTimeout - time.Minute)
	if err := global.GVA_DB.Model(&invoice).Updates(map[string]any{
		"verification_status":    model.InvoiceVerificationVerifying,
		"active_verification_id": oldAttempt.ID, "verification_started_at": &startedAt,
	}).Error; err != nil {
		t.Fatal(err)
	}

	previousFactory := newInvoiceVerifier
	newInvoiceVerifier = func(config.InvoiceRecognition) (provider.VerificationAdapter, error) {
		return testVerificationAdapter(verifierFunc(func(_ context.Context, request provider.VerificationRequest) (provider.VerificationResult, error) {
			return successfulVerificationResult(request), nil
		})), nil
	}
	t.Cleanup(func() { newInvoiceVerifier = previousFactory })

	outcome, err := (VerificationService{}).Verify(t.Context(), invoice.ID, AccessScope{UserID: 303, AuthorityID: 100})
	if err != nil {
		t.Fatal(err)
	}
	var expired model.InvoiceVerification
	if err = global.GVA_DB.First(&expired, oldAttempt.ID).Error; err != nil {
		t.Fatal(err)
	}
	if expired.Status != model.InvoiceVerificationUnavailable || expired.CompletedAt == nil {
		t.Fatalf("expired attempt was not completed: %#v", expired)
	}
	lateResult := successfulVerificationResult(provider.VerificationRequest{
		InvoiceCode: invoice.InvoiceCode, InvoiceNum: invoice.InvoiceNumber,
		InvoiceDate: invoice.IssueDate.Format("20060102"), AmountMode: model.VerificationAmountModeAmount,
	})
	if err = finishVerificationAttempt(&oldAttempt, lateResult, model.InvoiceVerificationVerifiedValid, nil); !errors.Is(err, errVerificationLeaseLost) {
		t.Fatalf("late attempt should lose its lease, got %v", err)
	}
	var current model.Invoice
	if err = global.GVA_DB.First(&current, invoice.ID).Error; err != nil {
		t.Fatal(err)
	}
	if current.LatestVerificationID == nil || *current.LatestVerificationID != outcome.Attempt.ID ||
		current.VerificationStatus != model.InvoiceVerificationVerifiedValid {
		t.Fatalf("late attempt overwrote current verification: %#v", current)
	}
}

func TestConfirmCASRejectsVerificationChangedAfterValidation(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	category := createInvoiceTestCategory(t)
	invoice := createReviewableInvoice(t, 304, 100, category.ID, "CONFIRM-CAS")
	previousHook := beforeInvoiceConfirmPersist
	beforeInvoiceConfirmPersist = func(tx *gorm.DB, current model.Invoice) error {
		return tx.Model(&model.Invoice{}).Where("id = ?", current.ID).Updates(map[string]any{
			"verification_status":      model.InvoiceVerificationUnverified,
			"verification_fingerprint": "", "invoice_number": "CONFIRM-CAS-EDITED",
		}).Error
	}
	t.Cleanup(func() { beforeInvoiceConfirmPersist = previousHook })

	if _, err := (InvoiceService{}).Confirm(invoice.ID, AccessScope{UserID: 304, AuthorityID: 100}); err == nil || !strings.Contains(err.Error(), "状态已变更") {
		t.Fatalf("confirm CAS did not reject changed verification: %v", err)
	}
	var current model.Invoice
	if err := global.GVA_DB.First(&current, invoice.ID).Error; err != nil {
		t.Fatal(err)
	}
	if current.Status == model.InvoiceStatusConfirmed {
		t.Fatal("invoice was confirmed after verification changed")
	}
}
