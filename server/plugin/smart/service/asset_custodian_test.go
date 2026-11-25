package service

import (
	"context"
	"testing"

	assetModel "github.com/WangMi2022/mit-assets-admin/server/plugin/asset/model"
)

func TestExtractAssetCustodianBoundaries(t *testing.T) {
	cases := map[string]string{
		"行政部的王磊 名下都认领了哪些资产": "行政部-王磊",
		"王磊名下有哪些资产":         "王磊",
		"保管人是王磊的资产有哪些":      "王磊",
		"按保管人汇总资产":          "",
		"各保管人有哪些资产":         "",
	}
	for question, want := range cases {
		if got := extractAssetCustodian(question); got != want {
			t.Errorf("extractAssetCustodian(%q) = %q, want %q", question, got, want)
		}
	}
}

func TestCopilotFindsAssetsHeldByNamedCustodian(t *testing.T) {
	database := setupSmartTestDB(t)
	previousRegistry := defaultToolRegistry
	defaultToolRegistry = NewToolRegistry(func(_ uint, path, _ string) bool { return path == "/asset/list" })
	t.Cleanup(func() { defaultToolRegistry = previousRegistry })

	assets := []assetModel.Asset{
		{AssetCode: "CUSTODIAN-001", Name: "会议椅", Custodian: "行政部-王磊", Quantity: 23},
		{AssetCode: "CUSTODIAN-002", Name: "新能源公务车", Custodian: "行政部-王磊", Quantity: 2},
		{AssetCode: "CUSTODIAN-003", Name: "会议交互大屏", Custodian: "行政部-王磊", Quantity: 1},
		{AssetCode: "CUSTODIAN-004", Name: "12 人会议桌", Custodian: "行政部-王磊", Quantity: 2},
		{AssetCode: "CUSTODIAN-OTHER-DEPT", Name: "笔记本电脑", Custodian: "技术部-王磊", Quantity: 1},
		{AssetCode: "CUSTODIAN-OTHER-PERSON", Name: "打印机", Custodian: "行政部-李明", Quantity: 1},
		{AssetCode: "CUSTODIAN-NAME-MENTION", Name: "行政部-王磊会议室电脑", Custodian: "技术部-李明", Quantity: 1},
	}
	if err := database.Create(&assets).Error; err != nil {
		t.Fatal(err)
	}

	for question, expected := range map[string]struct {
		custodian string
		total     int64
	}{
		"行政部的王磊 名下都认领了哪些资产": {custodian: "行政部-王磊", total: 4},
		"行政部的王磊 名下都有哪些资产":   {custodian: "行政部-王磊", total: 4},
		"行政部的王磊领用了哪些资产？":    {custodian: "行政部-王磊", total: 4},
		"王磊名下有哪些资产":         {custodian: "王磊", total: 5},
	} {
		t.Run(question, func(t *testing.T) {
			plan, err := NewRulePlanner(nil).Plan(context.Background(), PlanRequest{Question: question})
			if err != nil || len(plan.Calls) != 1 || plan.Calls[0].Name != "asset.search" || stringArgument(plan.Calls[0].Arguments, "custodian") != expected.custodian {
				t.Fatalf("expected an asset search for custodian %s, got plan=%#v err=%v", expected.custodian, plan, err)
			}
			result, err := Smart.Query(context.Background(), 1, 888, question, 0)
			if err != nil {
				t.Fatalf("query: %v", err)
			}
			data, ok := result.Data.(map[string]any)
			if !ok || data["total"] != expected.total {
				t.Fatalf("expected %d assets for %s, got tools=%v data=%#v answer=%q", expected.total, expected.custodian, result.Tools, result.Data, result.Answer)
			}
			list, ok := data["list"].([]assetModel.Asset)
			if !ok || int64(len(list)) != expected.total || int64(len(result.Citations)) != expected.total {
				t.Fatalf("expected %d asset records and citations, got list=%#v citations=%#v", expected.total, data["list"], result.Citations)
			}
			for _, item := range list {
				if expected.custodian == "行政部-王磊" && item.Custodian != expected.custodian {
					t.Fatalf("unrelated asset returned: %s (%s)", item.AssetCode, item.Custodian)
				}
				if expected.custodian == "王磊" && (item.Custodian != "行政部-王磊" && item.Custodian != "技术部-王磊") {
					t.Fatalf("unrelated asset returned for bare name: %s (%s)", item.AssetCode, item.Custodian)
				}
			}
		})
	}
}
