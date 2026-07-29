package api

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model"
	invoiceRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model/request"
	"github.com/gin-gonic/gin"
)

type categoryAPI struct{}

func (categoryAPI) Create(c *gin.Context) {
	var category model.InvoiceCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceCategory.Create(&category); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(category, "发票分类创建成功", c)
}

func (categoryAPI) Update(c *gin.Context) {
	var category model.InvoiceCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceCategory.Update(&category); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("发票分类更新成功", c)
}

func (categoryAPI) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := serviceCategory.Delete(id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("发票分类删除成功", c)
}

func (categoryAPI) List(c *gin.Context) {
	var search invoiceRequest.CategorySearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceCategory.List(search)
	if err != nil {
		response.FailWithMessage("获取发票分类失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, "获取成功", c)
}

func (categoryAPI) Options(c *gin.Context) {
	list, err := serviceCategory.Options()
	if err != nil {
		response.FailWithMessage("获取发票分类选项失败", c)
		return
	}
	response.OkWithData(list, c)
}
