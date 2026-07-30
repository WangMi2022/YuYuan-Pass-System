package provider

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
)

type VerificationOutcome string

const (
	VerificationOutcomeValid       VerificationOutcome = "valid"
	VerificationOutcomeVoided      VerificationOutcome = "voided"
	VerificationOutcomeRedFlushed  VerificationOutcome = "red_flushed"
	VerificationOutcomeMismatch    VerificationOutcome = "mismatch"
	VerificationOutcomeNotFound    VerificationOutcome = "not_found"
	VerificationOutcomeDeferred    VerificationOutcome = "deferred"
	VerificationOutcomeExpired     VerificationOutcome = "expired"
	VerificationOutcomeUnavailable VerificationOutcome = "unavailable"
)

type VerificationRequest struct {
	InvoiceCode string `json:"invoiceCode"`
	InvoiceNum  string `json:"invoiceNumber"`
	InvoiceDate string `json:"invoiceDate"`
	InvoiceType string `json:"invoiceType"`
	CheckCode   string `json:"checkCode"`
	TotalAmount string `json:"totalAmount"`
	AmountMode  string `json:"amountMode"`
}

func (r VerificationRequest) Snapshot() map[string]string {
	return map[string]string{
		"invoiceCode": r.InvoiceCode, "invoiceNumber": r.InvoiceNum,
		"invoiceDate": r.InvoiceDate, "verificationType": r.InvoiceType,
		"checkCode": r.CheckCode, "totalAmount": r.TotalAmount, "amountMode": r.AmountMode,
	}
}

type VerificationResult struct {
	Outcome         VerificationOutcome `json:"outcome"`
	VerifyResult    string              `json:"verifyResult"`
	VerifyMessage   string              `json:"verifyMessage"`
	VerifyFrequency string              `json:"verifyFrequency"`
	InvalidSign     string              `json:"invalidSign"`
	ProviderLogID   string              `json:"providerLogId"`
	Official        map[string]string   `json:"official"`
	RawPayload      string              `json:"-"`
}

type Verifier interface {
	Verify(context.Context, VerificationRequest) (VerificationResult, error)
}

type VerificationAdapter struct {
	Provider string
	Protocol string
	Verifier Verifier
}

func NewVerificationAdapter(configuration config.InvoiceRecognition) (VerificationAdapter, error) {
	configuration.Normalize()
	if !configuration.Verification.Enabled {
		return VerificationAdapter{}, errors.New("请先在运行配置中启用权威发票验真")
	}
	if err := configuration.Validate(); err != nil {
		return VerificationAdapter{}, err
	}
	providerConfiguration := configuration.Verification
	timeout := time.Duration(providerConfiguration.TimeoutSeconds) * time.Second
	switch providerConfiguration.Provider {
	case config.VerificationProviderBaidu:
		return VerificationAdapter{
			Provider: providerConfiguration.Provider,
			Protocol: providerConfiguration.Protocol,
			Verifier: NewBaiduClient(providerConfiguration.APIKey, providerConfiguration.SecretKey, timeout),
		}, nil
	case config.VerificationProviderHTTPCompatible:
		return VerificationAdapter{
			Provider: providerConfiguration.Provider,
			Protocol: providerConfiguration.Protocol,
			Verifier: &HTTPVerifier{
				Endpoint:              providerConfiguration.Endpoint,
				APIKey:                providerConfiguration.APIKey,
				SecretKey:             providerConfiguration.SecretKey,
				Timeout:               timeout,
				AllowPrivateEndpoints: configuration.AllowPrivateEndpoints,
			},
		}, nil
	default:
		return VerificationAdapter{}, errors.New("不支持的权威验真供应商")
	}
}

// DetectVerificationAdapter chooses from connection shape, verifies the
// selected external protocol, and returns metadata safe to persist. A blank
// endpoint means Baidu OAuth; a configured endpoint means the stable HTTP
// verification gateway contract. The browser never chooses either value.
func DetectVerificationAdapter(ctx context.Context, configuration config.InvoiceRecognition) (ConnectionInfo, error) {
	configuration.Normalize()
	if !configuration.Verification.Enabled {
		return ConnectionInfo{}, errors.New("请先启用权威发票验真")
	}
	providerConfiguration := configuration.Verification
	if strings.TrimSpace(providerConfiguration.Endpoint) == "" {
		configuration.Verification.Provider = config.VerificationProviderBaidu
		configuration.Verification.Protocol = config.VerificationProtocolBaiduVATV1
		if err := configuration.Validate(); err != nil {
			return ConnectionInfo{}, err
		}
		err := NewBaiduClient(
			providerConfiguration.APIKey,
			providerConfiguration.SecretKey,
			time.Duration(providerConfiguration.TimeoutSeconds)*time.Second,
		).Probe(ctx)
		return ConnectionInfo{
			Provider: config.VerificationProviderBaidu,
			Protocol: config.VerificationProtocolBaiduVATV1,
		}, err
	}
	configuration.Verification.Provider = config.VerificationProviderHTTPCompatible
	configuration.Verification.Protocol = config.VerificationProtocolHTTPJSONV1
	if err := configuration.Validate(); err != nil {
		return ConnectionInfo{}, err
	}
	return (&HTTPVerifier{
		Endpoint:              providerConfiguration.Endpoint,
		APIKey:                providerConfiguration.APIKey,
		SecretKey:             providerConfiguration.SecretKey,
		Timeout:               time.Duration(providerConfiguration.TimeoutSeconds) * time.Second,
		AllowPrivateEndpoints: configuration.AllowPrivateEndpoints,
	}).Detect(ctx)
}

func validVerificationOutcome(outcome VerificationOutcome) bool {
	switch outcome {
	case VerificationOutcomeValid, VerificationOutcomeVoided, VerificationOutcomeRedFlushed,
		VerificationOutcomeMismatch, VerificationOutcomeNotFound, VerificationOutcomeDeferred,
		VerificationOutcomeExpired, VerificationOutcomeUnavailable:
		return true
	default:
		return false
	}
}

func baiduVerificationOutcome(verifyResult, invalidSign string) VerificationOutcome {
	if verifyResult != "0001" {
		switch verifyResult {
		case "0006":
			return VerificationOutcomeMismatch
		case "0009":
			return VerificationOutcomeNotFound
		case "1014":
			return VerificationOutcomeDeferred
		case "1015":
			return VerificationOutcomeExpired
		default:
			return VerificationOutcomeUnavailable
		}
	}
	switch invalidSign {
	case "N":
		return VerificationOutcomeValid
	case "Y":
		return VerificationOutcomeVoided
	case "H", "BH", "QH":
		return VerificationOutcomeRedFlushed
	default:
		return VerificationOutcomeUnavailable
	}
}
