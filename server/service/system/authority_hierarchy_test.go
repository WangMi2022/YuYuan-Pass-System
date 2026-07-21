package system

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	systemModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func authorityParent(authorityID uint) *uint { return &authorityID }

func TestAuthorityHierarchyBuildsDeterministicTree(t *testing.T) {
	hierarchy, err := newAuthorityHierarchy([]systemModel.SysAuthority{
		{AuthorityId: 300, AuthorityName: "孙角色", ParentId: authorityParent(200)},
		{AuthorityId: 200, AuthorityName: "子角色 B", ParentId: authorityParent(100)},
		{AuthorityId: 100, AuthorityName: "根角色", ParentId: authorityParent(0)},
		{AuthorityId: 150, AuthorityName: "子角色 A", ParentId: authorityParent(100)},
	})
	if err != nil {
		t.Fatalf("build authority hierarchy: %v", err)
	}
	forest, err := hierarchy.forest([]uint{100})
	if err != nil {
		t.Fatalf("build authority tree: %v", err)
	}
	if len(forest) != 1 || len(forest[0].Children) != 2 {
		t.Fatalf("unexpected authority forest: %#v", forest)
	}
	if forest[0].Children[0].AuthorityId != 150 || forest[0].Children[1].AuthorityId != 200 {
		t.Fatalf("children are not sorted by authority ID: %#v", forest[0].Children)
	}
	if len(forest[0].Children[1].Children) != 1 || forest[0].Children[1].Children[0].AuthorityId != 300 {
		t.Fatalf("grandchild is missing: %#v", forest[0].Children[1])
	}
	descendants, err := hierarchy.descendantIDs(100)
	if err != nil {
		t.Fatalf("get descendant IDs: %v", err)
	}
	if want := []uint{150, 200, 300}; !reflect.DeepEqual(descendants, want) {
		t.Fatalf("descendant IDs = %v, want %v", descendants, want)
	}
}

func TestAuthorityHierarchyRejectsCycles(t *testing.T) {
	hierarchy, err := newAuthorityHierarchy([]systemModel.SysAuthority{
		{AuthorityId: 100, ParentId: authorityParent(200)},
		{AuthorityId: 200, ParentId: authorityParent(100)},
	})
	if err != nil {
		t.Fatalf("index cyclic hierarchy: %v", err)
	}
	if _, err := hierarchy.forest([]uint{100}); !errors.Is(err, errAuthorityHierarchyCycle) {
		t.Fatalf("expected cycle error, got %v", err)
	}
	if _, err := hierarchy.descendantIDs(100); !errors.Is(err, errAuthorityHierarchyCycle) {
		t.Fatalf("expected descendant cycle error, got %v", err)
	}
}

type authoritySQLCounter struct {
	logger.Interface
	statements []string
}

func (c *authoritySQLCounter) Trace(_ context.Context, _ time.Time, fc func() (string, int64), _ error) {
	sql, _ := fc()
	c.statements = append(c.statements, sql)
}

func (c *authoritySQLCounter) reset() { c.statements = nil }

func setupAuthorityHierarchyTestDB(t *testing.T) *authoritySQLCounter {
	t.Helper()
	previousDB := global.GVA_DB
	previousStrictAuth := global.GVA_CONFIG.System.UseStrictAuth
	counter := &authoritySQLCounter{Interface: logger.Default.LogMode(logger.Silent)}
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{TranslateError: true, Logger: counter})
	if err != nil {
		t.Fatalf("open authority test database: %v", err)
	}
	if err := db.AutoMigrate(&systemModel.SysAuthority{}); err != nil {
		t.Fatalf("migrate authority tables: %v", err)
	}
	global.GVA_DB = db
	global.GVA_CONFIG.System.UseStrictAuth = true
	t.Cleanup(func() {
		global.GVA_DB = previousDB
		global.GVA_CONFIG.System.UseStrictAuth = previousStrictAuth
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return counter
}

func TestAuthorityServiceLoadsHierarchyWithBoundedQueries(t *testing.T) {
	counter := setupAuthorityHierarchyTestDB(t)
	authorities := []systemModel.SysAuthority{
		{AuthorityId: 100, AuthorityName: "根角色", ParentId: authorityParent(0)},
		{AuthorityId: 200, AuthorityName: "子角色", ParentId: authorityParent(100)},
		{AuthorityId: 300, AuthorityName: "孙角色", ParentId: authorityParent(200)},
		{AuthorityId: 400, AuthorityName: "其他根角色", ParentId: authorityParent(0)},
	}
	if err := global.GVA_DB.Create(&authorities).Error; err != nil {
		t.Fatalf("create authorities: %v", err)
	}

	counter.reset()
	forest, err := AuthorityServiceApp.GetAuthorityInfoList(100)
	if err != nil {
		t.Fatalf("load authority tree: %v", err)
	}
	if len(counter.statements) > 3 {
		t.Fatalf("authority tree executed %d SQL statements, want at most 3: %v", len(counter.statements), counter.statements)
	}
	if len(forest) != 1 || forest[0].AuthorityId != 100 || len(forest[0].Children) != 1 || len(forest[0].Children[0].Children) != 1 {
		t.Fatalf("unexpected strict authority tree: %#v", forest)
	}
	forest, err = AuthorityServiceApp.GetAuthorityInfoList(200)
	if err != nil {
		t.Fatalf("load child authority tree: %v", err)
	}
	if len(forest) != 1 || forest[0].AuthorityId != 300 {
		t.Fatalf("non-root role should only see lower roles: %#v", forest)
	}

	global.GVA_CONFIG.System.UseStrictAuth = false
	forest, err = AuthorityServiceApp.GetAuthorityInfoList(200)
	if err != nil {
		t.Fatalf("load non-strict authority tree: %v", err)
	}
	if len(forest) != 2 || forest[0].AuthorityId != 100 || forest[1].AuthorityId != 400 {
		t.Fatalf("unexpected non-strict authority tree: %#v", forest)
	}
	global.GVA_CONFIG.System.UseStrictAuth = true

	counter.reset()
	allowed, err := AuthorityServiceApp.GetStructAuthorityList(100)
	if err != nil {
		t.Fatalf("load authority scope: %v", err)
	}
	if want := []uint{200, 300, 100}; !reflect.DeepEqual(allowed, want) {
		t.Fatalf("authority scope = %v, want %v", allowed, want)
	}
	if len(counter.statements) != 1 {
		t.Fatalf("authority scope executed %d SQL statements, want 1: %v", len(counter.statements), counter.statements)
	}

	counter.reset()
	if err := AuthorityServiceApp.checkAuthorityIDsAuth(100, []uint{200, 300}); err != nil {
		t.Fatalf("check multiple authority IDs: %v", err)
	}
	if len(counter.statements) != 1 {
		t.Fatalf("multi-ID authority check executed %d SQL statements, want 1: %v", len(counter.statements), counter.statements)
	}
}
