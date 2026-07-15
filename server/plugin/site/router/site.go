package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

var Site = new(siteRouter)

type siteRouter struct{}

func (r *siteRouter) Init(_ *gin.RouterGroup, private *gin.RouterGroup) {
	write := private.Group("site").Use(middleware.OperationRecord())
	{
		write.POST("create", apiSite.Create)
		write.PUT("update", apiSite.Update)
		write.DELETE("delete", apiSite.Delete)
	}
	read := private.Group("site")
	{
		read.GET("list", apiSite.List)
		read.GET("detail", apiSite.Detail)
		read.GET("categories", apiSite.Categories)
		read.POST("visit", apiSite.Visit)
	}
}
