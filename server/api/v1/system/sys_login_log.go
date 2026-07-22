package system

import (
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LoginLogApi struct{}

func (s *LoginLogApi) DeleteLoginLog(c *gin.Context) {
	var loginLog system.SysLoginLog
	err := c.ShouldBindJSON(&loginLog)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = loginLogService.DeleteLoginLog(loginLog)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (s *LoginLogApi) DeleteLoginLogByIds(c *gin.Context) {
	var req request.LogDeleteReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.ClearAll {
		deleted, clearErr := loginLogService.ClearLoginLogs(c.Request.Context())
		if clearErr != nil {
			global.GVA_LOG.Error("清空失败!", zap.Error(clearErr))
			response.FailWithMessage("清空失败", c)
			return
		}
		response.OkWithDetailed(
			gin.H{"deleted": deleted},
			fmt.Sprintf("已清空 %d 条登录日志", deleted),
			c,
		)
		return
	}
	if len(req.Ids) == 0 {
		response.FailWithMessage("请选择要删除的登录日志", c)
		return
	}
	err = loginLogService.DeleteLoginLogByIds(request.IdsReq{Ids: req.Ids})
	if err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败", c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

func (s *LoginLogApi) FindLoginLog(c *gin.Context) {
	var loginLog system.SysLoginLog
	err := c.ShouldBindQuery(&loginLog)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	reLoginLog, err := loginLogService.GetLoginLog(loginLog.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithDetailed(reLoginLog, "查询成功", c)
}

func (s *LoginLogApi) GetLoginLogList(c *gin.Context) {
	var pageInfo systemReq.SysLoginLogSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := loginLogService.GetLoginLogInfoList(pageInfo)
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
