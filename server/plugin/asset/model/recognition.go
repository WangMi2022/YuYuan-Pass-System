package model

import (
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/global"
)

const (
	AssetRecognitionPending    = "pending"
	AssetRecognitionProcessing = "processing"
	AssetRecognitionReviewing  = "reviewing"
	AssetRecognitionCompleted  = "completed"
	AssetRecognitionFailed     = "failed"
	AssetRecognitionDeleting   = "deleting"
)

type AssetRecognitionResult struct {
	Name                      string     `json:"name"`
	Brand                     string     `json:"brand"`
	Model                     string     `json:"model"`
	SerialNumber              string     `json:"serialNumber"`
	Specifications            string     `json:"specifications"`
	ProductionDate            *time.Time `json:"productionDate"`
	RecommendedCategoryCode   string     `json:"recommendedCategoryCode"`
	RecommendedCategoryID     uint       `json:"recommendedCategoryId"`
	RecommendedUnit           string     `json:"recommendedUnit"`
	RecommendedWarrantyMonths int        `json:"recommendedWarrantyMonths"`
	RawText                   string     `json:"rawText"`
}

type AssetRecognitionDraft struct {
	AssetCode                 string     `json:"assetCode"`
	Name                      string     `json:"name"`
	CategoryID                uint       `json:"categoryId"`
	Brand                     string     `json:"brand"`
	Model                     string     `json:"model"`
	SerialNumber              string     `json:"serialNumber"`
	Specifications            string     `json:"specifications"`
	ProductionDate            *time.Time `json:"productionDate"`
	Quantity                  int        `json:"quantity"`
	Unit                      string     `json:"unit"`
	UnitPrice                 float64    `json:"unitPrice"`
	CurrentValue              float64    `json:"currentValue"`
	Supplier                  string     `json:"supplier"`
	PurchaseDate              *time.Time `json:"purchaseDate"`
	WarrantyEndDate           *time.Time `json:"warrantyEndDate"`
	RecommendedWarrantyMonths int        `json:"recommendedWarrantyMonths"`
	Photos                    []Photo    `json:"photos"`
	Remarks                   string     `json:"remarks"`
}

type AssetRecognitionWarning struct {
	Code     string `json:"code"`
	Field    string `json:"field"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type AssetDuplicateCandidate struct {
	AssetID      uint   `json:"assetId"`
	AssetCode    string `json:"assetCode"`
	Name         string `json:"name"`
	CategoryName string `json:"categoryName"`
	Brand        string `json:"brand"`
	Model        string `json:"model"`
	SerialNumber string `json:"serialNumber"`
	MatchType    string `json:"matchType"`
}

type AssetRecognitionJob struct {
	global.GVA_MODEL
	Status              string                    `json:"status" gorm:"size:30;not null;default:pending;index;comment:任务状态"`
	Attempts            int                       `json:"attempts" gorm:"not null;default:0;comment:执行次数"`
	MaxAttempts         int                       `json:"maxAttempts" gorm:"not null;default:3;comment:最大执行次数"`
	Provider            string                    `json:"provider" gorm:"size:80;index;comment:模型供应商"`
	Model               string                    `json:"model" gorm:"size:200;index;comment:模型名称"`
	PromptVersion       int                       `json:"promptVersion" gorm:"not null;default:0;comment:Prompt版本"`
	InputPhotos         []Photo                   `json:"fileKeys" gorm:"column:file_keys;serializer:json;type:jsonb;comment:输入图片对象"`
	Result              AssetRecognitionResult    `json:"result" gorm:"serializer:json;type:jsonb;comment:结构化识别结果"`
	Draft               AssetRecognitionDraft     `json:"draft" gorm:"serializer:json;type:jsonb;comment:人工修正草稿"`
	FieldConfidences    map[string]float64        `json:"fieldConfidences" gorm:"serializer:json;type:jsonb;comment:字段置信度"`
	Warnings            []AssetRecognitionWarning `json:"warnings" gorm:"serializer:json;type:jsonb;comment:校验警告"`
	DuplicateCandidates []AssetDuplicateCandidate `json:"duplicateCandidates" gorm:"serializer:json;type:jsonb;comment:重复候选"`
	StorageType         string                    `json:"-" gorm:"size:30;not null;comment:存储类型"`
	StorageRoot         string                    `json:"-" gorm:"size:500;comment:本地存储根目录"`
	StorageEndpoint     string                    `json:"-" gorm:"size:300;comment:对象存储端点"`
	StorageBucket       string                    `json:"-" gorm:"size:200;comment:对象存储桶"`
	StorageUseSSL       bool                      `json:"-" gorm:"comment:对象存储是否启用TLS"`
	NextRunAt           *time.Time                `json:"nextRunAt" gorm:"index;comment:下次执行时间"`
	LockedAt            *time.Time                `json:"lockedAt" gorm:"index;comment:任务领取时间"`
	LockToken           string                    `json:"-" gorm:"size:36;not null;default:'';index;comment:任务租约令牌"`
	LastError           string                    `json:"lastError" gorm:"size:1000;comment:最后错误"`
	CreatedBy           uint                      `json:"createdBy" gorm:"not null;index;comment:创建用户"`
	AuthorityID         uint                      `json:"authorityId" gorm:"not null;index;comment:创建角色"`
	DraftUpdatedBy      uint                      `json:"draftUpdatedBy" gorm:"index;comment:草稿更新用户"`
	DraftUpdatedAt      *time.Time                `json:"draftUpdatedAt" gorm:"index;comment:草稿更新时间"`
	CompletedAt         *time.Time                `json:"completedAt" gorm:"index;comment:识别或确认完成时间"`
	ConfirmedAssetID    *uint                     `json:"confirmedAssetId" gorm:"uniqueIndex;comment:最终资产ID"`
	ConfirmedBy         uint                      `json:"confirmedBy" gorm:"index;comment:确认用户"`
	ConfirmedAt         *time.Time                `json:"confirmedAt" gorm:"index;comment:确认时间"`
}

func (AssetRecognitionJob) TableName() string { return "asset_recognition_jobs" }
