package model

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type InvoiceReviewCorrection struct {
	global.GVA_MODEL
	InvoiceID        uint      `json:"invoiceId" gorm:"not null;index;uniqueIndex:idx_invoice_review_field;comment:发票ID"`
	RecognitionJobID *uint     `json:"recognitionJobId" gorm:"index;comment:识别任务ID"`
	InvocationID     *uint     `json:"invocationId" gorm:"index;comment:AI调用审计ID"`
	FieldName        string    `json:"fieldName" gorm:"size:80;not null;uniqueIndex:idx_invoice_review_field;comment:字段名"`
	RecognizedValue  string    `json:"recognizedValue" gorm:"type:text;not null;comment:模型识别值或脱敏摘要"`
	CorrectedValue   string    `json:"correctedValue" gorm:"type:text;not null;comment:人工修正值或脱敏摘要"`
	Confidence       float64   `json:"confidence" gorm:"type:numeric(5,4);not null;default:0;comment:字段识别置信度"`
	Provider         string    `json:"provider" gorm:"size:50;index;comment:识别提供方"`
	Model            string    `json:"model" gorm:"size:200;index;comment:识别模型"`
	PromptVersion    int       `json:"promptVersion" gorm:"not null;default:0;comment:Prompt版本"`
	CorrectedBy      uint      `json:"correctedBy" gorm:"not null;index;comment:修正人ID"`
	CorrectedAt      time.Time `json:"correctedAt" gorm:"not null;index;comment:修正时间"`
	Confirmed        bool      `json:"confirmed" gorm:"not null;default:false;index;comment:是否进入最终确认"`
}

func (InvoiceReviewCorrection) TableName() string { return "invoice_review_corrections" }
