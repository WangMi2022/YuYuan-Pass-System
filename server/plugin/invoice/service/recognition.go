package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/provider"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormLogger "gorm.io/gorm/logger"
)

type RecognitionService struct{}

var recognitionWorkerOnce sync.Once
var newRecheckRecognizer = func(configuration config.InvoiceRecognition) (provider.Recognizer, error) {
	return provider.NewMultimodal(configuration)
}

var errRecognitionLeaseLost = errors.New("发票识别任务租约已失效")

const (
	recognitionLeaseTimeout = 5 * time.Minute
	recognitionBatchWorkers = 4
	maxRecognitionItems     = 200
	maxRecognitionRawText   = 200 << 10
	maxRecognitionPayload   = 1 << 20
)

var recognitionSlots = make(chan struct{}, recognitionBatchWorkers)
var recheckRequests singleflight.Group

func acquireRecognitionSlot(ctx context.Context) (func(), error) {
	select {
	case recognitionSlots <- struct{}{}:
		return func() { <-recognitionSlots }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func recognizeWithSlot(ctx context.Context, recognizer provider.Recognizer, input provider.Input) (provider.Result, error) {
	release, err := acquireRecognitionSlot(ctx)
	if err != nil {
		return provider.Result{}, err
	}
	defer release()
	return recognizer.Recognize(ctx, input)
}

func (RecognitionService) StartWorker() {
	recognitionWorkerOnce.Do(func() {
		go runRecognitionWorker()
		go runInvoiceFileCleanupWorker()
	})
}

func runRecognitionWorker() {
	if global.GVA_DB == nil {
		return
	}
	recoverStaleRecognitionJobs()
	processRecognitionBatch()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		recoverStaleRecognitionJobs()
		processRecognitionBatch()
	}
}

func recoverStaleRecognitionJobs() {
	staleBefore := time.Now().Add(-recognitionLeaseTimeout)
	if err := global.GVA_DB.Model(&model.RecognitionJob{}).
		Where("status = ? AND (locked_at IS NULL OR locked_at < ?)", model.RecognitionJobProcessing, staleBefore).
		Updates(map[string]any{
			"status": model.RecognitionJobPending, "locked_at": nil,
			"lock_token": "", "next_run_at": time.Now(),
		}).Error; err != nil {
		global.GVA_LOG.Warn("恢复发票识别任务失败", zap.Error(err))
	}
}

func processRecognitionBatch() {
	jobs := make([]model.RecognitionJob, 0, recognitionBatchWorkers)
	for len(jobs) < recognitionBatchWorkers {
		job, err := claimRecognitionJob()
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				global.GVA_LOG.Error("领取发票识别任务失败", zap.Error(err))
			}
			break
		}
		jobs = append(jobs, job)
	}
	processClaimedRecognitionJobs(jobs, processRecognitionJob)
}

func processClaimedRecognitionJobs(jobs []model.RecognitionJob, process func(model.RecognitionJob)) {
	var workers sync.WaitGroup
	workers.Add(len(jobs))
	for _, claimed := range jobs {
		job := claimed
		go func() {
			defer workers.Done()
			process(job)
		}()
	}
	workers.Wait()
}

func claimRecognitionJob() (model.RecognitionJob, error) {
	var claimed model.RecognitionJob
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var candidate model.RecognitionJob
		now := time.Now()
		candidateQuery := tx.Session(&gorm.Session{Logger: gormLogger.Discard}).
			Where("status = ? AND attempts < max_attempts AND (next_run_at IS NULL OR next_run_at <= ?)", model.RecognitionJobPending, now).
			Order("next_run_at ASC, id ASC")
		if tx.Dialector.Name() == "postgres" {
			candidateQuery = candidateQuery.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})
		}
		if err := candidateQuery.First(&candidate).Error; err != nil {
			return err
		}
		lockToken := uuid.NewString()
		result := tx.Model(&model.RecognitionJob{}).
			Where("id = ? AND status = ?", candidate.ID, model.RecognitionJobPending).
			Updates(map[string]any{
				"status": model.RecognitionJobProcessing, "attempts": gorm.Expr("attempts + 1"),
				"locked_at": &now, "lock_token": lockToken,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.First(&claimed, candidate.ID).Error
	})
	return claimed, err
}

func processRecognitionJob(job model.RecognitionJob) {
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
	defer cancel()
	var invoice model.Invoice
	if err := global.GVA_DB.First(&invoice, job.InvoiceID).Error; err != nil {
		finishRecognitionFailure(job, err)
		return
	}
	marking := global.GVA_DB.Model(&model.Invoice{}).
		Where("id = ? AND status IN ?", invoice.ID, []string{
			model.InvoiceStatusUploaded, model.InvoiceStatusRecognitionFailed, model.InvoiceStatusRecognizing,
		}).Updates(map[string]any{"status": model.InvoiceStatusRecognizing, "recognition_error": ""})
	if marking.Error != nil {
		finishRecognitionFailure(job, marking.Error)
		return
	}
	if marking.RowsAffected == 0 {
		completeRecognitionJob(job)
		return
	}
	reader, err := openInvoiceObject(ctx, invoice)
	if err != nil {
		finishRecognitionFailure(job, fmt.Errorf("读取发票原图失败: %w", err))
		return
	}
	data, readErr := io.ReadAll(io.LimitReader(reader, maxInvoiceFileSize+1))
	_ = reader.Close()
	if readErr != nil || len(data) > maxInvoiceFileSize {
		if readErr == nil {
			readErr = errors.New("发票文件超过识别大小限制")
		}
		finishRecognitionFailure(job, readErr)
		return
	}
	result, err := recognizeWithSlot(ctx, provider.New(global.GVA_CONFIG.InvoiceRecognition), provider.Input{
		FileName: invoice.FileName, ContentType: invoice.MimeType, Data: data,
	})
	if err != nil {
		finishRecognitionFailure(job, err)
		return
	}
	if err = saveRecognitionResult(invoice, job, result); err != nil {
		finishRecognitionFailure(job, err)
	}
}

func (RecognitionService) TestProviderConnection(
	ctx context.Context,
	target string,
	configuration config.InvoiceRecognition,
) (string, error) {
	configuration = configuration.MergeSecrets(global.GVA_CONFIG.InvoiceRecognition, true)
	configuration.Normalize()
	return provider.TestConnection(ctx, target, configuration)
}

// Recheck runs the configured multimodal model once and returns a candidate
// result for the review form. It deliberately does not persist any field.
func (RecognitionService) Recheck(ctx context.Context, id uint, scope AccessScope) (provider.Result, error) {
	invoice, err := InvoiceService{}.Get(id, scope)
	if err != nil {
		return provider.Result{}, err
	}
	if invoice.Status == model.InvoiceStatusConfirmed {
		return provider.Result{}, errors.New("已确认发票不能重新核对")
	}
	configuration := global.GVA_CONFIG.InvoiceRecognition
	configuration.Normalize()
	deadline := time.Duration(configuration.Multimodal.TimeoutSeconds+5) * time.Second
	requestKey := fmt.Sprintf(
		"%d:%s:%s:%s:%s",
		invoice.ID,
		invoice.FileHash,
		configuration.Multimodal.Protocol,
		configuration.Multimodal.BaseURL,
		configuration.Multimodal.Model,
	)
	request := recheckRequests.DoChan(requestKey, func() (any, error) {
		// The shared model call must not be canceled by any single HTTP waiter.
		sharedCtx, cancel := context.WithTimeout(context.Background(), deadline)
		defer cancel()
		recognizer, createErr := newRecheckRecognizer(configuration)
		if createErr != nil {
			return provider.Result{}, createErr
		}
		release, slotErr := acquireRecognitionSlot(sharedCtx)
		if slotErr != nil {
			return provider.Result{}, slotErr
		}
		defer release()
		reader, openErr := openInvoiceObject(sharedCtx, invoice)
		if openErr != nil {
			return provider.Result{}, fmt.Errorf("读取发票原图失败: %w", openErr)
		}
		data, readErr := io.ReadAll(io.LimitReader(reader, maxInvoiceFileSize+1))
		_ = reader.Close()
		if readErr != nil {
			return provider.Result{}, readErr
		}
		if len(data) > maxInvoiceFileSize {
			return provider.Result{}, errors.New("发票文件超过识别大小限制")
		}
		return recognizer.Recognize(sharedCtx, provider.Input{
			FileName: invoice.FileName, ContentType: invoice.MimeType, Data: data,
		})
	})
	var candidate any
	select {
	case <-ctx.Done():
		return provider.Result{}, ctx.Err()
	case response := <-request:
		if response.Err != nil {
			return provider.Result{}, response.Err
		}
		candidate = response.Val
	}
	result, ok := candidate.(provider.Result)
	if !ok {
		return provider.Result{}, errors.New("模型核对结果类型不正确")
	}
	if err = normalizeRecognitionResult(&result); err != nil {
		return provider.Result{}, err
	}
	result.RawPayload = ""
	result.RawText = ""
	for index := range result.Items {
		result.Items[index].GVA_MODEL = global.GVA_MODEL{}
		result.Items[index].InvoiceID = 0
	}
	return result, nil
}

func saveRecognitionResult(invoice model.Invoice, job model.RecognitionJob, result provider.Result) error {
	if err := normalizeRecognitionResult(&result); err != nil {
		return err
	}
	invoice.InvoiceType = result.InvoiceType
	invoice.InvoiceCode = result.InvoiceCode
	invoice.InvoiceNumber = result.InvoiceNumber
	invoice.IssueDate = result.IssueDate
	invoice.BuyerName = result.BuyerName
	invoice.BuyerTaxNo = result.BuyerTaxNo
	invoice.SellerName = result.SellerName
	invoice.SellerTaxNo = result.SellerTaxNo
	invoice.AmountCents = result.AmountCents
	invoice.TaxCents = result.TaxCents
	invoice.TotalCents = result.TotalCents
	invoice.RawText = result.RawText
	invoice.Items = result.Items

	var rules []model.ClassificationRule
	if err := global.GVA_DB.Where("enabled = ?", true).Order("priority DESC, id ASC").Find(&rules).Error; err != nil {
		return err
	}
	classification := classifyInvoice(invoice, rules)
	fieldConfidences, err := json.Marshal(result.FieldConfidences)
	if err != nil {
		return fmt.Errorf("序列化字段置信度失败: %w", err)
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		lease := tx.Model(&model.RecognitionJob{}).
			Where("id = ? AND status = ? AND lock_token = ?", job.ID, model.RecognitionJobProcessing, job.LockToken).
			Update("locked_at", &now)
		if lease.Error != nil {
			return lease.Error
		}
		if lease.RowsAffected == 0 {
			return errRecognitionLeaseLost
		}
		updates := map[string]any{
			"invoice_type": result.InvoiceType, "invoice_code": result.InvoiceCode,
			"invoice_number": result.InvoiceNumber, "issue_date": result.IssueDate,
			"buyer_name": result.BuyerName, "buyer_tax_no": result.BuyerTaxNo,
			"seller_name": result.SellerName, "seller_tax_no": result.SellerTaxNo,
			"amount_cents": invoice.AmountCents, "tax_cents": result.TaxCents,
			"total_cents": result.TotalCents, "recognition_provider": result.Provider,
			"recognition_confidence": result.Confidence, "field_confidences": string(fieldConfidences),
			"recognition_error": "", "raw_text": result.RawText, "raw_payload": result.RawPayload,
			"classification_confidence": classification.Confidence,
			"classification_reason":     classification.Reason,
			"status":                    model.InvoiceStatusPendingReview,
		}
		if classification.CandidateID > 0 {
			candidateID := classification.CandidateID
			updates["suggested_category_id"] = &candidateID
		} else {
			updates["suggested_category_id"] = nil
		}
		if invoice.ClassificationSource != model.ClassificationManual {
			if classification.CategoryID > 0 {
				categoryID := classification.CategoryID
				updates["category_id"] = &categoryID
				updates["classification_source"] = model.ClassificationAuto
			} else {
				updates["category_id"] = nil
				updates["classification_source"] = ""
			}
		}
		invoiceUpdate := tx.Model(&model.Invoice{}).
			Where("id = ? AND status = ?", invoice.ID, model.InvoiceStatusRecognizing).
			Updates(updates)
		if invoiceUpdate.Error != nil {
			return invoiceUpdate.Error
		}
		if invoiceUpdate.RowsAffected == 0 {
			return completeOwnedRecognitionJob(tx, job)
		}
		if err := tx.Where("invoice_id = ?", invoice.ID).Delete(&model.InvoiceItem{}).Error; err != nil {
			return err
		}
		if len(result.Items) > 0 {
			items := make([]model.InvoiceItem, 0, len(result.Items))
			for _, item := range result.Items {
				item.GVA_MODEL = global.GVA_MODEL{}
				item.InvoiceID = invoice.ID
				items = append(items, item)
			}
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return completeOwnedRecognitionJob(tx, job)
	})
}

func normalizeRecognitionResult(result *provider.Result) error {
	result.Provider = strings.TrimSpace(result.Provider)
	result.InvoiceType = strings.TrimSpace(result.InvoiceType)
	result.InvoiceCode = strings.TrimSpace(result.InvoiceCode)
	result.InvoiceNumber = strings.TrimSpace(result.InvoiceNumber)
	result.BuyerName = strings.TrimSpace(result.BuyerName)
	result.BuyerTaxNo = strings.TrimSpace(result.BuyerTaxNo)
	result.SellerName = strings.TrimSpace(result.SellerName)
	result.SellerTaxNo = strings.TrimSpace(result.SellerTaxNo)
	if result.Provider == "" || len(result.Provider) > 50 || len(result.InvoiceType) > 60 ||
		len(result.InvoiceCode) > 80 || len(result.InvoiceNumber) > 80 ||
		len(result.BuyerName) > 200 || len(result.BuyerTaxNo) > 80 ||
		len(result.SellerName) > 200 || len(result.SellerTaxNo) > 80 {
		return errors.New("OCR 返回的发票字段长度超出限制")
	}
	if result.AmountCents == 0 && result.TotalCents >= result.TaxCents && result.TaxCents >= 0 {
		result.AmountCents = result.TotalCents - result.TaxCents
	}
	if err := validateInvoiceAmounts(result.AmountCents, result.TaxCents, result.TotalCents); err != nil {
		return err
	}
	if math.IsNaN(result.Confidence) || math.IsInf(result.Confidence, 0) || result.Confidence < 0 || result.Confidence > 1 {
		return errors.New("OCR 返回的识别置信度不正确")
	}
	if len(result.Items) > maxRecognitionItems {
		return errors.New("OCR 返回的发票明细数量超出限制")
	}
	if len(result.RawText) > maxRecognitionRawText {
		result.RawText = truncateUTF8(result.RawText, maxRecognitionRawText)
	}
	if len(result.RawPayload) > maxRecognitionPayload {
		result.RawPayload = truncateUTF8(result.RawPayload, maxRecognitionPayload)
	}
	for index := range result.Items {
		item := &result.Items[index]
		item.Name = strings.TrimSpace(item.Name)
		item.Specification = strings.TrimSpace(item.Specification)
		item.Unit = strings.TrimSpace(item.Unit)
		item.QuantityText = strings.TrimSpace(item.QuantityText)
		item.TaxRate = strings.TrimSpace(item.TaxRate)
		if item.Name == "" || len(item.Name) > 300 || len(item.Specification) > 200 ||
			len(item.Unit) > 30 || len(item.QuantityText) > 50 || len(item.TaxRate) > 30 ||
			item.UnitPriceCents < 0 || item.AmountCents < 0 || item.TaxCents < 0 {
			return errors.New("OCR 返回的发票明细不正确")
		}
	}
	if len(result.FieldConfidences) > 64 {
		return errors.New("OCR 返回的字段置信度数量超出限制")
	}
	for field, confidence := range result.FieldConfidences {
		if len(field) > 80 || math.IsNaN(confidence) || math.IsInf(confidence, 0) || confidence < 0 || confidence > 1 {
			return errors.New("OCR 返回的字段置信度不正确")
		}
	}
	return nil
}

func truncateUTF8(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	end := maxBytes
	for end > 0 && !utf8.RuneStart(value[end]) {
		end--
	}
	return value[:end]
}

func completeOwnedRecognitionJob(tx *gorm.DB, job model.RecognitionJob) error {
	result := tx.Model(&model.RecognitionJob{}).
		Where("id = ? AND status = ? AND lock_token = ?", job.ID, model.RecognitionJobProcessing, job.LockToken).
		Updates(map[string]any{
			"status": model.RecognitionJobCompleted, "locked_at": nil, "lock_token": "",
			"next_run_at": nil, "last_error": "",
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errRecognitionLeaseLost
	}
	return nil
}

func completeRecognitionJob(job model.RecognitionJob) {
	if err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		return completeOwnedRecognitionJob(tx, job)
	}); err != nil && !errors.Is(err, errRecognitionLeaseLost) {
		global.GVA_LOG.Warn("结束过期发票识别任务失败", zap.Error(err), zap.Uint("invoiceID", job.InvoiceID))
	}
}

func finishRecognitionFailure(job model.RecognitionJob, jobErr error) {
	message := jobErr.Error()
	if len(message) > 1000 {
		message = truncateUTF8(message, 1000)
	}
	jobStatus := model.RecognitionJobFailed
	invoiceStatus := model.InvoiceStatusRecognitionFailed
	var nextRunAt *time.Time
	if job.Attempts < job.MaxAttempts {
		jobStatus = model.RecognitionJobPending
		invoiceStatus = model.InvoiceStatusRecognizing
		delay := time.Duration(30*(1<<(job.Attempts-1))) * time.Second
		next := time.Now().Add(delay)
		nextRunAt = &next
	}
	if err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		jobUpdate := tx.Model(&model.RecognitionJob{}).
			Where("id = ? AND status = ? AND lock_token = ?", job.ID, model.RecognitionJobProcessing, job.LockToken).
			Updates(map[string]any{
				"status": jobStatus, "locked_at": nil, "lock_token": "",
				"next_run_at": nextRunAt, "last_error": message,
			})
		if jobUpdate.Error != nil {
			return jobUpdate.Error
		}
		if jobUpdate.RowsAffected == 0 {
			return errRecognitionLeaseLost
		}
		return tx.Model(&model.Invoice{}).
			Where("id = ? AND status = ?", job.InvoiceID, model.InvoiceStatusRecognizing).
			Updates(map[string]any{"status": invoiceStatus, "recognition_error": message}).Error
	}); err != nil && !errors.Is(err, errRecognitionLeaseLost) {
		global.GVA_LOG.Error("记录发票识别失败状态失败", zap.Error(err), zap.Uint("invoiceID", job.InvoiceID))
	}
}

func (RecognitionService) Retry(id uint, scope AccessScope) error {
	invoice, err := InvoiceService{}.Get(id, scope)
	if err != nil {
		return err
	}
	if invoice.Status == model.InvoiceStatusConfirmed {
		return errors.New("已确认发票不能重新识别")
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		job := model.RecognitionJob{InvoiceID: invoice.ID}
		updates := map[string]any{
			"status": model.RecognitionJobPending, "attempts": 0, "max_attempts": 3,
			"next_run_at": time.Now(), "locked_at": nil, "lock_token": "", "last_error": "",
		}
		result := tx.Model(&model.RecognitionJob{}).Where("invoice_id = ?", invoice.ID).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			job.Status = model.RecognitionJobPending
			job.MaxAttempts = 3
			if err := tx.Create(&job).Error; err != nil {
				return err
			}
		}
		return tx.Model(&model.Invoice{}).Where("id = ?", invoice.ID).Updates(map[string]any{
			"status": model.InvoiceStatusUploaded, "recognition_error": "",
		}).Error
	})
}
