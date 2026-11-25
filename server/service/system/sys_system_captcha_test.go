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
	currentConfig := global.MAA_CONFIG
	currentViper := global.MAA_VP
	defer func() {
		global.MAA_CONFIG = currentConfig
		global.MAA_VP = currentViper
	}()

	global.MAA_CONFIG = config.Server{Captcha: config.Captcha{
		KeyLong: 4, ImgWidth: 240, ImgHeight: 80,
		OpenCaptcha: 3, OpenCaptchaTimeOut: 3600,
	}}
	global.MAA_VP = viper.New()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	global.MAA_VP.SetConfigFile(configPath)
	global.MAA_VP.SetConfigType("yaml")
	if err := global.MAA_VP.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	incoming := global.MAA_CONFIG
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
	if global.MAA_CONFIG.Captcha != incoming.Captcha {
		t.Fatalf("runtime captcha config = %#v, want %#v", global.MAA_CONFIG.Captcha, incoming.Captcha)
	}
}
