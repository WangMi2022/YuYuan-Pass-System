package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/provider"
)

const (
	invoicePDFContentType   = "application/pdf"
	maxInvoicePDFPages      = 10
	maxRenderedPDFPageBytes = 8 << 20
	maxRenderedPDFBytes     = 40 << 20
	maxRenderedPDFPixels    = 8_000_000
	pdfRenderTimeout        = 60 * time.Second
	maxPDFCommandOutput     = 64 << 10
)

var pdfPagesPattern = regexp.MustCompile(`(?m)^Pages:\s*([0-9]+)\s*\r?$`)

type invoicePDFCommandRunner func(context.Context, string, ...string) ([]byte, error)

var runInvoicePDFCommand invoicePDFCommandRunner = executeInvoicePDFCommand

type limitedPDFCommandOutput struct {
	bytes.Buffer
}

func (output *limitedPDFCommandOutput) Write(data []byte) (int, error) {
	remaining := maxPDFCommandOutput - output.Len()
	if remaining > 0 {
		writeLength := len(data)
		if writeLength > remaining {
			writeLength = remaining
		}
		_, _ = output.Buffer.Write(data[:writeLength])
	}
	return len(data), nil
}

type pdfPageRecognition struct {
	Page   int
	Result provider.Result
	Err    error
}

func executeInvoicePDFCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Env = append(os.Environ(), "LC_ALL=C", "LANG=C")
	var output limitedPDFCommandOutput
	command.Stdout = &output
	command.Stderr = &output
	err := command.Run()
	return output.Bytes(), err
}

func parsePDFPageCount(output []byte) (int, error) {
	match := pdfPagesPattern.FindSubmatch(output)
	if len(match) != 2 {
		return 0, errors.New("无法读取 PDF 页数")
	}
	pages, err := strconv.Atoi(string(match[1]))
	if err != nil || pages <= 0 {
		return 0, errors.New("PDF 页数不正确")
	}
	if pages > maxInvoicePDFPages {
		return 0, fmt.Errorf("PDF 最多支持 %d 页，请拆分后上传", maxInvoicePDFPages)
	}
	return pages, nil
}

func pdfCommandFailure(ctx context.Context, operation string, err error) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("%s超时", operation)
	}
	var execError *exec.Error
	if errors.As(err, &execError) || errors.Is(err, exec.ErrNotFound) {
		return errors.New("PDF 识别组件未安装，请联系管理员")
	}
	return fmt.Errorf("%s失败", operation)
}

func readRenderedPDFPage(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New("读取 PDF 转换页面失败")
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.Size() <= 0 || info.Size() > maxRenderedPDFPageBytes {
		return nil, errors.New("PDF 转换页面大小超出限制")
	}
	data, err := io.ReadAll(io.LimitReader(file, maxRenderedPDFPageBytes+1))
	if err != nil || len(data) > maxRenderedPDFPageBytes {
		return nil, errors.New("读取 PDF 转换页面失败")
	}
	configuration, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || configuration.Width <= 0 || configuration.Height <= 0 ||
		int64(configuration.Width)*int64(configuration.Height) > maxRenderedPDFPixels {
		return nil, errors.New("PDF 转换页面尺寸超出限制")
	}
	return data, nil
}

func renderInvoicePDF(ctx context.Context, input provider.Input) ([]provider.Input, error) {
	if input.ContentType != invoicePDFContentType {
		return []provider.Input{input}, nil
	}
	if len(input.Data) == 0 || !bytes.HasPrefix(input.Data, []byte("%PDF-")) {
		return nil, errors.New("PDF 文件格式不正确")
	}
	renderCtx, cancel := context.WithTimeout(ctx, pdfRenderTimeout)
	defer cancel()
	temporaryDirectory, err := os.MkdirTemp("", "invoice-pdf-")
	if err != nil {
		return nil, errors.New("创建 PDF 转换目录失败")
	}
	defer os.RemoveAll(temporaryDirectory)
	inputPath := filepath.Join(temporaryDirectory, "source.pdf")
	if err = os.WriteFile(inputPath, input.Data, 0o600); err != nil {
		return nil, errors.New("写入 PDF 临时文件失败")
	}
	infoOutput, commandErr := runInvoicePDFCommand(renderCtx, "pdfinfo", inputPath)
	if commandErr != nil {
		return nil, pdfCommandFailure(renderCtx, "读取 PDF 信息", commandErr)
	}
	pageCount, err := parsePDFPageCount(infoOutput)
	if err != nil {
		return nil, err
	}
	baseName := strings.TrimSpace(strings.TrimSuffix(filepath.Base(input.FileName), filepath.Ext(input.FileName)))
	if baseName == "" {
		baseName = "invoice"
	}
	pages := make([]provider.Input, 0, pageCount)
	totalBytes := 0
	for page := 1; page <= pageCount; page++ {
		pageText := strconv.Itoa(page)
		outputPrefix := filepath.Join(temporaryDirectory, fmt.Sprintf("page-%02d", page))
		_, commandErr = runInvoicePDFCommand(
			renderCtx,
			"pdftoppm",
			"-f", pageText,
			"-l", pageText,
			"-singlefile",
			"-jpeg",
			"-jpegopt", "quality=90",
			"-scale-to", "2400",
			inputPath,
			outputPrefix,
		)
		if commandErr != nil {
			return nil, pdfCommandFailure(renderCtx, fmt.Sprintf("转换 PDF 第 %d 页", page), commandErr)
		}
		data, readErr := readRenderedPDFPage(outputPrefix + ".jpg")
		if readErr != nil {
			return nil, fmt.Errorf("PDF 第 %d 页处理失败: %w", page, readErr)
		}
		totalBytes += len(data)
		if totalBytes > maxRenderedPDFBytes {
			return nil, errors.New("PDF 转换后的总大小超出限制")
		}
		pages = append(pages, provider.Input{
			FileName:    fmt.Sprintf("%s-page-%02d.jpg", baseName, page),
			ContentType: "image/jpeg",
			Data:        data,
		})
	}
	return pages, nil
}

func normalizedInvoiceIdentity(value string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(value), " ", ""))
}

func mergeMissingRecognitionFields(target *provider.Result, source provider.Result) {
	if target.Provider == "" {
		target.Provider = source.Provider
	}
	if target.InvoiceType == "" {
		target.InvoiceType = source.InvoiceType
	}
	if target.VerificationType == "" {
		target.VerificationType = source.VerificationType
	}
	if target.VerificationAmountMode == "" {
		target.VerificationAmountMode = source.VerificationAmountMode
	}
	if target.InvoiceCode == "" {
		target.InvoiceCode = source.InvoiceCode
	}
	if target.InvoiceNumber == "" {
		target.InvoiceNumber = source.InvoiceNumber
	}
	if target.CheckCode == "" {
		target.CheckCode = source.CheckCode
	}
	if target.IssueDate == nil {
		target.IssueDate = source.IssueDate
	}
	if target.BuyerName == "" {
		target.BuyerName = source.BuyerName
	}
	if target.BuyerTaxNo == "" {
		target.BuyerTaxNo = source.BuyerTaxNo
	}
	if target.SellerName == "" {
		target.SellerName = source.SellerName
	}
	if target.SellerTaxNo == "" {
		target.SellerTaxNo = source.SellerTaxNo
	}
	if target.AmountCents == 0 {
		target.AmountCents = source.AmountCents
	}
	if target.TaxCents == 0 {
		target.TaxCents = source.TaxCents
	}
	if target.TotalCents == 0 {
		target.TotalCents = source.TotalCents
	}
}

func aggregatePDFRecognitionResults(pages []pdfPageRecognition) (provider.Result, error) {
	successful := make([]pdfPageRecognition, 0, len(pages))
	var firstError error
	for _, page := range pages {
		if page.Err != nil {
			if firstError == nil {
				firstError = page.Err
			}
			continue
		}
		successful = append(successful, page)
	}
	if len(successful) == 0 {
		if firstError == nil {
			firstError = errors.New("没有可用页面")
		}
		return provider.Result{}, fmt.Errorf("PDF 未识别到发票信息: %w", firstError)
	}
	primaryIndex := 0
	for index, page := range successful {
		if normalizedInvoiceIdentity(page.Result.InvoiceNumber) != "" {
			primaryIndex = index
			break
		}
	}
	referenceNumber := ""
	referenceCode := ""
	for _, page := range successful {
		number := normalizedInvoiceIdentity(page.Result.InvoiceNumber)
		code := normalizedInvoiceIdentity(page.Result.InvoiceCode)
		if referenceNumber != "" && number != "" && referenceNumber != number {
			return provider.Result{}, errors.New("PDF 中识别到多张不同发票，请拆分后上传")
		}
		if referenceCode != "" && code != "" && referenceCode != code {
			return provider.Result{}, errors.New("PDF 中识别到多张不同发票，请拆分后上传")
		}
		if referenceNumber == "" {
			referenceNumber = number
		}
		if referenceCode == "" {
			referenceCode = code
		}
	}
	result := successful[primaryIndex].Result
	result.Items = nil
	result.RawText = ""
	result.RawPayload = ""
	result.FieldConfidences = map[string]float64{}
	rawTexts := make([]string, 0, len(successful))
	type pagePayload struct {
		Page    int    `json:"page"`
		Payload string `json:"payload"`
	}
	payloads := make([]pagePayload, 0, len(successful))
	confidenceTotal := 0.0
	for _, page := range successful {
		mergeMissingRecognitionFields(&result, page.Result)
		result.Items = append(result.Items, page.Result.Items...)
		for field, confidence := range page.Result.FieldConfidences {
			if current, exists := result.FieldConfidences[field]; !exists || confidence > current {
				result.FieldConfidences[field] = confidence
			}
		}
		if rawText := strings.TrimSpace(page.Result.RawText); rawText != "" {
			rawTexts = append(rawTexts, fmt.Sprintf("第 %d 页\n%s", page.Page, rawText))
		}
		if payload := strings.TrimSpace(page.Result.RawPayload); payload != "" {
			payloads = append(payloads, pagePayload{Page: page.Page, Payload: payload})
		}
		confidenceTotal += page.Result.Confidence
	}
	result.RawText = strings.Join(rawTexts, "\n\n")
	if len(payloads) > 0 {
		if encoded, err := json.Marshal(payloads); err == nil {
			result.RawPayload = string(encoded)
		}
	}
	result.Confidence = confidenceTotal / float64(len(successful))
	if len(pages) > len(successful) {
		result.Confidence *= float64(len(successful)) / float64(len(pages))
	}
	return result, nil
}
