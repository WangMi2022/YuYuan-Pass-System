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

type operationSQLCounter struct {
	logger.Interface
	statements []string
}

func (c *operationSQLCounter) Trace(_ context.Context, _ time.Time, fc func() (string, int64), _ error) {
	sql, _ := fc()
	c.statements = append(c.statements, sql)
}

func (c *operationSQLCounter) reset() { c.statements = nil }

func setupOperationTestDB(t *testing.T) *operationSQLCounter {
	t.Helper()
	previous := global.GVA_DB
	counter := &operationSQLCounter{Interface: logger.Default.LogMode(logger.Silent)}
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{TranslateError: true, Logger: counter})
	if err != nil {
		t.Fatalf("open operation test database: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Category{}, &model.Asset{}, &model.AssetOperationOrder{},
		&model.AssetOperationItem{}, &model.AssetOperationRecord{},
	); err != nil {
		t.Fatalf("migrate operation tables: %v", err)
	}
	global.GVA_DB = db
	t.Cleanup(func() {
		global.GVA_DB = previous
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return counter
}

func createOperationTestAssets(t *testing.T, count int) []model.Asset {
	t.Helper()
	category := model.Category{Name: "办公设备", Code: fmt.Sprintf("OFFICE-%d", time.Now().UnixNano()), Enabled: true}
	if err := global.GVA_DB.Create(&category).Error; err != nil {
		t.Fatalf("create asset category: %v", err)
	}
	assets := make([]model.Asset, count)
	for i := range assets {
		assets[i] = model.Asset{
			AssetCode: fmt.Sprintf("TEST-%03d", i+1), Name: fmt.Sprintf("测试资产 %d", i+1),
			CategoryID: category.ID, Quantity: i + 1, Unit: "件", CurrentValue: float64((i + 1) * 100),
			Status: model.AssetStatusIdle, Location: fmt.Sprintf("原位置 %d", i+1),
		}
	}
	if err := global.GVA_DB.Create(&assets).Error; err != nil {
		t.Fatalf("create operation assets: %v", err)
	}
	return assets
}

func TestTransitionStatus(t *testing.T) {
	tests := []struct {
		name, operationType, from, want string
		wantErr                         bool
	}{
		{name: "inbound pending asset", operationType: "inbound", from: "pending_inbound", want: "idle"},
		{name: "reject repeated inbound", operationType: "inbound", from: "idle", wantErr: true},
		{name: "issue idle asset", operationType: "issue", from: "idle", want: "in_use"},
		{name: "return used asset", operationType: "return", from: "in_use", want: "idle"},
		{name: "return repaired asset", operationType: "return", from: "maintenance", want: "idle"},
		{name: "maintenance used asset", operationType: "maintenance", from: "in_use", want: "maintenance"},
		{name: "scrap maintenance asset", operationType: "scrap", from: "maintenance", want: "retired"},
		{name: "reject retired asset", operationType: "issue", from: "retired", wantErr: true},
		{name: "reject invalid operation", operationType: "unknown", from: "idle", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transitionStatus(tt.operationType, tt.from)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got status %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSubmitOperationBatchesAssetChangesAndSnapshots(t *testing.T) {
	counter := setupOperationTestDB(t)
	assets := createOperationTestAssets(t, 3)
	order, err := Operation.Create(assetRequest.SaveOperation{
		Type: "issue", TargetLocation: "研发部", TargetCustodian: "张三",
		AssetIDs: []uint{assets[2].ID, assets[0].ID, assets[1].ID},
	}, 1, "创建人")
	if err != nil {
		t.Fatalf("create draft operation: %v", err)
	}

	counter.reset()
	if err := Operation.Submit(order.ID, 2, "提交人"); err != nil {
		t.Fatalf("submit operation: %v", err)
	}
	if len(counter.statements) > 8 {
		t.Fatalf("submit executed %d SQL statements, want at most 8: %v", len(counter.statements), counter.statements)
	}
	lockedInIDOrder := false
	for _, statement := range counter.statements {
		normalized := strings.ToLower(statement)
		if strings.Contains(normalized, "from `assets`") && strings.Contains(normalized, "order by id asc") {
			lockedInIDOrder = true
			break
		}
	}
	if !lockedInIDOrder {
		t.Fatalf("asset lock query is not ordered by ID: %v", counter.statements)
	}

	var persistedOrder model.AssetOperationOrder
	if err := global.GVA_DB.First(&persistedOrder, order.ID).Error; err != nil {
		t.Fatalf("reload operation order: %v", err)
	}
	if persistedOrder.Status != model.OperationStatusCompleted || persistedOrder.CompletedBy != 2 || persistedOrder.CompletedAt == nil {
		t.Fatalf("unexpected completed order: %#v", persistedOrder)
	}
	var persistedAssets []model.Asset
	if err := global.GVA_DB.Order("id ASC").Find(&persistedAssets).Error; err != nil {
		t.Fatalf("reload assets: %v", err)
	}
	for _, asset := range persistedAssets {
		if asset.Status != model.AssetStatusInUse || asset.Location != "研发部" || asset.Custodian != "张三" {
			t.Fatalf("asset was not transitioned: %#v", asset)
		}
	}
	var items []model.AssetOperationItem
	if err := global.GVA_DB.Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
		t.Fatalf("reload operation items: %v", err)
	}
	if len(items) != len(assets) {
		t.Fatalf("got %d items, want %d", len(items), len(assets))
	}
	for _, item := range items {
		if item.FromStatus != model.AssetStatusIdle || item.ToStatus != model.AssetStatusInUse || item.ToLocation != "研发部" || item.ToCustodian != "张三" {
			t.Fatalf("unexpected item snapshot: %#v", item)
		}
	}
	var recordCount int64
	if err := global.GVA_DB.Model(&model.AssetOperationRecord{}).Where("order_id = ?", order.ID).Count(&recordCount).Error; err != nil {
		t.Fatalf("count operation records: %v", err)
	}
	if recordCount != int64(len(assets)) {
		t.Fatalf("got %d operation records, want %d", recordCount, len(assets))
	}
}

func TestSubmitOperationRollsBackWhenAnyAssetIsInvalid(t *testing.T) {
	setupOperationTestDB(t)
	assets := createOperationTestAssets(t, 2)
	order, err := Operation.Create(assetRequest.SaveOperation{
		Type: "issue", TargetCustodian: "李四", AssetIDs: []uint{assets[0].ID, assets[1].ID},
	}, 1, "创建人")
	if err != nil {
		t.Fatalf("create draft operation: %v", err)
	}
	if err := global.GVA_DB.Model(&model.Asset{}).Where("id = ?", assets[1].ID).Update("status", model.AssetStatusRetired).Error; err != nil {
		t.Fatalf("retire second asset: %v", err)
	}

	err = Operation.Submit(order.ID, 2, "提交人")
	if err == nil || !strings.Contains(err.Error(), "当前状态不允许执行领用") {
		t.Fatalf("expected invalid transition error, got %v", err)
	}
	var persistedOrder model.AssetOperationOrder
	if err := global.GVA_DB.First(&persistedOrder, order.ID).Error; err != nil {
		t.Fatalf("reload operation order: %v", err)
	}
	if persistedOrder.Status != model.OperationStatusDraft || persistedOrder.CompletedAt != nil {
		t.Fatalf("invalid submission changed order: %#v", persistedOrder)
	}
	var firstAsset model.Asset
	if err := global.GVA_DB.First(&firstAsset, assets[0].ID).Error; err != nil {
		t.Fatalf("reload first asset: %v", err)
	}
	if firstAsset.Status != model.AssetStatusIdle || firstAsset.Custodian != "" {
		t.Fatalf("invalid submission partially changed asset: %#v", firstAsset)
	}
	var recordCount int64
	if err := global.GVA_DB.Model(&model.AssetOperationRecord{}).Where("order_id = ?", order.ID).Count(&recordCount).Error; err != nil {
		t.Fatalf("count operation records: %v", err)
	}
	if recordCount != 0 {
		t.Fatalf("invalid submission created %d operation records", recordCount)
	}
}
