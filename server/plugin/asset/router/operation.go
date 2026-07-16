package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type operationRouter struct{}

func (r *operationRouter) Init(_ *gin.RouterGroup, private *gin.RouterGroup) {
	read := private.Group("assetOperation")
	{
		read.GET("list", apiOperation.List)
		read.GET("detail", apiOperation.Detail)
		read.GET("assetOptions", apiOperation.AssetOptions)
	}
	write := private.Group("assetOperation").Use(middleware.OperationRecord())
	{
		write.POST("create", apiOperation.Create)
		write.PUT("update", apiOperation.Update)
		write.PUT("submit", apiOperation.Submit)
		write.DELETE("delete", apiOperation.Delete)
	}
}
