package model

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common"
)

const (
	MessageRoleUser       = "user"
	MessageRoleAssistant  = "assistant"
	DraftTypeSchedule     = "schedule"
	DraftTypeOperation    = "operation"
	DraftStatusDraft      = "draft"
	DraftStatusProcessing = "processing"
	DraftStatusAccepted   = "accepted"
	DraftStatusDiscarded  = "discarded"
)

// CopilotSession and CopilotMessage keep a traceable, user-and-authority-scoped read-only
// assistant history. Business records are never modified by this module.
type CopilotSession struct {
	global.GVA_MODEL
	UserID        uint      `json:"userId" gorm:"not null;index"`
	AuthorityID   uint      `json:"authorityId" gorm:"not null;index"`
	Title         string    `json:"title" gorm:"size:160;not null"`
	LastMessageAt time.Time `json:"lastMessageAt" gorm:"index"`
}

func (CopilotSession) TableName() string { return "ai_copilot_sessions" }

type CopilotMessage struct {
	global.GVA_MODEL
	SessionID   uint           `json:"sessionId" gorm:"not null;index"`
	UserID      uint           `json:"userId" gorm:"not null;index"`
	AuthorityID uint           `json:"authorityId" gorm:"not null;index"`
	Role        string         `json:"role" gorm:"size:20;not null"`
	Content     string         `json:"content" gorm:"type:text;not null"`
	Intent      string         `json:"intent" gorm:"size:40;index"`
	Tool        string         `json:"tool" gorm:"size:80;index"`
	Citations   common.JSONMap `json:"citations" gorm:"serializer:json;type:jsonb"`
}

func (CopilotMessage) TableName() string { return "ai_copilot_messages" }

type SmartDailyReport struct {
	global.GVA_MODEL
	UserID      uint           `json:"userId" gorm:"not null;uniqueIndex:idx_smart_report_user_authority_date"`
	AuthorityID uint           `json:"authorityId" gorm:"not null;uniqueIndex:idx_smart_report_user_authority_date"`
	ReportDate  time.Time      `json:"reportDate" gorm:"type:date;not null;uniqueIndex:idx_smart_report_user_authority_date"`
	Metrics     common.JSONMap `json:"metrics" gorm:"serializer:json;type:jsonb"`
	Summary     string         `json:"summary" gorm:"type:text"`
	GeneratedBy string         `json:"generatedBy" gorm:"size:30;not null;default:deterministic"`
	GeneratedAt time.Time      `json:"generatedAt" gorm:"not null"`
}

func (SmartDailyReport) TableName() string { return "smart_daily_reports" }

type SmartReportSubscription struct {
	global.GVA_MODEL
	UserID       uint   `json:"userId" gorm:"not null;uniqueIndex"`
	Enabled      bool   `json:"enabled" gorm:"not null;default:true"`
	DeliveryTime string `json:"deliveryTime" gorm:"size:5;not null;default:09:00"`
	Channels     string `json:"channels" gorm:"type:text;not null;default:in_app"`
}

func (SmartReportSubscription) TableName() string { return "smart_report_subscriptions" }

type SmartReportDelivery struct {
	global.GVA_MODEL
	UserID      uint       `json:"userId" gorm:"not null;index;uniqueIndex:idx_smart_delivery_once"`
	ReportID    uint       `json:"reportId" gorm:"not null;index;uniqueIndex:idx_smart_delivery_once"`
	Channel     string     `json:"channel" gorm:"size:30;not null;uniqueIndex:idx_smart_delivery_once"`
	Status      string     `json:"status" gorm:"size:20;not null"`
	RetryCount  int        `json:"retryCount" gorm:"not null;default:0"`
	LastAttempt *time.Time `json:"lastAttempt,omitempty" gorm:"index"`
	SentAt      *time.Time `json:"sentAt" gorm:"index"`
	Error       string     `json:"error" gorm:"size:500"`
}

func (SmartReportDelivery) TableName() string { return "smart_report_deliveries" }

type SmartDraft struct {
	global.GVA_MODEL
	UserID     uint           `json:"userId" gorm:"not null;index"`
	DedupKey   *string        `json:"-" gorm:"size:180;uniqueIndex"`
	DraftType  string         `json:"draftType" gorm:"size:30;not null;index"`
	SourceID   uint           `json:"sourceId" gorm:"index"`
	Status     string         `json:"status" gorm:"size:20;not null;default:draft;index"`
	Payload    common.JSONMap `json:"payload" gorm:"serializer:json;type:jsonb"`
	Confidence float64        `json:"confidence" gorm:"type:numeric(5,4);default:0"`
	ExpiresAt  *time.Time     `json:"expiresAt" gorm:"index"`
	LastError  string         `json:"lastError,omitempty" gorm:"size:500"`
}

func (SmartDraft) TableName() string { return "smart_drafts" }
