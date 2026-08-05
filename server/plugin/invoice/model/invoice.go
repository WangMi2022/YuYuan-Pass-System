package model

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

const (
	InvoiceStatusUploaded          = "uploaded"
	InvoiceStatusRecognizing       = "recognizing"
	InvoiceStatusPendingReview     = "pending_review"
	InvoiceStatusConfirmed         = "confirmed"
	InvoiceStatusRecognitionFailed = "recognition_failed"

	InvoiceVerificationUnverified    = "unverified"
	InvoiceVerificationVerifying     = "verifying"
	InvoiceVerificationVerifiedValid = "verified_valid"
	InvoiceVerificationVoided        = "verified_voided"
	InvoiceVerificationRed           = "verified_red"
	InvoiceVerificationInconsistent  = "inconsistent"
	InvoiceVerificationNotFound      = "not_found"
	InvoiceVerificationDeferred      = "deferred"
	InvoiceVerificationExpired       = "expired"
	InvoiceVerificationUnavailable   = "unavailable"

	VerificationAmountModeAmount = "amount"
	VerificationAmountModeTotal  = "total"

	RecognitionJobPending    = "pending"
	RecognitionJobProcessing = "processing"
	RecognitionJobCompleted  = "completed"
	RecognitionJobFailed     = "failed"
	FileCleanupJobPending    = "pending"
	FileCleanupJobProcessing = "processing"

	ClassificationAuto   = "auto"
	ClassificationManual = "manual"
)

// InvoiceCategory is a reporting category for confirmed invoices.
type InvoiceCategory struct {
	global.GVA_MODEL
	Name        string `json:"name" form:"name" gorm:"size:100;not null;uniqueIndex;comment:分类名称"`
	Code        string `json:"code" form:"code" gorm:"size:50;not null;uniqueIndex;comment:分类编码"`
	Description string `json:"description" form:"description" gorm:"size:500;comment:分类说明"`
	Color       string `json:"color" form:"color" gorm:"size:20;default:#6d5dfb;comment:展示颜色"`
	Sort        int    `json:"sort" form:"sort" gorm:"default:0;comment:排序"`
	Enabled     bool   `json:"enabled" form:"enabled" gorm:"default:true;comment:是否启用"`
}

func (InvoiceCategory) TableName() string { return "invoice_categories" }

// Invoice keeps financial amounts in integer cents to avoid floating-point drift.
type Invoice struct {
	global.GVA_MODEL
	Direction                      string             `json:"direction" form:"direction" gorm:"size:20;not null;default:expense;index;comment:流水方向"`
	InvoiceType                    string             `json:"invoiceType" form:"invoiceType" gorm:"size:60;index;comment:发票类型"`
	VerificationType               string             `json:"verificationType" form:"verificationType" gorm:"size:60;index;comment:验真标准票种"`
	VerificationAmountMode         string             `json:"verificationAmountMode" form:"verificationAmountMode" gorm:"size:20;comment:验真金额口径"`
	InvoiceCode                    string             `json:"invoiceCode" form:"invoiceCode" gorm:"size:80;index;comment:发票代码"`
	InvoiceNumber                  string             `json:"invoiceNumber" form:"invoiceNumber" gorm:"size:80;index;comment:发票号码"`
	CheckCode                      string             `json:"checkCode" form:"checkCode" gorm:"size:80;comment:发票校验码"`
	DuplicateKey                   *string            `json:"-" gorm:"size:64;uniqueIndex;comment:已确认发票防重键"`
	IssueDate                      *time.Time         `json:"issueDate" form:"issueDate" gorm:"type:date;index;index:idx_invoice_status_issue_date;index:idx_invoices_created_status_date,priority:3;index:idx_invoices_authority_status_date,priority:3;comment:开票日期"`
	BuyerName                      string             `json:"buyerName" form:"buyerName" gorm:"size:200;comment:购买方名称"`
	BuyerTaxNo                     string             `json:"buyerTaxNo" form:"buyerTaxNo" gorm:"size:80;comment:购买方税号"`
	SellerName                     string             `json:"sellerName" form:"sellerName" gorm:"size:200;index;comment:销售方名称"`
	SellerTaxNo                    string             `json:"sellerTaxNo" form:"sellerTaxNo" gorm:"size:80;index;comment:销售方税号"`
	AmountCents                    int64              `json:"amountCents" form:"amountCents" gorm:"not null;default:0;comment:不含税金额分"`
	TaxCents                       int64              `json:"taxCents" form:"taxCents" gorm:"not null;default:0;comment:税额分"`
	TotalCents                     int64              `json:"totalCents" form:"totalCents" gorm:"not null;default:0;index;comment:价税合计分"`
	Currency                       string             `json:"currency" form:"currency" gorm:"size:10;not null;default:CNY;comment:币种"`
	CategoryID                     *uint              `json:"categoryId" form:"categoryId" gorm:"index;comment:发票分类ID"`
	Category                       *InvoiceCategory   `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	ClassificationSource           string             `json:"classificationSource" gorm:"size:20;comment:分类来源"`
	ClassificationConfidence       float64            `json:"classificationConfidence" gorm:"type:numeric(5,4);default:0;comment:分类置信度"`
	ClassificationReason           string             `json:"classificationReason" gorm:"size:1000;comment:分类依据"`
	SuggestedCategoryID            *uint              `json:"suggestedCategoryId" gorm:"index;comment:建议分类ID"`
	SuggestedCategory              *InvoiceCategory   `json:"suggestedCategory,omitempty" gorm:"foreignKey:SuggestedCategoryID"`
	Status                         string             `json:"status" form:"status" gorm:"size:30;not null;default:uploaded;index;index:idx_invoice_status_issue_date;index:idx_invoices_created_status_date,priority:2;index:idx_invoices_authority_status_date,priority:2;comment:处理状态"`
	RecognitionProvider            string             `json:"recognitionProvider" gorm:"size:50;comment:识别提供方"`
	RecognitionConfidence          float64            `json:"recognitionConfidence" gorm:"type:numeric(5,4);default:0;comment:识别置信度"`
	FieldConfidences               map[string]float64 `json:"fieldConfidences" gorm:"serializer:json;type:text;comment:字段识别置信度"`
	RecognitionError               string             `json:"recognitionError" gorm:"size:1000;comment:识别错误"`
	VerificationStatus             string             `json:"verificationStatus" gorm:"size:30;not null;default:unverified;index;comment:权威查验状态"`
	VerificationProvider           string             `json:"verificationProvider" gorm:"size:50;comment:查验提供方"`
	VerificationMessage            string             `json:"verificationMessage" gorm:"size:1000;comment:最近查验结果"`
	VerificationInvalidSign        string             `json:"verificationInvalidSign" gorm:"size:10;comment:发票作废红冲标识"`
	VerificationFingerprint        string             `json:"-" gorm:"size:64;index;comment:最近查验对应字段指纹"`
	LatestVerificationID           *uint              `json:"latestVerificationId" gorm:"index;comment:最近查验记录ID"`
	VerificationCheckedAt          *time.Time         `json:"verificationCheckedAt" gorm:"index;comment:最近查验时间"`
	ActiveVerificationID           *uint              `json:"activeVerificationId" gorm:"index;comment:当前查验任务ID"`
	VerificationStartedAt          *time.Time         `json:"verificationStartedAt" gorm:"index;comment:当前查验开始时间"`
	RawText                        string             `json:"rawText" gorm:"type:text;comment:识别原文"`
	RawPayload                     string             `json:"-" gorm:"type:text;comment:识别原始响应"`
	FileName                       string             `json:"fileName" gorm:"size:255;not null;comment:原文件名"`
	FileKey                        string             `json:"-" gorm:"size:500;not null;comment:私有对象键"`
	FileHash                       string             `json:"fileHash" gorm:"size:64;not null;index;comment:文件SHA256"`
	FileDedupKey                   *string            `json:"-" gorm:"size:64;uniqueIndex;comment:上传人文件防重键"`
	MimeType                       string             `json:"mimeType" gorm:"size:100;not null;comment:文件类型"`
	FileSize                       int64              `json:"fileSize" gorm:"not null;comment:文件大小"`
	StorageType                    string             `json:"storageType" gorm:"size:30;not null;comment:存储类型"`
	StorageRoot                    string             `json:"-" gorm:"size:500;comment:本地存储根目录"`
	StorageEndpoint                string             `json:"-" gorm:"size:300;comment:对象存储端点"`
	StorageBucket                  string             `json:"-" gorm:"size:200;comment:对象存储桶"`
	StorageUseSSL                  bool               `json:"-" gorm:"comment:对象存储是否启用TLS"`
	CreatedBy                      uint               `json:"createdBy" gorm:"not null;index;index:idx_invoices_created_status_date,priority:1;comment:上传人ID"`
	AuthorityID                    uint               `json:"authorityId" gorm:"not null;index;index:idx_invoices_authority_status_date,priority:1;comment:上传人主角色ID"`
	ConfirmedBy                    uint               `json:"confirmedBy" gorm:"index;comment:确认人ID"`
	ConfirmedAt                    *time.Time         `json:"confirmedAt" gorm:"index;comment:确认时间"`
	VerificationBypassed           bool               `json:"verificationBypassed" gorm:"not null;default:false;index;comment:是否绕过权威查验确认"`
	VerificationBypassReason       string             `json:"verificationBypassReason" gorm:"size:500;comment:绕过权威查验原因"`
	VerificationBypassedBy         uint               `json:"verificationBypassedBy" gorm:"index;comment:绕过权威查验操作人ID"`
	VerificationBypassedAt         *time.Time         `json:"verificationBypassedAt" gorm:"index;comment:绕过权威查验时间"`
	ConfirmationVerificationStatus string             `json:"confirmationVerificationStatus" gorm:"size:30;comment:确认时权威查验状态"`
	ReviewNotes                    string             `json:"reviewNotes" form:"reviewNotes" gorm:"size:1000;comment:核对备注"`
	Items                          []InvoiceItem      `json:"items,omitempty" gorm:"foreignKey:InvoiceID;constraint:OnDelete:CASCADE"`
}

func (Invoice) TableName() string { return "invoices" }

type InvoiceVerificationDifference struct {
	Field         string `json:"field"`
	Label         string `json:"label"`
	LocalValue    string `json:"localValue"`
	OfficialValue string `json:"officialValue"`
}

// InvoiceVerification is append-only business history. Application services
// create attempts but never update or delete prior attempts after completion.
type InvoiceVerification struct {
	global.GVA_MODEL
	InvoiceID        uint                            `json:"invoiceId" gorm:"not null;index;comment:发票ID"`
	Provider         string                          `json:"provider" gorm:"size:50;not null;comment:查验提供方"`
	Status           string                          `json:"status" gorm:"size:30;not null;index;comment:查验状态"`
	VerifyResult     string                          `json:"verifyResult" gorm:"size:20;index;comment:供应商查验结果码"`
	VerifyMessage    string                          `json:"verifyMessage" gorm:"size:1000;comment:供应商查验消息"`
	VerifyFrequency  string                          `json:"verifyFrequency" gorm:"size:30;comment:历史查验次数"`
	InvalidSign      string                          `json:"invalidSign" gorm:"size:10;index;comment:作废红冲标识"`
	ProviderLogID    string                          `json:"providerLogId" gorm:"size:100;index;comment:供应商日志ID"`
	LocalFingerprint string                          `json:"localFingerprint" gorm:"size:64;not null;index;comment:本地字段指纹"`
	RequestSnapshot  map[string]string               `json:"requestSnapshot" gorm:"serializer:json;type:text;comment:查验请求快照"`
	OfficialSnapshot map[string]string               `json:"officialSnapshot" gorm:"serializer:json;type:text;comment:权威返回快照"`
	Differences      []InvoiceVerificationDifference `json:"differences" gorm:"serializer:json;type:text;comment:字段差异"`
	RawPayload       string                          `json:"-" gorm:"type:text;comment:供应商原始响应"`
	RequestedBy      uint                            `json:"requestedBy" gorm:"not null;index;comment:发起人ID"`
	CompletedAt      *time.Time                      `json:"completedAt" gorm:"index;comment:完成时间"`
}

func (InvoiceVerification) TableName() string { return "invoice_verifications" }

type InvoiceItem struct {
	global.GVA_MODEL
	InvoiceID      uint   `json:"invoiceId" gorm:"not null;index;comment:发票ID"`
	Name           string `json:"name" gorm:"size:300;not null;comment:商品或服务名称"`
	Specification  string `json:"specification" gorm:"size:200;comment:规格型号"`
	Unit           string `json:"unit" gorm:"size:30;comment:单位"`
	QuantityText   string `json:"quantityText" gorm:"size:50;comment:数量原文"`
	UnitPriceCents int64  `json:"unitPriceCents" gorm:"default:0;comment:单价分"`
	AmountCents    int64  `json:"amountCents" gorm:"default:0;comment:金额分"`
	TaxRate        string `json:"taxRate" gorm:"size:30;comment:税率"`
	TaxCents       int64  `json:"taxCents" gorm:"default:0;comment:税额分"`
}

func (InvoiceItem) TableName() string { return "invoice_items" }

type ClassificationRule struct {
	global.GVA_MODEL
	Name       string          `json:"name" form:"name" gorm:"size:120;not null;comment:规则名称"`
	Field      string          `json:"field" form:"field" gorm:"size:30;not null;default:all;index;comment:匹配字段"`
	MatchType  string          `json:"matchType" form:"matchType" gorm:"size:20;not null;default:contains;comment:匹配方式"`
	Pattern    string          `json:"pattern" form:"pattern" gorm:"size:300;not null;index;comment:匹配内容"`
	Weight     int             `json:"weight" form:"weight" gorm:"not null;default:80;comment:分类分值"`
	Priority   int             `json:"priority" form:"priority" gorm:"not null;default:0;index;comment:优先级"`
	Enabled    bool            `json:"enabled" form:"enabled" gorm:"not null;default:true;index;comment:是否启用"`
	CategoryID uint            `json:"categoryId" form:"categoryId" gorm:"not null;index;comment:目标分类"`
	Category   InvoiceCategory `json:"category" gorm:"foreignKey:CategoryID"`
}

func (ClassificationRule) TableName() string { return "invoice_classification_rules" }

type RecognitionJob struct {
	global.GVA_MODEL
	InvoiceID   uint       `json:"invoiceId" gorm:"not null;uniqueIndex;comment:发票ID"`
	Status      string     `json:"status" gorm:"size:30;not null;default:pending;index;comment:任务状态"`
	Attempts    int        `json:"attempts" gorm:"not null;default:0;comment:执行次数"`
	MaxAttempts int        `json:"maxAttempts" gorm:"not null;default:3;comment:最大执行次数"`
	NextRunAt   *time.Time `json:"nextRunAt" gorm:"index;comment:下次执行时间"`
	LockedAt    *time.Time `json:"lockedAt" gorm:"index;comment:领取时间"`
	LockToken   string     `json:"-" gorm:"size:36;not null;default:'';index;comment:任务租约令牌"`
	LastError   string     `json:"lastError" gorm:"size:1000;comment:最后错误"`
}

func (RecognitionJob) TableName() string { return "invoice_recognition_jobs" }

type InvoiceFileCleanupJob struct {
	global.GVA_MODEL
	FileKey         string     `json:"-" gorm:"size:500;not null;uniqueIndex;comment:待清理对象键"`
	StorageType     string     `json:"-" gorm:"size:30;not null;comment:存储类型"`
	StorageRoot     string     `json:"-" gorm:"size:500;comment:本地存储根目录"`
	StorageEndpoint string     `json:"-" gorm:"size:300;comment:对象存储端点"`
	StorageBucket   string     `json:"-" gorm:"size:200;comment:对象存储桶"`
	StorageUseSSL   bool       `json:"-" gorm:"comment:对象存储是否启用TLS"`
	Status          string     `json:"status" gorm:"size:30;not null;default:pending;index;comment:清理状态"`
	Attempts        int        `json:"attempts" gorm:"not null;default:0;comment:清理次数"`
	NextRunAt       *time.Time `json:"nextRunAt" gorm:"index;comment:下次执行时间"`
	LockedAt        *time.Time `json:"lockedAt" gorm:"index;comment:领取时间"`
	LockToken       string     `json:"-" gorm:"size:36;not null;default:'';index;comment:任务租约令牌"`
	LastError       string     `json:"lastError" gorm:"size:1000;comment:最后错误"`
}

func (InvoiceFileCleanupJob) TableName() string { return "invoice_file_cleanup_jobs" }
