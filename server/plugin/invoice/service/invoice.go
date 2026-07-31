package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model"
	invoiceRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model/request"
	invoiceResponse "github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model/response"
	systemService "github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/upload"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	maxInvoiceFileSize                = 10 << 20
	maxVerificationBypassReasonLength = 500
	defaultAdminRoleID                = 888
)

type AccessScope struct {
	UserID       uint
	AuthorityID  uint
	AuthorityIDs []uint
	All          bool
	RoleWide     bool
}

type InvoiceService struct{}

var beforeInvoiceConfirmPersist = func(*gorm.DB, model.Invoice) error { return nil }

func hashInvoiceKey(parts ...string) string {
	hasher := sha256.New()
	for _, part := range parts {
		_, _ = hasher.Write([]byte(strings.ToLower(strings.TrimSpace(part))))
		_, _ = hasher.Write([]byte{0})
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func invoiceDuplicateKey(invoice model.Invoice) string {
	discriminatorType, discriminator := "seller-name", invoice.SellerName
	if strings.TrimSpace(invoice.InvoiceCode) != "" {
		discriminatorType, discriminator = "invoice-code", invoice.InvoiceCode
	} else if strings.TrimSpace(invoice.SellerTaxNo) != "" {
		discriminatorType, discriminator = "seller-tax-no", invoice.SellerTaxNo
	}
	return hashInvoiceKey("confirmed-invoice", invoice.InvoiceNumber, discriminatorType, discriminator)
}

func invoiceFileDedupKey(userID uint, hash string) string {
	return hashInvoiceKey("invoice-file", fmt.Sprint(userID), hash)
}

func isUniqueConstraintError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint") ||
		strings.Contains(message, "duplicate entry") ||
		strings.Contains(message, "duplicated key") ||
		strings.Contains(message, "sqlstate 23505")
}

func ResolveAccessScope(userID, authorityID uint) AccessScope {
	scope := AccessScope{UserID: userID, AuthorityID: authorityID, All: authorityID == defaultAdminRoleID}
	if scope.All || authorityID == 0 {
		return scope
	}
	authority, err := systemService.AuthorityServiceApp.GetAuthorityInfo(system.SysAuthority{AuthorityId: authorityID})
	if err != nil {
		return scope
	}
	seen := map[uint]struct{}{}
	for _, item := range authority.DataAuthorityId {
		if item == nil || item.AuthorityId == 0 {
			continue
		}
		seen[item.AuthorityId] = struct{}{}
		if item.AuthorityId != authorityID {
			scope.RoleWide = true
		}
	}
	if scope.RoleWide {
		seen[authorityID] = struct{}{}
		for id := range seen {
			scope.AuthorityIDs = append(scope.AuthorityIDs, id)
		}
	}
	return scope
}

func applyInvoiceScope(db *gorm.DB, scope AccessScope) *gorm.DB {
	switch {
	case scope.All:
		return db
	case scope.RoleWide && len(scope.AuthorityIDs) > 0:
		return db.Where("authority_id IN ?", scope.AuthorityIDs)
	default:
		return db.Where("created_by = ?", scope.UserID)
	}
}

func inspectInvoiceFile(file *multipart.FileHeader) (hash, mimeType, ext string, err error) {
	if file == nil {
		return "", "", "", errors.New("请选择发票文件")
	}
	if file.Size <= 0 || file.Size > maxInvoiceFileSize {
		return "", "", "", errors.New("发票文件大小必须在 10MB 以内")
	}
	reader, err := file.Open()
	if err != nil {
		return "", "", "", errors.New("读取发票文件失败")
	}
	defer reader.Close()

	buffer := make([]byte, 512)
	n, readErr := io.ReadFull(reader, buffer)
	if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) {
		return "", "", "", errors.New("读取发票文件失败")
	}
	mimeType = http.DetectContentType(buffer[:n])
	allowed := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", invoicePDFContentType: ".pdf"}
	ext, ok := allowed[mimeType]
	if !ok {
		return "", "", "", errors.New("仅支持 JPG、PNG 或 PDF 发票文件")
	}
	if _, err = reader.Seek(0, io.SeekStart); err != nil {
		return "", "", "", errors.New("读取发票文件失败")
	}
	hasher := sha256.New()
	if _, err = io.Copy(hasher, reader); err != nil {
		return "", "", "", errors.New("计算发票文件摘要失败")
	}
	return hex.EncodeToString(hasher.Sum(nil)), mimeType, ext, nil
}

func (InvoiceService) Upload(file *multipart.FileHeader, userID, authorityID uint) (model.Invoice, error) {
	if userID == 0 || authorityID == 0 {
		return model.Invoice{}, errors.New("无法识别当前用户权限")
	}
	hash, mimeType, ext, err := inspectInvoiceFile(file)
	if err != nil {
		return model.Invoice{}, err
	}
	dedupKey := invoiceFileDedupKey(userID, hash)
	var existing int64
	if err = global.GVA_DB.Model(&model.Invoice{}).Where("file_dedup_key = ?", dedupKey).Count(&existing).Error; err != nil {
		return model.Invoice{}, err
	}
	if existing > 0 {
		return model.Invoice{}, errors.New("你已上传过该发票文件，请勿重复提交")
	}
	storageType := strings.TrimSpace(global.GVA_CONFIG.System.OssType)
	if storageType == "" {
		storageType = "local"
	}
	if storageType != "local" && storageType != "minio" {
		return model.Invoice{}, errors.New("发票私有存储仅支持 local 或 minio")
	}
	storageRoot, storageEndpoint, storageBucket := "", "", ""
	storageUseSSL := false
	if storageType == "local" {
		storageRoot, err = filepath.Abs(filepath.Clean(global.GVA_CONFIG.Local.StorePath))
		if err != nil {
			return model.Invoice{}, errors.New("发票本地存储路径不正确")
		}
	} else {
		storageEndpoint = strings.TrimSpace(global.GVA_CONFIG.Minio.Endpoint)
		storageBucket = strings.TrimSpace(global.GVA_CONFIG.Minio.BucketName)
		storageUseSSL = global.GVA_CONFIG.Minio.UseSSL
	}

	uploadHeader := *file
	uploadHeader.Filename = "invoice-" + hash[:20] + "-" + uuid.NewString()[:8] + ext
	_, key, err := upload.NewOss().UploadFile(&uploadHeader)
	if err != nil {
		return model.Invoice{}, fmt.Errorf("发票文件保存失败: %w", err)
	}

	invoice := model.Invoice{
		Direction: "expense", Currency: "CNY", Status: model.InvoiceStatusUploaded,
		FileName: filepath.Base(file.Filename), FileKey: key, FileHash: hash,
		FileDedupKey: &dedupKey,
		MimeType:     mimeType, FileSize: file.Size, StorageType: storageType,
		StorageRoot: storageRoot, StorageEndpoint: storageEndpoint,
		StorageBucket: storageBucket, StorageUseSSL: storageUseSSL,
		CreatedBy: userID, AuthorityID: authorityID,
	}
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if createErr := tx.Create(&invoice).Error; createErr != nil {
			return createErr
		}
		job := model.RecognitionJob{InvoiceID: invoice.ID, Status: model.RecognitionJobPending, MaxAttempts: 3}
		return tx.Create(&job).Error
	})
	if err != nil {
		if cleanupErr := deleteInvoiceObject(context.Background(), invoice); cleanupErr != nil && global.GVA_LOG != nil {
			global.GVA_LOG.Error("发票记录创建失败且对象清理失败", zap.Error(cleanupErr), zap.String("fileKey", key))
		}
		if isUniqueConstraintError(err) {
			return model.Invoice{}, errors.New("你已上传过该发票文件，请勿重复提交")
		}
		return model.Invoice{}, err
	}
	return invoice, nil
}

func (InvoiceService) List(search invoiceRequest.InvoiceSearch, scope AccessScope) ([]model.Invoice, int64, error) {
	var list []model.Invoice
	var total int64
	search.Normalize()
	db := applyInvoiceScope(global.GVA_DB.Model(&model.Invoice{}), scope)
	if keyword := strings.TrimSpace(search.Keyword); keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		db = db.Where(`LOWER(invoice_number) LIKE ? OR LOWER(invoice_code) LIKE ? OR
			LOWER(seller_name) LIKE ? OR LOWER(buyer_name) LIKE ?`, like, like, like, like)
	}
	if search.Status != "" {
		db = db.Where("status = ?", search.Status)
	}
	if search.CategoryID > 0 {
		db = db.Where("category_id = ?", search.CategoryID)
	}
	if search.Direction != "" {
		db = db.Where("direction = ?", search.Direction)
	}
	if search.StartDate != nil {
		db = db.Where("issue_date >= ?", *search.StartDate)
	}
	if search.EndDate != nil {
		db = db.Where("issue_date < ?", search.EndDate.AddDate(0, 0, 1))
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Preload("Category").Preload("SuggestedCategory").
		Order("created_at DESC, id DESC").Scopes(search.Paginate()).Find(&list).Error
	return list, total, err
}

func (InvoiceService) Get(id uint, scope AccessScope) (model.Invoice, error) {
	var invoice model.Invoice
	if id == 0 {
		return invoice, errors.New("缺少发票 ID")
	}
	err := applyInvoiceScope(global.GVA_DB.Model(&model.Invoice{}), scope).
		Preload("Category").Preload("SuggestedCategory").Preload("Items").
		First(&invoice, id).Error
	return invoice, err
}

func validateInvoiceAmounts(amountCents, taxCents, totalCents int64) error {
	if amountCents < 0 || taxCents < 0 || totalCents < 0 {
		return errors.New("发票金额不能小于 0")
	}
	if amountCents != 0 || taxCents != 0 || totalCents != 0 {
		if totalCents != amountCents+taxCents {
			return errors.New("价税合计必须等于金额与税额之和")
		}
	}
	return nil
}

func normalizeInvoiceUpdate(req *invoiceRequest.InvoiceUpdate) error {
	req.Direction = strings.TrimSpace(req.Direction)
	if req.Direction != "expense" && req.Direction != "income" {
		return errors.New("流水方向不正确")
	}
	req.InvoiceType = strings.TrimSpace(req.InvoiceType)
	req.VerificationType = strings.TrimSpace(req.VerificationType)
	req.VerificationAmountMode = strings.TrimSpace(req.VerificationAmountMode)
	req.InvoiceCode = strings.TrimSpace(req.InvoiceCode)
	req.InvoiceNumber = strings.TrimSpace(req.InvoiceNumber)
	req.CheckCode = strings.TrimSpace(req.CheckCode)
	req.BuyerName = strings.TrimSpace(req.BuyerName)
	req.BuyerTaxNo = strings.TrimSpace(req.BuyerTaxNo)
	req.SellerName = strings.TrimSpace(req.SellerName)
	req.SellerTaxNo = strings.TrimSpace(req.SellerTaxNo)
	req.ReviewNotes = strings.TrimSpace(req.ReviewNotes)
	if req.VerificationAmountMode != "" &&
		req.VerificationAmountMode != model.VerificationAmountModeAmount &&
		req.VerificationAmountMode != model.VerificationAmountModeTotal {
		return errors.New("验真金额口径不正确")
	}
	if err := validateInvoiceAmounts(req.AmountCents, req.TaxCents, req.TotalCents); err != nil {
		return err
	}
	for index := range req.Items {
		item := &req.Items[index]
		item.Name = strings.TrimSpace(item.Name)
		item.Specification = strings.TrimSpace(item.Specification)
		item.Unit = strings.TrimSpace(item.Unit)
		item.QuantityText = strings.TrimSpace(item.QuantityText)
		item.TaxRate = strings.TrimSpace(item.TaxRate)
		if item.Name == "" || item.UnitPriceCents < 0 || item.AmountCents < 0 || item.TaxCents < 0 {
			return errors.New("发票明细名称和金额不正确")
		}
	}
	return nil
}

func (InvoiceService) Update(req invoiceRequest.InvoiceUpdate, scope AccessScope) (model.Invoice, error) {
	if req.ID == 0 {
		return model.Invoice{}, errors.New("缺少发票 ID")
	}
	if err := normalizeInvoiceUpdate(&req); err != nil {
		return model.Invoice{}, err
	}
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var current model.Invoice
		if err := applyInvoiceScope(tx.Model(&model.Invoice{}), scope).Preload("Items").First(&current, req.ID).Error; err != nil {
			return err
		}
		if current.Status == model.InvoiceStatusConfirmed {
			return errors.New("发票已确认，请先重新打开后再编辑")
		}
		if req.CategoryID != nil {
			var count int64
			if err := tx.Model(&model.InvoiceCategory{}).Where("id = ? AND enabled = ?", *req.CategoryID, true).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return errors.New("所选分类不存在或已停用")
			}
		}
		candidate := current
		candidate.Direction = req.Direction
		candidate.InvoiceType = req.InvoiceType
		candidate.VerificationType = req.VerificationType
		candidate.VerificationAmountMode = req.VerificationAmountMode
		candidate.InvoiceCode = req.InvoiceCode
		candidate.InvoiceNumber = req.InvoiceNumber
		candidate.CheckCode = req.CheckCode
		candidate.IssueDate = req.IssueDate
		candidate.BuyerName = req.BuyerName
		candidate.BuyerTaxNo = req.BuyerTaxNo
		candidate.SellerName = req.SellerName
		candidate.SellerTaxNo = req.SellerTaxNo
		candidate.AmountCents = req.AmountCents
		candidate.TaxCents = req.TaxCents
		candidate.TotalCents = req.TotalCents
		candidate.Items = req.Items
		verificationChanged := invoiceVerificationFingerprint(candidate) != invoiceVerificationFingerprint(current)

		updates := map[string]any{
			"direction": req.Direction, "invoice_type": req.InvoiceType, "invoice_code": req.InvoiceCode,
			"verification_type": req.VerificationType, "verification_amount_mode": req.VerificationAmountMode,
			"check_code":     req.CheckCode,
			"invoice_number": req.InvoiceNumber, "issue_date": req.IssueDate, "buyer_name": req.BuyerName,
			"buyer_tax_no": req.BuyerTaxNo, "seller_name": req.SellerName, "seller_tax_no": req.SellerTaxNo,
			"amount_cents": req.AmountCents, "tax_cents": req.TaxCents, "total_cents": req.TotalCents,
			"category_id": req.CategoryID, "classification_source": model.ClassificationManual,
			"review_notes": req.ReviewNotes, "status": model.InvoiceStatusPendingReview,
			"confirmed_by": 0, "confirmed_at": nil, "duplicate_key": nil,
		}
		if verificationChanged {
			updates["verification_fingerprint"] = ""
			updates["verification_invalid_sign"] = ""
			updates["verification_checked_at"] = nil
			if current.ActiveVerificationID == nil {
				updates["verification_status"] = model.InvoiceVerificationUnverified
				updates["verification_message"] = "发票关键字段已修改，请重新查验"
			} else {
				updates["verification_status"] = model.InvoiceVerificationVerifying
				updates["verification_message"] = "查验期间发票关键字段已修改，当前结果完成后需重新查验"
			}
		}
		result := tx.Model(&current).Where("status <> ?", model.InvoiceStatusConfirmed).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("发票状态已变更，请刷新后重试")
		}
		if err := tx.Where("invoice_id = ?", current.ID).Delete(&model.InvoiceItem{}).Error; err != nil {
			return err
		}
		if len(req.Items) > 0 {
			items := make([]model.InvoiceItem, 0, len(req.Items))
			for _, item := range req.Items {
				item.GVA_MODEL = global.GVA_MODEL{}
				item.InvoiceID = current.ID
				items = append(items, item)
			}
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return model.Invoice{}, err
	}
	return InvoiceService{}.Get(req.ID, scope)
}

func validateInvoiceForConfirmation(invoice model.Invoice) error {
	if err := validateInvoiceAmounts(invoice.AmountCents, invoice.TaxCents, invoice.TotalCents); err != nil {
		return err
	}
	switch {
	case strings.TrimSpace(invoice.InvoiceNumber) == "":
		return errors.New("请补充发票号码")
	case invoice.IssueDate == nil:
		return errors.New("请补充开票日期")
	case strings.TrimSpace(invoice.SellerName) == "":
		return errors.New("请补充销售方名称")
	case invoice.TotalCents <= 0:
		return errors.New("价税合计必须大于 0")
	case invoice.CategoryID == nil || *invoice.CategoryID == 0:
		return errors.New("请选择发票分类")
	}
	return nil
}

func validateInvoiceVerificationForConfirmation(invoice model.Invoice) error {
	switch {
	case invoice.VerificationStatus != model.InvoiceVerificationVerifiedValid:
		return errors.New("发票尚未通过权威查验，不能确认入账")
	case invoice.VerificationFingerprint == "" || invoice.VerificationFingerprint != invoiceVerificationFingerprint(invoice):
		return errors.New("发票字段已在查验后变更，请重新查验")
	case invoice.VerificationCheckedAt == nil:
		return errors.New("发票缺少有效查验时间，请重新查验")
	}
	return nil
}

func (InvoiceService) Confirm(id uint, scope AccessScope) (model.Invoice, error) {
	return (InvoiceService{}).ConfirmWithOptions(invoiceRequest.InvoiceConfirm{ID: id}, scope)
}

func normalizeInvoiceConfirm(request *invoiceRequest.InvoiceConfirm, scope AccessScope) error {
	if request.ID == 0 {
		return errors.New("缺少发票 ID")
	}
	request.VerificationBypassReason = strings.TrimSpace(request.VerificationBypassReason)
	if !request.VerificationBypass {
		request.VerificationBypassReason = ""
		return nil
	}
	if !scope.All || scope.AuthorityID != defaultAdminRoleID {
		return errors.New("仅超级管理员可以绕过权威查验确认发票")
	}
	if request.VerificationBypassReason == "" {
		return errors.New("请填写绕过权威查验的原因")
	}
	if utf8.RuneCountInString(request.VerificationBypassReason) > maxVerificationBypassReasonLength {
		return errors.New("绕过权威查验原因不能超过 500 个字符")
	}
	return nil
}

func (InvoiceService) ConfirmWithOptions(request invoiceRequest.InvoiceConfirm, scope AccessScope) (model.Invoice, error) {
	if err := normalizeInvoiceConfirm(&request, scope); err != nil {
		return model.Invoice{}, err
	}
	var confirmed model.Invoice
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := applyInvoiceScope(tx.Model(&model.Invoice{}), scope).Preload("Items").First(&confirmed, request.ID).Error; err != nil {
			return err
		}
		if confirmed.Status == model.InvoiceStatusConfirmed {
			return errors.New("该发票已经确认")
		}
		if err := validateInvoiceForConfirmation(confirmed); err != nil {
			return err
		}
		verificationErr := validateInvoiceVerificationForConfirmation(confirmed)
		verificationBypassed := verificationErr != nil && request.VerificationBypass
		if verificationErr != nil && !verificationBypassed {
			return verificationErr
		}
		duplicate := tx.Model(&model.Invoice{}).Where("id <> ? AND status = ? AND invoice_number = ?", confirmed.ID, model.InvoiceStatusConfirmed, confirmed.InvoiceNumber)
		if confirmed.InvoiceCode != "" {
			duplicate = duplicate.Where("invoice_code = ?", confirmed.InvoiceCode)
		} else if confirmed.SellerTaxNo != "" {
			duplicate = duplicate.Where("seller_tax_no = ?", confirmed.SellerTaxNo)
		} else {
			duplicate = duplicate.Where("seller_name = ?", confirmed.SellerName)
		}
		var count int64
		if err := duplicate.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("系统中已存在相同号码的已确认发票")
		}
		now := time.Now()
		duplicateKey := invoiceDuplicateKey(confirmed)
		if err := beforeInvoiceConfirmPersist(tx, confirmed); err != nil {
			return err
		}
		confirmQuery := tx.Model(&model.Invoice{}).Where(
			"id = ? AND status <> ? AND updated_at = ?",
			confirmed.ID, model.InvoiceStatusConfirmed, confirmed.UpdatedAt,
		)
		if !verificationBypassed {
			confirmQuery = confirmQuery.Where(
				"verification_status = ? AND verification_fingerprint = ?",
				model.InvoiceVerificationVerifiedValid, confirmed.VerificationFingerprint,
			)
		}
		updates := map[string]any{
			"status": model.InvoiceStatusConfirmed, "confirmed_by": scope.UserID, "confirmed_at": &now,
			"duplicate_key": &duplicateKey, "confirmation_verification_status": confirmed.VerificationStatus,
			"verification_bypassed": false, "verification_bypass_reason": "",
			"verification_bypassed_by": 0, "verification_bypassed_at": nil,
		}
		if verificationBypassed {
			updates["verification_bypassed"] = true
			updates["verification_bypass_reason"] = request.VerificationBypassReason
			updates["verification_bypassed_by"] = scope.UserID
			updates["verification_bypassed_at"] = &now
		}
		result := confirmQuery.Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("发票字段或查验状态已变更，请刷新后重新查验")
		}
		return nil
	})
	if err != nil {
		if isUniqueConstraintError(err) {
			return model.Invoice{}, errors.New("系统中已存在相同号码的已确认发票")
		}
		return model.Invoice{}, err
	}
	return InvoiceService{}.Get(request.ID, scope)
}

func (InvoiceService) Reopen(id uint, scope AccessScope) (model.Invoice, error) {
	if id == 0 {
		return model.Invoice{}, errors.New("缺少发票 ID")
	}
	if !scope.All {
		return model.Invoice{}, errors.New("已确认发票只能由管理员重新打开")
	}
	result := applyInvoiceScope(global.GVA_DB.Model(&model.Invoice{}), scope).
		Where("id = ? AND status = ?", id, model.InvoiceStatusConfirmed).
		Updates(map[string]any{
			"status": model.InvoiceStatusPendingReview, "confirmed_by": 0,
			"confirmed_at": nil, "duplicate_key": nil,
		})
	if result.Error != nil {
		return model.Invoice{}, result.Error
	}
	if result.RowsAffected == 0 {
		var invoice model.Invoice
		if err := applyInvoiceScope(global.GVA_DB.Model(&model.Invoice{}), scope).First(&invoice, id).Error; err != nil {
			return model.Invoice{}, err
		}
		return model.Invoice{}, errors.New("该发票尚未确认，无需重新打开")
	}
	return InvoiceService{}.Get(id, scope)
}

func (InvoiceService) Delete(id uint, scope AccessScope) error {
	invoice, err := InvoiceService{}.Get(id, scope)
	if err != nil {
		return err
	}
	if invoice.Status == model.InvoiceStatusConfirmed && !scope.All {
		return errors.New("已确认发票只能由管理员删除")
	}
	var cleanupJob model.InvoiceFileCleanupJob
	nextRunAt := time.Now()
	if err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if invoice.FileKey != "" {
			cleanupJob = model.InvoiceFileCleanupJob{
				FileKey: invoice.FileKey, StorageType: invoice.StorageType,
				StorageRoot: invoice.StorageRoot, StorageEndpoint: invoice.StorageEndpoint,
				StorageBucket: invoice.StorageBucket, StorageUseSSL: invoice.StorageUseSSL,
				Status: model.FileCleanupJobPending, NextRunAt: &nextRunAt,
			}
			if err := tx.Create(&cleanupJob).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&model.Invoice{}).Where("id = ?", id).Updates(map[string]any{
			"duplicate_key": nil, "file_dedup_key": nil,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("invoice_id = ?", id).Delete(&model.InvoiceItem{}).Error; err != nil {
			return err
		}
		if err := tx.Where("invoice_id = ?", id).Delete(&model.RecognitionJob{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Invoice{}, id).Error
	}); err != nil {
		return err
	}
	if cleanupJob.ID != 0 {
		processInvoiceFileCleanupJob(cleanupJob.ID)
	}
	return nil
}

func (InvoiceService) OpenFile(ctx context.Context, id uint, scope AccessScope) (model.Invoice, io.ReadCloser, error) {
	invoice, err := InvoiceService{}.Get(id, scope)
	if err != nil {
		return invoice, nil, err
	}
	reader, err := openInvoiceObject(ctx, invoice)
	return invoice, reader, err
}

func (InvoiceService) Dashboard(scope AccessScope) (invoiceResponse.Dashboard, error) {
	var result invoiceResponse.Dashboard
	db := applyInvoiceScope(global.GVA_DB.Model(&model.Invoice{}), scope)
	var counts struct {
		Confirmed int64
		Pending   int64
		Failed    int64
	}
	if err := db.Select(`
		SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS confirmed,
		SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS pending,
		SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS failed`,
		model.InvoiceStatusConfirmed, model.InvoiceStatusPendingReview, model.InvoiceStatusRecognitionFailed,
	).Scan(&counts).Error; err != nil {
		return result, err
	}
	result.ConfirmedCount, result.PendingCount, result.FailedCount = counts.Confirmed, counts.Pending, counts.Failed
	var totals struct{ TotalCents, AmountCents, TaxCents int64 }
	if err := db.Where("status = ?", model.InvoiceStatusConfirmed).
		Select("COALESCE(SUM(total_cents), 0) AS total_cents, COALESCE(SUM(amount_cents), 0) AS amount_cents, COALESCE(SUM(tax_cents), 0) AS tax_cents").
		Scan(&totals).Error; err != nil {
		return result, err
	}
	result.TotalCents, result.AmountCents, result.TaxCents = totals.TotalCents, totals.AmountCents, totals.TaxCents

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -11, 0)
	var trendRows []struct {
		IssueDate  time.Time
		TotalCents int64
	}
	if err := db.Where("status = ? AND issue_date >= ?", model.InvoiceStatusConfirmed, start).
		Select("issue_date, total_cents").Find(&trendRows).Error; err != nil {
		return result, err
	}
	trendMap := map[string]*invoiceResponse.MonthlySummary{}
	for index := 0; index < 12; index++ {
		month := start.AddDate(0, index, 0).Format("2006-01")
		result.MonthlyTrend = append(result.MonthlyTrend, invoiceResponse.MonthlySummary{Month: month})
		trendMap[month] = &result.MonthlyTrend[len(result.MonthlyTrend)-1]
	}
	for _, row := range trendRows {
		if item := trendMap[row.IssueDate.Format("2006-01")]; item != nil {
			item.Count++
			item.TotalCents += row.TotalCents
		}
	}

	categoryDB := applyInvoiceScope(global.GVA_DB.Model(&model.Invoice{}), scope)
	if err := categoryDB.Where("invoices.status = ?", model.InvoiceStatusConfirmed).
		Joins("JOIN invoice_categories ON invoice_categories.id = invoices.category_id").
		Select("invoice_categories.id AS category_id, invoice_categories.name, invoice_categories.color, COUNT(invoices.id) AS count, COALESCE(SUM(invoices.total_cents), 0) AS total_cents").
		Group("invoice_categories.id, invoice_categories.name, invoice_categories.color").
		Order("total_cents DESC").Scan(&result.Categories).Error; err != nil {
		return result, err
	}
	supplierDB := applyInvoiceScope(global.GVA_DB.Model(&model.Invoice{}), scope)
	if err := supplierDB.Where("status = ? AND seller_name <> ''", model.InvoiceStatusConfirmed).
		Select("seller_name AS name, COUNT(id) AS count, COALESCE(SUM(total_cents), 0) AS total_cents").
		Group("seller_name").Order("total_cents DESC").Limit(8).Scan(&result.Suppliers).Error; err != nil {
		return result, err
	}
	recentDB := applyInvoiceScope(global.GVA_DB.Model(&model.Invoice{}), scope)
	if err := recentDB.Preload("Category").Order("created_at DESC, id DESC").Limit(6).Find(&result.Recent).Error; err != nil {
		return result, err
	}
	return result, nil
}
