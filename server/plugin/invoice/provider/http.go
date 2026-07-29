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
	"time"
)

const maxOCRResponseSize = 2 << 20

// HTTPRecognizer is the stable seam for a cloud OCR gateway or local PaddleOCR service.
// The endpoint receives multipart field "file" and returns Result-compatible JSON.
type HTTPRecognizer struct {
	Endpoint string
	Token    string
	Timeout  time.Duration
}

func (r *HTTPRecognizer) Recognize(ctx context.Context, input Input) (Result, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", input.FileName)
	if err != nil {
		return Result{}, err
	}
	if _, err = part.Write(input.Data); err != nil {
		return Result{}, err
	}
	if err = writer.Close(); err != nil {
		return Result{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.Endpoint, &body)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if r.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.Token)
	}
	client := &http.Client{Timeout: r.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("OCR服务请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{}, fmt.Errorf("OCR服务返回状态 %d", resp.StatusCode)
	}
	var envelope struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data Result `json:"data"`
	}
	payload, readErr := io.ReadAll(io.LimitReader(resp.Body, maxOCRResponseSize+1))
	if readErr != nil {
		return Result{}, fmt.Errorf("OCR响应读取失败: %w", readErr)
	}
	if len(payload) > maxOCRResponseSize {
		return Result{}, errors.New("OCR响应超过 2MB 限制")
	}
	if err = json.Unmarshal(payload, &envelope); err != nil {
		return Result{}, fmt.Errorf("OCR响应解析失败: %w", err)
	}
	if envelope.Code != 0 {
		return Result{}, errors.New(envelope.Msg)
	}
	if envelope.Data.Provider == "" {
		envelope.Data.Provider = "http-ocr"
	}
	return envelope.Data, nil
}
