package service

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"testing"
)

func multipartFileHeader(t *testing.T, fileName string, data []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = part.Write(data); err != nil {
		t.Fatal(err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(http.MethodPost, "/invoice/upload", &body)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if err = request.ParseMultipartForm(int64(len(data) + 1024)); err != nil {
		t.Fatal(err)
	}
	return request.MultipartForm.File["file"][0]
}

func TestInspectInvoiceFileAcceptsDetectedPDF(t *testing.T) {
	header := multipartFileHeader(t, "invoice.jpg", []byte("%PDF-1.4\nexample"))
	hash, mimeType, extension, err := inspectInvoiceFile(header)
	if err != nil {
		t.Fatal(err)
	}
	if len(hash) != 64 || mimeType != invoicePDFContentType || extension != ".pdf" {
		t.Fatalf("unexpected PDF inspection: hash=%q mime=%q ext=%q", hash, mimeType, extension)
	}
}

func TestInspectInvoiceFileRejectsUnsupportedContent(t *testing.T) {
	header := multipartFileHeader(t, "invoice.pdf", []byte("not a pdf"))
	if _, _, _, err := inspectInvoiceFile(header); err == nil {
		t.Fatal("unsupported content was accepted by file name")
	}
}
