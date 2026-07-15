package api

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/systemsetting/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/systemsetting/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/systemsetting/service"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

var LoginBackground = new(loginBackgroundAPI)

type loginBackgroundAPI struct{}

func (a *loginBackgroundAPI) Current(c *gin.Context) {
	item, err := service.LoginBackground.Current()
	if err != nil {
		response.FailWithMessage("获取登录背景失败", c)
		return
	}
	response.OkWithData(item, c)
}

func (a *loginBackgroundAPI) List(c *gin.Context) {
	list, err := service.LoginBackground.List()
	if err != nil {
		response.FailWithMessage("获取背景图库失败", c)
		return
	}
	response.OkWithData(list, c)
}

func (a *loginBackgroundAPI) Create(c *gin.Context) {
	var req request.CreateBackground
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("图片信息不完整", c)
		return
	}
	item := model.LoginBackground{Name: req.Name, URL: req.URL, CreatedBy: utils.GetUserID(c)}
	if err := service.LoginBackground.Create(&item); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(item, "背景图片已加入图库", c)
}

func (a *loginBackgroundAPI) Activate(c *gin.Context) {
	var req request.ActivateBackground
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("请选择背景图片", c)
		return
	}
	if err := service.LoginBackground.Activate(req.ID); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("登录页背景已更新", c)
}

func (a *loginBackgroundAPI) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("背景图片参数不正确", c)
		return
	}
	if err := service.LoginBackground.Delete(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("背景图片已删除", c)
}
