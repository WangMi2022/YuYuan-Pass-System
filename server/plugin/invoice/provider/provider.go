package provider

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

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
	qr   Recognizer
	http Recognizer
}

func NewFromEnvironment() *Chain {
	endpoint := strings.TrimSpace(os.Getenv("INVOICE_OCR_ENDPOINT"))
	var httpRecognizer Recognizer
	if endpoint != "" {
		httpRecognizer = &HTTPRecognizer{
			Endpoint: endpoint,
			Token:    strings.TrimSpace(os.Getenv("INVOICE_OCR_TOKEN")),
			Timeout:  30 * time.Second,
		}
	}
	return &Chain{qr: QRRecognizer{}, http: httpRecognizer}
}

func (c *Chain) Recognize(ctx context.Context, input Input) (Result, error) {
	qrResult, qrErr := c.qr.Recognize(ctx, input)
	if c.http == nil {
		if qrErr == nil {
			return qrResult, nil
		}
		return Result{Provider: "manual", RawText: "", Confidence: 0}, nil
	}

	httpResult, httpErr := c.http.Recognize(ctx, input)
	if httpErr == nil {
		if qrErr == nil {
			mergeMissing(&httpResult, qrResult)
		}
		return httpResult, nil
	}
	if qrErr == nil {
		qrResult.RawPayload = httpErr.Error()
		return qrResult, nil
	}
	return Result{}, errors.Join(qrErr, httpErr)
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
	if target.TotalCents == 0 {
		target.TotalCents = fallback.TotalCents
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
