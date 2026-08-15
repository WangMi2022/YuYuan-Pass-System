package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type smartRouter struct{}

func (smartRouter) Init(private *gin.RouterGroup) {
	smartRead := private.Group("smart")
	smartRead.GET("copilot/sessions", apiSmart.Sessions)
	smartRead.GET("copilot/session", apiSmart.Session)
	smartRead.GET("copilot/tools", apiSmart.Tools)
	smartRead.GET("drafts", apiSmart.Drafts)
	smartRead.GET("operation/assets", apiSmart.OperationAssetCandidates)
	reportRead := private.Group("smartReport")
	reportRead.GET("today", apiSmart.TodayReport)
	reportRead.GET("list", apiSmart.Reports)
	reportRead.GET("detail", apiSmart.Report)
	reportRead.GET("subscription", apiSmart.Subscription)
	reportRead.GET("deliveries", apiSmart.Deliveries)

	smartWrite := private.Group("smart").Use(middleware.OperationRecord())
	smartWrite.POST("copilot/query", apiSmart.Query)
	smartWrite.POST("copilot/queryStream", apiSmart.QueryStream)
	smartWrite.DELETE("copilot/session", apiSmart.DeleteSession)
	smartWrite.POST("announcement/extract", apiSmart.ExtractAnnouncement)
	smartWrite.POST("announcement/draft-schedule", apiSmart.ExtractAnnouncement)
	smartWrite.POST("operation/draft", apiSmart.OperationDraft)
	smartWrite.POST("draft/accept", apiSmart.AcceptDraft)
	reportWrite := private.Group("smartReport").Use(middleware.OperationRecord())
	reportWrite.POST("generate", apiSmart.GenerateReport)
	reportWrite.PUT("subscription", apiSmart.SaveSubscription)
}
