package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/ai"
	"github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	assetService "github.com/WangMi2022/mit-assets-admin/server/plugin/asset/service"
	"github.com/spf13/viper"
)

// Opt-in provider integration: only reads the explicitly supplied runtime AI
// configuration. Business data and invocation audits use an in-memory fixture.
func TestAssetQueryConfiguredProvider(t *testing.T) {
	path := os.Getenv("SMART_QUERY_PROVIDER_CONFIG")
	if path == "" {
		t.Skip("set SMART_QUERY_PROVIDER_CONFIG to opt into configured provider calls")
	}
	db := setupSmartTestDB(t)
	if err := db.AutoMigrate(&ai.UsageQuota{}); err != nil {
		t.Fatal("prepare isolated quota table")
	}
	reader := viper.New()
	reader.SetConfigFile(path)
	if reader.ReadInConfig() != nil {
		t.Fatal("read explicitly supplied provider configuration")
	}
	var settings struct {
		AI config.AI `mapstructure:"ai"`
	}
	if reader.Unmarshal(&settings) != nil || !settings.AI.Enabled {
		t.Fatal("configured AI provider is unavailable")
	}
	global.GVA_CONFIG.AI = settings.AI
	planner := assetQueryPlanner{rules: NewRulePlanner(func() time.Time { return time.Date(2026, 9, 11, 10, 0, 0, 0, mustShanghaiLocation()) }), gateway: ai.Default, registry: NewToolRegistry(func(uint, string, string) bool { return true })}
	for _, tc := range []struct {
		name, question string
		fields         []string
		maintenance    bool
	}{
		{"composite", "行政部王磊名下超过5000元且2026年购入的电脑，按价值降序", []string{"custodian", "currentValue", "purchaseDate", "name"}, false},
		{"maintenance", "最近一年维修超过三次的设备", []string{"maintenanceCount"}, true},
		{"disjunction", "查找闲置或者当前估值低于1000元的资产", []string{"status", "currentValue"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := planner.Plan(ai.WithInternalActor(context.Background()), PlanRequest{Question: tc.question})
			if err != nil || plan.Clarification != "" || plan.Planner != llmPlannerName || len(plan.Calls) != 1 {
				t.Fatalf("configured provider did not produce a complete valid %s query", tc.name)
			}
			query, err := assetQueryFromCall(plan.Calls[0])
			if err != nil {
				t.Fatal("configured query failed validation")
			}
			for _, field := range tc.fields {
				if !assetQueryHasField(query.Where, field) {
					t.Fatalf("configured provider omitted required field %s", field)
				}
			}
			if tc.maintenance && (query.MaintenanceFrom != "2025-09-11" || query.MaintenanceTo != "2026-09-11" || !assetService.AssetQueryUsesMaintenance(query)) {
				t.Fatal("configured provider lost the maintenance date window")
			}
			if tc.name == "disjunction" && (query.Where == nil || len(query.Where.Any) != 2) {
				t.Fatal("configured provider lost OR semantics")
			}
			if tc.name == "composite" && (len(query.OrderBy) != 1 || query.OrderBy[0].Field != "currentValue" || query.OrderBy[0].Direction != "desc") {
				t.Fatal("configured provider lost requested sort")
			}
			t.Log("configured provider returned validated conditions for", tc.name)
		})
	}
}

func assetQueryHasField(filter *assetService.AssetFilter, field string) bool {
	if filter == nil {
		return false
	}
	if filter.Field == field {
		return true
	}
	for _, child := range append(append([]assetService.AssetFilter{}, filter.All...), filter.Any...) {
		if assetQueryHasField(&child, field) {
			return true
		}
	}
	return false
}
