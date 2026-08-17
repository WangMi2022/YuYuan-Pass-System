package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/WangMi2022/mit-assets-admin/server/ai"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	commonModel "github.com/WangMi2022/mit-assets-admin/server/model/common"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/asset/model"
	assetRequest "github.com/WangMi2022/mit-assets-admin/server/plugin/asset/model/request"
	"github.com/WangMi2022/mit-assets-admin/server/utils/upload"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	maxAssetRecognitionPhotos           = 6
	maxAssetRecognitionPhotoSize        = 10 << 20
	assetRecognitionPromptKey           = "asset-draft"
	assetRecognitionConfidenceThreshold = 0.7
	defaultAssetAdminAuthorityID        = 888
)

type assetRecognitionService struct{}

var assetRecognitionGateway ai.Gateway = ai.Default
var assetRecognitionStorage = func() upload.OSS { return upload.NewOss() }

type assetVisionOutput struct {
	Name                      string             `json:"name"`
	ProductName               string             `json:"productName"`
	Brand                     string             `json:"brand"`
	Manufacturer              string             `json:"manufacturer"`
	Model                     string             `json:"model"`
	SerialNumber              string             `json:"serialNumber"`
	Specifications            string             `json:"specifications"`
	ProductionDate            string             `json:"productionDate"`
	RecommendedCategoryCode   string             `json:"recommendedCategoryCode"`
	RecommendedUnit           string             `json:"recommendedUnit"`
	RecommendedWarrantyMonths int                `json:"recommendedWarrantyMonths"`
	WarrantyMonths            int                `json:"warrantyMonths"`
	RawText                   string             `json:"rawText"`
	FieldConfidences          map[string]float64 `json:"fieldConfidences"`
}

func inspectAssetRecognitionPhoto(file *multipart.FileHeader) (mimeType, extension string, err error) {
	if file == nil || file.Size <= 0 || file.Size > maxAssetRecognitionPhotoSize {
		return "", "", errors.New("资产图片大小必须在 10MB 以内")
	}
	reader, err := file.Open()
	if err != nil {
		return "", "", errors.New("读取资产图片失败")
	}
	defer reader.Close()
	buffer := make([]byte, 512)
	read, readErr := io.ReadFull(reader, buffer)
	if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) {
		return "", "", errors.New("读取资产图片失败")
	}
	mimeType = http.DetectContentType(buffer[:read])
	extensions := map[string]string{
		"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp", "image/gif": ".gif",
	}
	extension, exists := extensions[mimeType]
	if !exists {
		return "", "", errors.New("仅支持 JPG、PNG、WebP 或 GIF 资产图片")
	}
	return mimeType, extension, nil
}

func assetRecognitionStorageSnapshot() (model.AssetRecognitionJob, error) {
	storageType := strings.TrimSpace(global.GVA_CONFIG.System.OssType)
	if storageType == "" {
		storageType = "local"
	}
	job := model.AssetRecognitionJob{StorageType: storageType}
	switch storageType {
	case "local":
		if strings.TrimSpace(global.GVA_CONFIG.Local.StorePath) == "" {
			return job, errors.New("资产图片本地存储路径未配置")
		}
		root, err := filepath.Abs(filepath.Clean(global.GVA_CONFIG.Local.StorePath))
		if err != nil {
			return job, errors.New("资产图片本地存储路径不正确")
		}
		job.StorageRoot = root
	case "minio":
		job.StorageEndpoint = strings.TrimSpace(global.GVA_CONFIG.Minio.Endpoint)
		job.StorageBucket = strings.TrimSpace(global.GVA_CONFIG.Minio.BucketName)
		job.StorageUseSSL = global.GVA_CONFIG.Minio.UseSSL
		if job.StorageEndpoint == "" || job.StorageBucket == "" {
			return job, errors.New("资产图片对象存储配置不完整")
		}
	default:
		return job, errors.New("资产智能建档仅支持 local 或 minio 私有存储")
	}
	return job, nil
}

func cleanupAssetRecognitionPhotos(job model.AssetRecognitionJob) {
	for _, photo := range job.InputPhotos {
		if err := deleteAssetRecognitionPhoto(context.Background(), job, photo.Key); err != nil && global.GVA_LOG != nil {
			global.GVA_LOG.Warn("清理资产识别临时图片失败", zap.String("fileKey", photo.Key), zap.Error(err))
		}
	}
}

func (assetRecognitionService) Create(files []*multipart.FileHeader, userID, authorityID uint) (model.AssetRecognitionJob, error) {
	if userID == 0 || authorityID == 0 {
		return model.AssetRecognitionJob{}, errors.New("无法识别当前用户权限")
	}
	if len(files) == 0 || len(files) > maxAssetRecognitionPhotos {
		return model.AssetRecognitionJob{}, errors.New("请选择 1 到 6 张资产照片或铭牌图片")
	}
	job, err := assetRecognitionStorageSnapshot()
	if err != nil {
		return model.AssetRecognitionJob{}, err
	}
	storage := assetRecognitionStorage()
	for _, file := range files {
		_, extension, inspectErr := inspectAssetRecognitionPhoto(file)
		if inspectErr != nil {
			cleanupAssetRecognitionPhotos(job)
			return model.AssetRecognitionJob{}, inspectErr
		}
		originalName := filepath.Base(file.Filename)
		hash := sha256.Sum256([]byte(originalName + strconv.FormatInt(file.Size, 10) + uuid.NewString()))
		uploadHeader := *file
		uploadHeader.Filename = "asset-recognition-" + hex.EncodeToString(hash[:])[:20] + extension
		_, key, uploadErr := storage.UploadFile(&uploadHeader)
		if uploadErr != nil {
			cleanupAssetRecognitionPhotos(job)
			return model.AssetRecognitionJob{}, fmt.Errorf("资产图片保存失败: %w", uploadErr)
		}
		accessToken, tokenErr := model.CreatePhotoAccessToken(userID, key)
		if tokenErr != nil {
			cleanupAssetRecognitionPhotos(job)
			return model.AssetRecognitionJob{}, tokenErr
		}
		job.InputPhotos = append(job.InputPhotos, model.Photo{
			Name: originalName, Key: key, URL: model.BuildPhotoURL(0, key, accessToken), AccessToken: accessToken,
		})
	}
	job.Status = model.AssetRecognitionPending
	job.MaxAttempts = 3
	job.CreatedBy = userID
	job.AuthorityID = authorityID
	job.Draft = model.AssetRecognitionDraft{Quantity: 1, Unit: "件", Photos: append([]model.Photo(nil), job.InputPhotos...)}
	if err = global.GVA_DB.Create(&job).Error; err != nil {
		cleanupAssetRecognitionPhotos(job)
		return model.AssetRecognitionJob{}, err
	}
	return job, nil
}

func applyAssetRecognitionScope(db *gorm.DB, userID, authorityID uint) *gorm.DB {
	if authorityID == defaultAssetAdminAuthorityID {
		return db
	}
	return db.Where("created_by = ?", userID)
}

func (assetRecognitionService) List(search assetRequest.AssetRecognitionSearch, userID, authorityID uint) ([]model.AssetRecognitionJob, int64, error) {
	search.Normalize()
	var list []model.AssetRecognitionJob
	var total int64
	db := applyAssetRecognitionScope(global.GVA_DB.Model(&model.AssetRecognitionJob{}), userID, authorityID)
	if search.Status != "" {
		db = db.Where("status = ?", search.Status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("created_at DESC, id DESC").Scopes(search.Paginate()).Find(&list).Error
	return list, total, err
}

func (assetRecognitionService) Get(id, userID, authorityID uint) (model.AssetRecognitionJob, error) {
	var job model.AssetRecognitionJob
	if id == 0 {
		return job, errors.New("缺少识别任务 ID")
	}
	err := applyAssetRecognitionScope(global.GVA_DB.Model(&model.AssetRecognitionJob{}), userID, authorityID).First(&job, id).Error
	return job, err
}

func normalizeAssetRecognitionDraft(draft *model.AssetRecognitionDraft) error {
	draft.AssetCode = strings.ToUpper(strings.TrimSpace(draft.AssetCode))
	draft.Name = strings.TrimSpace(draft.Name)
	draft.Brand = strings.TrimSpace(draft.Brand)
	draft.Model = strings.TrimSpace(draft.Model)
	draft.SerialNumber = strings.TrimSpace(draft.SerialNumber)
	draft.Specifications = strings.TrimSpace(draft.Specifications)
	draft.Unit = strings.TrimSpace(draft.Unit)
	draft.Supplier = strings.TrimSpace(draft.Supplier)
	draft.Remarks = strings.TrimSpace(draft.Remarks)
	if draft.Quantity <= 0 {
		draft.Quantity = 1
	}
	if draft.Unit == "" {
		draft.Unit = "件"
	}
	if draft.UnitPrice < 0 || draft.CurrentValue < 0 {
		return errors.New("资产价格不能为负数")
	}
	if draft.ProductionDate != nil && draft.ProductionDate.After(time.Now().AddDate(0, 0, 1)) {
		return errors.New("生产日期不能晚于当前日期")
	}
	if draft.PurchaseDate != nil && draft.PurchaseDate.After(time.Now().AddDate(0, 0, 1)) {
		return errors.New("购置日期不能晚于当前日期")
	}
	if draft.PurchaseDate != nil && draft.WarrantyEndDate != nil && draft.WarrantyEndDate.Before(*draft.PurchaseDate) {
		return errors.New("质保到期日不能早于购置日期")
	}
	if utf8.RuneCountInString(draft.Specifications) > 1000 || utf8.RuneCountInString(draft.Remarks) > 1000 {
		return errors.New("规格参数或备注不能超过 1000 个字符")
	}
	return nil
}

func normalizedAssetSerial(value string) string {
	var builder strings.Builder
	for _, char := range strings.ToUpper(strings.TrimSpace(value)) {
		if unicode.IsLetter(char) || unicode.IsNumber(char) {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func findAssetDuplicateCandidates(db *gorm.DB, serialNumber string) ([]model.AssetDuplicateCandidate, error) {
	serialNumber = strings.TrimSpace(serialNumber)
	if serialNumber == "" {
		return []model.AssetDuplicateCandidate{}, nil
	}
	var assets []model.Asset
	if err := db.Select("id", "asset_code", "name", "category_id", "brand", "model", "serial_number").
		Preload("Category").Where("TRIM(serial_number) <> ''").Order("id DESC").Find(&assets).Error; err != nil {
		return nil, err
	}
	normalized := normalizedAssetSerial(serialNumber)
	candidates := make([]model.AssetDuplicateCandidate, 0, 10)
	for _, asset := range assets {
		matchType := ""
		if strings.EqualFold(strings.TrimSpace(asset.SerialNumber), serialNumber) {
			matchType = "exact"
		} else if normalized != "" && normalizedAssetSerial(asset.SerialNumber) == normalized {
			matchType = "normalized"
		}
		if matchType == "" {
			continue
		}
		candidates = append(candidates, model.AssetDuplicateCandidate{
			AssetID: asset.ID, AssetCode: asset.AssetCode, Name: asset.Name, CategoryName: asset.Category.Name,
			Brand: asset.Brand, Model: asset.Model, SerialNumber: asset.SerialNumber, MatchType: matchType,
		})
		if len(candidates) == 10 {
			break
		}
	}
	return candidates, nil
}

func activeAssetCategories(db *gorm.DB) ([]model.Category, error) {
	var categories []model.Category
	err := db.Where("enabled = ?", true).Order("sort ASC, id ASC").Find(&categories).Error
	return categories, err
}

func categoryByID(categories []model.Category, id uint) (model.Category, bool) {
	for _, category := range categories {
		if category.ID == id {
			return category, true
		}
	}
	return model.Category{}, false
}

func categoryByCode(categories []model.Category, code string) (model.Category, bool) {
	code = strings.TrimSpace(code)
	for _, category := range categories {
		if strings.EqualFold(category.Code, code) {
			return category, true
		}
	}
	return model.Category{}, false
}

func recognitionWarnings(
	result model.AssetRecognitionResult,
	confidences map[string]float64,
	draft model.AssetRecognitionDraft,
	categories []model.Category,
	duplicates []model.AssetDuplicateCandidate,
) []model.AssetRecognitionWarning {
	warnings := make([]model.AssetRecognitionWarning, 0)
	for field, confidence := range confidences {
		if confidence > 0 && confidence < assetRecognitionConfidenceThreshold {
			warnings = append(warnings, model.AssetRecognitionWarning{
				Code: "low_confidence", Field: field, Severity: "warning",
				Message: "该字段识别置信度较低，需要人工确认",
			})
		}
	}
	if draft.CategoryID == 0 {
		warnings = append(warnings, model.AssetRecognitionWarning{
			Code: "category_required", Field: "categoryId", Severity: "warning", Message: "未匹配到有效资产分类",
		})
	} else if _, exists := categoryByID(categories, draft.CategoryID); !exists {
		warnings = append(warnings, model.AssetRecognitionWarning{
			Code: "category_invalid", Field: "categoryId", Severity: "error", Message: "资产分类不存在或已停用",
		})
	}
	if result.ProductionDate != nil && result.ProductionDate.After(time.Now().AddDate(0, 0, 1)) {
		warnings = append(warnings, model.AssetRecognitionWarning{
			Code: "future_production_date", Field: "productionDate", Severity: "error", Message: "模型识别的生产日期晚于当前日期",
		})
	}
	if len(duplicates) > 0 {
		warnings = append(warnings, model.AssetRecognitionWarning{
			Code: "duplicate_serial", Field: "serialNumber", Severity: "error", Message: "序列号与现有资产重复，禁止直接确认建档",
		})
		candidate := duplicates[0]
		if draft.Brand != "" && candidate.Brand != "" && !strings.EqualFold(draft.Brand, candidate.Brand) {
			warnings = append(warnings, model.AssetRecognitionWarning{
				Code: "duplicate_brand_conflict", Field: "brand", Severity: "warning", Message: "重复候选资产的品牌与当前草稿不一致",
			})
		}
		if draft.Model != "" && candidate.Model != "" && !strings.EqualFold(draft.Model, candidate.Model) {
			warnings = append(warnings, model.AssetRecognitionWarning{
				Code: "duplicate_model_conflict", Field: "model", Severity: "warning", Message: "重复候选资产的型号与当前草稿不一致",
			})
		}
	}
	sort.SliceStable(warnings, func(left, right int) bool {
		if warnings[left].Severity == warnings[right].Severity {
			return warnings[left].Field < warnings[right].Field
		}
		return warnings[left].Severity == "error"
	})
	return warnings
}

func marshalRecognitionColumn(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func (assetRecognitionService) SaveDraft(input assetRequest.AssetRecognitionDraftUpdate, userID, authorityID uint) (model.AssetRecognitionJob, error) {
	if input.ID == 0 {
		return model.AssetRecognitionJob{}, errors.New("缺少识别任务 ID")
	}
	if err := normalizeAssetRecognitionDraft(&input.Draft); err != nil {
		return model.AssetRecognitionJob{}, err
	}
	var updated model.AssetRecognitionJob
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var job model.AssetRecognitionJob
		if err := applyAssetRecognitionScope(tx.Model(&model.AssetRecognitionJob{}), userID, authorityID).First(&job, input.ID).Error; err != nil {
			return err
		}
		if job.Status != model.AssetRecognitionReviewing {
			return errors.New("当前识别任务不能保存草稿")
		}
		input.Draft.Photos = append([]model.Photo(nil), job.InputPhotos...)
		categories, err := activeAssetCategories(tx)
		if err != nil {
			return err
		}
		duplicates, err := findAssetDuplicateCandidates(tx, input.Draft.SerialNumber)
		if err != nil {
			return err
		}
		warnings := recognitionWarnings(job.Result, job.FieldConfidences, input.Draft, categories, duplicates)
		draftJSON, err := marshalRecognitionColumn(input.Draft)
		if err != nil {
			return err
		}
		warningsJSON, err := marshalRecognitionColumn(warnings)
		if err != nil {
			return err
		}
		duplicatesJSON, err := marshalRecognitionColumn(duplicates)
		if err != nil {
			return err
		}
		now := time.Now()
		result := tx.Model(&model.AssetRecognitionJob{}).
			Where("id = ? AND status = ?", job.ID, model.AssetRecognitionReviewing).
			Updates(map[string]any{
				"draft": draftJSON, "warnings": warningsJSON, "duplicate_candidates": duplicatesJSON,
				"draft_updated_by": userID, "draft_updated_at": &now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("识别任务状态已变化，请刷新后重试")
		}
		return tx.First(&updated, job.ID).Error
	})
	return updated, err
}

func assetFromRecognitionDraft(draft model.AssetRecognitionDraft) model.Asset {
	return model.Asset{
		AssetCode: draft.AssetCode, Name: draft.Name, CategoryID: draft.CategoryID,
		Brand: draft.Brand, Model: draft.Model, SerialNumber: draft.SerialNumber,
		Specifications: draft.Specifications, ProductionDate: draft.ProductionDate,
		Quantity: draft.Quantity, Unit: draft.Unit, UnitPrice: draft.UnitPrice, CurrentValue: draft.CurrentValue,
		Supplier: draft.Supplier, PurchaseDate: draft.PurchaseDate, WarrantyEndDate: draft.WarrantyEndDate,
		Photos: append([]model.Photo(nil), draft.Photos...), Remarks: draft.Remarks,
	}
}

func (assetRecognitionService) Confirm(id, userID, authorityID uint) (model.Asset, error) {
	if id == 0 {
		return model.Asset{}, errors.New("缺少识别任务 ID")
	}
	var asset model.Asset
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var job model.AssetRecognitionJob
		query := applyAssetRecognitionScope(tx.Model(&model.AssetRecognitionJob{}), userID, authorityID)
		if tx.Dialector.Name() == "postgres" {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := query.First(&job, id).Error; err != nil {
			return err
		}
		if job.ConfirmedAssetID != nil || job.Status == model.AssetRecognitionCompleted {
			return errors.New("该识别任务已经创建正式资产")
		}
		if job.Status != model.AssetRecognitionReviewing {
			return errors.New("识别任务尚未进入人工确认阶段")
		}
		if err := normalizeAssetRecognitionDraft(&job.Draft); err != nil {
			return err
		}
		duplicates, err := findAssetDuplicateCandidates(tx, job.Draft.SerialNumber)
		if err != nil {
			return err
		}
		if len(duplicates) > 0 {
			return errors.New("序列号与现有资产重复，请核对后修改，不能直接确认")
		}
		asset = assetFromRecognitionDraft(job.Draft)
		if err = Asset.createWithDB(tx, &asset); err != nil {
			return err
		}
		now := time.Now()
		result := tx.Model(&model.AssetRecognitionJob{}).
			Where("id = ? AND status = ? AND confirmed_asset_id IS NULL", job.ID, model.AssetRecognitionReviewing).
			Updates(map[string]any{
				"status": model.AssetRecognitionCompleted, "confirmed_asset_id": &asset.ID,
				"confirmed_by": userID, "confirmed_at": &now, "completed_at": &now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("识别任务已被其他操作确认，请刷新后重试")
		}
		return nil
	})
	return asset, err
}

func (assetRecognitionService) Retry(id, userID, authorityID uint) error {
	if id == 0 {
		return errors.New("缺少识别任务 ID")
	}
	now := time.Now()
	result := applyAssetRecognitionScope(global.GVA_DB.Model(&model.AssetRecognitionJob{}), userID, authorityID).
		Where("id = ? AND status = ? AND confirmed_asset_id IS NULL", id, model.AssetRecognitionFailed).
		Updates(map[string]any{
			"status": model.AssetRecognitionPending, "attempts": 0, "next_run_at": &now,
			"locked_at": nil, "lock_token": "", "last_error": "", "completed_at": nil,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("仅失败且未确认的识别任务可以重试")
	}
	return nil
}

func (assetRecognitionService) Delete(id, userID, authorityID uint) error {
	job, err := (assetRecognitionService{}).Get(id, userID, authorityID)
	if err != nil {
		return err
	}
	if job.Status == model.AssetRecognitionCompleted || job.ConfirmedAssetID != nil {
		return errors.New("已创建正式资产的识别任务不能删除")
	}
	deleteToken := uuid.NewString()
	claim := applyAssetRecognitionScope(global.GVA_DB.Model(&model.AssetRecognitionJob{}), userID, authorityID).
		Where("id = ? AND confirmed_asset_id IS NULL AND (status IN ? OR (status = ? AND lock_token = ''))", id, []string{
			model.AssetRecognitionPending, model.AssetRecognitionReviewing, model.AssetRecognitionFailed,
		}, model.AssetRecognitionDeleting).Updates(map[string]any{
		"status": model.AssetRecognitionDeleting, "lock_token": deleteToken, "locked_at": time.Now(),
	})
	if claim.Error != nil {
		return claim.Error
	}
	if claim.RowsAffected == 0 {
		return errors.New("识别任务正在处理或状态已变化，请刷新后重试")
	}
	for _, photo := range job.InputPhotos {
		if err = deleteAssetRecognitionPhoto(context.Background(), job, photo.Key); err != nil {
			global.GVA_DB.Model(&model.AssetRecognitionJob{}).
				Where("id = ? AND lock_token = ?", id, deleteToken).
				Updates(map[string]any{
					"status": model.AssetRecognitionDeleting, "lock_token": "", "locked_at": nil,
					"last_error": "删除临时图片失败: " + truncateAssetRecognitionError(err.Error(), 900),
				})
			return fmt.Errorf("删除临时图片失败: %w", err)
		}
	}
	result := applyAssetRecognitionScope(global.GVA_DB.Model(&model.AssetRecognitionJob{}), userID, authorityID).
		Where("id = ? AND status = ? AND lock_token = ? AND confirmed_asset_id IS NULL", id, model.AssetRecognitionDeleting, deleteToken).
		Delete(&model.AssetRecognitionJob{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("识别任务状态已变化，请刷新后重试")
	}
	return nil
}

func assetRecognitionPromptPayload(categories []model.Category, imageIndex, imageCount int) commonModel.JSONMap {
	categoryOptions := make([]map[string]string, 0, len(categories))
	for _, category := range categories {
		categoryOptions = append(categoryOptions, map[string]string{
			"code": category.Code, "name": category.Name, "description": category.Description,
		})
	}
	return commonModel.JSONMap{
		"categories": categoryOptions, "imageIndex": imageIndex, "imageCount": imageCount,
	}
}
