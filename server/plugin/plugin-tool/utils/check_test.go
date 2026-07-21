package utils

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type metadataSQLCounter struct {
	logger.Interface
	statements []string
}

func (c *metadataSQLCounter) Trace(_ context.Context, _ time.Time, fc func() (string, int64), _ error) {
	sql, _ := fc()
	c.statements = append(c.statements, sql)
}

func (c *metadataSQLCounter) reset() { c.statements = nil }

func setupMetadataTestDB(t *testing.T) (*gorm.DB, *metadataSQLCounter) {
	t.Helper()
	counter := &metadataSQLCounter{Interface: logger.Default.LogMode(logger.Silent)}
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		TranslateError: true, Logger: counter, DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open metadata test database: %v", err)
	}
	if err := db.AutoMigrate(
		&system.SysApi{}, &system.SysBaseMenu{},
		&system.SysDictionary{}, &system.SysDictionaryDetail{},
	); err != nil {
		t.Fatalf("migrate metadata tables: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return db, counter
}

func TestRegisterApisBatchesMissingMetadata(t *testing.T) {
	db, counter := setupMetadataTestDB(t)
	apis := []system.SysApi{
		{Path: "/asset/list", Method: "GET", ApiGroup: "资产"},
		{Path: "/asset/create", Method: "POST", ApiGroup: "资产"},
		{Path: "/asset/list", Method: "POST", ApiGroup: "资产"},
		{Path: "/asset/list", Method: "GET", ApiGroup: "资产"},
	}

	counter.reset()
	if err := registerApis(db, apis); err != nil {
		t.Fatalf("register APIs: %v", err)
	}
	if len(counter.statements) > 2 {
		t.Fatalf("initial API registration executed %d SQL statements, want at most 2: %v", len(counter.statements), counter.statements)
	}
	var count int64
	if err := db.Model(&system.SysApi{}).Count(&count).Error; err != nil {
		t.Fatalf("count APIs: %v", err)
	}
	if count != 3 {
		t.Fatalf("registered %d APIs, want 3", count)
	}

	counter.reset()
	if err := registerApis(db, apis); err != nil {
		t.Fatalf("repeat API registration: %v", err)
	}
	if len(counter.statements) != 1 {
		t.Fatalf("repeat API registration executed %d SQL statements, want 1: %v", len(counter.statements), counter.statements)
	}
}

func TestRegisterMenusBatchesChildrenAndPreservesParent(t *testing.T) {
	db, counter := setupMetadataTestDB(t)
	menus := []system.SysBaseMenu{
		{Path: "asset", Name: "asset", Meta: system.Meta{Title: "资产管理"}},
		{Path: "list", Name: "assetList", Meta: system.Meta{Title: "资产列表"}},
		{Path: "category", Name: "assetCategory", Meta: system.Meta{Title: "资产分类"}},
	}

	counter.reset()
	if err := registerMenus(db, menus); err != nil {
		t.Fatalf("register menus: %v", err)
	}
	if len(counter.statements) > 4 {
		t.Fatalf("initial menu registration executed %d SQL statements, want at most 4: %v", len(counter.statements), counter.statements)
	}
	var parent system.SysBaseMenu
	if err := db.Where("name = ?", "asset").First(&parent).Error; err != nil {
		t.Fatalf("load parent menu: %v", err)
	}
	var children []system.SysBaseMenu
	if err := db.Where("parent_id = ?", parent.ID).Order("name ASC").Find(&children).Error; err != nil {
		t.Fatalf("load child menus: %v", err)
	}
	if len(children) != 2 {
		t.Fatalf("registered %d child menus, want 2", len(children))
	}

	counter.reset()
	if err := registerMenus(db, menus); err != nil {
		t.Fatalf("repeat menu registration: %v", err)
	}
	if len(counter.statements) != 2 {
		t.Fatalf("repeat menu registration executed %d SQL statements, want 2: %v", len(counter.statements), counter.statements)
	}
}

func TestRegisterDictionariesBatchesDefinitionsAndDetails(t *testing.T) {
	db, counter := setupMetadataTestDB(t)
	enabled := true
	dictionaries := []system.SysDictionary{
		{
			Name: "公告类型", Type: "announcement_type", Status: &enabled,
			SysDictionaryDetails: []system.SysDictionaryDetail{
				{Label: "通知", Value: "notice", Status: &enabled},
				{Label: "新闻", Value: "news", Status: &enabled},
				{Label: "重复通知", Value: "notice", Status: &enabled},
			},
		},
	}

	counter.reset()
	if err := registerDictionaries(db, dictionaries); err != nil {
		t.Fatalf("register dictionaries: %v", err)
	}
	if len(counter.statements) > 4 {
		t.Fatalf("initial dictionary registration executed %d SQL statements, want at most 4: %v", len(counter.statements), counter.statements)
	}
	var dictionary system.SysDictionary
	if err := db.Where("type = ?", "announcement_type").First(&dictionary).Error; err != nil {
		t.Fatalf("load dictionary: %v", err)
	}
	var detailCount int64
	if err := db.Model(&system.SysDictionaryDetail{}).Where("sys_dictionary_id = ?", dictionary.ID).Count(&detailCount).Error; err != nil {
		t.Fatalf("count dictionary details: %v", err)
	}
	if detailCount != 2 {
		t.Fatalf("registered %d dictionary details, want 2", detailCount)
	}

	counter.reset()
	if err := registerDictionaries(db, dictionaries); err != nil {
		t.Fatalf("repeat dictionary registration: %v", err)
	}
	if len(counter.statements) != 2 {
		t.Fatalf("repeat dictionary registration executed %d SQL statements, want 2: %v", len(counter.statements), counter.statements)
	}
}
