package service

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/provider"
)

func renderedJPEG(t *testing.T) []byte {
	t.Helper()
	canvas := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			canvas.Set(x, y, color.RGBA{R: 240, G: 245, B: 250, A: 255})
		}
	}
	var output bytes.Buffer
	if err := jpeg.Encode(&output, canvas, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func TestRenderInvoicePDFProducesBoundedPageInputs(t *testing.T) {
	previous := runInvoicePDFCommand
	jpegData := renderedJPEG(t)
	runInvoicePDFCommand = func(_ context.Context, name string, args ...string) ([]byte, error) {
		if name == "pdfinfo" {
			return []byte("Pages:          2\n"), nil
		}
		outputPrefix := args[len(args)-1]
		return nil, os.WriteFile(outputPrefix+".jpg", jpegData, 0o600)
	}
	t.Cleanup(func() { runInvoicePDFCommand = previous })

	pages, err := renderInvoicePDF(t.Context(), provider.Input{
		FileName: "采购发票.pdf", ContentType: invoicePDFContentType, Data: []byte("%PDF-1.4\nmock"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 || pages[0].ContentType != "image/jpeg" ||
		pages[0].FileName != "采购发票-page-01.jpg" || len(pages[1].Data) == 0 {
		t.Fatalf("unexpected rendered pages: %#v", pages)
	}
}

func TestParsePDFPageCountRejectsOversizedDocument(t *testing.T) {
	if _, err := parsePDFPageCount([]byte("Pages: 11\n")); err == nil || !strings.Contains(err.Error(), "最多支持 10 页") {
		t.Fatalf("unexpected page limit error: %v", err)
	}
}

func TestAggregatePDFRecognitionResultsMergesPages(t *testing.T) {
	result, err := aggregatePDFRecognitionResults([]pdfPageRecognition{
		{Page: 1, Result: provider.Result{
			Provider: "test", InvoiceNumber: "INV-001", SellerName: "供应商",
			Confidence: 0.8, FieldConfidences: map[string]float64{"sellerName": 0.8},
			RawText: "第一页", RawPayload: `{"page":1}`,
			Items: []model.InvoiceItem{{Name: "商品 A"}},
		}},
		{Page: 2, Result: provider.Result{
			Provider: "test", InvoiceNumber: "INV-001", TotalCents: 10600,
			Confidence: 0.6, FieldConfidences: map[string]float64{"sellerName": 0.9},
			RawText: "第二页", RawPayload: `{"page":2}`,
			Items: []model.InvoiceItem{{Name: "商品 B"}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.InvoiceNumber != "INV-001" || result.SellerName != "供应商" || result.TotalCents != 10600 ||
		len(result.Items) != 2 || math.Abs(result.Confidence-0.7) > 1e-9 || result.FieldConfidences["sellerName"] != 0.9 ||
		!strings.Contains(result.RawText, "第 1 页") || !strings.Contains(result.RawText, "第 2 页") || result.RawPayload == "" {
		t.Fatalf("unexpected aggregate result: %#v", result)
	}
}

func TestRecognizeInvoiceEvidencePenalizesFailedPDFPages(t *testing.T) {
	previous := runInvoicePDFCommand
	jpegData := renderedJPEG(t)
	runInvoicePDFCommand = func(_ context.Context, name string, args ...string) ([]byte, error) {
		if name == "pdfinfo" {
			return []byte("Pages: 3\n"), nil
		}
		outputPrefix := args[len(args)-1]
		return nil, os.WriteFile(outputPrefix+".jpg", jpegData, 0o600)
	}
	t.Cleanup(func() { runInvoicePDFCommand = previous })

	result, err := recognizeInvoiceEvidence(t.Context(), recognizerFunc(func(_ context.Context, input provider.Input) (provider.Result, error) {
		switch {
		case strings.Contains(input.FileName, "page-02"):
			return provider.Result{}, errors.New("page recognition failed")
		case strings.Contains(input.FileName, "page-03"):
			return provider.Result{Provider: "test", InvoiceNumber: "INV-001", TotalCents: 10600, Confidence: 0.6}, nil
		default:
			return provider.Result{Provider: "test", InvoiceNumber: "INV-001", SellerName: "供应商", Confidence: 0.9}, nil
		}
	}), provider.Input{
		FileName: "采购发票.pdf", ContentType: invoicePDFContentType, Data: []byte("%PDF-1.4\nmock"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.InvoiceNumber != "INV-001" || result.SellerName != "供应商" || result.TotalCents != 10600 ||
		math.Abs(result.Confidence-0.5) > 1e-9 {
		t.Fatalf("unexpected partial PDF result: %#v", result)
	}
}

func TestAggregatePDFRecognitionResultsRejectsMultipleInvoices(t *testing.T) {
	_, err := aggregatePDFRecognitionResults([]pdfPageRecognition{
		{Page: 1, Result: provider.Result{Provider: "test", InvoiceNumber: "INV-001"}},
		{Page: 2, Result: provider.Result{Provider: "test", InvoiceNumber: "INV-002"}},
	})
	if err == nil || !strings.Contains(err.Error(), "多张不同发票") {
		t.Fatalf("unexpected conflict result: %v", err)
	}
}
