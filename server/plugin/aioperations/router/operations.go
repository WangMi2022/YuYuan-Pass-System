package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type operationsRouter struct{}

func (operationsRouter) Init(private *gin.RouterGroup) {
	read := private.Group("ai")
	{
		read.GET("providers", apiOperations.Providers)
		read.GET("usage/summary", apiOperations.UsageSummary)
		read.GET("invocations", apiOperations.Invocations)
		read.GET("quotas", apiOperations.Quotas)
		read.GET("prompts", apiOperations.Prompts)
	}
	write := private.Group("ai").Use(middleware.OperationRecord())
	{
		write.PUT("providers", apiOperations.UpdateProviders)
		write.PUT("quotas", apiOperations.SaveQuota)
		write.POST("prompts", apiOperations.CreatePrompt)
		write.PUT("prompts/activate", apiOperations.ActivatePrompt)
	}
}
