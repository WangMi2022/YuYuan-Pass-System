package provider

import (
	"context"
	"errors"
	"testing"
)

type staticRecognizer struct {
	result Result
	err    error
	calls  int
}

func (r *staticRecognizer) Recognize(context.Context, Input) (Result, error) {
	r.calls++
	return r.result, r.err
}

func TestChainUsesMultimodalForLowConfidenceOCR(t *testing.T) {
	qr := &staticRecognizer{err: errors.New("no qr")}
	ocr := &staticRecognizer{result: Result{Provider: "ocr", InvoiceNumber: "OCR-1", BuyerName: "购买方", Confidence: 0.4}}
	multimodal := &staticRecognizer{result: Result{Provider: "multimodal", InvoiceNumber: "AI-1", SellerName: "示例公司", Confidence: 0.9}}
	chain := Chain{qr: qr, ocr: ocr, multimodal: multimodal, fallbackThreshold: 0.82}
	result, err := chain.Recognize(context.Background(), Input{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Provider != "multimodal" || result.InvoiceNumber != "AI-1" || result.BuyerName != "购买方" || result.SellerName != "示例公司" {
		t.Fatalf("unexpected fallback result: %#v", result)
	}
}

func TestChainMergesSuccessfulOCRWhenAnotherOCRFails(t *testing.T) {
	chain := Chain{
		qr: &staticRecognizer{err: errors.New("no qr")},
		ocr: &staticRecognizer{result: Result{
			Provider: "baidu-vat-invoice", InvoiceNumber: "OCR-1",
			BuyerTaxNo: "BUYER-TAX-NO", SellerTaxNo: "SELLER-TAX-NO", Confidence: 0.6,
		}},
		additionalOCRs: []Recognizer{&staticRecognizer{err: errors.New("secondary OCR unavailable")}},
		multimodal: &staticRecognizer{result: Result{
			Provider: "multimodal-ai", InvoiceNumber: "AI-1", Confidence: 0.95,
		}},
		fallbackThreshold: 0.82,
	}
	result, err := chain.Recognize(context.Background(), Input{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Provider != "multimodal-ai" || result.BuyerTaxNo != "BUYER-TAX-NO" || result.SellerTaxNo != "SELLER-TAX-NO" {
		t.Fatalf("successful OCR fields were not merged: %#v", result)
	}
}

func TestChainReturnsSuccessfulOCRWhenAnotherOCRFailsWithoutMultimodal(t *testing.T) {
	chain := Chain{
		qr: &staticRecognizer{err: errors.New("no qr")},
		ocr: &staticRecognizer{result: Result{
			Provider: "baidu-vat-invoice", InvoiceNumber: "OCR-1", Confidence: 0.6,
		}},
		additionalOCRs:    []Recognizer{&staticRecognizer{err: errors.New("secondary OCR unavailable")}},
		fallbackThreshold: 0.82,
	}
	result, err := chain.Recognize(context.Background(), Input{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Provider != "baidu-vat-invoice" || result.InvoiceNumber != "OCR-1" {
		t.Fatalf("successful OCR result was discarded: %#v", result)
	}
}

func TestChainSkipsMultimodalForHighConfidenceOCR(t *testing.T) {
	multimodal := &staticRecognizer{result: Result{Provider: "multimodal", Confidence: 0.99}}
	chain := Chain{
		qr:         &staticRecognizer{err: errors.New("no qr")},
		ocr:        &staticRecognizer{result: Result{Provider: "ocr", InvoiceNumber: "OCR-1", Confidence: 0.95}},
		multimodal: multimodal, fallbackThreshold: 0.82,
	}
	result, err := chain.Recognize(context.Background(), Input{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Provider != "ocr" || multimodal.calls != 0 {
		t.Fatalf("multimodal should not run: result=%#v calls=%d", result, multimodal.calls)
	}
}

func TestChainKeepsVerifiedQRCodeFields(t *testing.T) {
	chain := Chain{
		qr: &staticRecognizer{result: Result{
			Provider: "qrcode", InvoiceNumber: "QR-1", Confidence: 0.76,
			FieldConfidences: map[string]float64{"invoiceNumber": 0.98},
		}},
		ocr:               &staticRecognizer{result: Result{Provider: "ocr", InvoiceNumber: "OCR-1", Confidence: 0.95}},
		fallbackThreshold: 0.82,
	}
	result, err := chain.Recognize(context.Background(), Input{})
	if err != nil {
		t.Fatal(err)
	}
	if result.InvoiceNumber != "QR-1" {
		t.Fatalf("verified QR field was overwritten: %#v", result)
	}
}

func TestChainReturnsErrorInsteadOfManualProviderWhenUnconfigured(t *testing.T) {
	chain := Chain{qr: &staticRecognizer{err: errors.New("no qr")}, fallbackThreshold: 0.82}
	if result, err := chain.Recognize(context.Background(), Input{}); err == nil || result.Provider == "manual" {
		t.Fatalf("expected explicit unconfigured error, result=%#v err=%v", result, err)
	}
}

func TestChainUsesMultimodalWhenOCRHasOnlyRawText(t *testing.T) {
	chain := Chain{
		qr:  &staticRecognizer{err: errors.New("no qr")},
		ocr: &staticRecognizer{result: Result{Provider: "ocr", RawText: "only text", Confidence: 0.99}},
		multimodal: &staticRecognizer{result: Result{
			Provider: "multimodal", InvoiceNumber: "AI-1", Confidence: 0.8,
		}},
		fallbackThreshold: 0.82,
	}
	result, err := chain.Recognize(context.Background(), Input{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Provider != "multimodal" || result.InvoiceNumber != "AI-1" {
		t.Fatalf("raw text incorrectly bypassed multimodal fallback: %#v", result)
	}
}

func TestChainUsesMultimodalWhenOCRConfidenceIsInvalid(t *testing.T) {
	chain := Chain{
		qr:  &staticRecognizer{err: errors.New("no qr")},
		ocr: &staticRecognizer{result: Result{Provider: "ocr", InvoiceNumber: "OCR-1", Confidence: 2}},
		multimodal: &staticRecognizer{result: Result{
			Provider: "multimodal", InvoiceNumber: "AI-1", Confidence: 0.8,
		}},
		fallbackThreshold: 0.82,
	}
	result, err := chain.Recognize(context.Background(), Input{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Provider != "multimodal" {
		t.Fatalf("invalid OCR confidence bypassed fallback: %#v", result)
	}
}
