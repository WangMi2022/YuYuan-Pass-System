package service

import (
	"strings"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/systemsetting/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestLoginBackgroundResponsesDoNotExposeManagedStorageOrigin(t *testing.T) {
	previousDB := global.GVA_DB
	previousConfig := global.GVA_CONFIG
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open login background database: %v", err)
	}
	if err = database.AutoMigrate(&model.LoginBackground{}); err != nil {
		t.Fatalf("migrate login background table: %v", err)
	}
	global.GVA_DB = database
	global.GVA_CONFIG.System.OssType = "minio"
	global.GVA_CONFIG.Minio.BucketUrl = "http://172.30.3.135:9000/gva-assets"
	global.GVA_CONFIG.JWT.SigningKey = "login-background-preview-test-key"
	t.Cleanup(func() {
		global.GVA_DB = previousDB
		global.GVA_CONFIG = previousConfig
	})

	const rawURL = "http://172.30.3.135:9000/gva-assets/assets/2026-07-16/login-background.jpg"
	item := model.LoginBackground{Name: "login-background.jpg", URL: rawURL, IsActive: true}
	if err = database.Create(&item).Error; err != nil {
		t.Fatalf("create login background: %v", err)
	}

	current, err := LoginBackground.Current()
	if err != nil {
		t.Fatalf("read current login background: %v", err)
	}
	assertBackgroundUsesProxy(t, current.URL)

	list, err := LoginBackground.List()
	if err != nil || len(list) != 1 {
		t.Fatalf("list login backgrounds: len=%d err=%v", len(list), err)
	}
	assertBackgroundUsesProxy(t, list[0].URL)
}

func assertBackgroundUsesProxy(t *testing.T, previewURL string) {
	t.Helper()
	if strings.Contains(previewURL, "172.30.3.135") || !strings.HasPrefix(previewURL, "/api/fileUploadAndDownload/preview?") {
		t.Fatalf("login background exposed storage URL: %q", previewURL)
	}
}
