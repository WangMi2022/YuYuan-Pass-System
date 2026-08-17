package api

import (
	"github.com/WangMi2022/mit-assets-admin/server/model/common/response"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/systemsetting/model"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/systemsetting/model/request"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/systemsetting/service"
	"github.com/WangMi2022/mit-assets-admin/server/utils"
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
		response.FailWithMessage("品牌信息不完整", c)
		return
	}
	updatedBy := utils.GetUserID(c)
	hasLogo := req.Name != "" || req.URL != ""
	hasBranding := req.SystemName != "" || req.Subtitle != ""
	if !hasLogo && !hasBranding {
		response.FailWithMessage("品牌信息不完整", c)
		return
	}
	if hasLogo && hasBranding {
		response.FailWithMessage("系统标识与Logo请分别保存", c)
		return
	}
	if hasBranding {
		if err := service.LoginLogo.SaveBranding(req.SystemName, req.Subtitle, updatedBy); err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
	} else {
		item := model.LoginLogo{Name: req.Name, URL: req.URL, UpdatedBy: updatedBy}
		if err := service.LoginLogo.Save(&item); err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
	}
	response.OkWithMessage("品牌外观已更新", c)
}

func (a *loginLogoAPI) Reset(c *gin.Context) {
	if err := service.LoginLogo.Reset(); err != nil {
		response.FailWithMessage("恢复默认登录图标失败", c)
		return
	}
	response.OkWithMessage("已恢复默认系统Logo", c)
}
