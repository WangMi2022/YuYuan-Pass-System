package api

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
	assetRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Location = new(locationAPI)

type locationAPI struct{}

func (a *locationAPI) Create(c *gin.Context) {
	var location model.Location
	if err := c.ShouldBindJSON(&location); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceLocation.Create(&location); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(location, "位置创建成功", c)
}

func (a *locationAPI) Update(c *gin.Context) {
	var location model.Location
	if err := c.ShouldBindJSON(&location); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceLocation.Update(&location); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("位置更新成功", c)
}

func (a *locationAPI) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := serviceLocation.Delete(id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("位置删除成功", c)
}

func (a *locationAPI) List(c *gin.Context) {
	var search assetRequest.LocationSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceLocation.List(search)
	if err != nil {
		global.GVA_LOG.Error("获取资产位置列表失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: search.Page, PageSize: search.PageSize,
	}, "获取成功", c)
}

func (a *locationAPI) Options(c *gin.Context) {
	var search assetRequest.LocationOptionsSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := serviceLocation.Options(search.Type)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}
