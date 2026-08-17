package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/ai"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model"
	invoiceRequest "github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model/request"
	invoiceResponse "github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type QualityService struct{}

var qualityFieldLabels = map[string]string{
	"invoiceType": "发票类型", "invoiceCode": "发票代码", "invoiceNumber": "发票号码",
	"checkCode": "校验码", "issueDate": "开票日期", "buyerName": "购买方名称",
	"buyerTaxNo": "购买方税号", "sellerName": "销售方名称", "sellerTaxNo": "销售方税号",
	"amountCents": "不含税金额", "taxCents": "税额", "totalCents": "价税合计",
	"categoryId": "发票分类", "items": "发票明细",
}

var qualityFieldOrder = []string{
	"invoiceType", "invoiceCode", "invoiceNumber", "checkCode", "issueDate", "buyerName",
	"buyerTaxNo", "sellerName", "sellerTaxNo", "amountCents", "taxCents", "totalCents",
	"categoryId", "items",
}

func qualitySensitiveValue(field, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if field != "checkCode" && field != "buyerTaxNo" && field != "sellerTaxNo" {
		return value
	}
	digest := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func qualityDateValue(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02")
}

func qualityCategoryValue(value *uint) string {
	if value == nil || *value == 0 {
		return ""
	}
	return strconv.FormatUint(uint64(*value), 10)
}

func qualityItemsValue(items []model.InvoiceItem) string {
	type itemSnapshot struct {
		Name           string `json:"name"`
		Specification  string `json:"specification"`
		Unit           string `json:"unit"`
		QuantityText   string `json:"quantityText"`
		UnitPriceCents int64  `json:"unitPriceCents"`
		AmountCents    int64  `json:"amountCents"`
		TaxRate        string `json:"taxRate"`
		TaxCents       int64  `json:"taxCents"`
	}
	snapshots := make([]itemSnapshot, 0, len(items))
	for _, item := range items {
		snapshots = append(snapshots, itemSnapshot{
			Name: item.Name, Specification: item.Specification, Unit: item.Unit,
			QuantityText: item.QuantityText, UnitPriceCents: item.UnitPriceCents,
			AmountCents: item.AmountCents, TaxRate: item.TaxRate, TaxCents: item.TaxCents,
		})
	}
	encoded, _ := json.Marshal(snapshots)
	return string(encoded)
}

func invoiceQualityValues(invoice model.Invoice) map[string]string {
	return map[string]string{
		"invoiceType":   invoice.InvoiceType,
		"invoiceCode":   invoice.InvoiceCode,
		"invoiceNumber": invoice.InvoiceNumber,
		"checkCode":     qualitySensitiveValue("checkCode", invoice.CheckCode),
		"issueDate":     qualityDateValue(invoice.IssueDate),
		"buyerName":     invoice.BuyerName,
		"buyerTaxNo":    qualitySensitiveValue("buyerTaxNo", invoice.BuyerTaxNo),
		"sellerName":    invoice.SellerName,
		"sellerTaxNo":   qualitySensitiveValue("sellerTaxNo", invoice.SellerTaxNo),
		"amountCents":   strconv.FormatInt(invoice.AmountCents, 10),
		"taxCents":      strconv.FormatInt(invoice.TaxCents, 10),
		"totalCents":    strconv.FormatInt(invoice.TotalCents, 10),
		"categoryId":    qualityCategoryValue(invoice.CategoryID),
		"items":         qualityItemsValue(invoice.Items),
	}
}

func invoiceUpdateQualityValues(request invoiceRequest.InvoiceUpdate) map[string]string {
	return map[string]string{
		"invoiceType":   request.InvoiceType,
		"invoiceCode":   request.InvoiceCode,
		"invoiceNumber": request.InvoiceNumber,
		"checkCode":     qualitySensitiveValue("checkCode", request.CheckCode),
		"issueDate":     qualityDateValue(request.IssueDate),
		"buyerName":     request.BuyerName,
		"buyerTaxNo":    qualitySensitiveValue("buyerTaxNo", request.BuyerTaxNo),
		"sellerName":    request.SellerName,
		"sellerTaxNo":   qualitySensitiveValue("sellerTaxNo", request.SellerTaxNo),
		"amountCents":   strconv.FormatInt(request.AmountCents, 10),
		"taxCents":      strconv.FormatInt(request.TaxCents, 10),
		"totalCents":    strconv.FormatInt(request.TotalCents, 10),
		"categoryId":    qualityCategoryValue(request.CategoryID),
		"items":         qualityItemsValue(request.Items),
	}
}

func syncInvoiceReviewCorrections(tx *gorm.DB, invoice model.Invoice, request invoiceRequest.InvoiceUpdate, userID uint, correctedAt time.Time) error {
	var existing []model.InvoiceReviewCorrection
	if err := tx.Where("invoice_id = ?", invoice.ID).Find(&existing).Error; err != nil {
		return err
	}
	existingByField := make(map[string]model.InvoiceReviewCorrection, len(existing))
	for _, correction := range existing {
		existingByField[correction.FieldName] = correction
	}
	recognized := invoiceQualityValues(invoice)
	corrected := invoiceUpdateQualityValues(request)
	for field, correction := range existingByField {
		if _, tracked := recognized[field]; tracked {
			recognized[field] = correction.RecognizedValue
		}
	}
	var job model.RecognitionJob
	var jobID *uint
	if err := tx.Where("invoice_id = ?", invoice.ID).First(&job).Error; err == nil {
		jobID = &job.ID
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	for _, field := range qualityFieldOrder {
		if recognized[field] == corrected[field] {
			if correction, exists := existingByField[field]; exists {
				if err := tx.Delete(&model.InvoiceReviewCorrection{}, correction.ID).Error; err != nil {
					return err
				}
			}
			continue
		}
		correction := model.InvoiceReviewCorrection{
			InvoiceID: invoice.ID, RecognitionJobID: jobID, FieldName: field,
			RecognizedValue: recognized[field], CorrectedValue: corrected[field],
			Confidence: invoice.FieldConfidences[field], Provider: invoice.RecognitionProvider,
			Model: invoice.RecognitionModel, PromptVersion: invoice.RecognitionPromptVersion,
			CorrectedBy: userID, CorrectedAt: correctedAt, Confirmed: false,
		}
		if prior, exists := existingByField[field]; exists {
			correction.ID = prior.ID
			correction.CreatedAt = prior.CreatedAt
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "invoice_id"}, {Name: "field_name"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"recognition_job_id", "recognized_value", "corrected_value", "confidence", "provider",
				"model", "prompt_version", "corrected_by", "corrected_at", "confirmed", "updated_at",
			}),
		}).Create(&correction).Error; err != nil {
			return err
		}
	}
	return nil
}

type qualityDataset struct {
	Invoices    []model.Invoice
	Jobs        map[uint]model.RecognitionJob
	Corrections []model.InvoiceReviewCorrection
}

func loadQualityDataset(search invoiceRequest.QualitySearch, scope AccessScope) (qualityDataset, error) {
	var dataset qualityDataset
	db := applyInvoiceScope(global.GVA_DB.Model(&model.Invoice{}), scope).
		Where("created_at >= ? AND created_at <= ?", *search.StartDate, *search.EndDate)
	if search.Provider != "" {
		db = db.Where("recognition_provider = ?", search.Provider)
	}
	if search.Model != "" {
		db = db.Where("recognition_model = ?", search.Model)
	}
	if search.FileType != "" {
		db = db.Where("mime_type = ?", search.FileType)
	}
	if err := db.Order("created_at DESC, id DESC").Find(&dataset.Invoices).Error; err != nil {
		return dataset, err
	}
	dataset.Jobs = make(map[uint]model.RecognitionJob, len(dataset.Invoices))
	if len(dataset.Invoices) == 0 {
		return dataset, nil
	}
	ids := make([]uint, 0, len(dataset.Invoices))
	for _, invoice := range dataset.Invoices {
		ids = append(ids, invoice.ID)
	}
	var jobs []model.RecognitionJob
	if err := global.GVA_DB.Where("invoice_id IN ?", ids).Find(&jobs).Error; err != nil {
		return dataset, err
	}
	for _, job := range jobs {
		dataset.Jobs[job.InvoiceID] = job
	}
	if err := global.GVA_DB.Where("invoice_id IN ?", ids).Find(&dataset.Corrections).Error; err != nil {
		return dataset, err
	}
	return dataset, nil
}

func qualityRate(numerator, denominator int64) float64 {
	if denominator == 0 {
		return 0
	}
	value := float64(numerator) * 100 / float64(denominator)
	return float64(int(value*100+0.5)) / 100
}

func averageFloat(total float64, count int64) float64 {
	if count == 0 {
		return 0
	}
	value := total / float64(count)
	return float64(int(value*100+0.5)) / 100
}

func (QualityService) Dashboard(search invoiceRequest.QualitySearch, scope AccessScope) (invoiceResponse.QualityDashboard, error) {
	var response invoiceResponse.QualityDashboard
	dataset, err := loadQualityDataset(search, scope)
	if err != nil {
		return response, err
	}
	var totalAttempts, totalDuration int64
	var fallback int64
	for _, invoice := range dataset.Invoices {
		job := dataset.Jobs[invoice.ID]
		attempted := job.Attempts > 0 || invoice.RecognitionProvider != "" || invoice.Status == model.InvoiceStatusRecognitionFailed
		if !attempted {
			continue
		}
		response.TotalRecognitions++
		if invoice.Status == model.InvoiceStatusRecognitionFailed || job.Status == model.RecognitionJobFailed {
			response.FailedRecognitions++
		} else if invoice.RecognitionProvider != "" {
			response.SuccessfulRecognitions++
		}
		totalAttempts += int64(job.Attempts)
		duration := job.DurationMS
		if duration == 0 {
			duration = invoice.RecognitionDurationMS
		}
		totalDuration += duration
		if job.FallbackUsed || invoice.RecognitionProvider == "multimodal-ai" {
			fallback++
		}
		if invoice.ReviewCapturedAt != nil {
			response.ReviewedInvoices++
		} else if invoice.RecognitionProvider != "" && (invoice.Status == model.InvoiceStatusPendingReview || invoice.Status == model.InvoiceStatusConfirmed) {
			response.LegacyWithoutFieldData++
		}
	}
	response.CorrectedFields = int64(len(dataset.Corrections))
	response.SuccessRate = qualityRate(response.SuccessfulRecognitions, response.TotalRecognitions)
	response.FailureRate = qualityRate(response.FailedRecognitions, response.TotalRecognitions)
	response.MultimodalFallbackRate = qualityRate(fallback, response.SuccessfulRecognitions)
	if response.TotalRecognitions > 0 {
		response.AverageDurationMS = totalDuration / response.TotalRecognitions
		response.AverageAttempts = averageFloat(float64(totalAttempts), response.TotalRecognitions)
	}
	costDB := global.GVA_DB.Model(&ai.ModelInvocation{}).
		Where("module = ? AND created_at >= ? AND created_at <= ?", "invoice", *search.StartDate, *search.EndDate)
	if !scope.All {
		if scope.RoleWide && len(scope.AuthorityIDs) > 0 {
			costDB = costDB.Where("authority_id IN ?", scope.AuthorityIDs)
		} else {
			costDB = costDB.Where("user_id = ?", scope.UserID)
		}
	}
	if err := costDB.Select("COALESCE(SUM(estimated_cost_micros), 0)").Scan(&response.EstimatedCostMicros).Error; err != nil {
		return response, err
	}
	return response, nil
}

func (QualityService) ProviderMetrics(search invoiceRequest.QualitySearch, scope AccessScope) ([]invoiceResponse.ProviderMetric, error) {
	dataset, err := loadQualityDataset(search, scope)
	if err != nil {
		return nil, err
	}
	type accumulator struct {
		metric          invoiceResponse.ProviderMetric
		confidenceTotal float64
		durationTotal   int64
		attemptTotal    int64
	}
	groups := map[string]*accumulator{}
	correctionsByInvoice := map[uint]int64{}
	for _, correction := range dataset.Corrections {
		correctionsByInvoice[correction.InvoiceID]++
	}
	for _, invoice := range dataset.Invoices {
		job := dataset.Jobs[invoice.ID]
		if job.Attempts == 0 && invoice.RecognitionProvider == "" && invoice.Status != model.InvoiceStatusRecognitionFailed {
			continue
		}
		providerName := invoice.RecognitionProvider
		if providerName == "" {
			providerName = job.Provider
		}
		if providerName == "" {
			providerName = "unknown"
		}
		modelName := invoice.RecognitionModel
		if modelName == "" {
			modelName = job.Model
		}
		key := providerName + "\x00" + modelName + "\x00" + invoice.MimeType
		group := groups[key]
		if group == nil {
			group = &accumulator{metric: invoiceResponse.ProviderMetric{Provider: providerName, Model: modelName, FileType: invoice.MimeType}}
			groups[key] = group
		}
		group.metric.Total++
		if invoice.Status == model.InvoiceStatusRecognitionFailed || job.Status == model.RecognitionJobFailed {
			group.metric.Failed++
		} else {
			group.metric.Success++
		}
		group.confidenceTotal += invoice.RecognitionConfidence
		group.attemptTotal += int64(job.Attempts)
		if job.DurationMS > 0 {
			group.durationTotal += job.DurationMS
		} else {
			group.durationTotal += invoice.RecognitionDurationMS
		}
		group.metric.CorrectedFields += correctionsByInvoice[invoice.ID]
	}
	metrics := make([]invoiceResponse.ProviderMetric, 0, len(groups))
	for _, group := range groups {
		group.metric.SuccessRate = qualityRate(group.metric.Success, group.metric.Total)
		group.metric.AverageConfidence = averageFloat(group.confidenceTotal, group.metric.Success)
		if group.metric.Total > 0 {
			group.metric.AverageDurationMS = group.durationTotal / group.metric.Total
			group.metric.AverageAttempts = averageFloat(float64(group.attemptTotal), group.metric.Total)
		}
		metrics = append(metrics, group.metric)
	}
	sort.Slice(metrics, func(i, j int) bool {
		if metrics[i].Total == metrics[j].Total {
			return metrics[i].Provider < metrics[j].Provider
		}
		return metrics[i].Total > metrics[j].Total
	})
	return metrics, nil
}

func (QualityService) FieldMetrics(search invoiceRequest.QualitySearch, scope AccessScope) ([]invoiceResponse.FieldMetric, error) {
	dataset, err := loadQualityDataset(search, scope)
	if err != nil {
		return nil, err
	}
	modified := map[uint]map[string]struct{}{}
	for _, correction := range dataset.Corrections {
		if modified[correction.InvoiceID] == nil {
			modified[correction.InvoiceID] = map[string]struct{}{}
		}
		modified[correction.InvoiceID][correction.FieldName] = struct{}{}
	}
	type accumulator struct {
		metric          invoiceResponse.FieldMetric
		confidenceTotal float64
		confidenceCount int64
	}
	groups := make(map[string]*accumulator, len(qualityFieldOrder))
	for _, field := range qualityFieldOrder {
		groups[field] = &accumulator{metric: invoiceResponse.FieldMetric{FieldName: field, Label: qualityFieldLabels[field]}}
	}
	for _, invoice := range dataset.Invoices {
		if invoice.ReviewCapturedAt == nil {
			continue
		}
		values := invoiceQualityValues(invoice)
		for _, field := range qualityFieldOrder {
			_, hasConfidence := invoice.FieldConfidences[field]
			_, wasModified := modified[invoice.ID][field]
			if values[field] == "" && !hasConfidence && !wasModified {
				continue
			}
			group := groups[field]
			group.metric.Reviewed++
			if wasModified {
				group.metric.Modified++
			}
			if confidence, exists := invoice.FieldConfidences[field]; exists {
				group.confidenceTotal += confidence
				group.confidenceCount++
			}
		}
	}
	metrics := make([]invoiceResponse.FieldMetric, 0, len(groups))
	for _, field := range qualityFieldOrder {
		group := groups[field]
		if group.metric.Reviewed == 0 {
			continue
		}
		group.metric.ModificationRate = qualityRate(group.metric.Modified, group.metric.Reviewed)
		group.metric.AccuracyRate = qualityRate(group.metric.Reviewed-group.metric.Modified, group.metric.Reviewed)
		group.metric.AverageConfidence = averageFloat(group.confidenceTotal, group.confidenceCount)
		metrics = append(metrics, group.metric)
	}
	return metrics, nil
}

func (QualityService) Failures(search invoiceRequest.QualitySearch, scope AccessScope) ([]invoiceResponse.QualityFailure, int64, error) {
	dataset, err := loadQualityDataset(search, scope)
	if err != nil {
		return nil, 0, err
	}
	list := make([]invoiceResponse.QualityFailure, 0)
	for _, invoice := range dataset.Invoices {
		job := dataset.Jobs[invoice.ID]
		if invoice.Status != model.InvoiceStatusRecognitionFailed && job.Status != model.RecognitionJobFailed {
			continue
		}
		providerName, modelName := invoice.RecognitionProvider, invoice.RecognitionModel
		if providerName == "" {
			providerName = job.Provider
		}
		if modelName == "" {
			modelName = job.Model
		}
		message := strings.TrimSpace(job.LastError)
		if message == "" {
			message = strings.TrimSpace(invoice.RecognitionError)
		}
		list = append(list, invoiceResponse.QualityFailure{
			InvoiceID: invoice.ID, FileName: invoice.FileName, FileType: invoice.MimeType,
			Provider: providerName, Model: modelName, Attempts: job.Attempts, MaxAttempts: job.MaxAttempts,
			Error: message, CreatedAt: invoice.CreatedAt, UpdatedAt: invoice.UpdatedAt,
		})
	}
	total := int64(len(list))
	start := (search.Page - 1) * search.PageSize
	if start >= len(list) {
		return []invoiceResponse.QualityFailure{}, total, nil
	}
	end := start + search.PageSize
	if end > len(list) {
		end = len(list)
	}
	return list[start:end], total, nil
}

func (QualityService) ClassificationMetrics(search invoiceRequest.QualitySearch, scope AccessScope) (invoiceResponse.ClassificationMetric, error) {
	var metric invoiceResponse.ClassificationMetric
	dataset, err := loadQualityDataset(search, scope)
	if err != nil {
		return metric, err
	}
	for _, invoice := range dataset.Invoices {
		if invoice.SuggestedCategoryID == nil || *invoice.SuggestedCategoryID == 0 {
			continue
		}
		metric.Suggested++
		if invoice.ReviewCapturedAt == nil || invoice.CategoryID == nil {
			metric.Pending++
			continue
		}
		if *invoice.CategoryID == *invoice.SuggestedCategoryID {
			metric.Accepted++
		} else {
			metric.Overridden++
		}
	}
	decided := metric.Accepted + metric.Overridden
	metric.AcceptanceRate = qualityRate(metric.Accepted, decided)
	metric.OverrideRate = qualityRate(metric.Overridden, decided)
	return metric, nil
}
