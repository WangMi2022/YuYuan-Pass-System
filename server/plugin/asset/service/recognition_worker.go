package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/WangMi2022/mit-assets-admin/server/ai"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/asset/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormLogger "gorm.io/gorm/logger"
)

const (
	assetRecognitionLeaseTimeout = 5 * time.Minute
	assetRecognitionJobTimeout   = 4*time.Minute + 30*time.Second
	assetRecognitionWorkers      = 2
	maxAssetRecognitionRawText   = 10000
)

var assetRecognitionWorkerOnce sync.Once
var errAssetRecognitionLeaseLost = errors.New("资产识别任务租约已失效")

func (assetRecognitionService) StartWorker() {
	assetRecognitionWorkerOnce.Do(func() { go runAssetRecognitionWorker() })
}

func runAssetRecognitionWorker() {
	if global.GVA_DB == nil {
		return
	}
	recoverStaleAssetRecognitionJobs()
	processAssetRecognitionBatch()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		recoverStaleAssetRecognitionJobs()
		processAssetRecognitionBatch()
	}
}

func recoverStaleAssetRecognitionJobs() {
	staleBefore := time.Now().Add(-assetRecognitionLeaseTimeout)
	if err := global.GVA_DB.Model(&model.AssetRecognitionJob{}).
		Where("status = ? AND (locked_at IS NULL OR locked_at < ?)", model.AssetRecognitionProcessing, staleBefore).
		Updates(map[string]any{
			"status": model.AssetRecognitionPending, "locked_at": nil,
			"lock_token": "", "next_run_at": time.Now(),
		}).Error; err != nil && global.GVA_LOG != nil {
		global.GVA_LOG.Warn("恢复资产识别任务失败", zap.Error(err))
	}
	if err := global.GVA_DB.Model(&model.AssetRecognitionJob{}).
		Where("status = ? AND lock_token <> '' AND (locked_at IS NULL OR locked_at < ?)", model.AssetRecognitionDeleting, staleBefore).
		Updates(map[string]any{"locked_at": nil, "lock_token": ""}).Error; err != nil && global.GVA_LOG != nil {
		global.GVA_LOG.Warn("恢复资产识别图片清理任务失败", zap.Error(err))
	}
}

func processAssetRecognitionBatch() {
	jobs := make([]model.AssetRecognitionJob, 0, assetRecognitionWorkers)
	for len(jobs) < assetRecognitionWorkers {
		job, err := claimAssetRecognitionJob()
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) && global.GVA_LOG != nil {
				global.GVA_LOG.Error("领取资产识别任务失败", zap.Error(err))
			}
			break
		}
		jobs = append(jobs, job)
	}
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(jobs))
	for _, claimed := range jobs {
		job := claimed
		go func() {
			defer waitGroup.Done()
			processAssetRecognitionJob(job)
		}()
	}
	waitGroup.Wait()
}

func claimAssetRecognitionJob() (model.AssetRecognitionJob, error) {
	var claimed model.AssetRecognitionJob
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var candidate model.AssetRecognitionJob
		now := time.Now()
		query := tx.Session(&gorm.Session{Logger: gormLogger.Discard}).
			Where("status = ? AND attempts < max_attempts AND (next_run_at IS NULL OR next_run_at <= ?)", model.AssetRecognitionPending, now).
			Order("next_run_at ASC, id ASC")
		if tx.Dialector.Name() == "postgres" {
			query = query.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})
		}
		if err := query.First(&candidate).Error; err != nil {
			return err
		}
		lockToken := uuid.NewString()
		result := tx.Model(&model.AssetRecognitionJob{}).
			Where("id = ? AND status = ?", candidate.ID, model.AssetRecognitionPending).
			Updates(map[string]any{
				"status": model.AssetRecognitionProcessing, "attempts": gorm.Expr("attempts + 1"),
				"locked_at": &now, "lock_token": lockToken, "last_error": "",
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

func parseAssetVisionOutput(content string) (assetVisionOutput, error) {
	var output assetVisionOutput
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &output); err != nil {
		return output, fmt.Errorf("资产识别结果不是合法 JSON: %w", err)
	}
	output.Name = strings.TrimSpace(output.Name)
	output.ProductName = strings.TrimSpace(output.ProductName)
	output.Brand = strings.TrimSpace(output.Brand)
	output.Manufacturer = strings.TrimSpace(output.Manufacturer)
	output.Model = strings.TrimSpace(output.Model)
	output.SerialNumber = strings.TrimSpace(output.SerialNumber)
	output.Specifications = strings.TrimSpace(output.Specifications)
	output.ProductionDate = strings.TrimSpace(output.ProductionDate)
	output.RecommendedCategoryCode = strings.TrimSpace(output.RecommendedCategoryCode)
	output.RecommendedUnit = strings.TrimSpace(output.RecommendedUnit)
	output.RawText = strings.TrimSpace(output.RawText)
	useProductName := output.Name == "" && output.ProductName != ""
	useManufacturer := output.Brand == "" && output.Manufacturer != ""
	useWarrantyMonths := output.RecommendedWarrantyMonths == 0 && output.WarrantyMonths > 0
	if useProductName {
		output.Name = output.ProductName
	}
	if useManufacturer {
		output.Brand = output.Manufacturer
	}
	if useWarrantyMonths {
		output.RecommendedWarrantyMonths = output.WarrantyMonths
	}
	output.ProductName = ""
	output.Manufacturer = ""
	output.WarrantyMonths = 0
	normalizeAssetRecognitionConfidenceAlias(output.FieldConfidences, "name", "productName", useProductName)
	normalizeAssetRecognitionConfidenceAlias(output.FieldConfidences, "brand", "manufacturer", useManufacturer)
	normalizeAssetRecognitionConfidenceAlias(output.FieldConfidences, "recommendedWarrantyMonths", "warrantyMonths", useWarrantyMonths)
	if utf8.RuneCountInString(output.Name) > 150 || utf8.RuneCountInString(output.Brand) > 100 ||
		utf8.RuneCountInString(output.Model) > 120 || utf8.RuneCountInString(output.SerialNumber) > 120 ||
		utf8.RuneCountInString(output.Specifications) > 1000 || utf8.RuneCountInString(output.RawText) > maxAssetRecognitionRawText {
		return output, errors.New("资产识别结果字段长度超出限制")
	}
	if output.RecommendedWarrantyMonths < 0 || output.RecommendedWarrantyMonths > 120 {
		return output, errors.New("模型返回的建议质保月数不正确")
	}
	if len(output.FieldConfidences) > 32 {
		return output, errors.New("模型返回的字段置信度数量超出限制")
	}
	for field, confidence := range output.FieldConfidences {
		if len(field) > 80 || math.IsNaN(confidence) || math.IsInf(confidence, 0) || confidence < 0 || confidence > 1 {
			return output, errors.New("模型返回的字段置信度不正确")
		}
	}
	return output, nil
}

func normalizeAssetRecognitionConfidenceAlias(confidences map[string]float64, field, alias string, aliasUsed bool) {
	if confidences == nil {
		return
	}
	if confidence, aliasExists := confidences[alias]; aliasExists {
		if _, fieldExists := confidences[field]; aliasUsed || !fieldExists {
			confidences[field] = confidence
		}
	}
	delete(confidences, alias)
}

func parseAssetProductionDate(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	for _, layout := range []string{"2006-01-02", "2006/01/02", time.RFC3339} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return &parsed, nil
		}
	}
	return nil, errors.New("模型返回的生产日期格式不正确")
}

func chooseAssetRecognitionString(target *string, field, candidate string, aggregate, candidateConfidences map[string]float64) {
	if candidate == "" {
		return
	}
	confidence := candidateConfidences[field]
	if *target == "" || confidence >= aggregate[field] {
		*target = candidate
		aggregate[field] = confidence
	}
}

func mergeAssetVisionOutput(
	target *model.AssetRecognitionResult,
	aggregate map[string]float64,
	output assetVisionOutput,
	imageIndex int,
) error {
	chooseAssetRecognitionString(&target.Name, "name", output.Name, aggregate, output.FieldConfidences)
	chooseAssetRecognitionString(&target.Brand, "brand", output.Brand, aggregate, output.FieldConfidences)
	chooseAssetRecognitionString(&target.Model, "model", output.Model, aggregate, output.FieldConfidences)
	chooseAssetRecognitionString(&target.SerialNumber, "serialNumber", output.SerialNumber, aggregate, output.FieldConfidences)
	chooseAssetRecognitionString(&target.Specifications, "specifications", output.Specifications, aggregate, output.FieldConfidences)
	chooseAssetRecognitionString(&target.RecommendedCategoryCode, "recommendedCategoryCode", output.RecommendedCategoryCode, aggregate, output.FieldConfidences)
	chooseAssetRecognitionString(&target.RecommendedUnit, "recommendedUnit", output.RecommendedUnit, aggregate, output.FieldConfidences)
	productionDate, err := parseAssetProductionDate(output.ProductionDate)
	if err != nil {
		return err
	}
	if productionDate != nil && (target.ProductionDate == nil || output.FieldConfidences["productionDate"] >= aggregate["productionDate"]) {
		target.ProductionDate = productionDate
		aggregate["productionDate"] = output.FieldConfidences["productionDate"]
	}
	if output.RecommendedWarrantyMonths > 0 && (target.RecommendedWarrantyMonths == 0 || output.FieldConfidences["recommendedWarrantyMonths"] >= aggregate["recommendedWarrantyMonths"]) {
		target.RecommendedWarrantyMonths = output.RecommendedWarrantyMonths
		aggregate["recommendedWarrantyMonths"] = output.FieldConfidences["recommendedWarrantyMonths"]
	}
	if output.RawText != "" {
		if target.RawText != "" {
			target.RawText += "\n\n"
		}
		target.RawText += fmt.Sprintf("[图片 %d]\n%s", imageIndex, output.RawText)
	}
	return nil
}

func buildAssetRecognitionDraft(job model.AssetRecognitionJob, result model.AssetRecognitionResult) model.AssetRecognitionDraft {
	unit := strings.TrimSpace(result.RecommendedUnit)
	if unit == "" {
		unit = "件"
	}
	return model.AssetRecognitionDraft{
		Name: result.Name, CategoryID: result.RecommendedCategoryID, Brand: result.Brand,
		Model: result.Model, SerialNumber: result.SerialNumber, Specifications: result.Specifications,
		ProductionDate: result.ProductionDate, Quantity: 1, Unit: unit,
		RecommendedWarrantyMonths: result.RecommendedWarrantyMonths,
		Photos:                    append([]model.Photo(nil), job.InputPhotos...),
	}
}

func latestAssetRecognitionInvocation(jobID uint) ai.ModelInvocation {
	var invocation ai.ModelInvocation
	_ = global.GVA_DB.Where("object_type = ? AND object_id = ?", "asset_recognition_job", strconv.FormatUint(uint64(jobID), 10)).
		Order("id DESC").First(&invocation).Error
	return invocation
}

func processAssetRecognitionJob(job model.AssetRecognitionJob) {
	ctx, cancel := context.WithTimeout(context.Background(), assetRecognitionJobTimeout)
	defer cancel()
	categories, err := activeAssetCategories(global.GVA_DB)
	if err != nil {
		finishAssetRecognitionFailure(job, err)
		return
	}
	actorContext := ai.WithActorPermission(ctx, job.CreatedBy, job.AuthorityID, "/assetRecognition/create", http.MethodPost)
	aggregate := model.AssetRecognitionResult{}
	fieldConfidences := map[string]float64{}
	warnings := make([]model.AssetRecognitionWarning, 0)
	providerName, modelName := "", ""
	var firstError error
	successes := 0
	for index, photo := range job.InputPhotos {
		reader, openErr := openAssetRecognitionPhoto(actorContext, job, photo.Key)
		if openErr != nil {
			if firstError == nil {
				firstError = openErr
			}
			warnings = append(warnings, model.AssetRecognitionWarning{Code: "image_read_failed", Field: "photos", Severity: "warning", Message: fmt.Sprintf("第 %d 张图片读取失败", index+1)})
			continue
		}
		data, readErr := io.ReadAll(io.LimitReader(reader, maxAssetRecognitionPhotoSize+1))
		_ = reader.Close()
		if readErr != nil || len(data) > maxAssetRecognitionPhotoSize {
			if readErr == nil {
				readErr = errors.New("资产图片超过识别大小限制")
			}
			if firstError == nil {
				firstError = readErr
			}
			warnings = append(warnings, model.AssetRecognitionWarning{Code: "image_read_failed", Field: "photos", Severity: "warning", Message: fmt.Sprintf("第 %d 张图片读取失败", index+1)})
			continue
		}
		mimeType := http.DetectContentType(data)
		promptPayload := assetRecognitionPromptPayload(categories, index+1, len(job.InputPhotos))
		visionResult, visionErr := assetRecognitionGateway.Vision(actorContext, ai.VisionRequest{
			CompletionRequest: ai.CompletionRequest{
				Module: "asset", Operation: "recognize-draft", PromptKey: assetRecognitionPromptKey,
				ObjectType: "asset_recognition_job", ObjectID: strconv.FormatUint(uint64(job.ID), 10),
				Prompt:  "本次识别上下文：" + ai.DecodePayload(promptPayload),
				Payload: promptPayload, MaxOutputTokens: 2400,
			},
			Image: data, MIMEType: mimeType,
		})
		if visionErr != nil {
			if firstError == nil {
				firstError = visionErr
			}
			warnings = append(warnings, model.AssetRecognitionWarning{Code: "image_recognition_failed", Field: "photos", Severity: "warning", Message: fmt.Sprintf("第 %d 张图片识别失败", index+1)})
			continue
		}
		output, parseErr := parseAssetVisionOutput(visionResult.Content)
		if parseErr != nil {
			if firstError == nil {
				firstError = parseErr
			}
			warnings = append(warnings, model.AssetRecognitionWarning{Code: "image_result_invalid", Field: "photos", Severity: "warning", Message: fmt.Sprintf("第 %d 张图片返回结果无效", index+1)})
			continue
		}
		if mergeErr := mergeAssetVisionOutput(&aggregate, fieldConfidences, output, index+1); mergeErr != nil {
			if firstError == nil {
				firstError = mergeErr
			}
			warnings = append(warnings, model.AssetRecognitionWarning{Code: "image_result_invalid", Field: "photos", Severity: "warning", Message: fmt.Sprintf("第 %d 张图片返回结果无效", index+1)})
			continue
		}
		successes++
		providerName, modelName = visionResult.Provider, visionResult.Model
	}
	if successes == 0 {
		if firstError == nil {
			firstError = errors.New("没有资产图片完成识别")
		}
		finishAssetRecognitionFailure(job, firstError)
		return
	}
	if category, exists := categoryByCode(categories, aggregate.RecommendedCategoryCode); exists {
		aggregate.RecommendedCategoryID = category.ID
	}
	draft := buildAssetRecognitionDraft(job, aggregate)
	duplicates, err := findAssetDuplicateCandidates(global.GVA_DB, draft.SerialNumber)
	if err != nil {
		finishAssetRecognitionFailure(job, err)
		return
	}
	warnings = append(warnings, recognitionWarnings(aggregate, fieldConfidences, draft, categories, duplicates)...)
	invocation := latestAssetRecognitionInvocation(job.ID)
	resultJSON, err := marshalRecognitionColumn(aggregate)
	if err != nil {
		finishAssetRecognitionFailure(job, err)
		return
	}
	draftJSON, err := marshalRecognitionColumn(draft)
	if err != nil {
		finishAssetRecognitionFailure(job, err)
		return
	}
	confidenceJSON, err := marshalRecognitionColumn(fieldConfidences)
	if err != nil {
		finishAssetRecognitionFailure(job, err)
		return
	}
	warningsJSON, err := marshalRecognitionColumn(warnings)
	if err != nil {
		finishAssetRecognitionFailure(job, err)
		return
	}
	duplicatesJSON, err := marshalRecognitionColumn(duplicates)
	if err != nil {
		finishAssetRecognitionFailure(job, err)
		return
	}
	now := time.Now()
	update := global.GVA_DB.Model(&model.AssetRecognitionJob{}).
		Where("id = ? AND status = ? AND lock_token = ?", job.ID, model.AssetRecognitionProcessing, job.LockToken).
		Updates(map[string]any{
			"status": model.AssetRecognitionReviewing, "provider": providerName, "model": modelName,
			"prompt_version": invocation.PromptVersion, "result": resultJSON, "draft": draftJSON,
			"field_confidences": confidenceJSON, "warnings": warningsJSON, "duplicate_candidates": duplicatesJSON,
			"locked_at": nil, "lock_token": "", "next_run_at": nil, "last_error": "", "completed_at": &now,
		})
	if update.Error != nil {
		finishAssetRecognitionFailure(job, update.Error)
		return
	}
	if update.RowsAffected == 0 && global.GVA_LOG != nil {
		global.GVA_LOG.Warn("资产识别任务结果因租约失效未写入", zap.Uint("jobID", job.ID))
	}
}

func truncateAssetRecognitionError(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	end := maxBytes
	for end > 0 && !utf8.RuneStart(value[end]) {
		end--
	}
	return value[:end]
}

func finishAssetRecognitionFailure(job model.AssetRecognitionJob, jobErr error) {
	message := truncateAssetRecognitionError(jobErr.Error(), 1000)
	status := model.AssetRecognitionFailed
	var nextRunAt *time.Time
	errorType := ai.ErrorKind(jobErr)
	retryable := errorType != ai.ErrorTypeDisabled && errorType != ai.ErrorTypePolicy &&
		errorType != ai.ErrorTypeValidation && errorType != ai.ErrorTypeSchema
	if retryable && job.Attempts < job.MaxAttempts {
		status = model.AssetRecognitionPending
		delay := time.Duration(30*(1<<(job.Attempts-1))) * time.Second
		next := time.Now().Add(delay)
		nextRunAt = &next
	}
	var completedAt *time.Time
	if status == model.AssetRecognitionFailed {
		now := time.Now()
		completedAt = &now
	}
	result := global.GVA_DB.Model(&model.AssetRecognitionJob{}).
		Where("id = ? AND status = ? AND lock_token = ?", job.ID, model.AssetRecognitionProcessing, job.LockToken).
		Updates(map[string]any{
			"status": status, "locked_at": nil, "lock_token": "", "next_run_at": nextRunAt,
			"last_error": message, "completed_at": completedAt,
		})
	if result.Error != nil && global.GVA_LOG != nil {
		global.GVA_LOG.Error("记录资产识别失败状态失败", zap.Uint("jobID", job.ID), zap.Error(result.Error))
	}
}
