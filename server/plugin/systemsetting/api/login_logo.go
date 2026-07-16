package api

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/systemsetting/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/systemsetting/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/systemsetting/service"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

var LoginLogo = new(loginLogoAPI)

type loginLogoAPI struct{}

func (a *loginLogoAPI) Current(c *gin.Context) {
	item, err := service.LoginLogo.Current()
	if err != nil {
		response.FailWithMessage("获取登录图标失败", c)
		return
	}
	response.OkWithData(item, c)
}

func (a *loginLogoAPI) Save(c *gin.Context) {
	var req request.SaveLoginLogo
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("图标信息不完整", c)
		return
	}
	item := model.LoginLogo{Name: req.Name, URL: req.URL, UpdatedBy: utils.GetUserID(c)}
	if err := service.LoginLogo.Save(&item); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("登录图标已更新", c)
}

func (a *loginLogoAPI) Reset(c *gin.Context) {
	if err := service.LoginLogo.Reset(); err != nil {
		response.FailWithMessage("恢复默认登录图标失败", c)
		return
	}
	response.OkWithMessage("已恢复默认登录图标", c)
}
