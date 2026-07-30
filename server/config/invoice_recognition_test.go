package config

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestInvoiceRecognitionRedactsAPIKeys(t *testing.T) {
	configuration := InvoiceRecognition{
		Baidu:      InvoiceBaiduProvider{APIKey: "baidu-ak", SecretKey: "baidu-sk"},
		PublicOCR:  InvoicePublicOCR{APIKey: "ocr-secret"},
		Multimodal: InvoiceMultimodalProvider{APIKey: "ai-secret"},
	}.Redacted()
	payload, err := json.Marshal(configuration)
	if err != nil {
		t.Fatal(err)
	}
	serialized := string(payload)
	if strings.Contains(serialized, "ocr-secret") || strings.Contains(serialized, "ai-secret") ||
		strings.Contains(serialized, "baidu-ak") || strings.Contains(serialized, "baidu-sk") {
		t.Fatalf("API key leaked in JSON: %s", serialized)
	}
	if !configuration.Baidu.APIKeyConfigured || !configuration.Baidu.SecretKeyConfigured ||
		!configuration.PublicOCR.APIKeyConfigured || !configuration.Multimodal.APIKeyConfigured {
		t.Fatal("configured flags were not set")
	}
}

func TestInvoiceRecognitionMergeSecrets(t *testing.T) {
	current := InvoiceRecognition{
		Baidu:     InvoiceBaiduProvider{APIKey: "old-baidu-ak", SecretKey: "old-baidu-sk"},
		PublicOCR: InvoicePublicOCR{Endpoint: "https://old.example", APIKey: "old-ocr"},
		Multimodal: InvoiceMultimodalProvider{
			BaseURL: "https://old-ai.example/v1", APIKey: "old-ai", Model: "vision-model",
			Protocol: MultimodalProtocolOpenAICompatible,
		},
	}
	request := InvoiceRecognition{
		Baidu:      InvoiceBaiduProvider{APIKeyInput: "new-baidu-ak"},
		PublicOCR:  InvoicePublicOCR{Endpoint: "https://new.example"},
		Multimodal: InvoiceMultimodalProvider{BaseURL: "https://new-ai.example/v1", APIKeyInput: "new-ai"},
	}
	merged := request.MergeSecrets(current, true)
	if merged.Baidu.APIKey != "new-baidu-ak" || merged.Baidu.SecretKey != "old-baidu-sk" ||
		merged.PublicOCR.APIKey != "old-ocr" || merged.Multimodal.APIKey != "new-ai" {
		t.Fatalf("unexpected merged secrets: %#v", merged)
	}
	request.PublicOCR.ClearAPIKey = true
	merged = request.MergeSecrets(current, true)
	if merged.PublicOCR.APIKey != "" {
		t.Fatal("clear API key was ignored")
	}
	blocked := request.MergeSecrets(current, false)
	if blocked.PublicOCR.Endpoint != current.PublicOCR.Endpoint || blocked.Multimodal.BaseURL != current.Multimodal.BaseURL {
		t.Fatal("non-admin changed invoice recognition configuration")
	}
}

func TestInvoiceRecognitionPreservesDetectedProtocolForOlderClients(t *testing.T) {
	current := InvoiceRecognition{Multimodal: InvoiceMultimodalProvider{
		BaseURL: "https://ai.example/v1", Model: "vision-model",
		APIKey: "stored-key", Protocol: MultimodalProtocolAnthropic,
	}}
	request := InvoiceRecognition{Multimodal: InvoiceMultimodalProvider{
		BaseURL: "https://ai.example/v1", Model: "vision-model",
	}}

	merged := request.MergeSecrets(current, true)
	if merged.Multimodal.Protocol != MultimodalProtocolAnthropic {
		t.Fatalf("detected protocol was not preserved: %#v", merged.Multimodal)
	}

	request.Multimodal.Model = "new-model"
	merged = request.MergeSecrets(current, true)
	if merged.Multimodal.Protocol != "" {
		t.Fatalf("protocol should be cleared after connection identity changes: %#v", merged.Multimodal)
	}

	request.Multimodal.Model = current.Multimodal.Model
	request.Multimodal.APIKeyInput = "new-key"
	merged = request.MergeSecrets(current, true)
	if merged.Multimodal.Protocol != "" {
		t.Fatalf("protocol should be cleared after API key changes: %#v", merged.Multimodal)
	}
}

func TestInvoiceRecognitionNormalizesAndValidatesProtocol(t *testing.T) {
	configuration := InvoiceRecognition{Multimodal: InvoiceMultimodalProvider{Protocol: " Anthropic "}}
	configuration.Normalize()
	if configuration.Multimodal.Protocol != MultimodalProtocolAnthropic {
		t.Fatalf("protocol was not normalized: %q", configuration.Multimodal.Protocol)
	}

	configuration.Multimodal.Protocol = "custom"
	if err := configuration.Validate(); err == nil {
		t.Fatal("invalid protocol was accepted")
	}
}
