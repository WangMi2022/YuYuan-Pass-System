package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type loginBackgroundRouter struct{}

func (r *loginBackgroundRouter) Init(public, private *gin.RouterGroup) {
	public.GET("appearance/login-background", apiLoginBackground.Current)

	read := private.Group("appearance")
	read.GET("login-backgrounds", apiLoginBackground.List)

	write := private.Group("appearance").Use(middleware.OperationRecord())
	write.POST("login-background", apiLoginBackground.Create)
	write.PUT("login-background/activate", apiLoginBackground.Activate)
	write.DELETE("login-background", apiLoginBackground.Delete)
}
