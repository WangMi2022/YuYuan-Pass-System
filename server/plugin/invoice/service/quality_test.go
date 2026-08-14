package service

import (
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/ai"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model"
	invoiceRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model/request"
)

func TestQualityMetricsAggregateRecognitionReviewAndCost(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	if err := global.GVA_DB.AutoMigrate(&ai.ModelInvocation{}); err != nil {
		t.Fatalf("migrate model invocation table: %v", err)
	}
	category := createInvoiceTestCategory(t)
	reviewedAt := time.Now()
	success := createReviewableInvoice(t, 41, 410, category.ID, "QUALITY-SUCCESS")
	success.RecognitionProvider = "ocr-provider"
	success.RecognitionModel = "ocr-model"
	success.RecognitionConfidence = 0.9
	success.RecognitionDurationMS = 100
	success.FieldConfidences = map[string]float64{"invoiceNumber": 0.9, "sellerName": 0.6}
	success.ReviewCapturedAt = &reviewedAt
	success.SuggestedCategoryID = &category.ID
	success.ClassificationConfidence = 0.88
	if err := global.GVA_DB.Save(&success).Error; err != nil {
		t.Fatal(err)
	}
	failure := createReviewableInvoice(t, 41, 410, category.ID, "QUALITY-FAILED")
	if err := global.GVA_DB.Model(&failure).Updates(map[string]any{
		"status": model.InvoiceStatusRecognitionFailed, "recognition_error": "provider timeout",
	}).Error; err != nil {
		t.Fatal(err)
	}
	legacy := createReviewableInvoice(t, 41, 410, category.ID, "QUALITY-LEGACY")
	if err := global.GVA_DB.Model(&legacy).Updates(map[string]any{
		"recognition_provider": "ocr-provider", "recognition_model": "ocr-model",
		"recognition_confidence": 0.8, "recognition_duration_ms": 200,
	}).Error; err != nil {
		t.Fatal(err)
	}
	jobs := []model.RecognitionJob{
		{InvoiceID: success.ID, Status: model.RecognitionJobCompleted, Attempts: 1, MaxAttempts: 3, Provider: "ocr-provider", Model: "ocr-model", DurationMS: 100},
		{InvoiceID: failure.ID, Status: model.RecognitionJobFailed, Attempts: 2, MaxAttempts: 3, Provider: "ocr-provider", Model: "ocr-model", DurationMS: 300, LastError: "provider timeout"},
		{InvoiceID: legacy.ID, Status: model.RecognitionJobCompleted, Attempts: 1, MaxAttempts: 3, Provider: "ocr-provider", Model: "ocr-model", DurationMS: 200, FallbackUsed: true},
	}
	if err := global.GVA_DB.Create(&jobs).Error; err != nil {
		t.Fatal(err)
	}
	correction := model.InvoiceReviewCorrection{
		InvoiceID: success.ID, RecognitionJobID: &jobs[0].ID, FieldName: "sellerName",
		RecognizedValue: "识别销售方", CorrectedValue: "测试销售方", Confidence: 0.6,
		Provider: "ocr-provider", Model: "ocr-model", CorrectedBy: 41, CorrectedAt: reviewedAt,
	}
	if err := global.GVA_DB.Create(&correction).Error; err != nil {
		t.Fatal(err)
	}
	invocations := []ai.ModelInvocation{
		{RequestID: "quality-cost-owned", UserID: 41, AuthorityID: 410, Module: "invoice", Operation: "recognize", Provider: "ocr-provider", EstimatedCostMicros: 1200, Status: ai.InvocationStatusSuccess},
		{RequestID: "quality-cost-other", UserID: 42, AuthorityID: 420, Module: "invoice", Operation: "recognize", Provider: "ocr-provider", EstimatedCostMicros: 800, Status: ai.InvocationStatusSuccess},
	}
	if err := global.GVA_DB.Create(&invocations).Error; err != nil {
		t.Fatal(err)
	}
	start, end := time.Now().Add(-time.Hour), time.Now().Add(time.Hour)
	search := invoiceRequest.QualitySearch{StartDate: &start, EndDate: &end}
	if err := search.Normalize(); err != nil {
		t.Fatal(err)
	}
	scope := AccessScope{UserID: 41, AuthorityID: 410}

	dashboard, err := (QualityService{}).Dashboard(search, scope)
	if err != nil {
		t.Fatalf("load quality dashboard: %v", err)
	}
	if dashboard.TotalRecognitions != 3 || dashboard.SuccessfulRecognitions != 2 || dashboard.FailedRecognitions != 1 ||
		dashboard.SuccessRate != 66.67 || dashboard.FailureRate != 33.33 || dashboard.AverageDurationMS != 200 ||
		dashboard.AverageAttempts != 1.33 || dashboard.MultimodalFallbackRate != 50 || dashboard.ReviewedInvoices != 1 ||
		dashboard.CorrectedFields != 1 || dashboard.LegacyWithoutFieldData != 1 || dashboard.EstimatedCostMicros != 1200 {
		t.Fatalf("unexpected dashboard: %#v", dashboard)
	}
	providers, err := (QualityService{}).ProviderMetrics(search, scope)
	if err != nil || len(providers) != 1 {
		t.Fatalf("load provider metrics: metrics=%#v err=%v", providers, err)
	}
	provider := providers[0]
	if provider.Total != 3 || provider.Success != 2 || provider.Failed != 1 || provider.SuccessRate != 66.67 ||
		provider.AverageConfidence != 0.85 || provider.AverageDurationMS != 200 || provider.AverageAttempts != 1.33 ||
		provider.CorrectedFields != 1 {
		t.Fatalf("unexpected provider metric: %#v", provider)
	}
	fields, err := (QualityService{}).FieldMetrics(search, scope)
	if err != nil {
		t.Fatalf("load field metrics: %v", err)
	}
	fieldByName := make(map[string]struct {
		reviewed, modified   int64
		accuracy, confidence float64
	}, len(fields))
	for _, field := range fields {
		fieldByName[field.FieldName] = struct {
			reviewed, modified   int64
			accuracy, confidence float64
		}{field.Reviewed, field.Modified, field.AccuracyRate, field.AverageConfidence}
	}
	if field := fieldByName["invoiceNumber"]; field.reviewed != 1 || field.modified != 0 || field.accuracy != 100 || field.confidence != 0.9 {
		t.Fatalf("unexpected invoice number metric: %#v", field)
	}
	if field := fieldByName["sellerName"]; field.reviewed != 1 || field.modified != 1 || field.accuracy != 0 || field.confidence != 0.6 {
		t.Fatalf("unexpected seller name metric: %#v", field)
	}
}
