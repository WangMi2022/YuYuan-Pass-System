package model

import (
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/common"
)

// CopilotRun records planner/tool observability without changing business data.
type CopilotRun struct {
	global.GVA_MODEL
	RequestID      string         `json:"requestId" gorm:"size:36;not null;uniqueIndex"`
	SessionID      uint           `json:"sessionId" gorm:"not null;index"`
	UserID         uint           `json:"userId" gorm:"not null;index"`
	AuthorityID    uint           `json:"authorityId" gorm:"not null;index"`
	Planner        string         `json:"planner" gorm:"size:30;not null;index"`
	Intent         string         `json:"intent" gorm:"size:160;index"`
	PlannedTools   common.JSONMap `json:"plannedTools" gorm:"serializer:json;type:jsonb"`
	ExecutedTools  common.JSONMap `json:"executedTools" gorm:"serializer:json;type:jsonb"`
	Status         string         `json:"status" gorm:"size:20;not null;index"`
	Partial        bool           `json:"partial" gorm:"not null;default:false;index"`
	DurationMS     int64          `json:"durationMs" gorm:"not null;default:0"`
	ModelUsed      bool           `json:"modelUsed" gorm:"not null;default:false"`
	FallbackReason string         `json:"fallbackReason" gorm:"size:120"`
}

func (CopilotRun) TableName() string { return "ai_copilot_runs" }
