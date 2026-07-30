package request

import "github.com/flipped-aurora/gin-vue-admin/server/config"

type ProviderConnectionTest struct {
	Target string                    `json:"target" binding:"required,oneof=baidu public-ocr multimodal"`
	Config config.InvoiceRecognition `json:"config" binding:"required"`
}
