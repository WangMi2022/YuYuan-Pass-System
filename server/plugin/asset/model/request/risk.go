package request

import (
	"github.com/WangMi2022/mit-assets-admin/server/model/common"
	commonRequest "github.com/WangMi2022/mit-assets-admin/server/model/common/request"
)

type RiskSearch struct {
	commonRequest.PageInfo
	Status     string `json:"status" form:"status"`
	Severity   string `json:"severity" form:"severity"`
	Category   string `json:"category" form:"category"`
	RuleCode   string `json:"ruleCode" form:"ruleCode"`
	AssetID    uint   `json:"assetId" form:"assetId"`
	AssignedTo uint   `json:"assignedTo" form:"assignedTo"`
}

type RiskRuleSearch struct {
	commonRequest.PageInfo
	Paged bool `json:"paged" form:"paged"`
}

type RiskScanSearch struct {
	commonRequest.PageInfo
	Status      string `json:"status" form:"status"`
	TriggerType string `json:"triggerType" form:"triggerType"`
}

type StartRiskScan struct {
	RunID uint `json:"runId"`
}

type RiskScanDelete struct {
	IDs           []uint `json:"ids"`
	ClearFinished bool   `json:"clearFinished"`
}

type RiskEventDelete struct {
	IDs          []uint `json:"ids"`
	ClearHistory bool   `json:"clearHistory"`
}

type RiskRuleUpdate struct {
	ID         uint           `json:"ID"`
	Severity   string         `json:"severity"`
	Parameters common.JSONMap `json:"parameters"`
	Enabled    bool           `json:"enabled"`
}

type RiskAction struct {
	IDs  []uint `json:"ids"`
	Note string `json:"note"`
}

type RiskAssignment struct {
	IDs        []uint `json:"ids"`
	AssignedTo uint   `json:"assignedTo"`
}
