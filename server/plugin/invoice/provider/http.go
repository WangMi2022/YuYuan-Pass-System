package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/config"
)

const maxOCRResponseSize = 2 << 20

// HTTPRecognizer is the stable seam for a cloud OCR gateway or local PaddleOCR service.
// The endpoint receives multipart field "file" and returns Result-compatible JSON.
type HTTPRecognizer struct {
	Endpoint              string
	Token                 string
	Timeout               time.Duration
	Provider              string
	AllowPrivateEndpoints bool
}

func (r *HTTPRecognizer) Recognize(ctx context.Context, input Input) (Result, error) {
	status, payload, err := r.request(ctx, input)
	if err != nil {
		return Result{}, err
	}
	if status < 200 || status >= 300 {
		return Result{}, fmt.Errorf("OCR服务返回状态 %d", status)
	}
	var envelope struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data Result `json:"data"`
	}
	if err = json.Unmarshal(payload, &envelope); err != nil {
		return Result{}, fmt.Errorf("OCR响应解析失败: %w", err)
	}
	if envelope.Code != 0 {
		return Result{}, errors.New(envelope.Msg)
	}
	if envelope.Data.Provider == "" {
		envelope.Data.Provider = strings.TrimSpace(r.Provider)
		if envelope.Data.Provider == "" {
			envelope.Data.Provider = "http-ocr"
		}
	}
	return envelope.Data, nil
}

func (r *HTTPRecognizer) Probe(ctx context.Context) error {
	_, err := r.Detect(ctx)
	return err
}

func (r *HTTPRecognizer) Detect(ctx context.Context) (ConnectionInfo, error) {
	status, payload, err := r.request(ctx, Input{
		FileName: "connection-test.png", ContentType: "image/png", Data: probePNG,
	})
	if err != nil {
		return ConnectionInfo{}, err
	}
	switch {
	case status >= 200 && status < 300:
		var envelope struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
			Data Result `json:"data"`
		}
		if err = json.Unmarshal(payload, &envelope); err != nil {
			return ConnectionInfo{}, errors.New("OCR 服务可达，但响应协议不兼容")
		}
		if envelope.Code != 0 {
			if strings.TrimSpace(envelope.Msg) == "" {
				return ConnectionInfo{}, fmt.Errorf("OCR 服务可达，但业务校验失败（code=%d）", envelope.Code)
			}
			return ConnectionInfo{}, fmt.Errorf("OCR 服务可达，但业务校验失败: %s", envelope.Msg)
		}
		return ConnectionInfo{Provider: config.OCRProviderHTTPCompatible, Protocol: config.OCRProtocolMultipartJSONV1}, nil
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return ConnectionInfo{}, fmt.Errorf("OCR 服务鉴权失败（HTTP %d）", status)
	case status == http.StatusNotFound || status == http.StatusMethodNotAllowed:
		return ConnectionInfo{}, fmt.Errorf("OCR 接口地址不正确（HTTP %d）", status)
	case status == http.StatusBadRequest || status == http.StatusUnsupportedMediaType || status == http.StatusUnprocessableEntity:
		return ConnectionInfo{}, fmt.Errorf("OCR 服务可达，但测试图片或接口协议未通过验证（HTTP %d）", status)
	default:
		return ConnectionInfo{}, fmt.Errorf("OCR 服务连接失败（HTTP %d）", status)
	}
}

func (r *HTTPRecognizer) request(ctx context.Context, input Input) (int, []byte, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", input.FileName)
	if err != nil {
		return 0, nil, err
	}
	if _, err = part.Write(input.Data); err != nil {
		return 0, nil, err
	}
	if err = writer.Close(); err != nil {
		return 0, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.Endpoint, &body)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if r.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.Token)
	}
	client := providerHTTPClient(r.Timeout, r.AllowPrivateEndpoints)
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("OCR服务请求失败: %w", err)
	}
	defer resp.Body.Close()
	payload, readErr := io.ReadAll(io.LimitReader(resp.Body, maxOCRResponseSize+1))
	if readErr != nil {
		return 0, nil, fmt.Errorf("OCR响应读取失败: %w", readErr)
	}
	if len(payload) > maxOCRResponseSize {
		return 0, nil, errors.New("OCR响应超过 2MB 限制")
	}
	return resp.StatusCode, payload, nil
}

var probePNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x44, 0x41,
	0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0xf0,
	0x1f, 0x00, 0x05, 0x00, 0x01, 0xff, 0x89, 0x99,
	0x3d, 0x1d, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45,
	0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}
