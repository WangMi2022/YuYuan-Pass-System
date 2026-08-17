package router

import (
	"github.com/WangMi2022/mit-assets-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type recognitionRouter struct{}

func (recognitionRouter) Init(private *gin.RouterGroup) {
	write := private.Group("assetRecognition").Use(middleware.OperationRecord())
	{
		write.POST("create", apiRecognition.Create)
		write.POST("retry", apiRecognition.Retry)
		write.PUT("draft", apiRecognition.Draft)
		write.POST("confirm", apiRecognition.Confirm)
		write.DELETE("delete", apiRecognition.Delete)
	}
	read := private.Group("assetRecognition")
	{
		read.GET("list", apiRecognition.List)
		read.GET("detail", apiRecognition.Detail)
	}
}
