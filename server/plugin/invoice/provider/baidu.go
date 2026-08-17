package provider

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model"
)

const (
	baiduTokenURL        = "https://aip.baidubce.com/oauth/2.0/token"
	baiduVATInvoiceURL   = "https://aip.baidubce.com/rest/2.0/ocr/v1/vat_invoice"
	baiduVerificationURL = "https://aip.baidubce.com/rest/2.0/ocr/v1/vat_invoice_verification"
	maxBaiduRequestSize  = 8 << 20
)

type baiduTokenState struct {
	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

var baiduTokens sync.Map

type BaiduClient struct {
	APIKey          string
	SecretKey       string
	Timeout         time.Duration
	TokenURL        string
	OCRURL          string
	VerificationURL string
	HTTPClient      *http.Client
}

func NewBaiduClient(apiKey, secretKey string, timeout time.Duration) *BaiduClient {
	return &BaiduClient{APIKey: apiKey, SecretKey: secretKey, Timeout: timeout}
}

func (c *BaiduClient) client() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &http.Client{
		Timeout:       timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

func (c *BaiduClient) endpoint(configured, fallback string) string {
	if value := strings.TrimSpace(configured); value != "" {
		return value
	}
	return fallback
}

func (c *BaiduClient) tokenState() *baiduTokenState {
	sum := sha256.Sum256([]byte(strings.TrimSpace(c.APIKey) + "\x00" + strings.TrimSpace(c.SecretKey) + "\x00" + c.endpoint(c.TokenURL, baiduTokenURL)))
	key := hex.EncodeToString(sum[:])
	state, _ := baiduTokens.LoadOrStore(key, &baiduTokenState{})
	return state.(*baiduTokenState)
}

func (c *BaiduClient) accessToken(ctx context.Context, force bool) (string, error) {
	if strings.TrimSpace(c.APIKey) == "" || strings.TrimSpace(c.SecretKey) == "" {
		return "", errors.New("百度智能云 API Key 或 Secret Key 未配置")
	}
	state := c.tokenState()
	state.mu.Lock()
	defer state.mu.Unlock()
	if !force && state.token != "" && time.Now().Add(time.Minute).Before(state.expiresAt) {
		return state.token, nil
	}
	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {strings.TrimSpace(c.APIKey)},
		"client_secret": {strings.TrimSpace(c.SecretKey)},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint(c.TokenURL, baiduTokenURL), strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.client().Do(req)
	if err != nil {
		return "", fmt.Errorf("百度智能云认证失败: %w", err)
	}
	defer resp.Body.Close()
	payload, err := readBaiduResponse(resp)
	if err != nil {
		return "", err
	}
	var result struct {
		AccessToken      string `json:"access_token"`
		ExpiresIn        int64  `json:"expires_in"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err = json.Unmarshal(payload, &result); err != nil {
		return "", fmt.Errorf("百度智能云认证响应解析失败: %w", err)
	}
	if result.AccessToken == "" {
		message := strings.TrimSpace(result.ErrorDescription)
		if message == "" {
			message = strings.TrimSpace(result.Error)
		}
		if message == "" {
			message = "未返回 Access Token"
		}
		return "", errors.New("百度智能云认证失败: " + message)
	}
	if result.ExpiresIn <= 0 {
		result.ExpiresIn = 30 * 24 * 60 * 60
	}
	state.token = result.AccessToken
	state.expiresAt = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	return state.token, nil
}

func readBaiduResponse(resp *http.Response) ([]byte, error) {
	payload, err := io.ReadAll(io.LimitReader(resp.Body, maxOCRResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("读取百度智能云响应失败: %w", err)
	}
	if len(payload) > maxOCRResponseSize {
		return nil, errors.New("百度智能云响应超过 2MB 限制")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("百度智能云返回状态 %d", resp.StatusCode)
	}
	return payload, nil
}

type baiduEnvelope struct {
	ErrorCode   int64                      `json:"error_code"`
	ErrorMsg    string                     `json:"error_msg"`
	LogID       json.RawMessage            `json:"log_id"`
	WordsResult map[string]json.RawMessage `json:"words_result"`
}

func (c *BaiduClient) postForm(ctx context.Context, endpoint string, form url.Values) ([]byte, baiduEnvelope, error) {
	encoded := form.Encode()
	if len(encoded) > maxBaiduRequestSize {
		return nil, baiduEnvelope{}, errors.New("发票图片编码后超过百度 OCR 8MB 限制")
	}
	for attempt := 0; attempt < 2; attempt++ {
		token, err := c.accessToken(ctx, attempt > 0)
		if err != nil {
			return nil, baiduEnvelope{}, err
		}
		parsed, err := url.Parse(endpoint)
		if err != nil {
			return nil, baiduEnvelope{}, err
		}
		query := parsed.Query()
		query.Set("access_token", token)
		parsed.RawQuery = query.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, parsed.String(), strings.NewReader(encoded))
		if err != nil {
			return nil, baiduEnvelope{}, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := c.client().Do(req)
		if err != nil {
			return nil, baiduEnvelope{}, fmt.Errorf("调用百度智能云失败: %w", err)
		}
		payload, readErr := readBaiduResponse(resp)
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, baiduEnvelope{}, readErr
		}
		var envelope baiduEnvelope
		if err = json.Unmarshal(payload, &envelope); err != nil {
			return nil, baiduEnvelope{}, fmt.Errorf("百度智能云响应解析失败: %w", err)
		}
		if (envelope.ErrorCode == 110 || envelope.ErrorCode == 111) && attempt == 0 {
			continue
		}
		if envelope.ErrorCode != 0 {
			return nil, envelope, fmt.Errorf("百度智能云错误 %d: %s", envelope.ErrorCode, strings.TrimSpace(envelope.ErrorMsg))
		}
		return payload, envelope, nil
	}
	return nil, baiduEnvelope{}, errors.New("百度智能云 Access Token 刷新失败")
}

func (c *BaiduClient) Probe(ctx context.Context) error {
	_, err := c.accessToken(ctx, true)
	return err
}

func (c *BaiduClient) Recognize(ctx context.Context, input Input) (Result, error) {
	if len(input.Data) == 0 {
		return Result{}, errors.New("发票图片为空")
	}
	payload, envelope, err := c.postForm(ctx, c.endpoint(c.OCRURL, baiduVATInvoiceURL), url.Values{
		"image": {base64.StdEncoding.EncodeToString(input.Data)},
	})
	if err != nil {
		return Result{}, err
	}
	result := Result{
		Provider: "baidu-vat-invoice", RawPayload: string(payload),
		FieldConfidences: map[string]float64{},
	}
	result.InvoiceType = firstBaiduWord(envelope.WordsResult, "InvoiceTypeOrg", "InvoiceType")
	result.VerificationType = NormalizeInvoiceType(result.InvoiceType)
	if result.VerificationType == "motor_vehicle_invoice" {
		result.VerificationAmountMode = baiduMotorVehicleAmountMode(result.InvoiceType)
	}
	result.InvoiceCode = firstBaiduWord(envelope.WordsResult, "InvoiceCode")
	result.InvoiceNumber = firstBaiduWord(envelope.WordsResult, "InvoiceNum", "InvoiceNumber")
	result.CheckCode = firstBaiduWord(envelope.WordsResult, "CheckCode")
	result.BuyerName = firstBaiduWord(envelope.WordsResult, "PurchaserName", "BuyerName")
	result.BuyerTaxNo = firstBaiduWord(envelope.WordsResult, "PurchaserRegisterNum", "BuyerRegisterNum")
	result.SellerName = firstBaiduWord(envelope.WordsResult, "SellerName")
	result.SellerTaxNo = firstBaiduWord(envelope.WordsResult, "SellerRegisterNum", "SellerTaxNo")
	if date := firstBaiduWord(envelope.WordsResult, "InvoiceDate"); date != "" {
		result.IssueDate = parseBaiduDate(date)
	}
	result.AmountCents = parseBaiduCents(firstBaiduWord(envelope.WordsResult, "TotalAmount"))
	result.TaxCents = parseBaiduCents(firstBaiduWord(envelope.WordsResult, "TotalTax"))
	result.TotalCents = parseBaiduCents(firstBaiduWord(envelope.WordsResult, "AmountInFiguers", "AmountInFigures"))
	if result.TotalCents == 0 && result.AmountCents+result.TaxCents > 0 {
		result.TotalCents = result.AmountCents + result.TaxCents
	}
	result.Items = parseBaiduItems(envelope.WordsResult)
	result.RawText = baiduRawText(envelope.WordsResult)
	for field, value := range map[string]string{
		"invoiceType": result.InvoiceType, "verificationType": result.VerificationType,
		"verificationAmountMode": result.VerificationAmountMode,
		"invoiceCode":            result.InvoiceCode, "invoiceNumber": result.InvoiceNumber,
		"checkCode": result.CheckCode, "buyerName": result.BuyerName,
		"buyerTaxNo": result.BuyerTaxNo, "sellerName": result.SellerName,
		"sellerTaxNo": result.SellerTaxNo,
	} {
		if value != "" {
			result.FieldConfidences[field] = 0.9
		}
	}
	if result.IssueDate != nil {
		result.FieldConfidences["issueDate"] = 0.9
	}
	for field, value := range map[string]int64{"amountCents": result.AmountCents, "taxCents": result.TaxCents, "totalCents": result.TotalCents} {
		if value != 0 {
			result.FieldConfidences[field] = 0.9
		}
	}
	result.Confidence = 0.78
	if result.InvoiceNumber != "" && result.IssueDate != nil && result.TotalCents > 0 {
		result.Confidence = 0.9
	}
	return result, nil
}

func baiduMotorVehicleAmountMode(invoiceType string) string {
	if strings.Contains(invoiceType, "纸质机动车") || !strings.Contains(invoiceType, "电子发票") {
		return model.VerificationAmountModeAmount
	}
	return model.VerificationAmountModeTotal
}

func (c *BaiduClient) Verify(ctx context.Context, request VerificationRequest) (VerificationResult, error) {
	form := url.Values{
		"invoice_code": {request.InvoiceCode}, "invoice_num": {request.InvoiceNum},
		"invoice_date": {request.InvoiceDate}, "invoice_type": {request.InvoiceType},
		"check_code": {request.CheckCode}, "total_amount": {request.TotalAmount},
	}
	payload, envelope, err := c.postForm(ctx, c.endpoint(c.VerificationURL, baiduVerificationURL), form)
	if err != nil {
		return VerificationResult{}, err
	}
	official := map[string]string{
		"invoiceType":   firstBaiduWord(envelope.WordsResult, "InvoiceType"),
		"invoiceCode":   firstBaiduWord(envelope.WordsResult, "InvoiceCode"),
		"invoiceNumber": firstBaiduWord(envelope.WordsResult, "InvoiceNum"),
		"checkCode":     firstBaiduWord(envelope.WordsResult, "CheckCode"),
		"issueDate":     firstBaiduWord(envelope.WordsResult, "InvoiceDate"),
		"buyerName":     firstBaiduWord(envelope.WordsResult, "PurchaserName"),
		"buyerTaxNo":    firstBaiduWord(envelope.WordsResult, "PurchaserRegisterNum"),
		"sellerName":    firstBaiduWord(envelope.WordsResult, "SellerName"),
		"sellerTaxNo":   firstBaiduWord(envelope.WordsResult, "SellerRegisterNum"),
		"amountCents":   strconv.FormatInt(parseBaiduCents(firstBaiduWord(envelope.WordsResult, "TotalAmount")), 10),
		"taxCents":      strconv.FormatInt(parseBaiduCents(firstBaiduWord(envelope.WordsResult, "TotalTax")), 10),
		"totalCents":    strconv.FormatInt(parseBaiduCents(firstBaiduWord(envelope.WordsResult, "AmountInFiguers", "AmountInFigures")), 10),
	}
	for key, value := range official {
		if value == "" || ((key == "amountCents" || key == "taxCents" || key == "totalCents") && value == "0") {
			delete(official, key)
		}
	}
	return VerificationResult{
		Outcome:         baiduVerificationOutcome(firstBaiduWord(envelope.WordsResult, "VerifyResult"), firstBaiduWord(envelope.WordsResult, "InvalidSign")),
		VerifyResult:    firstBaiduWord(envelope.WordsResult, "VerifyResult"),
		VerifyMessage:   firstBaiduWord(envelope.WordsResult, "VerifyMessage"),
		VerifyFrequency: firstBaiduWord(envelope.WordsResult, "VerifyFrequency"),
		InvalidSign:     firstBaiduWord(envelope.WordsResult, "InvalidSign"),
		ProviderLogID:   rawScalar(envelope.LogID), Official: official, RawPayload: string(payload),
	}, nil
}

func firstBaiduWord(fields map[string]json.RawMessage, keys ...string) string {
	for _, key := range keys {
		if value := baiduWord(fields[key]); value != "" {
			return value
		}
	}
	return ""
}

func baiduWord(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var value string
	if json.Unmarshal(raw, &value) == nil {
		return strings.TrimSpace(value)
	}
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) == nil {
		for _, key := range []string{"words", "word", "value"} {
			if candidate := rawScalar(object[key]); candidate != "" {
				return candidate
			}
		}
	}
	values := baiduWords(raw)
	if len(values) > 0 {
		return strings.Join(values, " ")
	}
	return rawScalar(raw)
}

func rawScalar(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return ""
	}
}

func baiduWords(raw json.RawMessage) []string {
	var rows []json.RawMessage
	if json.Unmarshal(raw, &rows) != nil {
		return nil
	}
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		if value := baiduWord(row); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func parseBaiduDate(value string) *time.Time {
	normalized := strings.NewReplacer("年", "-", "月", "-", "日", "", "/", "-").Replace(strings.TrimSpace(value))
	for _, layout := range []string{"2006-01-02", "20060102", "2006-1-2"} {
		if parsed, err := time.ParseInLocation(layout, normalized, time.Local); err == nil {
			return &parsed
		}
	}
	return nil
}

func parseBaiduCents(value string) int64 {
	normalized := strings.NewReplacer(",", "", "￥", "", "¥", "", "元", "", " ", "").Replace(strings.TrimSpace(value))
	amount, err := strconv.ParseFloat(normalized, 64)
	if err != nil || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0
	}
	return int64(math.Round(amount * 100))
}

func parseBaiduItems(fields map[string]json.RawMessage) []model.InvoiceItem {
	columns := [][]string{
		baiduWords(fields["CommodityName"]), baiduWords(fields["CommodityType"]),
		baiduWords(fields["CommodityUnit"]), baiduWords(fields["CommodityNum"]),
		baiduWords(fields["CommodityPrice"]), baiduWords(fields["CommodityAmount"]),
		baiduWords(fields["CommodityTaxRate"]), baiduWords(fields["CommodityTax"]),
	}
	count := 0
	for _, column := range columns {
		if len(column) > count {
			count = len(column)
		}
	}
	items := make([]model.InvoiceItem, 0, count)
	for index := 0; index < count; index++ {
		item := model.InvoiceItem{
			Name: itemColumn(columns[0], index), Specification: itemColumn(columns[1], index),
			Unit: itemColumn(columns[2], index), QuantityText: itemColumn(columns[3], index),
			UnitPriceCents: parseBaiduCents(itemColumn(columns[4], index)),
			AmountCents:    parseBaiduCents(itemColumn(columns[5], index)),
			TaxRate:        itemColumn(columns[6], index), TaxCents: parseBaiduCents(itemColumn(columns[7], index)),
		}
		if item.Name != "" || item.AmountCents != 0 || item.TaxCents != 0 {
			items = append(items, item)
		}
	}
	return items
}

func itemColumn(column []string, index int) string {
	if index < len(column) {
		return column[index]
	}
	return ""
}

func baiduRawText(fields map[string]json.RawMessage) string {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		if value := baiduWord(fields[key]); value != "" {
			lines = append(lines, key+": "+value)
		}
	}
	return strings.Join(lines, "\n")
}

var canonicalInvoiceTypeAliases = map[string]string{
	"增值税专用发票": "special_vat_invoice", "专用发票": "special_vat_invoice", "special_vat_invoice": "special_vat_invoice",
	"增值税电子专用发票": "elec_special_vat_invoice", "电子专用发票": "elec_special_vat_invoice", "elec_special_vat_invoice": "elec_special_vat_invoice",
	"增值税普通发票": "normal_invoice", "普通发票": "normal_invoice", "normal_invoice": "normal_invoice",
	"增值税普通发票（电子）": "elec_normal_invoice", "增值税电子普通发票": "elec_normal_invoice", "电子普通发票": "elec_normal_invoice", "elec_normal_invoice": "elec_normal_invoice",
	"增值税普通发票（卷式）": "roll_normal_invoice", "roll_normal_invoice": "roll_normal_invoice",
	"通行费增值税电子普通发票": "toll_elec_normal_invoice", "toll_elec_normal_invoice": "toll_elec_normal_invoice",
	"区块链电子发票": "blockchain_invoice", "区块链发票": "blockchain_invoice", "blockchain_invoice": "blockchain_invoice",
	"全电发票（专用发票）": "elec_invoice_special", "电子发票（专用发票）": "elec_invoice_special", "电子发票(专用发票)": "elec_invoice_special", "elec_invoice_special": "elec_invoice_special",
	"全电发票（普通发票）": "elec_invoice_normal", "电子发票（普通发票）": "elec_invoice_normal", "电子发票(普通发票)": "elec_invoice_normal", "elec_invoice_normal": "elec_invoice_normal",
	"货物运输业增值税专用发票": "special_freight_transport_invoice", "货运运输业增值税专用发票": "special_freight_transport_invoice", "special_freight_transport_invoice": "special_freight_transport_invoice",
	"机动车销售发票": "motor_vehicle_invoice", "机动车销售统一发票": "motor_vehicle_invoice", "电子发票（纸质机动车销售统一发票）": "motor_vehicle_invoice", "电子发票（机动车销售统一发票）": "motor_vehicle_invoice", "电子发票(机动车销售统一发票)": "motor_vehicle_invoice", "motor_vehicle_invoice": "motor_vehicle_invoice",
	"二手车销售发票": "used_vehicle_invoice", "二手车销售统一发票": "used_vehicle_invoice", "电子发票（纸质二手车销售统一发票）": "used_vehicle_invoice", "电子发票（二手车销售统一发票）": "used_vehicle_invoice", "电子发票(二手车销售统一发票)": "used_vehicle_invoice", "used_vehicle_invoice": "used_vehicle_invoice",
	"电子发票（航空运输电子客票行程单）": "elec_flight_itinerary_invoice", "elec_flight_itinerary_invoice": "elec_flight_itinerary_invoice",
	"电子发票（铁路电子客票）": "elec_train_ticket_invoice", "elec_train_ticket_invoice": "elec_train_ticket_invoice",
	"全电发票（含通行费标识）": "elec_toll_invoice", "elec_toll_invoice": "elec_toll_invoice",
}

func NormalizeInvoiceType(value string) string {
	normalized := strings.TrimSpace(value)
	if result := canonicalInvoiceTypeAliases[normalized]; result != "" {
		return result
	}
	normalized = strings.ReplaceAll(normalized, " ", "")
	return canonicalInvoiceTypeAliases[normalized]
}

// NormalizeBaiduInvoiceType remains for source compatibility. New callers use
// the vendor-neutral canonical invoice type vocabulary.
func NormalizeBaiduInvoiceType(value string) string { return NormalizeInvoiceType(value) }
