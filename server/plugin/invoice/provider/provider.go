package provider

import (
	"context"
	"errors"
	"math"
	"os"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model"
)

type Input struct {
	FileName    string
	ContentType string
	Data        []byte
}

type Result struct {
	Provider         string              `json:"provider"`
	InvoiceType      string              `json:"invoiceType"`
	InvoiceCode      string              `json:"invoiceCode"`
	InvoiceNumber    string              `json:"invoiceNumber"`
	IssueDate        *time.Time          `json:"issueDate"`
	BuyerName        string              `json:"buyerName"`
	BuyerTaxNo       string              `json:"buyerTaxNo"`
	SellerName       string              `json:"sellerName"`
	SellerTaxNo      string              `json:"sellerTaxNo"`
	AmountCents      int64               `json:"amountCents"`
	TaxCents         int64               `json:"taxCents"`
	TotalCents       int64               `json:"totalCents"`
	RawText          string              `json:"rawText"`
	RawPayload       string              `json:"rawPayload"`
	Confidence       float64             `json:"confidence"`
	FieldConfidences map[string]float64  `json:"fieldConfidences"`
	Items            []model.InvoiceItem `json:"items"`
}

type Recognizer interface {
	Recognize(ctx context.Context, input Input) (Result, error)
}

type Chain struct {
	qr                Recognizer
	ocr               Recognizer
	multimodal        Recognizer
	fallbackThreshold float64
}

func NewFromEnvironment() *Chain {
	endpoint := strings.TrimSpace(os.Getenv("INVOICE_OCR_ENDPOINT"))
	var ocrRecognizer Recognizer
	if endpoint != "" {
		ocrRecognizer = &HTTPRecognizer{
			Endpoint: endpoint,
			Token:    strings.TrimSpace(os.Getenv("INVOICE_OCR_TOKEN")),
			Timeout:  30 * time.Second,
		}
	}
	return &Chain{qr: QRRecognizer{}, ocr: ocrRecognizer, fallbackThreshold: 0.82}
}

func New(configuration config.InvoiceRecognition) *Chain {
	configuration.Normalize()
	chain := &Chain{qr: QRRecognizer{}, fallbackThreshold: configuration.FallbackThreshold}
	if configuration.PublicOCR.Enabled && configuration.PublicOCR.Endpoint != "" {
		chain.ocr = &HTTPRecognizer{
			Endpoint: configuration.PublicOCR.Endpoint,
			Token:    configuration.PublicOCR.APIKey,
			Timeout:  time.Duration(configuration.PublicOCR.TimeoutSeconds) * time.Second,
			Provider: configuration.PublicOCR.Provider,
		}
	}
	if configuration.Multimodal.Enabled && configuration.Multimodal.BaseURL != "" {
		chain.multimodal = &MultimodalRecognizer{
			BaseURL:  configuration.Multimodal.BaseURL,
			APIKey:   configuration.Multimodal.APIKey,
			Model:    configuration.Multimodal.Model,
			Protocol: configuration.Multimodal.Protocol,
			Timeout:  time.Duration(configuration.Multimodal.TimeoutSeconds) * time.Second,
		}
	}
	return chain
}

func (c *Chain) Recognize(ctx context.Context, input Input) (Result, error) {
	qrResult, qrErr := c.qr.Recognize(ctx, input)

	var ocrResult Result
	var ocrErr error
	if c.ocr != nil {
		ocrResult, ocrErr = c.ocr.Recognize(ctx, input)
		if ocrErr == nil && !validConfidence(ocrResult.Confidence) {
			ocrErr = errors.New("公网 OCR 返回的识别置信度不正确")
		}
		if ocrErr == nil && !hasRecognitionData(ocrResult) {
			ocrErr = errors.New("公网 OCR 未返回可用发票字段")
		}
		if ocrErr == nil && ocrResult.Confidence >= c.fallbackThreshold {
			mergeVerifiedQR(&ocrResult, qrResult, qrErr)
			return ocrResult, nil
		}
	}

	var multimodalResult Result
	var multimodalErr error
	if c.multimodal != nil {
		multimodalResult, multimodalErr = c.multimodal.Recognize(ctx, input)
		if multimodalErr == nil && !validConfidence(multimodalResult.Confidence) {
			multimodalErr = errors.New("多模态模型返回的识别置信度不正确")
		}
		if multimodalErr == nil && !hasRecognitionData(multimodalResult) {
			multimodalErr = errors.New("多模态模型未返回可用发票字段")
		}
		if multimodalErr == nil {
			if ocrErr == nil {
				mergeMissing(&multimodalResult, ocrResult)
			}
			mergeVerifiedQR(&multimodalResult, qrResult, qrErr)
			return multimodalResult, nil
		}
	}

	if ocrErr == nil && hasRecognitionData(ocrResult) {
		mergeVerifiedQR(&ocrResult, qrResult, qrErr)
		return ocrResult, nil
	}
	if qrErr == nil && hasRecognitionData(qrResult) {
		return qrResult, nil
	}
	if c.ocr == nil && c.multimodal == nil {
		return Result{}, errors.Join(qrErr, errors.New("未启用公网 OCR 或多模态识别服务"))
	}
	return Result{}, errors.Join(qrErr, ocrErr, multimodalErr)
}

func mergeMissing(target *Result, fallback Result) {
	if target.InvoiceType == "" {
		target.InvoiceType = fallback.InvoiceType
	}
	if target.InvoiceCode == "" {
		target.InvoiceCode = fallback.InvoiceCode
	}
	if target.InvoiceNumber == "" {
		target.InvoiceNumber = fallback.InvoiceNumber
	}
	if target.IssueDate == nil {
		target.IssueDate = fallback.IssueDate
	}
	if target.BuyerName == "" {
		target.BuyerName = fallback.BuyerName
	}
	if target.BuyerTaxNo == "" {
		target.BuyerTaxNo = fallback.BuyerTaxNo
	}
	if target.SellerName == "" {
		target.SellerName = fallback.SellerName
	}
	if target.SellerTaxNo == "" {
		target.SellerTaxNo = fallback.SellerTaxNo
	}
	if target.AmountCents == 0 {
		target.AmountCents = fallback.AmountCents
	}
	if target.TaxCents == 0 {
		target.TaxCents = fallback.TaxCents
	}
	if target.TotalCents == 0 {
		target.TotalCents = fallback.TotalCents
	}
	if len(target.Items) == 0 {
		target.Items = fallback.Items
	}
	if target.RawText == "" {
		target.RawText = fallback.RawText
	}
	if fallback.RawText != "" && !strings.Contains(target.RawText, fallback.RawText) {
		target.RawText += "\n" + fallback.RawText
	}
	if target.FieldConfidences == nil {
		target.FieldConfidences = map[string]float64{}
	}
	for field, confidence := range fallback.FieldConfidences {
		if _, exists := target.FieldConfidences[field]; !exists {
			target.FieldConfidences[field] = confidence
		}
	}
}

func mergeVerifiedQR(target *Result, qr Result, qrErr error) {
	if qrErr != nil {
		return
	}
	mergeMissing(target, qr)
	if qr.FieldConfidences["invoiceType"] >= 0.95 {
		target.InvoiceType = qr.InvoiceType
	}
	if qr.FieldConfidences["invoiceCode"] >= 0.95 {
		target.InvoiceCode = qr.InvoiceCode
	}
	if qr.FieldConfidences["invoiceNumber"] >= 0.95 {
		target.InvoiceNumber = qr.InvoiceNumber
	}
	if qr.FieldConfidences["issueDate"] >= 0.95 {
		target.IssueDate = qr.IssueDate
	}
	if qr.FieldConfidences["totalCents"] >= 0.95 {
		target.TotalCents = qr.TotalCents
	}
}

func hasRecognitionData(result Result) bool {
	return result.InvoiceNumber != "" ||
		(result.InvoiceCode != "" && result.IssueDate != nil) ||
		(result.IssueDate != nil && result.TotalCents != 0 && (result.BuyerName != "" || result.SellerName != ""))
}

func validConfidence(confidence float64) bool {
	return !math.IsNaN(confidence) && !math.IsInf(confidence, 0) && confidence >= 0 && confidence <= 1
}

func TestConnection(ctx context.Context, target string, configuration config.InvoiceRecognition) (string, error) {
	configuration.Normalize()
	switch target {
	case "public-ocr":
		configuration.Multimodal.Enabled = false
		if err := configuration.Validate(); err != nil {
			return "", err
		}
		if !configuration.PublicOCR.Enabled {
			return "", errors.New("请先启用公网 OCR")
		}
		err := (&HTTPRecognizer{
			Endpoint: configuration.PublicOCR.Endpoint,
			Token:    configuration.PublicOCR.APIKey,
			Timeout:  time.Duration(configuration.PublicOCR.TimeoutSeconds) * time.Second,
			Provider: configuration.PublicOCR.Provider,
		}).Probe(ctx)
		return "", err
	case "multimodal":
		configuration.PublicOCR.Enabled = false
		if err := configuration.Validate(); err != nil {
			return "", err
		}
		if !configuration.Multimodal.Enabled {
			return "", errors.New("请先启用多模态模型")
		}
		return (&MultimodalRecognizer{
			BaseURL:  configuration.Multimodal.BaseURL,
			APIKey:   configuration.Multimodal.APIKey,
			Model:    configuration.Multimodal.Model,
			Protocol: configuration.Multimodal.Protocol,
			Timeout:  time.Duration(configuration.Multimodal.TimeoutSeconds) * time.Second,
		}).Probe(ctx)
	default:
		return "", errors.New("不支持的识别服务类型")
	}
}
