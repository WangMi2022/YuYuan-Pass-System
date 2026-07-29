package config

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestInvoiceRecognitionRedactsAPIKeys(t *testing.T) {
	configuration := InvoiceRecognition{
		PublicOCR:  InvoicePublicOCR{APIKey: "ocr-secret"},
		Multimodal: InvoiceMultimodalProvider{APIKey: "ai-secret"},
	}.Redacted()
	payload, err := json.Marshal(configuration)
	if err != nil {
		t.Fatal(err)
	}
	serialized := string(payload)
	if strings.Contains(serialized, "ocr-secret") || strings.Contains(serialized, "ai-secret") {
		t.Fatalf("API key leaked in JSON: %s", serialized)
	}
	if !configuration.PublicOCR.APIKeyConfigured || !configuration.Multimodal.APIKeyConfigured {
		t.Fatal("configured flags were not set")
	}
}

func TestInvoiceRecognitionMergeSecrets(t *testing.T) {
	current := InvoiceRecognition{
		PublicOCR:  InvoicePublicOCR{Endpoint: "https://old.example", APIKey: "old-ocr"},
		Multimodal: InvoiceMultimodalProvider{BaseURL: "https://old-ai.example/v1", APIKey: "old-ai"},
	}
	request := InvoiceRecognition{
		PublicOCR:  InvoicePublicOCR{Endpoint: "https://new.example"},
		Multimodal: InvoiceMultimodalProvider{BaseURL: "https://new-ai.example/v1", APIKeyInput: "new-ai"},
	}
	merged := request.MergeSecrets(current, true)
	if merged.PublicOCR.APIKey != "old-ocr" || merged.Multimodal.APIKey != "new-ai" {
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
