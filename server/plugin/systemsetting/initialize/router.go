package initialize

import (
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/middleware"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/systemsetting/router"
	"github.com/gin-gonic/gin"
)

func Router(engine *gin.Engine) {
	public := engine.Group(global.GVA_CONFIG.System.RouterPrefix)
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix)
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.Router.LoginBackground.Init(public, private)
	router.Router.LoginLogo.Init(public, private)
}
