package initialize

import (
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/middleware"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/schedule/router"
	"github.com/gin-gonic/gin"
)

func Router(engine *gin.Engine) {
	authenticated := engine.Group(global.GVA_CONFIG.System.RouterPrefix)
	authenticated.Use(middleware.JWTAuth())
	router.Router.Init(authenticated)
}
