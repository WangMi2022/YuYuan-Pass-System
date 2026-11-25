package provider

import (
	"bytes"
	"context"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

type QRRecognizer struct{}

func (QRRecognizer) Recognize(_ context.Context, input Input) (Result, error) {
	img, _, err := image.Decode(bytes.NewReader(input.Data))
	if err != nil {
		return Result{}, errors.New("发票图片解码失败")
	}
	bitmap, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return Result{}, errors.New("二维码图像处理失败")
	}
	decoded, err := qrcode.NewQRCodeReader().Decode(bitmap, nil)
	if err != nil {
		return Result{}, errors.New("未识别到发票二维码")
	}
	text := strings.TrimSpace(decoded.GetText())
	result := Result{
		Provider: "qrcode", RawText: text, RawPayload: text, Confidence: 0.76,
		FieldConfidences: map[string]float64{},
	}
	parts := strings.Split(text, ",")
	if len(parts) < 6 {
		return result, nil
	}
	result.InvoiceType = invoiceTypeName(strings.TrimSpace(parts[1]))
	result.InvoiceCode = strings.TrimSpace(parts[2])
	result.InvoiceNumber = strings.TrimSpace(parts[3])
	result.FieldConfidences["invoiceType"] = 0.98
	result.FieldConfidences["invoiceCode"] = 0.98
	result.FieldConfidences["invoiceNumber"] = 0.98
	if amount, parseErr := parseDecimalCents(parts[4]); parseErr == nil {
		result.TotalCents = amount
		result.FieldConfidences["totalCents"] = 0.98
	}
	if date, parseErr := time.Parse("20060102", strings.TrimSpace(parts[5])); parseErr == nil {
		result.IssueDate = &date
		result.FieldConfidences["issueDate"] = 0.98
	}
	return result, nil
}

func parseDecimalCents(value string) (int64, error) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, ".")
	if value == "" || len(parts) > 2 || parts[0] == "" {
		return 0, errors.New("金额格式不正确")
	}
	for _, digit := range parts[0] {
		if digit < '0' || digit > '9' {
			return 0, errors.New("金额格式不正确")
		}
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, errors.New("金额超出范围")
	}
	fraction := int64(0)
	if len(parts) == 2 {
		if len(parts[1]) == 0 || len(parts[1]) > 2 {
			return 0, errors.New("金额格式不正确")
		}
		for _, digit := range parts[1] {
			if digit < '0' || digit > '9' {
				return 0, errors.New("金额格式不正确")
			}
		}
		fraction, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, errors.New("金额格式不正确")
		}
		if len(parts[1]) == 1 {
			fraction *= 10
		}
	}
	if whole > (math.MaxInt64-fraction)/100 {
		return 0, errors.New("金额超出范围")
	}
	return whole*100 + fraction, nil
}

func invoiceTypeName(code string) string {
	switch code {
	case "01":
		return "增值税专用发票"
	case "04":
		return "增值税普通发票"
	case "10":
		return "增值税电子普通发票"
	case "11":
		return "增值税普通发票（卷票）"
	case "14":
		return "增值税电子专用发票"
	case "20":
		return "全电发票"
	default:
		return "电子发票"
	}
}
