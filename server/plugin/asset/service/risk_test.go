package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
	assetRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model/request"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupRiskTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	previous := global.GVA_DB
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{TranslateError: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open risk test database: %v", err)
	}
	if err := database.AutoMigrate(
		&model.Category{}, &model.Asset{}, &model.AssetOperationRecord{},
		&model.AssetRiskRule{}, &model.AssetRiskEvent{}, &model.AssetRiskEventLog{}, &model.AssetRiskScanRun{},
	); err != nil {
		t.Fatalf("migrate risk tables: %v", err)
	}
	global.GVA_DB = database
	t.Cleanup(func() {
		global.GVA_DB = previous
		if sqlDB, err := database.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return database
}

func createRiskTestAsset(t *testing.T, status string, createdAt time.Time) model.Asset {
	t.Helper()
	category := model.Category{Name: "风险测试分类", Code: fmt.Sprintf("RISK-%d", time.Now().UnixNano()), Enabled: true}
	if err := global.GVA_DB.Create(&category).Error; err != nil {
		t.Fatalf("create risk category: %v", err)
	}
	asset := model.Asset{
		AssetCode: fmt.Sprintf("RISK-%d", time.Now().UnixNano()), Name: "风险测试资产", CategoryID: category.ID,
		Quantity: 1, Unit: "件", OriginalValue: 100, CurrentValue: 100, Status: status,
	}
	asset.CreatedAt = createdAt
	asset.UpdatedAt = createdAt
	if err := global.GVA_DB.Create(&asset).Error; err != nil {
		t.Fatalf("create risk asset: %v", err)
	}
	return asset
}

func createRiskScanRun(t *testing.T, startedAt time.Time) model.AssetRiskScanRun {
	t.Helper()
	run := model.AssetRiskScanRun{TriggerType: model.RiskScanTriggerManual, Status: model.RiskScanStatusRunning, StartedAt: startedAt, HeartbeatAt: startedAt}
	if err := global.GVA_DB.Create(&run).Error; err != nil {
		t.Fatalf("create risk scan run: %v", err)
	}
	return run
}

func TestDefaultRiskRulesAreCompleteAndValid(t *testing.T) {
	rules := defaultRiskRules()
	if len(rules) != 17 {
		t.Fatalf("default risk rules = %d, want 17", len(rules))
	}
	seen := make(map[string]struct{}, len(rules))
	for _, rule := range rules {
		if _, exists := seen[rule.Code]; exists {
			t.Fatalf("duplicate default rule %s", rule.Code)
		}
		seen[rule.Code] = struct{}{}
		if !validRiskSeverity(rule.Severity) || rule.Version != 1 || !rule.Enabled {
			t.Fatalf("invalid default rule: %#v", rule)
		}
		if err := ValidateRiskRuleParameters(rule.Code, rule.Parameters); err != nil {
			t.Fatalf("validate %s: %v", rule.Code, err)
		}
	}
}

func TestEveryDefaultRiskRuleCanProduceEvidence(t *testing.T) {
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.Local)
	old := now.AddDate(0, 0, -200)
	warrantySoon := now.AddDate(0, 0, 10)
	warrantyExpired := now.AddDate(0, 0, -10)
	baseAsset := model.Asset{
		AssetCode: "ASSET-10001", Name: "规则测试资产", SerialNumber: "SN-001", Quantity: 1,
		OriginalValue: 20000, CurrentValue: 15000, Status: model.AssetStatusIdle,
		Category: model.Category{Code: "IT-COMPUTER"},
	}
	baseAsset.ID = 1
	baseAsset.CreatedAt = old
	baseAsset.UpdatedAt = old
	ruleByCode := make(map[string]model.AssetRiskRule)
	for _, rule := range defaultRiskRules() {
		ruleByCode[rule.Code] = rule
	}
	tests := []struct {
		code       string
		asset      model.Asset
		evaluation riskEvaluationContext
	}{
		{code: riskRulePendingTooLong, asset: func() model.Asset { item := baseAsset; item.Status = model.AssetStatusPendingInbound; return item }()},
		{code: riskRuleInUseWithoutCustodian, asset: func() model.Asset {
			item := baseAsset
			item.Status = model.AssetStatusInUse
			item.Custodian = ""
			return item
		}()},
		{code: riskRuleIdleWithCustodian, asset: func() model.Asset { item := baseAsset; item.Custodian = "张三"; return item }()},
		{code: riskRuleRetiredWithCustodian, asset: func() model.Asset {
			item := baseAsset
			item.Status = model.AssetStatusRetired
			item.Custodian = "张三"
			return item
		}()},
		{code: riskRuleRetiredValueNonzero, asset: func() model.Asset { item := baseAsset; item.Status = model.AssetStatusRetired; return item }()},
		{code: riskRuleValueOverOriginal, asset: func() model.Asset { item := baseAsset; item.CurrentValue = 30000; return item }()},
		{code: riskRuleHighValueIdle, asset: baseAsset, evaluation: riskEvaluationContext{recordsByAsset: map[uint][]model.AssetOperationRecord{1: {{Type: "return", ToStatus: model.AssetStatusIdle, OperatedAt: old}}}}},
		{code: riskRuleWarrantyMissing, asset: baseAsset},
		{code: riskRuleWarrantyExpiring, asset: func() model.Asset { item := baseAsset; item.WarrantyEndDate = &warrantySoon; return item }()},
		{code: riskRuleWarrantyExpired, asset: func() model.Asset { item := baseAsset; item.WarrantyEndDate = &warrantyExpired; return item }()},
		{code: riskRuleMaintenanceOverdue, asset: func() model.Asset { item := baseAsset; item.Status = model.AssetStatusMaintenance; return item }(), evaluation: riskEvaluationContext{recordsByAsset: map[uint][]model.AssetOperationRecord{1: {{Type: "maintenance", ToStatus: model.AssetStatusMaintenance, OperatedAt: old}}}}},
		{code: riskRuleFrequentMaintenance, asset: baseAsset, evaluation: riskEvaluationContext{recordsByAsset: map[uint][]model.AssetOperationRecord{1: {{Type: "maintenance", OperatedAt: now.AddDate(0, 0, -10)}, {Type: "maintenance", OperatedAt: now.AddDate(0, 0, -20)}, {Type: "maintenance", OperatedAt: now.AddDate(0, 0, -30)}}}}},
		{code: riskRuleFrequentTransfer, asset: baseAsset, evaluation: riskEvaluationContext{recordsByAsset: map[uint][]model.AssetOperationRecord{1: {{Type: "transfer", OperatedAt: now.AddDate(0, 0, -10)}, {Type: "transfer", OperatedAt: now.AddDate(0, 0, -20)}, {Type: "transfer", OperatedAt: now.AddDate(0, 0, -30)}}}}},
		{code: riskRuleLongTermInUse, asset: func() model.Asset {
			item := baseAsset
			item.Status = model.AssetStatusInUse
			item.Custodian = "张三"
			return item
		}(), evaluation: riskEvaluationContext{recordsByAsset: map[uint][]model.AssetOperationRecord{1: {{Type: "issue", ToStatus: model.AssetStatusInUse, OperatedAt: old}}}}},
		{code: riskRuleDuplicateSerial, asset: baseAsset, evaluation: riskEvaluationContext{serialGroups: map[string][]similarAsset{"SN001": {{ID: 1, Code: "ASSET-10001"}, {ID: 2, Code: "ASSET-20001"}}}}},
		{code: riskRuleSimilarCode, asset: baseAsset, evaluation: riskEvaluationContext{similarByAsset: map[uint][]similarAsset{1: {{ID: 2, Code: "ASSET-10002"}}}}},
		{code: riskRuleStateRecordMismatch, asset: baseAsset, evaluation: riskEvaluationContext{recordsByAsset: map[uint][]model.AssetOperationRecord{1: {{Type: "issue", ToStatus: model.AssetStatusInUse, ToCustodian: "张三", OperatedAt: now}}}}},
	}
	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			test.evaluation.now = now
			if test.evaluation.recordsByAsset == nil {
				test.evaluation.recordsByAsset = map[uint][]model.AssetOperationRecord{}
			}
			matches := evaluateRiskRule(ruleByCode[test.code], test.asset, test.evaluation)
			if len(matches) == 0 || len(matches[0].evidence) == 0 {
				t.Fatalf("rule %s did not produce evidence: %#v", test.code, matches)
			}
		})
	}
}

func TestRiskScanIsIdempotentAndAutoResolves(t *testing.T) {
	setupRiskTestDB(t)
	Risk.SeedRules(context.Background())
	now := time.Now()
	asset := createRiskTestAsset(t, model.AssetStatusPendingInbound, now.AddDate(0, 0, -10))

	first := createRiskScanRun(t, now)
	if err := Risk.runScan(context.Background(), first.ID); err != nil {
		t.Fatalf("first risk scan: %v", err)
	}
	var pendingEvents []model.AssetRiskEvent
	if err := global.GVA_DB.Where("asset_id = ? AND rule_code = ?", asset.ID, riskRulePendingTooLong).Find(&pendingEvents).Error; err != nil {
		t.Fatalf("list first risk events: %v", err)
	}
	if len(pendingEvents) != 1 || pendingEvents[0].Status != model.RiskStatusOpen {
		t.Fatalf("unexpected first events: %#v", pendingEvents)
	}

	second := createRiskScanRun(t, now.Add(time.Minute))
	if err := Risk.runScan(context.Background(), second.ID); err != nil {
		t.Fatalf("second risk scan: %v", err)
	}
	var count int64
	if err := global.GVA_DB.Model(&model.AssetRiskEvent{}).Where("asset_id = ? AND rule_code = ?", asset.ID, riskRulePendingTooLong).Count(&count).Error; err != nil {
		t.Fatalf("count idempotent events: %v", err)
	}
	if count != 1 {
		t.Fatalf("idempotent event count = %d, want 1", count)
	}
	if err := global.GVA_DB.First(&second, second.ID).Error; err != nil {
		t.Fatalf("reload second run: %v", err)
	}
	if second.NewEvents != 0 || second.UpdatedEvents == 0 || second.Status != model.RiskScanStatusSuccess {
		t.Fatalf("unexpected second run: %#v", second)
	}

	if err := global.GVA_DB.Model(&model.Asset{}).Where("id = ?", asset.ID).Updates(map[string]any{"status": model.AssetStatusIdle, "custodian": ""}).Error; err != nil {
		t.Fatalf("fix pending asset: %v", err)
	}
	third := createRiskScanRun(t, now.Add(2*time.Minute))
	if err := Risk.runScan(context.Background(), third.ID); err != nil {
		t.Fatalf("third risk scan: %v", err)
	}
	if err := global.GVA_DB.First(&pendingEvents[0], pendingEvents[0].ID).Error; err != nil {
		t.Fatalf("reload pending event: %v", err)
	}
	if pendingEvents[0].Status != model.RiskStatusResolved || pendingEvents[0].ResolutionNote == "" {
		t.Fatalf("stale event not auto-resolved: %#v", pendingEvents[0])
	}
}

func TestRiskRuleVersionPreservesHistoricalEvent(t *testing.T) {
	setupRiskTestDB(t)
	Risk.SeedRules(context.Background())
	now := time.Now()
	asset := createRiskTestAsset(t, model.AssetStatusPendingInbound, now.AddDate(0, 0, -20))
	first := createRiskScanRun(t, now)
	if err := Risk.runScan(context.Background(), first.ID); err != nil {
		t.Fatalf("first scan: %v", err)
	}
	var rule model.AssetRiskRule
	if err := global.GVA_DB.Where("code = ?", riskRulePendingTooLong).First(&rule).Error; err != nil {
		t.Fatalf("load pending rule: %v", err)
	}
	updated, err := Risk.UpdateRule(context.Background(), assetRequest.RiskRuleUpdate{ID: rule.ID, Severity: model.RiskSeverityHigh, Parameters: map[string]any{"days": 14}, Enabled: true})
	if err != nil {
		t.Fatalf("update pending rule: %v", err)
	}
	if updated.Version != 2 {
		t.Fatalf("updated rule version = %d, want 2", updated.Version)
	}
	second := createRiskScanRun(t, now.Add(time.Minute))
	if err := Risk.runScan(context.Background(), second.ID); err != nil {
		t.Fatalf("second scan: %v", err)
	}
	var events []model.AssetRiskEvent
	if err := global.GVA_DB.Where("asset_id = ? AND rule_code = ?", asset.ID, riskRulePendingTooLong).Order("rule_version ASC").Find(&events).Error; err != nil {
		t.Fatalf("list versioned events: %v", err)
	}
	if len(events) != 2 || events[0].Status != model.RiskStatusResolved || events[1].Status != model.RiskStatusOpen || events[1].RuleVersion != 2 {
		t.Fatalf("unexpected versioned events: %#v", events)
	}
}

func TestRiskTransitionsCreateAuditLogs(t *testing.T) {
	setupRiskTestDB(t)
	asset := createRiskTestAsset(t, model.AssetStatusIdle, time.Now())
	event := model.AssetRiskEvent{
		Fingerprint: strings.Repeat("a", 64), AssetID: asset.ID, RuleCode: riskRuleHighValueIdle, RuleVersion: 1,
		Category: "value", Severity: model.RiskSeverityHigh, Status: model.RiskStatusOpen,
		Title: "测试风险", FirstDetectedAt: time.Now(), LastDetectedAt: time.Now(),
	}
	if err := global.GVA_DB.Create(&event).Error; err != nil {
		t.Fatalf("create transition event: %v", err)
	}
	if err := Risk.Acknowledge(context.Background(), assetRequest.RiskAction{IDs: []uint{event.ID}, Note: "已接手"}, 7, "处理人"); err != nil {
		t.Fatalf("acknowledge risk: %v", err)
	}
	if err := Risk.Resolve(context.Background(), assetRequest.RiskAction{IDs: []uint{event.ID}, Note: "资产已重新调配"}, 7, "处理人"); err != nil {
		t.Fatalf("resolve risk: %v", err)
	}
	if err := Risk.Reopen(context.Background(), assetRequest.RiskAction{IDs: []uint{event.ID}}, 8, "复核人"); err == nil {
		t.Fatal("reopen risk without note should fail")
	}
	if err := Risk.Reopen(context.Background(), assetRequest.RiskAction{IDs: []uint{event.ID}, Note: "复核发现仍需处理"}, 8, "复核人"); err != nil {
		t.Fatalf("reopen risk: %v", err)
	}
	if err := global.GVA_DB.First(&event, event.ID).Error; err != nil {
		t.Fatalf("reload transition event: %v", err)
	}
	if event.Status != model.RiskStatusOpen || event.HandledAt != nil || event.ResolutionNote != "" {
		t.Fatalf("unexpected reopened event: %#v", event)
	}
	var logCount int64
	if err := global.GVA_DB.Model(&model.AssetRiskEventLog{}).Where("event_id = ?", event.ID).Count(&logCount).Error; err != nil {
		t.Fatalf("count transition logs: %v", err)
	}
	if logCount != 3 {
		t.Fatalf("transition log count = %d, want 3", logCount)
	}
}

func TestPrepareScanRunRejectsFreshDatabaseRun(t *testing.T) {
	setupRiskTestDB(t)
	now := time.Now()
	active := createRiskScanRun(t, now)

	if _, err := Risk.prepareScanRun(model.RiskScanTriggerManual, 1, "测试用户", 0); err == nil {
		t.Fatal("new scan should reject a fresh database run")
	}
	if err := global.GVA_DB.Model(&active).Update("heartbeat_at", now.Add(-riskScanHeartbeatTimeout-time.Minute)).Error; err != nil {
		t.Fatalf("expire active scan heartbeat: %v", err)
	}
	if _, err := Risk.prepareScanRun(model.RiskScanTriggerManual, 1, "测试用户", 0); err != nil {
		t.Fatalf("new scan should allow a stale database run: %v", err)
	}
}

func TestRiskScanRunAllowsOnlyOneRunningTask(t *testing.T) {
	setupRiskTestDB(t)
	now := time.Now()
	createRiskScanRun(t, now)
	second := model.AssetRiskScanRun{TriggerType: model.RiskScanTriggerManual, Status: model.RiskScanStatusRunning, StartedAt: now, HeartbeatAt: now}
	if err := global.GVA_DB.Create(&second).Error; err == nil {
		t.Fatal("database must reject a second running scan task")
	}
}

func TestPrepareScanRunRejectsHistoricalFailedRun(t *testing.T) {
	setupRiskTestDB(t)
	now := time.Now()
	failed := createRiskScanRun(t, now.Add(-time.Hour))
	if err := global.GVA_DB.Model(&failed).Updates(map[string]any{"status": model.RiskScanStatusFailed, "heartbeat_at": now.Add(-time.Hour)}).Error; err != nil {
		t.Fatalf("mark historical scan failed: %v", err)
	}
	latest := createRiskScanRun(t, now)
	if err := global.GVA_DB.Model(&latest).Update("status", model.RiskScanStatusSuccess).Error; err != nil {
		t.Fatalf("mark latest scan successful: %v", err)
	}
	if _, err := Risk.prepareScanRun(model.RiskScanTriggerManual, 1, "测试用户", failed.ID); err == nil {
		t.Fatal("historical failed scan should not be resumable")
	}
}

func TestSimilarAssetIndexFindsEditDistanceOne(t *testing.T) {
	items := []similarAsset{{ID: 1, Code: "ASSET-10001"}, {ID: 2, Code: "ASSET-10002"}, {ID: 3, Code: "OTHER-90001"}}
	index := buildSimilarAssetIndex(items, 1, 5)
	if len(index[1]) != 1 || index[1][0].ID != 2 || len(index[2]) != 1 || index[2][0].ID != 1 {
		t.Fatalf("unexpected similar code index: %#v", index)
	}
	if len(index[3]) != 0 {
		t.Fatalf("unrelated code should not match: %#v", index[3])
	}
}

func TestValidateRiskRuleParametersRejectsUnknownAndUnsafeValues(t *testing.T) {
	tests := []struct {
		code       string
		parameters map[string]any
	}{
		{code: riskRulePendingTooLong, parameters: map[string]any{"days": 0}},
		{code: riskRuleFrequentMaintenance, parameters: map[string]any{"windowDays": 30, "count": 1}},
		{code: riskRuleSimilarCode, parameters: map[string]any{"maxDistance": 9, "minLength": 5}},
		{code: riskRuleDuplicateSerial, parameters: map[string]any{"unexpected": true}},
	}
	for _, test := range tests {
		if err := ValidateRiskRuleParameters(test.code, test.parameters); err == nil {
			t.Fatalf("expected validation error for %s %#v", test.code, test.parameters)
		}
	}
}
