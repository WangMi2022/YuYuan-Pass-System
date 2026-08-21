package initialize

import (
	"context"

	"github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	systemService "github.com/WangMi2022/mit-assets-admin/server/service/system"
	"go.uber.org/zap"
)

// Reload 优雅地重新加载系统配置
func Reload() error {
	global.GVA_LOG.Info("正在重新加载系统配置...")

	// 重新加载配置文件
	if err := global.GVA_VP.ReadInConfig(); err != nil {
		global.GVA_LOG.Error("重新读取配置文件失败!", zap.Error(err))
		return err
	}
	var nextConfig config.Server
	if err := global.GVA_VP.Unmarshal(&nextConfig); err != nil {
		global.GVA_LOG.Error("重新解析配置文件失败!", zap.Error(err))
		return err
	}
	nextConfig.Captcha.Normalize()
	if err := nextConfig.Captcha.Validate(); err != nil {
		global.GVA_LOG.Error("验证码配置无效!", zap.Error(err))
		return err
	}
	global.GVA_CONFIG_LOCK.Lock()
	global.GVA_CONFIG.Captcha = nextConfig.Captcha
	global.GVA_CONFIG_LOCK.Unlock()

	// 重新初始化数据库连接
	if global.GVA_DB != nil {
		db, _ := global.GVA_DB.DB()
		err := db.Close()
		if err != nil {
			global.GVA_LOG.Error("关闭原数据库连接失败!", zap.Error(err))
			return err
		}
	}

	// 重新建立数据库连接
	global.GVA_DB = Gorm()

	// 重新初始化其他配置
	OtherInit()
	DBList()

	if global.GVA_DB != nil {
		// 确保数据库表结构是最新的
		RegisterTables()
		if err := systemService.CasbinServiceApp.EnsureContactVerificationPermissions(context.Background(), global.GVA_DB); err != nil {
			global.GVA_LOG.Error("reload contact verification permissions failed", zap.Error(err))
		}
		if err := systemService.CasbinServiceApp.FreshCasbin(); err != nil {
			global.GVA_LOG.Warn("reload casbin policies failed", zap.Error(err))
		}
	}

	// 重新初始化定时任务
	Timer()

	global.GVA_LOG.Info("系统配置重新加载完成")
	return nil
}
