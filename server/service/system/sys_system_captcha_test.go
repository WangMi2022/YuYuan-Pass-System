package system

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	systemModel "github.com/WangMi2022/mit-assets-admin/server/model/system"
	"github.com/spf13/viper"
)

func TestSetSystemConfigAppliesCaptchaRuntimeImmediately(t *testing.T) {
	currentConfig := global.GVA_CONFIG
	currentViper := global.GVA_VP
	defer func() {
		global.GVA_CONFIG = currentConfig
		global.GVA_VP = currentViper
	}()

	global.GVA_CONFIG = config.Server{Captcha: config.Captcha{
		KeyLong: 4, ImgWidth: 240, ImgHeight: 80,
		OpenCaptcha: 3, OpenCaptchaTimeOut: 3600,
	}}
	global.GVA_VP = viper.New()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	global.GVA_VP.SetConfigFile(configPath)
	global.GVA_VP.SetConfigType("yaml")
	if err := global.GVA_VP.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	incoming := global.GVA_CONFIG
	incoming.Captcha = config.Captcha{
		KeyLong: 6, ImgWidth: 300, ImgHeight: 90,
		OpenCaptcha: 0, OpenCaptchaTimeOut: 1800,
	}
	service := &SystemConfigService{}
	if _, err := service.SetSystemConfig(
		context.Background(),
		systemModel.System{Config: incoming},
		true,
	); err != nil {
		t.Fatal(err)
	}
	if global.GVA_CONFIG.Captcha != incoming.Captcha {
		t.Fatalf("runtime captcha config = %#v, want %#v", global.GVA_CONFIG.Captcha, incoming.Captcha)
	}
}
