package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model"
)

func TestBaiduClientRecognizesAndVerifiesWithCachedToken(t *testing.T) {
	var tokenCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/token":
			tokenCalls.Add(1)
			if err := request.ParseForm(); err != nil || request.Form.Get("client_id") != "api-key" || request.Form.Get("client_secret") != "secret-key" {
				t.Fatalf("unexpected token form: %#v err=%v", request.Form, err)
			}
			_ = json.NewEncoder(response).Encode(map[string]any{"access_token": "cached-token", "expires_in": 3600})
		case "/ocr":
			if request.URL.Query().Get("access_token") != "cached-token" {
				t.Fatal("OCR request did not use access token")
			}
			if err := request.ParseForm(); err != nil {
				t.Fatal(err)
			}
			decoded, err := base64.StdEncoding.DecodeString(request.Form.Get("image"))
			if err != nil || string(decoded) != "invoice-image" {
				t.Fatalf("unexpected OCR image: %q err=%v", string(decoded), err)
			}
			_, _ = response.Write([]byte(`{
				"log_id":123,
				"words_result":{
					"InvoiceTypeOrg":{"words":"增值税专用发票"},
					"InvoiceCode":{"words":"044001"},
					"InvoiceNum":{"words":"12345678"},
					"InvoiceDate":{"words":"2026年07月01日"},
					"SellerName":{"words":"测试销售方"},
					"TotalAmount":{"words":"100.00"},
					"TotalTax":{"words":"6.00"},
					"AmountInFiguers":{"words":"106.00"},
					"CommodityName":[{"word":"技术服务"}],
					"CommodityAmount":[{"word":"100.00"}]
				}
			}`))
		case "/verify":
			if request.URL.Query().Get("access_token") != "cached-token" {
				t.Fatal("verification request did not use access token")
			}
			if err := request.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if request.Form.Get("invoice_type") != "special_vat_invoice" || request.Form.Get("total_amount") != "100.00" {
				t.Fatalf("unexpected verification form: %#v", request.Form)
			}
			_, _ = response.Write([]byte(`{
				"log_id":456,
				"words_result":{
					"VerifyResult":"0001","VerifyMessage":"查验成功发票一致","VerifyFrequency":"2","InvalidSign":"N",
					"InvoiceType":"增值税专用发票","InvoiceCode":"044001","InvoiceNum":"12345678",
					"InvoiceDate":"20260701","SellerName":"测试销售方","TotalAmount":"100.00","TotalTax":"6.00","AmountInFiguers":"106.00"
				}
			}`))
		default:
			response.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := &BaiduClient{
		APIKey: "api-key", SecretKey: "secret-key", Timeout: time.Second,
		TokenURL: server.URL + "/token", OCRURL: server.URL + "/ocr",
		VerificationURL: server.URL + "/verify", HTTPClient: server.Client(),
	}
	recognized, err := client.Recognize(context.Background(), Input{Data: []byte("invoice-image")})
	if err != nil {
		t.Fatal(err)
	}
	if recognized.InvoiceNumber != "12345678" || recognized.VerificationType != "special_vat_invoice" || recognized.TotalCents != 10600 || len(recognized.Items) != 1 {
		t.Fatalf("unexpected OCR result: %#v", recognized)
	}
	verified, err := client.Verify(context.Background(), VerificationRequest{
		InvoiceCode: "044001", InvoiceNum: "12345678", InvoiceDate: "20260701",
		InvoiceType: "special_vat_invoice", TotalAmount: "100.00",
	})
	if err != nil {
		t.Fatal(err)
	}
	if verified.Outcome != VerificationOutcomeValid || verified.VerifyResult != "0001" || verified.InvalidSign != "N" ||
		verified.ProviderLogID != "456" || verified.Official["totalCents"] != "10600" {
		t.Fatalf("unexpected verification result: %#v", verified)
	}
	if tokenCalls.Load() != 1 {
		t.Fatalf("access token was not cached: calls=%d", tokenCalls.Load())
	}
}

func TestBaiduClientDoesNotForwardCredentialsThroughRedirect(t *testing.T) {
	redirectTarget := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("client_secret") != "" || strings.Contains(request.Header.Get("Authorization"), "secret") {
			t.Fatal("credentials were forwarded through redirect")
		}
	}))
	defer redirectTarget.Close()
	redirector := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Redirect(response, request, redirectTarget.URL, http.StatusFound)
	}))
	defer redirector.Close()
	client := NewBaiduClient("api-key", "secret-key", time.Second)
	client.TokenURL = redirector.URL
	if err := client.Probe(context.Background()); err == nil {
		t.Fatal("redirected token response was accepted")
	}
}

func TestBaiduMotorVehicleAmountModeDistinguishesPaperAndElectronicInvoices(t *testing.T) {
	tests := map[string]string{
		"机动车销售统一发票":         model.VerificationAmountModeAmount,
		"电子发票（纸质机动车销售统一发票）": model.VerificationAmountModeAmount,
		"电子发票（机动车销售统一发票）":   model.VerificationAmountModeTotal,
	}
	for invoiceType, expected := range tests {
		if actual := baiduMotorVehicleAmountMode(invoiceType); actual != expected {
			t.Errorf("invoice type %q amount mode = %q, want %q", invoiceType, actual, expected)
		}
	}
}
