package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type loginLogoRouter struct{}

func (r *loginLogoRouter) Init(public, private *gin.RouterGroup) {
	public.GET("appearance/login-logo", apiLoginLogo.Current)

	write := private.Group("appearance").Use(middleware.OperationRecord())
	write.PUT("login-logo", apiLoginLogo.Save)
	write.DELETE("login-logo", apiLoginLogo.Reset)
}
