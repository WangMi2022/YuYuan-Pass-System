package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

var Router = new(workScheduleRouter)

type workScheduleRouter struct{}

func (r *workScheduleRouter) Init(authenticated *gin.RouterGroup) {
	write := authenticated.Group("workSchedule").Use(middleware.OperationRecord())
	{
		write.POST("create", apiWorkSchedule.Create)
		write.PUT("update", apiWorkSchedule.Update)
		write.DELETE("delete", apiWorkSchedule.Delete)
		write.POST("import", apiWorkSchedule.ImportLegacy)
		write.POST("notifications/read", apiWorkSchedule.MarkNotificationRead)
		write.POST("notifications/readAll", apiWorkSchedule.MarkAllNotificationsRead)
	}
	read := authenticated.Group("workSchedule")
	{
		read.GET("list", apiWorkSchedule.List)
		read.GET("notifications", apiWorkSchedule.Notifications)
	}
}
