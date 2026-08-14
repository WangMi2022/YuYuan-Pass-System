package response

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
)

type RiskMetric struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}

type RiskTrendPoint struct {
	Date     string `json:"date"`
	New      int64  `json:"new"`
	Resolved int64  `json:"resolved"`
}

type RiskDashboard struct {
	TotalOpen    int64                   `json:"totalOpen"`
	HighOpen     int64                   `json:"highOpen"`
	TodayNew     int64                   `json:"todayNew"`
	Overdue      int64                   `json:"overdue"`
	ByCategory   []RiskMetric            `json:"byCategory"`
	BySeverity   []RiskMetric            `json:"bySeverity"`
	ByStatus     []RiskMetric            `json:"byStatus"`
	ByCustodian  []RiskMetric            `json:"byCustodian"`
	Trend        []RiskTrendPoint        `json:"trend"`
	RecentEvents []model.AssetRiskEvent  `json:"recentEvents"`
	LatestScan   *model.AssetRiskScanRun `json:"latestScan"`
	GeneratedAt  time.Time               `json:"generatedAt"`
}

type RiskDetail struct {
	Event model.AssetRiskEvent      `json:"event"`
	Logs  []model.AssetRiskEventLog `json:"logs"`
}
