package api

import (
	"github.com/WangMi2022/mit-assets-admin/server/model/common/response"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model"
	invoiceRequest "github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model/request"
	"github.com/gin-gonic/gin"
)

type ruleAPI struct{}

func (ruleAPI) Create(c *gin.Context) {
	var rule model.ClassificationRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceRule.Create(&rule); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(rule, "分类规则创建成功", c)
}

func (ruleAPI) Update(c *gin.Context) {
	var rule model.ClassificationRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceRule.Update(&rule); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("分类规则更新成功", c)
}

func (ruleAPI) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := serviceRule.Delete(id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("分类规则删除成功", c)
}

func (ruleAPI) List(c *gin.Context) {
	var search invoiceRequest.RuleSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceRule.List(search)
	if err != nil {
		response.FailWithMessage("获取分类规则失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, "获取成功", c)
}
