package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

var Category = new(categoryRouter)

type categoryRouter struct{}

func (r *categoryRouter) Init(_ *gin.RouterGroup, private *gin.RouterGroup) {
	write := private.Group("assetCategory").Use(middleware.OperationRecord())
	{
		write.POST("create", apiCategory.Create)
		write.PUT("update", apiCategory.Update)
		write.DELETE("delete", apiCategory.Delete)
	}
	private.GET("assetCategory/list", apiCategory.List)
}
