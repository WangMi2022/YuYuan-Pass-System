package api

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
	assetRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model/request"
	"github.com/gin-gonic/gin"
)

var Category = new(categoryAPI)

type categoryAPI struct{}

func (a *categoryAPI) Create(c *gin.Context) {
	var category model.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceCategory.Create(&category); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(category, "分类创建成功", c)
}

func (a *categoryAPI) Update(c *gin.Context) {
	var category model.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceCategory.Update(&category); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("分类更新成功", c)
}

func (a *categoryAPI) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := serviceCategory.Delete(id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("分类删除成功", c)
}

func (a *categoryAPI) List(c *gin.Context) {
	var search assetRequest.CategorySearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceCategory.List(search)
	if err != nil {
		response.FailWithMessage("获取分类列表失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: search.Page, PageSize: search.PageSize,
	}, "获取成功", c)
}

func (a *categoryAPI) Options(c *gin.Context) {
	list, err := serviceCategory.Options()
	if err != nil {
		response.FailWithMessage("获取分类选项失败", c)
		return
	}
	response.OkWithData(list, c)
}
