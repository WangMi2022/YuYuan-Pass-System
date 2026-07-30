package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
)

const maxVerificationResponseSize = 2 << 20

// HTTPVerifier is the stable adapter contract for vendor gateways. Vendor
// credentials and response codes stay behind the gateway; this application
// sends canonical invoice fields and receives a canonical outcome.
type HTTPVerifier struct {
	Endpoint              string
	APIKey                string
	SecretKey             string
	Timeout               time.Duration
	AllowPrivateEndpoints bool
}

type httpVerificationEnvelope struct {
	Code *int               `json:"code"`
	Msg  string             `json:"msg"`
	Data VerificationResult `json:"data"`
}

func (v *HTTPVerifier) Verify(ctx context.Context, request VerificationRequest) (VerificationResult, error) {
	status, payload, err := v.request(ctx, map[string]any{"invoice": request})
	if err != nil {
		return VerificationResult{}, err
	}
	if status < 200 || status >= 300 {
		return VerificationResult{}, verificationHTTPStatusError(status)
	}
	envelope, err := parseHTTPVerificationEnvelope(payload)
	if err != nil {
		return VerificationResult{}, err
	}
	if *envelope.Code != 0 {
		if strings.TrimSpace(envelope.Msg) == "" {
			return VerificationResult{}, fmt.Errorf("权威验真网关返回业务错误（code=%d）", *envelope.Code)
		}
		return VerificationResult{}, errors.New(strings.TrimSpace(envelope.Msg))
	}
	if !validVerificationOutcome(envelope.Data.Outcome) {
		return VerificationResult{}, errors.New("权威验真网关未返回有效的统一查验状态")
	}
	envelope.Data.RawPayload = string(payload)
	return envelope.Data, nil
}

func (v *HTTPVerifier) Detect(ctx context.Context) (ConnectionInfo, error) {
	status, payload, err := v.request(ctx, map[string]any{"probe": true})
	if err != nil {
		return ConnectionInfo{}, err
	}
	if status < 200 || status >= 300 {
		return ConnectionInfo{}, verificationHTTPStatusError(status)
	}
	envelope, err := parseHTTPVerificationEnvelope(payload)
	if err != nil {
		return ConnectionInfo{}, errors.New("权威验真服务可达，但响应协议不兼容")
	}
	if *envelope.Code != 0 {
		if strings.TrimSpace(envelope.Msg) == "" {
			return ConnectionInfo{}, fmt.Errorf("权威验真服务可达，但探测失败（code=%d）", *envelope.Code)
		}
		return ConnectionInfo{}, fmt.Errorf("权威验真服务可达，但探测失败: %s", strings.TrimSpace(envelope.Msg))
	}
	return ConnectionInfo{
		Provider: config.VerificationProviderHTTPCompatible,
		Protocol: config.VerificationProtocolHTTPJSONV1,
	}, nil
}

func parseHTTPVerificationEnvelope(payload []byte) (httpVerificationEnvelope, error) {
	var envelope httpVerificationEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil || envelope.Code == nil {
		return httpVerificationEnvelope{}, errors.New("权威验真响应协议不兼容")
	}
	return envelope, nil
}

func (v *HTTPVerifier) request(ctx context.Context, body any) (int, []byte, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return 0, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSpace(v.Endpoint), bytes.NewReader(encoded))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(v.APIKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(v.APIKey))
	}
	if strings.TrimSpace(v.SecretKey) != "" {
		req.Header.Set("X-API-Secret", strings.TrimSpace(v.SecretKey))
	}
	client := providerHTTPClient(v.Timeout, v.AllowPrivateEndpoints)
	resp, err := client.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return 0, nil, fmt.Errorf("权威验真服务请求失败: %w", err)
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, maxVerificationResponseSize+1))
	if err != nil {
		return 0, nil, fmt.Errorf("权威验真响应读取失败: %w", err)
	}
	if len(payload) > maxVerificationResponseSize {
		return 0, nil, errors.New("权威验真响应超过 2MB 限制")
	}
	return resp.StatusCode, payload, nil
}

func verificationHTTPStatusError(status int) error {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("权威验真服务鉴权失败（HTTP %d）", status)
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return fmt.Errorf("权威验真接口地址不正确（HTTP %d）", status)
	default:
		return fmt.Errorf("权威验真服务返回状态 %d", status)
	}
}
