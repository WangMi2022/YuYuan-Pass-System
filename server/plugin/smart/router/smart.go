package router

import (
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type smartRouter struct{}

var smartAIPolicy = middleware.AIRequestPolicy{
	Window:      time.Minute,
	MaxRequests: 30,
	BodyLimit:   64 * 1024,
	Timeout:     90 * time.Second,
}

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
	smartWrite.DELETE("copilot/session", apiSmart.DeleteSession)
	smartWrite.POST("draft/accept", apiSmart.AcceptDraft)
	smartAIWrite := private.Group("smart").Use(middleware.AISecurity(smartAIPolicy), middleware.AIOperationRecord())
	smartAIWrite.POST("copilot/query", apiSmart.Query)
	smartAIWrite.POST("announcement/extract", apiSmart.ExtractAnnouncement)
	smartAIWrite.POST("announcement/draft-schedule", apiSmart.ExtractAnnouncement)
	smartAIWrite.POST("operation/draft", apiSmart.OperationDraft)
	smartAIStreamWrite := private.Group("smart").Use(middleware.AISecurity(smartAIPolicy), middleware.AISSEConcurrency(2), middleware.AIOperationRecord())
	smartAIStreamWrite.POST("copilot/queryStream", apiSmart.QueryStream)
	reportWrite := private.Group("smartReport").Use(middleware.OperationRecord())
	reportWrite.PUT("subscription", apiSmart.SaveSubscription)
	reportAIWrite := private.Group("smartReport").Use(middleware.AISecurity(smartAIPolicy), middleware.AIOperationRecord())
	reportAIWrite.POST("generate", apiSmart.GenerateReport)
}
