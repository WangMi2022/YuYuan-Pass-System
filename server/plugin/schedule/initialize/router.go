package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/schedule/router"
	"github.com/gin-gonic/gin"
)

func Router(engine *gin.Engine) {
	authenticated := engine.Group(global.GVA_CONFIG.System.RouterPrefix)
	authenticated.Use(middleware.JWTAuth())
	router.Router.Init(authenticated)
}
