package service

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/ai"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	commonRequest "github.com/WangMi2022/mit-assets-admin/server/model/common/request"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestPromptPageUsesTenItemDefaultPages(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "prompts.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = database.AutoMigrate(&ai.PromptTemplate{}, &ai.UsageQuota{}); err != nil {
		t.Fatal(err)
	}
	previous := global.GVA_DB
	global.GVA_DB = database
	t.Cleanup(func() {
		global.GVA_DB = previous
		if sqlDB, dbErr := database.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	})

	for index := 0; index < 12; index++ {
		prompt := ai.PromptTemplate{
			PromptKey: fmt.Sprintf("prompt-%02d", index), Version: 1,
			Content: "测试", Status: ai.PromptStatusDraft, CreatedBy: 1,
		}
		if err = database.Create(&prompt).Error; err != nil {
			t.Fatalf("create prompt %d: %v", index, err)
		}
		quota := ai.UsageQuota{
			ScopeType: "user", ScopeID: fmt.Sprintf("user-%02d", index), Enabled: true,
		}
		if err = database.Create(&quota).Error; err != nil {
			t.Fatalf("create quota %d: %v", index, err)
		}
	}

	firstPage := commonRequest.PageInfo{Page: 1, PageSize: 10}
	list, total, err := (operationsService{}).PromptPage(t.Context(), &firstPage)
	if err != nil || total != 12 || len(list) != 10 {
		t.Fatalf("first page = %d, total=%d, err=%v", len(list), total, err)
	}
	secondPage := commonRequest.PageInfo{Page: 2, PageSize: 10}
	list, total, err = (operationsService{}).PromptPage(t.Context(), &secondPage)
	if err != nil || total != 12 || len(list) != 2 {
		t.Fatalf("second page = %d, total=%d, err=%v", len(list), total, err)
	}

	quotaPage := commonRequest.PageInfo{}
	quotas, quotaTotal, err := (operationsService{}).QuotaPage(t.Context(), &quotaPage)
	if err != nil || quotaTotal != 12 || len(quotas) != 10 || quotaPage.Page != 1 || quotaPage.PageSize != 10 {
		t.Fatalf("quota default page = %d/%d, total=%d, page=%d, err=%v", len(quotas), quotaPage.PageSize, quotaTotal, quotaPage.Page, err)
	}
}
