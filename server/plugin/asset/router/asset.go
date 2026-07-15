package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

var Asset = new(assetRouter)

type assetRouter struct{}

func (r *assetRouter) Init(public, private *gin.RouterGroup) {
	write := private.Group("asset").Use(middleware.OperationRecord())
	{
		write.POST("create", apiAsset.Create)
		write.PUT("update", apiAsset.Update)
		write.DELETE("delete", apiAsset.Delete)
		write.POST("uploadPhoto", apiAsset.UploadPhoto)
		write.DELETE("deletePhoto", apiAsset.DeletePhoto)
	}
	read := private.Group("asset")
	{
		read.GET("detail", apiAsset.Detail)
		read.GET("list", apiAsset.List)
		read.GET("dashboard", apiAsset.Dashboard)
		read.GET("categoryOptions", apiCategory.Options)
	}
	public.GET("asset/photo", apiAsset.Photo)
}
