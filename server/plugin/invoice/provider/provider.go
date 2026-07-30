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
	Provider               string              `json:"provider"`
	InvoiceType            string              `json:"invoiceType"`
	VerificationType       string              `json:"verificationType"`
	VerificationAmountMode string              `json:"verificationAmountMode"`
	InvoiceCode            string              `json:"invoiceCode"`
	InvoiceNumber          string              `json:"invoiceNumber"`
	CheckCode              string              `json:"checkCode"`
	IssueDate              *time.Time          `json:"issueDate"`
	BuyerName              string              `json:"buyerName"`
	BuyerTaxNo             string              `json:"buyerTaxNo"`
	SellerName             string              `json:"sellerName"`
	SellerTaxNo            string              `json:"sellerTaxNo"`
	AmountCents            int64               `json:"amountCents"`
	TaxCents               int64               `json:"taxCents"`
	TotalCents             int64               `json:"totalCents"`
	RawText                string              `json:"rawText"`
	RawPayload             string              `json:"rawPayload"`
	Confidence             float64             `json:"confidence"`
	FieldConfidences       map[string]float64  `json:"fieldConfidences"`
	Items                  []model.InvoiceItem `json:"items"`
}

type Recognizer interface {
	Recognize(ctx context.Context, input Input) (Result, error)
}

// ConnectionInfo is server-detected metadata. It is persisted only after a
// successful probe and is read-only to browser clients.
type ConnectionInfo struct {
	Provider string `json:"provider"`
	Protocol string `json:"protocol"`
}

type Chain struct {
	qr                Recognizer
	ocr               Recognizer
	additionalOCRs    []Recognizer
	multimodal        Recognizer
	fallbackThreshold float64
}

func NewFromEnvironment() *Chain {
	endpoint := strings.TrimSpace(os.Getenv("INVOICE_OCR_ENDPOINT"))
	var ocrRecognizer Recognizer
	if endpoint != "" {
		ocrRecognizer = &HTTPRecognizer{
			Endpoint:              endpoint,
			Token:                 strings.TrimSpace(os.Getenv("INVOICE_OCR_TOKEN")),
			Timeout:               30 * time.Second,
			AllowPrivateEndpoints: true,
		}
	}
	return &Chain{qr: QRRecognizer{}, ocr: ocrRecognizer, fallbackThreshold: 0.82}
}

func New(configuration config.InvoiceRecognition) *Chain {
	configuration.Normalize()
	chain := &Chain{qr: QRRecognizer{}, fallbackThreshold: configuration.FallbackThreshold}
	if configuration.Baidu.Enabled {
		chain.ocr = NewBaiduClient(
			configuration.Baidu.APIKey,
			configuration.Baidu.SecretKey,
			time.Duration(configuration.Baidu.TimeoutSeconds)*time.Second,
		)
	}
	if configuration.PublicOCR.Enabled && configuration.PublicOCR.Endpoint != "" {
		recognizer, err := newPublicOCRRecognizer(configuration.PublicOCR, configuration.AllowPrivateEndpoints)
		if err != nil {
			// Invalid runtime configuration is rejected on save. Keeping the
			// worker chain available here still allows QR/model fallback.
			recognizer = nil
		}
		if recognizer != nil {
			if chain.ocr == nil {
				chain.ocr = recognizer
			} else {
				chain.additionalOCRs = append(chain.additionalOCRs, recognizer)
			}
		}
	}
	if configuration.Multimodal.Enabled && configuration.Multimodal.BaseURL != "" {
		chain.multimodal = &MultimodalRecognizer{
			BaseURL:               configuration.Multimodal.BaseURL,
			APIKey:                configuration.Multimodal.APIKey,
			Model:                 configuration.Multimodal.Model,
			Protocol:              configuration.Multimodal.Protocol,
			Timeout:               time.Duration(configuration.Multimodal.TimeoutSeconds) * time.Second,
			AllowPrivateEndpoints: configuration.AllowPrivateEndpoints,
		}
	}
	return chain
}

// NewMultimodal creates the model-only recognizer used for an explicit human
// recheck. Unlike the automatic chain, it never short-circuits on QR or OCR.
func NewMultimodal(configuration config.InvoiceRecognition) (*MultimodalRecognizer, error) {
	configuration.Normalize()
	configuration.PublicOCR.Enabled = false
	if err := configuration.Validate(); err != nil {
		return nil, err
	}
	if !configuration.Multimodal.Enabled {
		return nil, errors.New("请先在运行配置中启用多模态模型")
	}
	return &MultimodalRecognizer{
		BaseURL:               configuration.Multimodal.BaseURL,
		APIKey:                configuration.Multimodal.APIKey,
		Model:                 configuration.Multimodal.Model,
		Protocol:              configuration.Multimodal.Protocol,
		Timeout:               time.Duration(configuration.Multimodal.TimeoutSeconds) * time.Second,
		AllowPrivateEndpoints: configuration.AllowPrivateEndpoints,
	}, nil
}

func (c *Chain) Recognize(ctx context.Context, input Input) (Result, error) {
	qrResult, qrErr := c.qr.Recognize(ctx, input)

	var ocrResult Result
	var ocrErrors []error
	ocrRecognizers := append([]Recognizer{}, c.ocr)
	ocrRecognizers = append(ocrRecognizers, c.additionalOCRs...)
	for _, recognizer := range ocrRecognizers {
		if recognizer == nil {
			continue
		}
		candidate, candidateErr := recognizer.Recognize(ctx, input)
		if candidateErr == nil && !validConfidence(candidate.Confidence) {
			candidateErr = errors.New("公网 OCR 返回的识别置信度不正确")
		}
		if candidateErr == nil && !hasRecognitionData(candidate) {
			candidateErr = errors.New("公网 OCR 未返回可用发票字段")
		}
		if candidateErr != nil {
			ocrErrors = append(ocrErrors, candidateErr)
			continue
		}
		if !hasRecognitionData(ocrResult) || candidate.Confidence > ocrResult.Confidence {
			ocrResult = candidate
		}
		if candidate.Confidence >= c.fallbackThreshold {
			mergeVerifiedQR(&candidate, qrResult, qrErr)
			return candidate, nil
		}
	}
	ocrErr := errors.Join(ocrErrors...)

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
	if target.VerificationType == "" {
		target.VerificationType = fallback.VerificationType
	}
	if target.VerificationAmountMode == "" {
		target.VerificationAmountMode = fallback.VerificationAmountMode
	}
	if target.CheckCode == "" {
		target.CheckCode = fallback.CheckCode
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

func newPublicOCRRecognizer(configuration config.InvoicePublicOCR, allowPrivateEndpoints bool) (*HTTPRecognizer, error) {
	if configuration.Protocol != config.OCRProtocolMultipartJSONV1 {
		return nil, errors.New("不支持的公网 OCR 接口协议")
	}
	return &HTTPRecognizer{
		Endpoint:              configuration.Endpoint,
		Token:                 configuration.APIKey,
		Timeout:               time.Duration(configuration.TimeoutSeconds) * time.Second,
		Provider:              configuration.Provider,
		AllowPrivateEndpoints: allowPrivateEndpoints,
	}, nil
}

func TestConnection(ctx context.Context, target string, configuration config.InvoiceRecognition) (ConnectionInfo, error) {
	configuration.Normalize()
	switch target {
	case "baidu":
		configuration.PublicOCR.Enabled = false
		configuration.Verification.Enabled = false
		configuration.Multimodal.Enabled = false
		if err := configuration.Validate(); err != nil {
			return ConnectionInfo{}, err
		}
		if !configuration.Baidu.Enabled {
			return ConnectionInfo{}, errors.New("请先启用百度发票 OCR")
		}
		err := NewBaiduClient(
			configuration.Baidu.APIKey,
			configuration.Baidu.SecretKey,
			time.Duration(configuration.Baidu.TimeoutSeconds)*time.Second,
		).Probe(ctx)
		return ConnectionInfo{Provider: config.OCRProviderBaidu, Protocol: config.OCRProtocolBaiduVATV1}, err
	case "public-ocr":
		configuration.Baidu.Enabled = false
		configuration.Baidu.VerificationEnabled = false
		configuration.Verification.Enabled = false
		configuration.Multimodal.Enabled = false
		configuration.PublicOCR.Provider = config.OCRProviderHTTPCompatible
		configuration.PublicOCR.Protocol = config.OCRProtocolMultipartJSONV1
		if err := configuration.Validate(); err != nil {
			return ConnectionInfo{}, err
		}
		if !configuration.PublicOCR.Enabled {
			return ConnectionInfo{}, errors.New("请先启用公网 OCR")
		}
		return (&HTTPRecognizer{
			Endpoint:              configuration.PublicOCR.Endpoint,
			Token:                 configuration.PublicOCR.APIKey,
			Timeout:               time.Duration(configuration.PublicOCR.TimeoutSeconds) * time.Second,
			Provider:              config.OCRProviderHTTPCompatible,
			AllowPrivateEndpoints: configuration.AllowPrivateEndpoints,
		}).Detect(ctx)
	case "verification":
		configuration.Baidu.Enabled = false
		configuration.Baidu.VerificationEnabled = false
		configuration.PublicOCR.Enabled = false
		configuration.Multimodal.Enabled = false
		return DetectVerificationAdapter(ctx, configuration)
	case "multimodal":
		configuration.Baidu.Enabled = false
		configuration.Baidu.VerificationEnabled = false
		configuration.PublicOCR.Enabled = false
		configuration.Verification.Enabled = false
		if err := configuration.Validate(); err != nil {
			return ConnectionInfo{}, err
		}
		if !configuration.Multimodal.Enabled {
			return ConnectionInfo{}, errors.New("请先启用多模态模型")
		}
		protocol, err := (&MultimodalRecognizer{
			BaseURL:               configuration.Multimodal.BaseURL,
			APIKey:                configuration.Multimodal.APIKey,
			Model:                 configuration.Multimodal.Model,
			Protocol:              configuration.Multimodal.Protocol,
			Timeout:               time.Duration(configuration.Multimodal.TimeoutSeconds) * time.Second,
			AllowPrivateEndpoints: configuration.AllowPrivateEndpoints,
		}).Probe(ctx)
		return ConnectionInfo{Provider: "multimodal", Protocol: protocol}, err
	default:
		return ConnectionInfo{}, errors.New("不支持的识别服务类型")
	}
}
