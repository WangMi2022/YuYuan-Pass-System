package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/provider"
	"gorm.io/gorm"
)

type VerificationService struct{}

type VerificationOutcome struct {
	Invoice model.Invoice             `json:"invoice"`
	Attempt model.InvoiceVerification `json:"attempt"`
}

const verificationLeaseTimeout = 5 * time.Minute

var errVerificationLeaseLost = errors.New("发票查验任务已被新的请求接管")

var newInvoiceVerifier = provider.NewVerificationAdapter

type verificationRule struct {
	checkCodeRequired   bool
	amountMode          string
	motorVehicle        bool
	invoiceCodeOptional bool
}

var verificationRules = map[string]verificationRule{
	"special_vat_invoice":               {amountMode: model.VerificationAmountModeAmount},
	"elec_special_vat_invoice":          {checkCodeRequired: true, amountMode: model.VerificationAmountModeAmount},
	"normal_invoice":                    {checkCodeRequired: true},
	"elec_normal_invoice":               {checkCodeRequired: true},
	"roll_normal_invoice":               {checkCodeRequired: true},
	"toll_elec_normal_invoice":          {checkCodeRequired: true},
	"blockchain_invoice":                {checkCodeRequired: true, amountMode: model.VerificationAmountModeAmount},
	"elec_invoice_special":              {amountMode: model.VerificationAmountModeTotal, invoiceCodeOptional: true},
	"elec_invoice_normal":               {amountMode: model.VerificationAmountModeTotal, invoiceCodeOptional: true},
	"special_freight_transport_invoice": {amountMode: model.VerificationAmountModeAmount},
	"motor_vehicle_invoice":             {motorVehicle: true},
	"used_vehicle_invoice":              {amountMode: model.VerificationAmountModeTotal, invoiceCodeOptional: true},
	"elec_flight_itinerary_invoice":     {amountMode: model.VerificationAmountModeTotal, invoiceCodeOptional: true},
	"elec_train_ticket_invoice":         {amountMode: model.VerificationAmountModeTotal, invoiceCodeOptional: true},
	"elec_toll_invoice":                 {amountMode: model.VerificationAmountModeTotal, invoiceCodeOptional: true},
}

func invoiceVerificationFingerprint(invoice model.Invoice) string {
	parts := []string{
		"invoice-verification-v2", invoice.VerificationType, invoice.VerificationAmountMode, invoice.InvoiceType,
		invoice.InvoiceCode, invoice.InvoiceNumber, invoice.CheckCode,
		invoice.BuyerName, invoice.BuyerTaxNo, invoice.SellerName, invoice.SellerTaxNo,
		strconv.FormatInt(invoice.AmountCents, 10), strconv.FormatInt(invoice.TaxCents, 10),
		strconv.FormatInt(invoice.TotalCents, 10),
	}
	if invoice.IssueDate != nil {
		parts = append(parts, invoice.IssueDate.Format("20060102"))
	}
	for _, item := range invoice.Items {
		parts = append(parts, item.Name, item.Specification, item.Unit, item.QuantityText,
			strconv.FormatInt(item.UnitPriceCents, 10), strconv.FormatInt(item.AmountCents, 10),
			item.TaxRate, strconv.FormatInt(item.TaxCents, 10))
	}
	return hashInvoiceKey(parts...)
}

func resolveVerificationType(invoice model.Invoice) string {
	if value := provider.NormalizeInvoiceType(invoice.VerificationType); value != "" {
		return value
	}
	return provider.NormalizeInvoiceType(invoice.InvoiceType)
}

func lastSix(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 6 {
		return value
	}
	return value[len(value)-6:]
}

func formatVerificationAmount(cents int64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

func buildVerificationRequest(invoice model.Invoice) (provider.VerificationRequest, error) {
	verificationType := resolveVerificationType(invoice)
	rule, ok := verificationRules[verificationType]
	if !ok {
		return provider.VerificationRequest{}, errors.New("请选择系统支持的发票验真类型")
	}
	if strings.TrimSpace(invoice.InvoiceNumber) == "" {
		return provider.VerificationRequest{}, errors.New("请补充发票号码")
	}
	if invoice.IssueDate == nil {
		return provider.VerificationRequest{}, errors.New("请补充开票日期")
	}
	invoiceCodeOptional := rule.invoiceCodeOptional ||
		(rule.motorVehicle && invoice.VerificationAmountMode == model.VerificationAmountModeTotal)
	if !invoiceCodeOptional && strings.TrimSpace(invoice.InvoiceCode) == "" {
		return provider.VerificationRequest{}, errors.New("该票种验真需要填写发票代码")
	}
	request := provider.VerificationRequest{
		InvoiceCode: strings.TrimSpace(invoice.InvoiceCode), InvoiceNum: strings.TrimSpace(invoice.InvoiceNumber),
		InvoiceDate: invoice.IssueDate.Format("20060102"), InvoiceType: verificationType,
	}
	if rule.checkCodeRequired {
		request.CheckCode = lastSix(invoice.CheckCode)
		if len(request.CheckCode) != 6 {
			return provider.VerificationRequest{}, errors.New("该票种验真需要填写校验码后 6 位")
		}
	}
	amountMode := rule.amountMode
	if rule.motorVehicle {
		amountMode = strings.TrimSpace(invoice.VerificationAmountMode)
		if amountMode != model.VerificationAmountModeAmount && amountMode != model.VerificationAmountModeTotal {
			return provider.VerificationRequest{}, errors.New("机动车销售统一发票请选择纸质票或电子票验真口径")
		}
	}
	request.AmountMode = amountMode
	switch amountMode {
	case model.VerificationAmountModeAmount:
		if invoice.AmountCents <= 0 {
			return provider.VerificationRequest{}, errors.New("该票种验真需要填写不含税金额")
		}
		request.TotalAmount = formatVerificationAmount(invoice.AmountCents)
	case model.VerificationAmountModeTotal:
		if invoice.TotalCents <= 0 {
			return provider.VerificationRequest{}, errors.New("该票种验真需要填写价税合计")
		}
		request.TotalAmount = formatVerificationAmount(invoice.TotalCents)
	}
	return request, nil
}

func localVerificationSnapshot(invoice model.Invoice) map[string]string {
	snapshot := map[string]string{
		"verificationType": invoice.VerificationType, "invoiceType": invoice.InvoiceType,
		"verificationAmountMode": invoice.VerificationAmountMode,
		"invoiceCode":            strings.TrimSpace(invoice.InvoiceCode), "invoiceNumber": strings.TrimSpace(invoice.InvoiceNumber),
		"checkCode": strings.TrimSpace(invoice.CheckCode), "buyerName": strings.TrimSpace(invoice.BuyerName),
		"buyerTaxNo": strings.TrimSpace(invoice.BuyerTaxNo), "sellerName": strings.TrimSpace(invoice.SellerName),
		"sellerTaxNo": strings.TrimSpace(invoice.SellerTaxNo), "amountCents": strconv.FormatInt(invoice.AmountCents, 10),
		"taxCents": strconv.FormatInt(invoice.TaxCents, 10), "totalCents": strconv.FormatInt(invoice.TotalCents, 10),
	}
	if invoice.IssueDate != nil {
		snapshot["issueDate"] = invoice.IssueDate.Format("20060102")
	}
	if items, err := json.Marshal(invoice.Items); err == nil {
		snapshot["items"] = string(items)
	}
	return snapshot
}

var verificationFieldLabels = map[string]string{
	"invoiceType": "发票类型", "verificationAmountMode": "验真金额口径",
	"invoiceCode": "发票代码", "invoiceNumber": "发票号码",
	"checkCode": "校验码", "issueDate": "开票日期", "buyerName": "购买方名称",
	"buyerTaxNo": "购买方税号", "sellerName": "销售方名称", "sellerTaxNo": "销售方税号",
	"amountCents": "不含税金额", "taxCents": "税额", "totalCents": "价税合计",
}

func requiredOfficialFields(request provider.VerificationRequest) []string {
	fields := []string{"invoiceNumber", "issueDate"}
	if request.InvoiceCode != "" {
		fields = append(fields, "invoiceCode")
	}
	if request.CheckCode != "" {
		fields = append(fields, "checkCode")
	}
	switch request.AmountMode {
	case model.VerificationAmountModeAmount:
		fields = append(fields, "amountCents")
	case model.VerificationAmountModeTotal:
		fields = append(fields, "totalCents")
	}
	return fields
}

func missingOfficialFields(request provider.VerificationRequest, official map[string]string) []string {
	missing := make([]string, 0)
	for _, field := range requiredOfficialFields(request) {
		if strings.TrimSpace(official[field]) == "" {
			missing = append(missing, verificationFieldLabels[field])
		}
	}
	return missing
}

func normalizeComparison(field, value string) string {
	value = strings.TrimSpace(value)
	switch field {
	case "invoiceType":
		if normalized := provider.NormalizeInvoiceType(value); normalized != "" {
			return normalized
		}
	case "issueDate":
		value = strings.NewReplacer("年", "", "月", "", "日", "", "-", "", "/", "").Replace(value)
	case "invoiceCode", "invoiceNumber", "buyerTaxNo", "sellerTaxNo":
		return strings.ToUpper(strings.ReplaceAll(value, " ", ""))
	case "buyerName", "sellerName":
		return strings.ReplaceAll(value, " ", "")
	}
	return value
}

func verificationDifferences(invoice model.Invoice, official map[string]string) []model.InvoiceVerificationDifference {
	local := localVerificationSnapshot(invoice)
	local["invoiceType"] = invoice.VerificationType
	keys := make([]string, 0, len(official))
	for key := range official {
		if _, comparable := verificationFieldLabels[key]; comparable {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	differences := make([]model.InvoiceVerificationDifference, 0)
	for _, key := range keys {
		localValue, officialValue := local[key], official[key]
		matches := normalizeComparison(key, localValue) == normalizeComparison(key, officialValue)
		if key == "checkCode" {
			matches = lastSix(localValue) != "" && lastSix(localValue) == lastSix(officialValue)
		}
		if !matches {
			differences = append(differences, model.InvoiceVerificationDifference{
				Field: key, Label: verificationFieldLabels[key], LocalValue: localValue, OfficialValue: officialValue,
			})
		}
	}
	return differences
}

func verificationStatus(result provider.VerificationResult, differences []model.InvoiceVerificationDifference, authorityFieldsMissing bool) string {
	if len(differences) > 0 {
		return model.InvoiceVerificationInconsistent
	}
	if result.Outcome == provider.VerificationOutcomeValid && authorityFieldsMissing {
		return model.InvoiceVerificationUnavailable
	}
	switch result.Outcome {
	case provider.VerificationOutcomeValid:
		return model.InvoiceVerificationVerifiedValid
	case provider.VerificationOutcomeVoided:
		return model.InvoiceVerificationVoided
	case provider.VerificationOutcomeRedFlushed:
		return model.InvoiceVerificationRed
	case provider.VerificationOutcomeMismatch:
		return model.InvoiceVerificationInconsistent
	case provider.VerificationOutcomeNotFound:
		return model.InvoiceVerificationNotFound
	case provider.VerificationOutcomeDeferred:
		return model.InvoiceVerificationDeferred
	case provider.VerificationOutcomeExpired:
		return model.InvoiceVerificationExpired
	default:
		return model.InvoiceVerificationUnavailable
	}
}

func finishVerificationAttempt(
	attempt *model.InvoiceVerification,
	result provider.VerificationResult,
	status string,
	differences []model.InvoiceVerificationDifference,
) error {
	completedAt := time.Now()
	verifyMessage := truncateUTF8(strings.TrimSpace(result.VerifyMessage), 1000)
	officialSnapshot, err := json.Marshal(result.Official)
	if err != nil {
		return fmt.Errorf("序列化权威查验结果失败: %w", err)
	}
	differenceSnapshot, err := json.Marshal(differences)
	if err != nil {
		return fmt.Errorf("序列化发票字段差异失败: %w", err)
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		attemptUpdate := tx.Model(&model.InvoiceVerification{}).
			Where("id = ? AND completed_at IS NULL", attempt.ID).
			Updates(map[string]any{
				"status": status, "verify_result": result.VerifyResult, "verify_message": verifyMessage,
				"verify_frequency": result.VerifyFrequency, "invalid_sign": result.InvalidSign,
				"provider_log_id": result.ProviderLogID, "official_snapshot": string(officialSnapshot),
				"differences": string(differenceSnapshot), "raw_payload": result.RawPayload, "completed_at": &completedAt,
			})
		if attemptUpdate.Error != nil {
			return attemptUpdate.Error
		}
		if attemptUpdate.RowsAffected == 0 {
			return errVerificationLeaseLost
		}
		var current model.Invoice
		if err := tx.Preload("Items").First(&current, attempt.InvoiceID).Error; err != nil {
			return err
		}
		updates := map[string]any{
			"latest_verification_id": attempt.ID, "verification_checked_at": &completedAt,
			"verification_provider": attempt.Provider, "verification_message": verifyMessage,
			"verification_invalid_sign": result.InvalidSign, "active_verification_id": nil,
			"verification_started_at": nil,
		}
		if invoiceVerificationFingerprint(current) != attempt.LocalFingerprint {
			updates["verification_status"] = model.InvoiceVerificationUnverified
			updates["verification_fingerprint"] = ""
			updates["verification_message"] = "查验期间发票关键字段已修改，请重新查验"
		} else {
			updates["verification_status"] = status
			updates["verification_fingerprint"] = attempt.LocalFingerprint
		}
		invoiceUpdate := tx.Model(&model.Invoice{}).
			Where("id = ? AND active_verification_id = ?", current.ID, attempt.ID).
			Updates(updates)
		if invoiceUpdate.Error != nil {
			return invoiceUpdate.Error
		}
		if invoiceUpdate.RowsAffected == 0 {
			return errVerificationLeaseLost
		}
		attempt.Status = status
		attempt.VerifyResult = result.VerifyResult
		attempt.VerifyMessage = verifyMessage
		attempt.VerifyFrequency = result.VerifyFrequency
		attempt.InvalidSign = result.InvalidSign
		attempt.ProviderLogID = result.ProviderLogID
		attempt.OfficialSnapshot = result.Official
		attempt.Differences = differences
		attempt.CompletedAt = &completedAt
		return nil
	})
}

func startVerificationAttempt(id uint, scope AccessScope, providerID string) (model.Invoice, provider.VerificationRequest, model.InvoiceVerification, error) {
	var invoice model.Invoice
	var request provider.VerificationRequest
	var attempt model.InvoiceVerification
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := applyInvoiceScope(tx.Model(&model.Invoice{}), scope).Preload("Items").First(&invoice, id).Error; err != nil {
			return err
		}
		if invoice.Status == model.InvoiceStatusConfirmed {
			return errors.New("发票已确认，请先重新打开后再查验")
		}
		verificationType := resolveVerificationType(invoice)
		invoice.VerificationType = verificationType
		var err error
		request, err = buildVerificationRequest(invoice)
		if err != nil {
			return err
		}
		fingerprint := invoiceVerificationFingerprint(invoice)
		requestSnapshot := localVerificationSnapshot(invoice)
		for key, value := range request.Snapshot() {
			requestSnapshot[key] = value
		}

		now := time.Now()
		if invoice.ActiveVerificationID != nil {
			leaseActive := invoice.VerificationStartedAt != nil && invoice.VerificationStartedAt.After(now.Add(-verificationLeaseTimeout))
			if leaseActive {
				return errors.New("该发票正在查验，请稍后重试")
			}
			completedAt := now
			if err = tx.Model(&model.InvoiceVerification{}).
				Where("id = ? AND completed_at IS NULL", *invoice.ActiveVerificationID).
				Updates(map[string]any{
					"status":         model.InvoiceVerificationUnavailable,
					"verify_message": "查验任务超时，已由新请求接管", "completed_at": &completedAt,
				}).Error; err != nil {
				return err
			}
		}

		attempt = model.InvoiceVerification{
			InvoiceID: invoice.ID, Provider: providerID,
			Status: model.InvoiceVerificationVerifying, LocalFingerprint: fingerprint,
			RequestSnapshot: requestSnapshot, RequestedBy: scope.UserID,
		}
		if err = tx.Create(&attempt).Error; err != nil {
			return err
		}
		claim := applyInvoiceScope(tx.Model(&model.Invoice{}), scope).
			Where("id = ? AND updated_at = ?", invoice.ID, invoice.UpdatedAt)
		if invoice.ActiveVerificationID == nil {
			claim = claim.Where("active_verification_id IS NULL")
		} else {
			claim = claim.Where("active_verification_id = ?", *invoice.ActiveVerificationID)
		}
		claim = claim.Updates(map[string]any{
			"verification_type": verificationType, "verification_status": model.InvoiceVerificationVerifying,
			"verification_message": "正在连接权威查验服务", "active_verification_id": attempt.ID,
			"verification_started_at": &now,
		})
		if claim.Error != nil {
			return claim.Error
		}
		if claim.RowsAffected == 0 {
			return errors.New("发票字段或查验状态已变更，请刷新后重试")
		}
		return nil
	})
	return invoice, request, attempt, err
}

func (VerificationService) Verify(ctx context.Context, id uint, scope AccessScope) (VerificationOutcome, error) {
	adapter, err := newInvoiceVerifier(global.GVA_CONFIG.InvoiceRecognition)
	if err != nil {
		return VerificationOutcome{}, err
	}
	invoice, request, attempt, err := startVerificationAttempt(id, scope, adapter.Provider)
	if err != nil {
		return VerificationOutcome{}, err
	}
	result, verifyErr := adapter.Verifier.Verify(ctx, request)
	if verifyErr != nil {
		result.VerifyMessage = verifyErr.Error()
		if finishErr := finishVerificationAttempt(&attempt, result, model.InvoiceVerificationUnavailable, nil); finishErr != nil {
			return VerificationOutcome{}, errors.Join(verifyErr, finishErr)
		}
		return VerificationOutcome{}, fmt.Errorf("发票查验暂不可用: %w", verifyErr)
	}
	differences := verificationDifferences(invoice, result.Official)
	missingFields := missingOfficialFields(request, result.Official)
	if result.Outcome == provider.VerificationOutcomeValid && len(missingFields) > 0 {
		message := strings.TrimSpace(result.VerifyMessage)
		if message != "" {
			message += "；"
		}
		result.VerifyMessage = message + "查验响应缺少权威字段：" + strings.Join(missingFields, "、")
	}
	status := verificationStatus(result, differences, len(missingFields) > 0)
	if err = finishVerificationAttempt(&attempt, result, status, differences); err != nil {
		return VerificationOutcome{}, err
	}
	updated, err := InvoiceService{}.Get(id, scope)
	if err != nil {
		return VerificationOutcome{}, err
	}
	return VerificationOutcome{Invoice: updated, Attempt: attempt}, nil
}

func (VerificationService) History(invoiceID uint, scope AccessScope) ([]model.InvoiceVerification, error) {
	if _, err := (InvoiceService{}).Get(invoiceID, scope); err != nil {
		return nil, err
	}
	var history []model.InvoiceVerification
	err := global.GVA_DB.Where("invoice_id = ?", invoiceID).
		Order("created_at DESC, id DESC").Limit(50).Find(&history).Error
	return history, err
}
