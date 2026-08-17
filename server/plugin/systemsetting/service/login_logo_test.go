package service

import (
	"strings"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/systemsetting/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupLoginLogoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	previousDB := global.GVA_DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err = db.AutoMigrate(&model.LoginLogo{}); err != nil {
		t.Fatalf("migrate login logo table: %v", err)
	}
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = previousDB })
	return db
}

func TestLoginLogoBrandingLifecyclePreservesIndependentFields(t *testing.T) {
	db := setupLoginLogoTestDB(t)

	current, err := LoginLogo.Current()
	if err != nil {
		t.Fatalf("read default branding: %v", err)
	}
	if current.SystemName != model.DefaultSystemName || current.Subtitle == nil || *current.Subtitle != model.DefaultSystemSubtitle {
		t.Fatalf("unexpected default branding: %#v", current)
	}

	if err = LoginLogo.SaveBranding("  MIT 资产管理平台  ", "  ASSET OPS  ", 7); err != nil {
		t.Fatalf("save branding: %v", err)
	}
	logo := model.LoginLogo{
		Name:      "  mark.png  ",
		URL:       "  https://assets.example.com/mark.png  ",
		UpdatedBy: 9,
	}
	if err = LoginLogo.Save(&logo); err != nil {
		t.Fatalf("save logo: %v", err)
	}

	current, err = LoginLogo.Current()
	if err != nil {
		t.Fatalf("read saved branding: %v", err)
	}
	if current.SystemName != "MIT 资产管理平台" || current.Subtitle == nil || *current.Subtitle != "ASSET OPS" {
		t.Fatalf("logo update changed branding: %#v", current)
	}
	if current.Name != "mark.png" || current.URL != "https://assets.example.com/mark.png" || current.UpdatedBy != 9 {
		t.Fatalf("unexpected saved logo: %#v", current)
	}

	if err = LoginLogo.Reset(); err != nil {
		t.Fatalf("reset logo: %v", err)
	}
	current, err = LoginLogo.Current()
	if err != nil {
		t.Fatalf("read reset branding: %v", err)
	}
	if current.Name != "" || current.URL != "" {
		t.Fatalf("logo reset did not clear logo fields: %#v", current)
	}
	if current.SystemName != "MIT 资产管理平台" || current.Subtitle == nil || *current.Subtitle != "ASSET OPS" {
		t.Fatalf("logo reset changed branding: %#v", current)
	}

	var count int64
	if err = db.Model(&model.LoginLogo{}).Count(&count).Error; err != nil {
		t.Fatalf("count branding rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one branding row, got %d", count)
	}
}

func TestSaveBrandingValidatesNameAndRuneLimits(t *testing.T) {
	setupLoginLogoTestDB(t)

	tests := []struct {
		name       string
		systemName string
		subtitle   string
		want       string
	}{
		{name: "blank name", systemName: "   ", want: "系统名称不能为空"},
		{name: "name too long", systemName: strings.Repeat("资", 81), want: "系统名称不能超过80个字符"},
		{name: "subtitle too long", systemName: "MIT", subtitle: strings.Repeat("产", 121), want: "品牌副标题不能超过120个字符"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := LoginLogo.SaveBranding(test.systemName, test.subtitle, 1)
			if err == nil || err.Error() != test.want {
				t.Fatalf("expected %q, got %v", test.want, err)
			}
		})
	}
}
