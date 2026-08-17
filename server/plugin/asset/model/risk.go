package model

import (
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/common"
)

const (
	RiskSeverityLow      = "low"
	RiskSeverityMedium   = "medium"
	RiskSeverityHigh     = "high"
	RiskSeverityCritical = "critical"

	RiskStatusOpen         = "open"
	RiskStatusAcknowledged = "acknowledged"
	RiskStatusResolved     = "resolved"
	RiskStatusIgnored      = "ignored"

	RiskScanStatusRunning = "running"
	RiskScanStatusSuccess = "success"
	RiskScanStatusFailed  = "failed"

	RiskScanTriggerManual    = "manual"
	RiskScanTriggerScheduled = "scheduled"
)

type AssetRiskRule struct {
	global.GVA_MODEL
	Code           string         `json:"code" gorm:"size:80;not null;uniqueIndex;comment:规则编码"`
	Name           string         `json:"name" gorm:"size:120;not null;comment:规则名称"`
	Category       string         `json:"category" gorm:"size:30;not null;index;comment:风险类别"`
	Severity       string         `json:"severity" gorm:"size:20;not null;index;comment:风险等级"`
	Description    string         `json:"description" gorm:"size:500;comment:规则说明"`
	Recommendation string         `json:"recommendation" gorm:"size:500;comment:默认建议动作"`
	Parameters     common.JSONMap `json:"parameters" gorm:"serializer:json;type:jsonb;comment:规则阈值"`
	Enabled        bool           `json:"enabled" gorm:"not null;default:true;index;comment:是否启用"`
	Version        int            `json:"version" gorm:"not null;default:1;comment:规则版本"`
}

func (AssetRiskRule) TableName() string { return "asset_risk_rules" }

type AssetRiskEvent struct {
	global.GVA_MODEL
	Fingerprint     string         `json:"fingerprint" gorm:"size:64;not null;uniqueIndex;comment:风险唯一指纹"`
	AssetID         uint           `json:"assetId" gorm:"not null;index;index:idx_asset_risk_active,priority:1;comment:资产ID"`
	Asset           Asset          `json:"asset" gorm:"foreignKey:AssetID"`
	RuleCode        string         `json:"ruleCode" gorm:"size:80;not null;index;comment:规则编码"`
	RuleVersion     int            `json:"ruleVersion" gorm:"not null;comment:规则版本"`
	Category        string         `json:"category" gorm:"size:30;not null;index;comment:风险类别"`
	Severity        string         `json:"severity" gorm:"size:20;not null;index;comment:风险等级"`
	Status          string         `json:"status" gorm:"size:20;not null;default:open;index;index:idx_asset_risk_active,priority:2;comment:处理状态"`
	Title           string         `json:"title" gorm:"size:200;not null;comment:风险标题"`
	Description     string         `json:"description" gorm:"type:text;comment:风险说明"`
	Evidence        common.JSONMap `json:"evidence" gorm:"serializer:json;type:jsonb;comment:证据快照"`
	Recommendation  string         `json:"recommendation" gorm:"size:500;comment:推荐动作"`
	FirstDetectedAt time.Time      `json:"firstDetectedAt" gorm:"not null;index;comment:首次命中时间"`
	LastDetectedAt  time.Time      `json:"lastDetectedAt" gorm:"not null;index;comment:最近命中时间"`
	LastScanRunID   uint           `json:"lastScanRunId" gorm:"not null;index;comment:最近扫描运行ID"`
	AssignedTo      uint           `json:"assignedTo" gorm:"index;comment:处理人ID"`
	AssignedToName  string         `json:"assignedToName" gorm:"size:100;comment:处理人名称"`
	HandledBy       uint           `json:"handledBy" gorm:"index;comment:最终处理人ID"`
	HandledByName   string         `json:"handledByName" gorm:"size:100;comment:最终处理人名称"`
	HandledAt       *time.Time     `json:"handledAt" gorm:"index;comment:最终处理时间"`
	ResolutionNote  string         `json:"resolutionNote" gorm:"type:text;comment:处理说明"`
}

func (AssetRiskEvent) TableName() string { return "asset_risk_events" }

type AssetRiskEventLog struct {
	global.GVA_MODEL
	EventID   uint   `json:"eventId" gorm:"not null;index;comment:风险事件ID"`
	Action    string `json:"action" gorm:"size:30;not null;index;comment:处理动作"`
	FromState string `json:"fromState" gorm:"size:20;comment:原状态"`
	ToState   string `json:"toState" gorm:"size:20;comment:新状态"`
	ActorID   uint   `json:"actorId" gorm:"index;comment:操作人ID"`
	ActorName string `json:"actorName" gorm:"size:100;comment:操作人名称"`
	Note      string `json:"note" gorm:"type:text;comment:操作说明"`
}

func (AssetRiskEventLog) TableName() string { return "asset_risk_event_logs" }

type AssetRiskScanRun struct {
	global.GVA_MODEL
	TriggerType     string     `json:"triggerType" gorm:"size:20;not null;index;comment:触发方式"`
	Status          string     `json:"status" gorm:"size:20;not null;index;uniqueIndex:idx_asset_risk_scan_single_running,where:status = 'running';comment:运行状态"`
	StartedAt       time.Time  `json:"startedAt" gorm:"not null;index;comment:开始时间"`
	FinishedAt      *time.Time `json:"finishedAt" gorm:"index;comment:结束时间"`
	HeartbeatAt     time.Time  `json:"heartbeatAt" gorm:"not null;index;comment:最近心跳"`
	ResumeCount     int        `json:"resumeCount" gorm:"not null;default:0;comment:续扫次数"`
	CursorAssetID   uint       `json:"cursorAssetId" gorm:"not null;default:0;comment:续扫游标"`
	ScannedAssets   int64      `json:"scannedAssets" gorm:"not null;default:0;comment:已扫描资产数"`
	NewEvents       int64      `json:"newEvents" gorm:"not null;default:0;comment:新增风险数"`
	UpdatedEvents   int64      `json:"updatedEvents" gorm:"not null;default:0;comment:更新风险数"`
	ClosedEvents    int64      `json:"closedEvents" gorm:"not null;default:0;comment:自动关闭风险数"`
	ErrorMessage    string     `json:"errorMessage" gorm:"type:text;comment:错误信息"`
	TriggeredBy     uint       `json:"triggeredBy" gorm:"index;comment:触发用户ID"`
	TriggeredByName string     `json:"triggeredByName" gorm:"size:100;comment:触发用户名称"`
}

func (AssetRiskScanRun) TableName() string { return "asset_risk_scan_runs" }
