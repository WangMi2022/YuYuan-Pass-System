package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

var Location = new(locationRouter)

type locationRouter struct{}

func (r *locationRouter) Init(_ *gin.RouterGroup, private *gin.RouterGroup) {
	write := private.Group("assetLocation").Use(middleware.OperationRecord())
	{
		write.POST("create", apiLocation.Create)
		write.PUT("update", apiLocation.Update)
		write.DELETE("delete", apiLocation.Delete)
	}
	read := private.Group("assetLocation")
	{
		read.GET("list", apiLocation.List)
		read.GET("options", apiLocation.Options)
	}
}
