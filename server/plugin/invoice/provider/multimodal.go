package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model"
)

const (
	maxMultimodalResponseSize = 4 << 20
	probeMaxTokens            = 256
)

const invoiceExtractionPrompt = `识别图片中的中国发票并只返回一个 JSON 对象，不要返回 Markdown。金额全部换算为整数分，日期格式为 YYYY-MM-DD。无法确认的文本字段返回空字符串，无法确认的金额返回 0。JSON 字段必须是：invoiceType、verificationType、verificationAmountMode、invoiceCode、invoiceNumber、checkCode、issueDate、buyerName、buyerTaxNo、sellerName、sellerTaxNo、amountCents、taxCents、totalCents、rawText、confidence、fieldConfidences、items。verificationType 使用系统统一票种代码；机动车销售统一发票的 verificationAmountMode，纸质票返回 amount，电子票返回 total，其他票种返回空字符串。confidence 和 fieldConfidences 的值范围为 0 到 1。items 每项字段必须是：name、specification、unit、quantityText、unitPriceCents、amountCents、taxRate、taxCents。`

type MultimodalRecognizer struct {
	BaseURL               string
	APIKey                string
	Model                 string
	Protocol              string
	Timeout               time.Duration
	AllowPrivateEndpoints bool
}

type multimodalCompletion struct {
	payload  []byte
	content  string
	protocol string
}

type multimodalInvoiceResult struct {
	InvoiceType            string                      `json:"invoiceType"`
	VerificationType       string                      `json:"verificationType"`
	VerificationAmountMode string                      `json:"verificationAmountMode"`
	InvoiceCode            string                      `json:"invoiceCode"`
	InvoiceNumber          string                      `json:"invoiceNumber"`
	CheckCode              string                      `json:"checkCode"`
	IssueDate              string                      `json:"issueDate"`
	BuyerName              string                      `json:"buyerName"`
	BuyerTaxNo             string                      `json:"buyerTaxNo"`
	SellerName             string                      `json:"sellerName"`
	SellerTaxNo            string                      `json:"sellerTaxNo"`
	AmountCents            int64                       `json:"amountCents"`
	TaxCents               int64                       `json:"taxCents"`
	TotalCents             int64                       `json:"totalCents"`
	RawText                string                      `json:"rawText"`
	Confidence             float64                     `json:"confidence"`
	FieldConfidences       map[string]float64          `json:"fieldConfidences"`
	Items                  []multimodalInvoiceLineItem `json:"items"`
}

type multimodalInvoiceLineItem struct {
	Name           string `json:"name"`
	Specification  string `json:"specification"`
	Unit           string `json:"unit"`
	QuantityText   string `json:"quantityText"`
	UnitPriceCents int64  `json:"unitPriceCents"`
	AmountCents    int64  `json:"amountCents"`
	TaxRate        string `json:"taxRate"`
	TaxCents       int64  `json:"taxCents"`
}

func (r *MultimodalRecognizer) Recognize(ctx context.Context, input Input) (Result, error) {
	completion, err := r.complete(ctx, input, invoiceExtractionPrompt, 4000)
	if err != nil {
		return Result{}, err
	}
	var extracted multimodalInvoiceResult
	if err = json.Unmarshal(extractJSONObject(completion.content), &extracted); err != nil {
		return Result{}, fmt.Errorf("多模态识别结果解析失败: %w", err)
	}
	issueDate, err := parseIssueDate(extracted.IssueDate)
	if err != nil {
		return Result{}, err
	}
	items := make([]model.InvoiceItem, 0, len(extracted.Items))
	for _, item := range extracted.Items {
		items = append(items, model.InvoiceItem{
			Name: item.Name, Specification: item.Specification, Unit: item.Unit,
			QuantityText: item.QuantityText, UnitPriceCents: item.UnitPriceCents,
			AmountCents: item.AmountCents, TaxRate: item.TaxRate, TaxCents: item.TaxCents,
		})
	}
	confidence := extracted.Confidence
	if confidence == 0 && len(extracted.FieldConfidences) > 0 {
		for _, value := range extracted.FieldConfidences {
			confidence += value
		}
		confidence /= float64(len(extracted.FieldConfidences))
	}
	return Result{
		Provider: "multimodal-ai", InvoiceType: extracted.InvoiceType,
		VerificationType: extracted.VerificationType, VerificationAmountMode: extracted.VerificationAmountMode,
		InvoiceCode: extracted.InvoiceCode, InvoiceNumber: extracted.InvoiceNumber,
		CheckCode: extracted.CheckCode,
		IssueDate: issueDate, BuyerName: extracted.BuyerName, BuyerTaxNo: extracted.BuyerTaxNo,
		SellerName: extracted.SellerName, SellerTaxNo: extracted.SellerTaxNo,
		AmountCents: extracted.AmountCents, TaxCents: extracted.TaxCents, TotalCents: extracted.TotalCents,
		RawText: extracted.RawText, RawPayload: string(completion.payload), Confidence: confidence,
		FieldConfidences: extracted.FieldConfidences, Items: items,
	}, nil
}

func (r *MultimodalRecognizer) Probe(ctx context.Context) (string, error) {
	completion, err := r.complete(ctx, Input{
		FileName: "connection-test.png", ContentType: "image/png", Data: probePNG,
	}, "请确认你能读取这张测试图片，只回复 OK。", probeMaxTokens)
	if err != nil {
		return "", err
	}
	return completion.protocol, nil
}

func (r *MultimodalRecognizer) complete(
	ctx context.Context,
	input Input,
	prompt string,
	maxTokens int,
) (multimodalCompletion, error) {
	protocols, err := multimodalProtocolCandidates(r.Protocol)
	if err != nil {
		return multimodalCompletion{}, err
	}
	attemptErrors := make([]error, 0, len(protocols))
	for _, protocol := range protocols {
		payload, requestErr := r.chat(ctx, input, prompt, maxTokens, protocol)
		if requestErr != nil {
			wrappedErr := fmt.Errorf("%s: %w", multimodalProtocolName(protocol), requestErr)
			if len(protocols) == 1 || !multimodalProtocolFallbackAllowed(requestErr) {
				return multimodalCompletion{}, wrappedErr
			}
			attemptErrors = append(attemptErrors, wrappedErr)
			continue
		}
		content, contentErr := multimodalResponseContent(payload, protocol)
		if contentErr == nil {
			r.Protocol = protocol
			return multimodalCompletion{payload: payload, content: content, protocol: protocol}, nil
		}
		attemptErrors = append(attemptErrors, fmt.Errorf("%s: %w", multimodalProtocolName(protocol), contentErr))
	}
	if len(protocols) == 1 {
		return multimodalCompletion{}, attemptErrors[0]
	}
	return multimodalCompletion{}, fmt.Errorf(
		"无法自动识别多模态接口协议（已尝试 OpenAI Compatible 和 Anthropic）: %w",
		errors.Join(attemptErrors...),
	)
}

func multimodalProtocolCandidates(protocol string) ([]string, error) {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "":
		return []string{
			config.MultimodalProtocolOpenAICompatible,
			config.MultimodalProtocolAnthropic,
		}, nil
	case config.MultimodalProtocolOpenAICompatible:
		return []string{config.MultimodalProtocolOpenAICompatible}, nil
	case config.MultimodalProtocolAnthropic:
		return []string{config.MultimodalProtocolAnthropic}, nil
	default:
		return nil, errors.New("多模态模型协议不正确，请重新测试连接")
	}
}

func multimodalProtocolName(protocol string) string {
	if protocol == config.MultimodalProtocolAnthropic {
		return "Anthropic"
	}
	return "OpenAI Compatible"
}

func (r *MultimodalRecognizer) chat(
	ctx context.Context,
	input Input,
	prompt string,
	maxTokens int,
	protocol string,
) ([]byte, error) {
	preparedInput, err := prepareMultimodalImage(ctx, input, maxMultimodalInitialImageSize)
	if err != nil {
		return nil, err
	}
	client := providerHTTPClient(r.Timeout, r.AllowPrivateEndpoints)
	response, err := r.chatOnce(ctx, client, preparedInput, prompt, maxTokens, protocol)
	if err != nil {
		return nil, err
	}
	if response.statusCode == http.StatusRequestEntityTooLarge {
		retryInput, prepareErr := prepareMultimodalRetryImage(ctx, preparedInput)
		if prepareErr != nil {
			return nil, fmt.Errorf("模型拒绝了过大的图片，自动压缩失败: %w", prepareErr)
		}
		if bytes.Equal(retryInput.Data, preparedInput.Data) {
			return nil, multimodalStatusError(response.statusCode)
		}
		response, err = r.chatOnce(ctx, client, retryInput, prompt, maxTokens, protocol)
		if err != nil {
			return nil, err
		}
	}
	if response.statusCode < 200 || response.statusCode >= 300 {
		return nil, multimodalStatusError(response.statusCode)
	}
	return response.payload, nil
}

type multimodalHTTPResponse struct {
	payload    []byte
	statusCode int
}

type multimodalHTTPStatusError struct {
	statusCode int
	message    string
}

func (e *multimodalHTTPStatusError) Error() string {
	return e.message
}

func (r *MultimodalRecognizer) chatOnce(
	ctx context.Context,
	client *http.Client,
	input Input,
	prompt string,
	maxTokens int,
	protocol string,
) (multimodalHTTPResponse, error) {
	mimeType := strings.TrimSpace(input.ContentType)
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		mimeType = "image/png"
	}
	encodedImage := base64.StdEncoding.EncodeToString(input.Data)
	var requestBody map[string]any
	var endpoint string
	switch protocol {
	case config.MultimodalProtocolOpenAICompatible:
		endpoint = chatCompletionsURL(r.BaseURL)
		requestBody = map[string]any{
			"model": r.Model,
			"messages": []map[string]any{{
				"role": "user",
				"content": []map[string]any{
					{"type": "text", "text": prompt},
					{"type": "image_url", "image_url": map[string]any{
						"url": "data:" + mimeType + ";base64," + encodedImage, "detail": "high",
					}},
				},
			}},
			"temperature": 0,
			"max_tokens":  maxTokens,
		}
	case config.MultimodalProtocolAnthropic:
		endpoint = anthropicMessagesURL(r.BaseURL)
		requestBody = map[string]any{
			"model": r.Model,
			"messages": []map[string]any{{
				"role": "user",
				"content": []map[string]any{
					{"type": "image", "source": map[string]any{
						"type": "base64", "media_type": mimeType, "data": encodedImage,
					}},
					{"type": "text", "text": prompt},
				},
			}},
			"temperature": 0,
			"max_tokens":  maxTokens,
		}
	default:
		return multimodalHTTPResponse{}, errors.New("多模态模型协议不正确，请重新测试连接")
	}
	encoded, err := json.Marshal(requestBody)
	if err != nil {
		return multimodalHTTPResponse{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return multimodalHTTPResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(r.APIKey) != "" && protocol == config.MultimodalProtocolOpenAICompatible {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(r.APIKey))
	}
	if strings.TrimSpace(r.APIKey) != "" && protocol == config.MultimodalProtocolAnthropic {
		req.Header.Set("x-api-key", strings.TrimSpace(r.APIKey))
	}
	if protocol == config.MultimodalProtocolAnthropic {
		req.Header.Set("anthropic-version", "2023-06-01")
	}
	resp, err := client.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return multimodalHTTPResponse{}, fmt.Errorf("多模态模型请求失败: %w", err)
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, maxMultimodalResponseSize+1))
	if err != nil {
		return multimodalHTTPResponse{}, fmt.Errorf("多模态模型响应读取失败: %w", err)
	}
	if len(payload) > maxMultimodalResponseSize {
		return multimodalHTTPResponse{}, errors.New("多模态模型响应超过 4MB 限制")
	}
	return multimodalHTTPResponse{payload: payload, statusCode: resp.StatusCode}, nil
}

func multimodalStatusError(statusCode int) error {
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		return &multimodalHTTPStatusError{
			statusCode: statusCode,
			message:    fmt.Sprintf("多模态模型鉴权失败（HTTP %d）", statusCode),
		}
	}
	if statusCode == http.StatusRequestEntityTooLarge {
		return &multimodalHTTPStatusError{
			statusCode: statusCode,
			message:    "模型服务允许的图片大小过小，系统压缩后仍被拒绝，请联系管理员调整接口网关限制",
		}
	}
	return &multimodalHTTPStatusError{
		statusCode: statusCode,
		message:    fmt.Sprintf("多模态模型返回状态 %d", statusCode),
	}
}

func multimodalProtocolFallbackAllowed(err error) bool {
	var statusErr *multimodalHTTPStatusError
	if !errors.As(err, &statusErr) {
		return false
	}
	switch statusErr.statusCode {
	case http.StatusBadRequest, http.StatusNotFound, http.StatusMethodNotAllowed,
		http.StatusUnsupportedMediaType, http.StatusUnprocessableEntity:
		return true
	default:
		return false
	}
}

func chatCompletionsURL(baseURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return strings.TrimSpace(baseURL)
	}
	path := strings.TrimRight(parsed.Path, "/")
	if !strings.HasSuffix(path, "/chat/completions") {
		if strings.HasSuffix(path, "/v1") {
			path += "/chat/completions"
		} else {
			path += "/v1/chat/completions"
		}
	}
	parsed.Path = path
	parsed.RawPath = ""
	parsed.Fragment = ""
	return parsed.String()
}

func anthropicMessagesURL(baseURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return strings.TrimSpace(baseURL)
	}
	path := strings.TrimRight(parsed.Path, "/")
	if !strings.HasSuffix(path, "/messages") {
		if strings.HasSuffix(path, "/v1") {
			path += "/messages"
		} else {
			path += "/v1/messages"
		}
	}
	parsed.Path = path
	parsed.RawPath = ""
	parsed.Fragment = ""
	return parsed.String()
}

func multimodalResponseContent(payload []byte, protocol string) (string, error) {
	if protocol == config.MultimodalProtocolAnthropic {
		return anthropicResponseContent(payload)
	}
	return openAIResponseContent(payload)
}

func openAIResponseContent(payload []byte) (string, error) {
	var response struct {
		Choices []struct {
			Message struct {
				Content          json.RawMessage `json:"content"`
				ReasoningContent json.RawMessage `json:"reasoning_content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(payload, &response); err != nil {
		return "", fmt.Errorf("多模态模型响应解析失败: %w", err)
	}
	if len(response.Choices) == 0 {
		return "", errors.New("多模态模型未返回结果")
	}
	var content string
	if err := json.Unmarshal(response.Choices[0].Message.Content, &content); err == nil {
		if strings.TrimSpace(content) == "" {
			return "", emptyOpenAIContentError(response.Choices[0].Message.ReasoningContent)
		}
		return content, nil
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(response.Choices[0].Message.Content, &parts); err != nil {
		return "", errors.New("多模态模型返回内容格式不兼容")
	}
	var builder strings.Builder
	for _, part := range parts {
		if part.Type == "text" || part.Type == "output_text" {
			builder.WriteString(part.Text)
		}
	}
	if strings.TrimSpace(builder.String()) == "" {
		return "", emptyOpenAIContentError(response.Choices[0].Message.ReasoningContent)
	}
	return builder.String(), nil
}

func emptyOpenAIContentError(reasoningContent json.RawMessage) error {
	if len(reasoningContent) > 0 && string(reasoningContent) != "null" && string(reasoningContent) != `""` {
		return errors.New("多模态模型只返回了推理内容，未返回正文")
	}
	return errors.New("多模态模型返回内容为空")
}

func anthropicResponseContent(payload []byte) (string, error) {
	var response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(payload, &response); err != nil {
		return "", fmt.Errorf("多模态模型响应解析失败: %w", err)
	}
	if len(response.Content) == 0 {
		return "", errors.New("多模态模型未返回结果")
	}
	var builder strings.Builder
	for _, part := range response.Content {
		if part.Type == "text" || part.Type == "output_text" {
			builder.WriteString(part.Text)
		}
	}
	if strings.TrimSpace(builder.String()) == "" {
		return "", errors.New("多模态模型返回内容为空")
	}
	return builder.String(), nil
}

func extractJSONObject(content string) []byte {
	content = strings.TrimSpace(content)
	if start := strings.Index(content, "{"); start >= 0 {
		if end := strings.LastIndex(content, "}"); end >= start {
			content = content[start : end+1]
		}
	}
	return []byte(content)
}

func parseIssueDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	for _, layout := range []string{"2006-01-02", "2006/01/02", time.RFC3339} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return &parsed, nil
		}
	}
	return nil, errors.New("多模态模型返回的开票日期格式不正确")
}
