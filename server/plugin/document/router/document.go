package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

var Document = new(documentRouter)

type documentRouter struct{}

func (r *documentRouter) Init(_ *gin.RouterGroup, private *gin.RouterGroup) {
	write := private.Group("document").Use(middleware.OperationRecord())
	{
		write.POST("upload", apiDocument.Upload)
		write.PUT("updateContent", apiDocument.UpdateContent)
		write.DELETE("delete", apiDocument.Delete)
	}
	read := private.Group("document")
	{
		read.GET("list", apiDocument.List)
		read.GET("detail", apiDocument.Detail)
		read.GET("file", apiDocument.File)
	}
}
