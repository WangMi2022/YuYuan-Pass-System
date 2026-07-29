package system

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"go.uber.org/zap"
)

//@author: [piexlmax](https://github.com/piexlmax)
//@function: GetSystemConfig
//@description: 读取配置文件
//@return: conf config.Server, err error

type SystemConfigService struct{}

var SystemConfigServiceApp = new(SystemConfigService)

func (systemConfigService *SystemConfigService) GetSystemConfig() (conf config.Server, err error) {
	conf = global.GVA_CONFIG
	conf.InvoiceRecognition.Normalize()
	conf.InvoiceRecognition = conf.InvoiceRecognition.Redacted()
	return conf, nil
}

// @description   set system config,
//@author: [piexlmax](https://github.com/piexlmax)
//@function: SetSystemConfig
//@description: 设置配置文件
//@param: system model.System
//@return: err error

func (systemConfigService *SystemConfigService) SetSystemConfig(system system.System, allowInvoiceRecognition bool) (err error) {
	system.Config.InvoiceRecognition = system.Config.InvoiceRecognition.MergeSecrets(
		global.GVA_CONFIG.InvoiceRecognition,
		allowInvoiceRecognition,
	)
	system.Config.InvoiceRecognition.Normalize()
	if err = system.Config.InvoiceRecognition.Validate(); err != nil {
		return err
	}
	cs := utils.StructToMap(system.Config)
	for k, v := range cs {
		global.GVA_VP.Set(k, v)
	}
	if err = global.GVA_VP.WriteConfig(); err != nil {
		return err
	}
	// The invoice worker reads this section for every job, so provider changes
	// take effect immediately without restarting database connections.
	global.GVA_CONFIG.InvoiceRecognition = system.Config.InvoiceRecognition
	return nil
}

//@author: [SliverHorn](https://github.com/SliverHorn)
//@function: GetServerInfo
//@description: 获取服务器信息
//@return: server *utils.Server, err error

func (systemConfigService *SystemConfigService) GetServerInfo() (server *utils.Server, err error) {
	var s utils.Server
	s.CollectedAt = time.Now()
	s.Os = utils.InitOS()
	if s.Host, err = utils.InitHost(); err != nil {
		global.GVA_LOG.Error("func utils.InitHost() Failed", zap.String("err", err.Error()))
		return &s, err
	}
	if s.Load, err = utils.InitLoad(); err != nil {
		global.GVA_LOG.Error("func utils.InitLoad() Failed", zap.String("err", err.Error()))
		return &s, err
	}
	if s.Cpu, err = utils.InitCPU(); err != nil {
		global.GVA_LOG.Error("func utils.InitCPU() Failed", zap.String("err", err.Error()))
		return &s, err
	}
	if s.Ram, err = utils.InitRAM(); err != nil {
		global.GVA_LOG.Error("func utils.InitRAM() Failed", zap.String("err", err.Error()))
		return &s, err
	}
	if s.Disk, err = utils.InitDisk(); err != nil {
		global.GVA_LOG.Error("func utils.InitDisk() Failed", zap.String("err", err.Error()))
		return &s, err
	}

	return &s, nil
}
