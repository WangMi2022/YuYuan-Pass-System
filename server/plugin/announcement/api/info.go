package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/announcement/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/announcement/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/announcement/service"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Info = new(info)

type info struct{}

// CreateInfo 创建公告
// @Tags Info
// @Summary 创建公告
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.Info true "创建公告"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /info/createInfo [post]
func (a *info) CreateInfo(c *gin.Context) {
	var info model.Info
	err := c.ShouldBindJSON(&info)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := int(utils.GetUserID(c))
	info.UserID = &userID
	published, err := serviceInfo.CreateInfo(&info)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	if published {
		service.NotificationHub.Publish(service.AnnouncementEvent{ID: info.ID, Title: info.Title})
	}
	response.OkWithDetailed(info, map[bool]string{true: "公告发布成功", false: "草稿保存成功"}[published], c)
}

// DeleteInfo 删除公告
// @Tags Info
// @Summary 删除公告
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.Info true "删除公告"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /info/deleteInfo [delete]
func (a *info) DeleteInfo(c *gin.Context) {
	ID := c.Query("ID")
	err := serviceInfo.DeleteInfo(ID)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteInfoByIds 批量删除公告
// @Tags Info
// @Summary 批量删除公告
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /info/deleteInfoByIds [delete]
func (a *info) DeleteInfoByIds(c *gin.Context) {
	IDs := c.QueryArray("IDs[]")
	if err := serviceInfo.DeleteInfoByIds(IDs); err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败", c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateInfo 更新公告
// @Tags Info
// @Summary 更新公告
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.Info true "更新公告"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /info/updateInfo [put]
func (a *info) UpdateInfo(c *gin.Context) {
	var info model.Info
	err := c.ShouldBindJSON(&info)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := int(utils.GetUserID(c))
	info.UserID = &userID
	published, err := serviceInfo.UpdateInfo(info)
	if err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	if published {
		service.NotificationHub.Publish(service.AnnouncementEvent{ID: info.ID, Title: info.Title})
	}
	response.OkWithMessage(map[bool]string{true: "公告发布成功", false: "更新成功"}[published], c)
}

// FindInfo 用id查询公告
// @Tags Info
// @Summary 用id查询公告
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query model.Info true "用id查询公告"
// @Success 200 {object} response.Response{data=model.Info,msg=string} "查询成功"
// @Router /info/findInfo [get]
func (a *info) FindInfo(c *gin.Context) {
	ID := c.Query("ID")
	reinfo, err := serviceInfo.GetInfo(ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(reinfo, c)
}

// GetInfoList 分页获取公告列表
// @Tags Info
// @Summary 分页获取公告列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query request.InfoSearch true "分页获取公告列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /info/getInfoList [get]
func (a *info) GetInfoList(c *gin.Context) {
	var pageInfo request.InfoSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceInfo.GetInfoInfoList(pageInfo)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}

// GetInfoDataSource 获取Info的数据源
// @Tags Info
// @Summary 获取Info的数据源
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "查询成功"
// @Router /info/getInfoDataSource [get]
func (a *info) GetInfoDataSource(c *gin.Context) {
	// 此接口为获取数据源定义的数据
	dataSource, err := serviceInfo.GetInfoDataSource()
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(dataSource, c)
}

// GetInfoPublic 不需要鉴权的公告接口
// @Tags Info
// @Summary 不需要鉴权的公告接口
// @accept application/json
// @Produce application/json
// @Param data query request.InfoSearch true "分页获取公告列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /info/getInfoPublic [get]
func (a *info) GetInfoPublic(c *gin.Context) {
	// 此接口不需要鉴权 示例为返回了一个固定的消息接口，一般本接口用于C端服务，需要自己实现业务逻辑
	response.OkWithDetailed(gin.H{"info": "不需要鉴权的公告接口信息"}, "获取成功", c)
}

// Notifications 获取当前用户最近公告及未读数量。
func (a *info) Notifications(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	result, err := serviceInfo.Notifications(utils.GetUserID(c), limit)
	if err != nil {
		global.GVA_LOG.Error("获取公告提醒失败", zap.Error(err))
		response.FailWithMessage("获取公告提醒失败", c)
		return
	}
	response.OkWithData(result, c)
}

// MarkRead 将指定公告标记为已读。
func (a *info) MarkRead(c *gin.Context) {
	var req request.ReadInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("公告参数不正确", c)
		return
	}
	if err := serviceInfo.MarkRead(utils.GetUserID(c), req.ID); err != nil {
		response.FailWithMessage("标记已读失败", c)
		return
	}
	response.OkWithMessage("已标记为已读", c)
}

// MarkAllRead 将当前用户所有已发布公告标记为已读。
func (a *info) MarkAllRead(c *gin.Context) {
	if err := serviceInfo.MarkAllRead(utils.GetUserID(c)); err != nil {
		response.FailWithMessage("全部已读失败", c)
		return
	}
	response.OkWithMessage("全部公告已读", c)
}

// Stream 使用 SSE 向在线用户实时推送新公告事件。
func (a *info) Stream(c *gin.Context) {
	w := c.Writer
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher, ok := w.(http.Flusher)
	if !ok {
		response.FailWithMessage("当前服务不支持实时推送", c)
		return
	}
	_, _ = fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	client := service.NotificationHub.Subscribe()
	defer service.NotificationHub.Unsubscribe(client)
	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case event := <-client:
			payload, _ := json.Marshal(event)
			_, _ = fmt.Fprintf(w, "event: announcement\ndata: %s\n\n", payload)
			flusher.Flush()
		case <-heartbeat.C:
			_, _ = fmt.Fprintf(w, "event: ping\ndata: {\"time\":%d}\n\n", time.Now().Unix())
			flusher.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}
