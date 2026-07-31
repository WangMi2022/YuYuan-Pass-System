package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type invoiceRouter struct{}

func (invoiceRouter) Init(private *gin.RouterGroup) {
	write := private.Group("invoice").Use(middleware.OperationRecord())
	{
		write.POST("upload", apiInvoice.Upload)
		write.PUT("update", apiInvoice.Update)
		write.PUT("confirm", apiInvoice.Confirm)
		write.PUT("reopen", apiInvoice.Reopen)
		write.PUT("retry", apiInvoice.Retry)
		write.POST("recheck", apiInvoice.Recheck)
		write.POST("verify", apiInvoice.Verify)
		write.POST("provider/test", apiInvoice.TestProviderConnection)
		write.DELETE("delete", apiInvoice.Delete)
	}
	read := private.Group("invoice")
	{
		read.GET("list", apiInvoice.List)
		read.GET("detail", apiInvoice.Detail)
		read.GET("capabilities", apiInvoice.Capabilities)
		read.GET("file", apiInvoice.File)
		read.GET("dashboard", apiInvoice.Dashboard)
		read.GET("verificationHistory", apiInvoice.VerificationHistory)
		read.GET("categoryOptions", apiCategory.Options)
	}
	categoryWrite := private.Group("invoiceCategory").Use(middleware.OperationRecord())
	{
		categoryWrite.POST("create", apiCategory.Create)
		categoryWrite.PUT("update", apiCategory.Update)
		categoryWrite.DELETE("delete", apiCategory.Delete)
	}
	private.GET("invoiceCategory/list", apiCategory.List)
	ruleWrite := private.Group("invoiceRule").Use(middleware.OperationRecord())
	{
		ruleWrite.POST("create", apiRule.Create)
		ruleWrite.PUT("update", apiRule.Update)
		ruleWrite.DELETE("delete", apiRule.Delete)
	}
	private.GET("invoiceRule/list", apiRule.List)
}
