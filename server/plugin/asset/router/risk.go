package router

import (
	"github.com/WangMi2022/mit-assets-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type riskRouter struct{}

func (r *riskRouter) Init(_ *gin.RouterGroup, private *gin.RouterGroup) {
	read := private.Group("assetRisk")
	{
		read.GET("dashboard", apiRisk.Dashboard)
		read.GET("list", apiRisk.List)
		read.GET("detail", apiRisk.Detail)
		read.GET("rules", apiRisk.Rules)
		read.GET("scans", apiRisk.ScanRuns)
	}
	write := private.Group("assetRisk").Use(middleware.OperationRecord())
	{
		write.POST("scan", apiRisk.StartScan)
		write.DELETE("events", apiRisk.DeleteEvents)
		write.DELETE("scans", apiRisk.DeleteScanRuns)
		write.PUT("rules", apiRisk.UpdateRule)
		write.PUT("acknowledge", apiRisk.Acknowledge)
		write.PUT("resolve", apiRisk.Resolve)
		write.PUT("ignore", apiRisk.Ignore)
		write.PUT("reopen", apiRisk.Reopen)
		write.PUT("assign", apiRisk.Assign)
	}
}
