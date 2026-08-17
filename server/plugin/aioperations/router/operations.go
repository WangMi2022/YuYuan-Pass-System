package router

import (
	"github.com/WangMi2022/mit-assets-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type operationsRouter struct{}

func (operationsRouter) Init(private *gin.RouterGroup) {
	read := private.Group("ai")
	{
		read.GET("providers", apiOperations.Providers)
		read.GET("invoice-recognition", apiOperations.InvoiceRecognition)
		read.GET("usage/summary", apiOperations.UsageSummary)
		read.GET("invocations", apiOperations.Invocations)
		read.GET("quotas", apiOperations.Quotas)
		read.GET("prompts", apiOperations.Prompts)
	}
	write := private.Group("ai").Use(middleware.OperationRecord())
	{
		write.PUT("providers", apiOperations.UpdateProviders)
		write.PUT("invoice-recognition", apiOperations.UpdateInvoiceRecognition)
		write.POST("invoice-recognition/test", apiOperations.TestInvoiceRecognition)
		write.PUT("quotas", apiOperations.SaveQuota)
		write.POST("prompts", apiOperations.CreatePrompt)
		write.PUT("prompts/activate", apiOperations.ActivatePrompt)
	}
}
