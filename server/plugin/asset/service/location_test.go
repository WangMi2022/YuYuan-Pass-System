package service

import (
	"strings"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
	assetRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model/request"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupLocationTestDB(t *testing.T) {
	t.Helper()
	previous := global.GVA_DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&model.Location{}); err != nil {
		t.Fatalf("migrate location table: %v", err)
	}
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = previous })
}

func TestNormalizeLocation(t *testing.T) {
	location := model.Location{
		Name: "  资产仓库 A 区  ", Type: " INBOUND ", Code: " rk-a ", Description: "  常规入库区  ",
	}
	if err := normalizeLocation(&location); err != nil {
		t.Fatalf("normalize location: %v", err)
	}
	if location.Name != "资产仓库 A 区" || location.Type != model.LocationTypeInbound || location.Code != "RK-A" || location.Description != "常规入库区" {
		t.Fatalf("unexpected normalized location: %#v", location)
	}
}

func TestNormalizeLocationRejectsInvalidInput(t *testing.T) {
	tests := []model.Location{
		{Name: "  ", Type: model.LocationTypeInbound},
		{Name: "测试位置", Type: "unknown"},
	}
	for _, location := range tests {
		if err := normalizeLocation(&location); err == nil {
			t.Fatalf("expected validation error for %#v", location)
		}
	}
}

func TestLocationCreateEnforcesTypeAndNameUniqueness(t *testing.T) {
	setupLocationTestDB(t)
	first := model.Location{Name: " 资产仓库 A 区 ", Type: model.LocationTypeInbound, Enabled: true}
	if err := Location.Create(&first); err != nil {
		t.Fatalf("create first location: %v", err)
	}
	duplicate := model.Location{Name: "资产仓库 A 区", Type: model.LocationTypeInbound, Enabled: true}
	if err := Location.Create(&duplicate); err == nil || !strings.Contains(err.Error(), "同名位置") {
		t.Fatalf("expected duplicate location error, got %v", err)
	}
	otherType := model.Location{Name: "资产仓库 A 区", Type: model.LocationTypeReturn, Enabled: true}
	if err := Location.Create(&otherType); err != nil {
		t.Fatalf("same name in another type should be allowed: %v", err)
	}
}

func TestLocationOptionsOnlyReturnsEnabledLocationsInSortOrder(t *testing.T) {
	setupLocationTestDB(t)
	items := []model.Location{
		{Name: "B 区", Type: model.LocationTypeUsage, Sort: 20, Enabled: true},
		{Name: "停用区", Type: model.LocationTypeUsage, Sort: 5, Enabled: false},
		{Name: "A 区", Type: model.LocationTypeUsage, Sort: 10, Enabled: true},
		{Name: "其他类型", Type: model.LocationTypeInbound, Sort: 1, Enabled: true},
	}
	for index := range items {
		if err := Location.Create(&items[index]); err != nil {
			t.Fatalf("create location %q: %v", items[index].Name, err)
		}
	}
	options, err := Location.Options(model.LocationTypeUsage)
	if err != nil {
		t.Fatalf("get location options: %v", err)
	}
	if len(options) != 2 || options[0].Name != "A 区" || options[1].Name != "B 区" {
		t.Fatalf("unexpected location options: %#v", options)
	}

	list, total, err := Location.List(assetRequest.LocationSearch{Type: model.LocationTypeUsage})
	if err != nil {
		t.Fatalf("list managed locations: %v", err)
	}
	if total != 3 || len(list) != 3 {
		t.Fatalf("management list should include disabled locations: total=%d list=%d", total, len(list))
	}
}
