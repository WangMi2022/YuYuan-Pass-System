package api

import (
	"strconv"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/common/response"
	scheduleRequest "github.com/WangMi2022/mit-assets-admin/server/plugin/schedule/model/request"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/schedule/service"
	"github.com/WangMi2022/mit-assets-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var WorkSchedule = new(workScheduleAPI)

type workScheduleAPI struct{}

func (a *workScheduleAPI) List(c *gin.Context) {
	list, err := service.WorkSchedule.List(c.Request.Context(), utils.GetUserID(c))
	if err != nil {
		global.GVA_LOG.Error("获取日程列表失败", zap.Error(err))
		response.FailWithMessage("获取日程列表失败", c)
		return
	}
	response.OkWithData(list, c)
}

func (a *workScheduleAPI) Create(c *gin.Context) {
	var payload scheduleRequest.WorkScheduleUpsert
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	schedule, err := service.WorkSchedule.Create(c.Request.Context(), utils.GetUserID(c), payload)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(schedule, "日程已创建", c)
}

func (a *workScheduleAPI) Update(c *gin.Context) {
	var payload scheduleRequest.WorkScheduleUpsert
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	schedule, err := service.WorkSchedule.Update(c.Request.Context(), utils.GetUserID(c), payload)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(schedule, "日程已更新", c)
}

func (a *workScheduleAPI) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("日程 ID 不正确", c)
		return
	}
	if err = service.WorkSchedule.Delete(c.Request.Context(), utils.GetUserID(c), uint(id)); err != nil {
		response.FailWithMessage("删除日程失败", c)
		return
	}
	response.OkWithMessage("日程已删除", c)
}

func (a *workScheduleAPI) ImportLegacy(c *gin.Context) {
	var payload scheduleRequest.LegacyScheduleImport
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	created, err := service.WorkSchedule.ImportLegacy(c.Request.Context(), utils.GetUserID(c), payload)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"created": created}, "本地日程已同步", c)
}

func (a *workScheduleAPI) Notifications(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	result, err := service.WorkSchedule.Notifications(c.Request.Context(), utils.GetUserID(c), limit)
	if err != nil {
		global.GVA_LOG.Error("获取日程提醒失败", zap.Error(err))
		response.FailWithMessage("获取日程提醒失败", c)
		return
	}
	response.OkWithData(result, c)
}

func (a *workScheduleAPI) MarkNotificationRead(c *gin.Context) {
	var payload scheduleRequest.NotificationRead
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.FailWithMessage("提醒参数不正确", c)
		return
	}
	if err := service.WorkSchedule.MarkNotificationRead(c.Request.Context(), utils.GetUserID(c), payload.ID); err != nil {
		response.FailWithMessage("标记已读失败", c)
		return
	}
	response.OkWithMessage("已标记为已读", c)
}

func (a *workScheduleAPI) MarkAllNotificationsRead(c *gin.Context) {
	if err := service.WorkSchedule.MarkAllNotificationsRead(c.Request.Context(), utils.GetUserID(c)); err != nil {
		response.FailWithMessage("全部日程提醒已读失败", c)
		return
	}
	response.OkWithMessage("全部日程提醒已读", c)
}
