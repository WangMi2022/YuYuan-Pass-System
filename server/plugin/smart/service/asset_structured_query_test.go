package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/ai"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	assetModel "github.com/WangMi2022/mit-assets-admin/server/plugin/asset/model"
	assetService "github.com/WangMi2022/mit-assets-admin/server/plugin/asset/service"
	smartModel "github.com/WangMi2022/mit-assets-admin/server/plugin/smart/model"
)

func structuredQueryFixture(t *testing.T) time.Time {
	t.Helper()
	db := setupSmartTestDB(t)
	if err := db.AutoMigrate(&assetModel.AssetOperationRecord{}); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 11, 23, 30, 0, 0, mustShanghaiLocation())
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday, tomorrow := today.AddDate(0, 0, -1), today.AddDate(0, 0, 1)
	category := assetModel.Category{Name: "办公设备", Code: "STRUCTURED-OFFICE", Enabled: true}
	if err := db.Create(&category).Error; err != nil {
		t.Fatal(err)
	}
	items := []assetModel.Asset{
		{AssetCode: "QUERY-EXPIRED", Name: "笔记本电脑", CategoryID: category.ID, Custodian: "行政部-王磊", Quantity: 3, CurrentValue: 6000, Status: "in_use", WarrantyEndDate: &yesterday},
		{AssetCode: "QUERY-TODAY", Name: "台式电脑", CategoryID: category.ID, Custodian: "行政部-王磊", Quantity: 2, CurrentValue: 8000, Status: "idle", WarrantyEndDate: &today},
		{AssetCode: "QUERY-FUTURE", Name: "打印机", CategoryID: category.ID, Custodian: "行政部-李明", Quantity: 1, CurrentValue: 2000, Status: "idle", WarrantyEndDate: &tomorrow},
		{AssetCode: "QUERY-UNKNOWN", Name: "未登记质保电脑", CategoryID: category.ID, Custodian: "技术部-王磊", Quantity: 1, CurrentValue: 9000, Status: "in_use"},
		{AssetCode: "QUERY-OTHER-DEPT", Name: "电脑", CategoryID: category.ID, Custodian: "技术部-王磊", Quantity: 1, CurrentValue: 7000, Status: "in_use", WarrantyEndDate: &yesterday},
	}
	if err := db.Create(&items).Error; err != nil {
		t.Fatal(err)
	}
	return now
}

func TestStructuredAssetQueryCombinesConditionsAndKeepsTotalsAcrossPages(t *testing.T) {
	now := structuredQueryFixture(t)
	query := assetService.AssetQuery{Where: &assetService.AssetFilter{All: []assetService.AssetFilter{
		{Field: "custodian", Op: "eq", Value: "行政部-王磊"},
		{Any: []assetService.AssetFilter{{Field: "status", Op: "eq", Value: "in_use"}, {Field: "status", Op: "eq", Value: "idle"}}},
		{Field: "currentValue", Op: "gte", Value: 6000},
		{Field: "warrantyEndDate", Op: "lte", Value: now.Format("2006-01-02")},
	}}, OrderBy: []assetService.AssetSort{{Field: "currentValue", Direction: "desc"}}, PageSize: 1}
	first, err := assetService.Asset.Query(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}
	if first.Total != 2 || first.Quantity != 5 || first.CurrentValue != 14000 || len(first.List) != 1 || first.List[0].AssetCode != "QUERY-TODAY" {
		t.Fatalf("wrong totals/page: %#v", first)
	}
	query.Page = 2
	second, err := assetService.Asset.Query(context.Background(), query)
	if err != nil || second.Total != 2 || len(second.List) != 1 || second.List[0].AssetCode != "QUERY-EXPIRED" {
		t.Fatalf("page 2: %#v, %v", second, err)
	}
	query.Page = 0
	query.PageSize = 0
	query.OrderBy = nil
	query.GroupBy = "status"
	grouped, err := assetService.Asset.Query(context.Background(), query)
	if err != nil || grouped.Total != 2 || grouped.GroupTotal != 2 || len(grouped.Groups) != 2 {
		t.Fatalf("groups: %#v, %v", grouped, err)
	}
}

func TestStructuredAssetQueryDatesTreatTodayAndUnknownSeparately(t *testing.T) {
	now := structuredQueryFixture(t)
	for _, tc := range []struct {
		op    string
		value any
		want  int64
	}{{"lt", now.Format("2006-01-02"), 2}, {"gte", now.Format("2006-01-02"), 2}, {"isNull", nil, 1}} {
		result, err := assetService.Asset.Query(context.Background(), assetService.AssetQuery{Where: &assetService.AssetFilter{Field: "warrantyEndDate", Op: tc.op, Value: tc.value}})
		if err != nil || result.Total != tc.want {
			t.Fatalf("date %s: total=%d, err=%v", tc.op, result.Total, err)
		}
	}
	result, err := assetService.Asset.Query(context.Background(), assetService.AssetQuery{Where: &assetService.AssetFilter{Field: "department", Op: "eq", Value: "行政部"}})
	if err != nil || result.Total != 3 {
		t.Fatalf("department: %#v, %v", result, err)
	}
	result, err = assetService.Asset.Query(context.Background(), assetService.AssetQuery{Where: &assetService.AssetFilter{Field: "category", Op: "contains", Value: "办公"}, GroupBy: "category"})
	if err != nil || result.Total != 5 || len(result.Groups) != 1 || result.Groups[0].Key != "办公设备" {
		t.Fatalf("category: %#v, %v", result, err)
	}
}

func TestStructuredAssetQueryEscapesWildcardsAndRejectsUnsafeQueries(t *testing.T) {
	structuredQueryFixture(t)
	for _, value := range []string{"%", "_", "' OR 1=1 --", "行政部%"} {
		result, err := assetService.Asset.Query(context.Background(), assetService.AssetQuery{Where: &assetService.AssetFilter{Field: "name", Op: "contains", Value: value}})
		if err != nil || result.Total != 0 {
			t.Fatalf("literal %q: %#v, %v", value, result, err)
		}
	}
	for _, query := range []assetService.AssetQuery{
		{Where: &assetService.AssetFilter{Field: "tenantId", Op: "eq", Value: 1}},
		{Where: &assetService.AssetFilter{Field: "name; DROP TABLE assets", Op: "eq", Value: "x"}},
		{Where: &assetService.AssetFilter{Field: "status", Op: "eq", Value: "expired"}},
		{Where: &assetService.AssetFilter{Field: "warrantyEndDate", Op: "lt", Value: "2026-02-30"}},
		{Where: &assetService.AssetFilter{Field: "currentValue", Op: "gt", Value: "5000"}},
		{Where: &assetService.AssetFilter{Field: "currentValue", Op: "gt", Value: nil}},
		{Where: &assetService.AssetFilter{Field: "department", Op: "contains", Value: "行政"}},
		{Where: &assetService.AssetFilter{}},
		{Where: &assetService.AssetFilter{All: []assetService.AssetFilter{{Field: "name", Op: "eq", Value: "a"}}, Field: "status", Op: "eq", Value: "idle"}},
		{OrderBy: []assetService.AssetSort{{Field: "name", Direction: "desc; SELECT 1"}}},
		{PageSize: 101},
		{MaintenanceFrom: "2026-01-01"},
	} {
		if _, err := assetService.Asset.Query(context.Background(), query); err == nil {
			t.Fatalf("unsafe query accepted: %#v", query)
		}
	}
}

func TestStructuredAssetQueryCountsCompletedMaintenanceWithinBusinessDates(t *testing.T) {
	structuredQueryFixture(t)
	var asset assetModel.Asset
	if err := global.GVA_DB.Where("asset_code = ?", "QUERY-EXPIRED").First(&asset).Error; err != nil {
		t.Fatal(err)
	}
	for index, fixture := range []struct{ date, status string }{{"2026-03-01", "completed"}, {"2026-04-01", "completed"}, {"2026-05-01", "completed"}, {"2025-03-01", "completed"}, {"2026-06-01", "draft"}} {
		date, _ := time.Parse("2006-01-02", fixture.date)
		order := assetModel.AssetOperationOrder{OrderNo: fmt.Sprintf("QUERY-MAINT-%d", index), Type: "maintenance", Status: fixture.status, BusinessDate: date}
		if err := global.GVA_DB.Create(&order).Error; err != nil {
			t.Fatal(err)
		}
		// Duplicate evidence for a single order must not inflate the repair count.
		for duplicate := 0; duplicate < 2; duplicate++ {
			record := assetModel.AssetOperationRecord{OrderID: order.ID, OrderNo: order.OrderNo, Type: "maintenance", AssetID: asset.ID, AssetCode: asset.AssetCode, AssetName: asset.Name, Quantity: 1, OperatedAt: date}
			if err := global.GVA_DB.Create(&record).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	query := assetService.AssetQuery{Where: &assetService.AssetFilter{Field: "maintenanceCount", Op: "eq", Value: 3}, MaintenanceFrom: "2026-01-01", MaintenanceTo: "2026-12-31"}
	result, err := assetService.Asset.Query(context.Background(), query)
	if err != nil || result.Total != 1 || result.List[0].ID != asset.ID {
		t.Fatalf("maintenance: %#v, %v", result, err)
	}
	registry := NewToolRegistry(func(_ uint, path, _ string) bool { return path == "/asset/list" })
	if _, err := registry.Execute(context.Background(), Smart, AssistantActor{UserID: 1, AuthorityID: 888}, ToolCall{Name: "asset.search", Arguments: map[string]any{"query": query}}); err == nil {
		t.Fatal("maintenance filter bypassed operation permission")
	}
}

func TestStructuredAssetRuleQueries(t *testing.T) {
	now := structuredQueryFixture(t)
	planner := NewRulePlanner(func() time.Time { return now })
	for question, want := range map[string]int64{
		"有哪些已经过保的资产":                    2,
		"哪些资产质保已过期":                     2,
		"查询未过保的资产":                      2,
		"查询未来 1 天内质保到期的资产":              2,
		"查询行政部的资产":                      3,
		"行政部王磊名下已经过保且价值超过5000元的电脑":      1,
		"查询行政部的王磊名下价值不超过8000元的电脑，按金额降序": 2,
		"查询所有闲置资产":                      2,
		"按保管人汇总资产数量和价值":                 5,
	} {
		t.Run(question, func(t *testing.T) {
			plan, err := planner.Plan(context.Background(), PlanRequest{Question: question})
			if err != nil || plan.Clarification != "" || len(plan.Calls) != 1 {
				t.Fatalf("plan: %#v, %v", plan, err)
			}
			result, err := Smart.executeRegisteredTool(context.Background(), AssistantActor{UserID: 1, AuthorityID: 888}, plan.Calls[0])
			if err != nil || result.Data.(map[string]any)["total"] != want {
				t.Fatalf("result: %#v, err=%v, want=%d", result, err, want)
			}
		})
	}
}

func TestStructuredAssetQueriesWithoutQueryVerbRemainReadOnly(t *testing.T) {
	for _, question := range []string{"最近一年维修超过三次的设备", "维修中的资产", "已报废资产", "王磊已领用的设备"} {
		if isWriteIntent(question) {
			t.Fatalf("read-only noun phrase rejected: %s", question)
		}
	}
	for _, question := range []string{"帮我维修资产12", "立即报废这台资产", "删除这张发票", "修改资产12"} {
		if !isWriteIntent(question) {
			t.Fatalf("write command accepted: %s", question)
		}
	}
}

func TestStructuredAssetQueryClarifiesBeforeCountingAndKeepsFollowupScope(t *testing.T) {
	structuredQueryFixture(t)
	oldRegistry := defaultToolRegistry
	defaultToolRegistry = NewToolRegistry(func(uint, string, string) bool { return true })
	t.Cleanup(func() { defaultToolRegistry = oldRegistry })
	first, err := Smart.Query(context.Background(), 1, 888, "行政部的王磊名下有哪些过期资产", 0)
	if err != nil || first.Intent != "clarification" || len(first.Tools) != 0 || strings.Contains(first.Answer, "0 项") {
		t.Fatalf("clarification: %#v, %v", first, err)
	}
	if len(first.ClarificationOptions) != 0 {
		t.Fatalf("rule fallback invented clarification options: %#v", first.ClarificationOptions)
	}
	var runs []smartModel.CopilotRun
	if err := global.GVA_DB.Where("session_id = ?", first.SessionID).Find(&runs).Error; err != nil || len(runs) != 1 || runs[0].Status != "clarification" {
		t.Fatalf("run: %#v, %v", runs, err)
	}
	// Use a new query planner clock by fixing all expiry dates relative to runtime.
	yesterday := time.Now().In(mustShanghaiLocation()).AddDate(0, 0, -1)
	if err := global.GVA_DB.Model(&assetModel.Asset{}).Where("asset_code = ?", "QUERY-EXPIRED").Update("warranty_end_date", yesterday).Error; err != nil {
		t.Fatal(err)
	}
	second, err := Smart.Query(context.Background(), 1, 888, "是的", first.SessionID)
	if err != nil || second.Intent == "clarification" || !strings.Contains(second.Answer, "行政部-王磊") {
		t.Fatalf("followup: %#v, %v", second, err)
	}
	if _, err := Smart.Query(context.Background(), 2, 888, "质保", first.SessionID); err == nil {
		t.Fatal("cross-user session accessed")
	}
}

func TestAssetClarificationRecognizesAffirmativeWarrantyReplies(t *testing.T) {
	for _, reply := range []string{"是的", "对", "对的", "没错", "是质保"} {
		if !isWarrantyClarificationReply(reply) {
			t.Errorf("reply %q was not recognized as warranty clarification", reply)
		}
	}
}

func TestRiskToolReturnsUserFacingRows(t *testing.T) {
	db := setupSmartTestDB(t)
	category := assetModel.Category{Name: "办公设备", Code: "RISK-OFFICE", Enabled: true}
	if err := db.Create(&category).Error; err != nil {
		t.Fatal(err)
	}
	asset := assetModel.Asset{AssetCode: "RISK-001", Name: "会议室投影仪", CategoryID: category.ID, Custodian: "行政部-王磊", Status: assetModel.AssetStatusInUse, Quantity: 1}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatal(err)
	}
	event := assetModel.AssetRiskEvent{Fingerprint: "risk-user-facing", AssetID: asset.ID, RuleCode: "WARRANTY_EXPIRED", RuleVersion: 1, Category: "warranty", Severity: assetModel.RiskSeverityHigh, Status: assetModel.RiskStatusAcknowledged, Title: "资产质保已过期", Description: "质保日期早于今天", FirstDetectedAt: time.Now(), LastDetectedAt: time.Now(), LastScanRunID: 1}
	if err := db.Create(&event).Error; err != nil {
		t.Fatal(err)
	}
	result, err := Smart.executeRegisteredTool(context.Background(), AssistantActor{UserID: 1, AuthorityID: 888}, ToolCall{Name: "asset.risk.list"})
	if err != nil {
		t.Fatal(err)
	}
	data := result.Data.(map[string]any)
	rows, ok := data["list"].([]assetRiskQueryRow)
	if !ok || len(rows) != 1 || rows[0].AssetCode != "RISK-001" || rows[0].AssetName != "会议室投影仪" || rows[0].Custodian != "行政部-王磊" {
		t.Fatalf("unexpected user-facing risk rows: %#v", data["list"])
	}
}

func TestWarrantyToolUsesRequestedWindowAndFullCount(t *testing.T) {
	db := setupSmartTestDB(t)
	today := time.Now().In(mustShanghaiLocation())
	inside, outside := today.AddDate(0, 0, 45), today.AddDate(0, 0, 70)
	items := make([]assetModel.Asset, 61)
	for index := range items {
		items[index] = assetModel.Asset{AssetCode: fmt.Sprintf("QUERY-WARRANTY-%d", index), Name: "设备", Quantity: 1, WarrantyEndDate: &inside}
	}
	items[60].WarrantyEndDate = &outside
	if err := db.Create(&items).Error; err != nil {
		t.Fatal(err)
	}
	result, err := Smart.executeRegisteredTool(context.Background(), AssistantActor{UserID: 1, AuthorityID: 888}, ToolCall{Name: "asset.warranty.expiring", Arguments: map[string]any{"days": 60}})
	if err != nil {
		t.Fatal(err)
	}
	data := result.Data.(map[string]any)
	if data["total"] != int64(60) || len(data["list"].([]assetModel.Asset)) != 20 {
		t.Fatalf("truncated count: %#v", data)
	}
}

type structuredQueryGateway struct {
	content string
	err     error
	calls   int
	request ai.CompletionRequest
}

func (g *structuredQueryGateway) Complete(_ context.Context, request ai.CompletionRequest) (ai.CompletionResult, error) {
	g.calls++
	g.request = request
	return ai.CompletionResult{Content: g.content}, g.err
}
func (*structuredQueryGateway) Vision(context.Context, ai.VisionRequest) (ai.VisionResult, error) {
	return ai.VisionResult{}, nil
}
func (*structuredQueryGateway) Stream(context.Context, ai.CompletionRequest) (ai.StreamResult, error) {
	return ai.StreamResult{}, nil
}

func TestStructuredAssetModelPlansValidatedQueryAndCannotRewriteFacts(t *testing.T) {
	now := structuredQueryFixture(t)
	global.GVA_CONFIG.AI.Enabled = true
	gateway := &structuredQueryGateway{content: `{"query":{"where":{"all":[{"field":"purchaseDate","op":"gte","value":"2026-01-01"},{"any":[{"field":"status","op":"eq","value":"idle"},{"field":"currentValue","op":"gt","value":5000}]}]},"orderBy":[{"field":"currentValue","direction":"desc"}]},"clarification":"","unhandled":[]}`}
	registry := NewToolRegistry(func(uint, string, string) bool { return true })
	planner := assetQueryPlanner{rules: NewRulePlanner(func() time.Time { return now }), gateway: gateway, registry: registry}
	result, err := NewAssistantOrchestrator(Smart, planner, registry).Ask(context.Background(), AssistantActor{UserID: 1, AuthorityID: 888}, "今年购入且闲置或者估值超过5000元的资产，按估值降序")
	if err != nil || result.Plan.Planner != llmPlannerName || result.Plan.Clarification != "" || gateway.calls != 1 || !result.ModelUsed {
		t.Fatalf("model result: %#v, calls=%d, %v", result, gateway.calls, err)
	}
	if !strings.Contains(result.Answer, "购置日期") || gateway.request.Operation != "asset-query-plan" || gateway.request.PermissionPath != "/smart/copilot/query" {
		t.Fatalf("facts/context lost: %#v", result)
	}
}

func TestStructuredAssetModelOptionsAreSanitizedAndReturned(t *testing.T) {
	now := structuredQueryFixture(t)
	global.GVA_CONFIG.AI.Enabled = true
	gateway := &structuredQueryGateway{content: `{"query":null,"clarification":"请明确要按哪种日期筛选。","unhandled":[],"options":[{"key":"X","label":"按合同到期","value":"合同已到期","description":"按合同到期日筛选"},{"key":"Y","label":"按合同到期","value":"合同已到期"},{"key":"Z","label":"按服务到期","value":"服务已到期"}]}`}
	registry := NewToolRegistry(func(uint, string, string) bool { return true })
	planner := assetQueryPlanner{rules: NewRulePlanner(func() time.Time { return now }), gateway: gateway, registry: registry}
	result, err := NewAssistantOrchestrator(Smart, planner, registry).Ask(context.Background(), AssistantActor{UserID: 1, AuthorityID: 888}, "行政部王磊名下有哪些过期资产")
	if err != nil || result.Plan.Clarification == "" || len(result.Plan.ClarificationOptions) != 2 {
		t.Fatalf("model options: %#v, %v", result, err)
	}
	if got := result.Plan.ClarificationOptions; got[0].Key != "A" || got[1].Key != "B" || got[1].Label != "按服务到期" {
		t.Fatalf("normalized options: %#v", got)
	}
	data, ok := result.Data.(map[string]any)
	if !ok || len(data["clarificationOptions"].([]smartModel.ClarificationOption)) != 2 {
		t.Fatalf("response options: %#v", result.Data)
	}
}

func TestStructuredAssetModelFailureOrUnknownConditionsNeverReturnFalseZero(t *testing.T) {
	now := structuredQueryFixture(t)
	global.GVA_CONFIG.AI.Enabled = true
	for _, gateway := range []*structuredQueryGateway{
		{err: errors.New("provider unavailable")},
		{content: `{"query":{"where":{"field":"usefulLife","op":"gt","value":5}},"clarification":"","unhandled":[]}`},
		{content: `{"query":{},"clarification":"","unhandled":[]}`},
		{content: `{"query":{"where":{"field":"name","op":"eq","value":"电脑"}},"clarification":"","unhandled":["维修费用"]}`},
		{content: `{"query":{"sql":"SELECT * FROM assets"},"clarification":"","unhandled":[]}`},
	} {
		registry := NewToolRegistry(func(uint, string, string) bool { return true })
		planner := assetQueryPlanner{rules: NewRulePlanner(func() time.Time { return now }), gateway: gateway, registry: registry}
		result, err := NewAssistantOrchestrator(Smart, planner, registry).Ask(context.Background(), AssistantActor{UserID: 1, AuthorityID: 888}, "查询今年维修费用超过5000元的资产")
		if err != nil || result.Plan.Clarification == "" || len(result.Executions) != 0 || strings.Contains(result.Answer, "匹配 0") {
			t.Fatalf("unsafe fallback: %#v, %v", result, err)
		}
	}
	encoded, _ := json.Marshal(assetService.AssetQuery{Where: &assetService.AssetFilter{Field: "currentValue", Op: "gt", Value: 1000}})
	var decoded assetService.AssetQuery
	if err := decodeQueryJSON(string(encoded)+` {}`, &decoded); err == nil {
		t.Fatal("accepted trailing JSON")
	}
}

func TestStructuredAssetPagingReusesSavedFilters(t *testing.T) {
	now := structuredQueryFixture(t)
	oldRegistry := defaultToolRegistry
	defaultToolRegistry = NewToolRegistry(func(uint, string, string) bool { return true })
	t.Cleanup(func() { defaultToolRegistry = oldRegistry })
	first, err := Smart.Query(context.Background(), 1, 888, "行政部王磊名下有哪些资产", 0)
	if err != nil {
		t.Fatal(err)
	}
	var run smartModel.CopilotRun
	if err := global.GVA_DB.Where("session_id = ?", first.SessionID).First(&run).Error; err != nil {
		t.Fatal(err)
	}
	query := assetService.AssetQuery{Where: &assetService.AssetFilter{Field: "custodian", Op: "eq", Value: "行政部-王磊"}, PageSize: 1, OrderBy: []assetService.AssetSort{{Field: "currentValue", Direction: "desc"}}}
	call := ToolCall{Name: "asset.search", Intent: "asset", Arguments: map[string]any{"query": query}}
	run.PlannedTools = map[string]any{"0": call}
	if err := global.GVA_DB.Model(&run).Select("PlannedTools").Updates(&run).Error; err != nil {
		t.Fatal(err)
	}
	second, err := Smart.Query(context.Background(), 1, 888, "下一页", first.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	data := second.Data.(map[string]any)
	if data["page"] != 2 || data["total"] != int64(2) || len(data["list"].([]assetModel.Asset)) != 1 || data["list"].([]assetModel.Asset)[0].AssetCode != "QUERY-EXPIRED" {
		t.Fatalf("paging scope changed: %#v", data)
	}
	plan := assetQueryPlanner{rules: NewRulePlanner(func() time.Time { return now }), gateway: &structuredQueryGateway{content: `{"query":{},"clarification":"","unhandled":[]}`}, registry: defaultToolRegistry}
	global.GVA_CONFIG.AI.Enabled = true
	unsafe, err := plan.Plan(context.Background(), PlanRequest{Actor: AssistantActor{UserID: 1, AuthorityID: 888}, Question: "查询所有今年购置的资产"})
	if err != nil || unsafe.Clarification == "" || len(unsafe.Calls) != 0 {
		t.Fatalf("empty query silently dropped constraints: %#v, %v", unsafe, err)
	}
}
