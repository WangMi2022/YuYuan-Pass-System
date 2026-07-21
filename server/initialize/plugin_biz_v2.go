package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin"
	systemService "github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/plugin/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func PluginInitV2(group *gin.Engine, plugins ...plugin.Plugin) {
	for i := 0; i < len(plugins); i++ {
		plugins[i].Register(group)
	}
}
func bizPluginV2(engine *gin.Engine) {
	PluginInitV2(engine, plugin.Registered()...)
	if err := systemService.CasbinServiceApp.FreshCasbin(); err != nil {
		global.GVA_LOG.Warn("刷新插件接口权限失败", zap.Error(err))
	}
}
