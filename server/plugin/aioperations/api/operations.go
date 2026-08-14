package api

import (
	"github.com/flipped-aurora/gin-vue-admin/server/ai"
	"github.com/flipped-aurora/gin-vue-admin/server/config"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	aiService "github.com/flipped-aurora/gin-vue-admin/server/plugin/aioperations/service"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

type operationsAPI struct{}

func (operationsAPI) Providers(c *gin.Context) {
	response.OkWithData(serviceOperations.Providers(), c)
}

func (operationsAPI) UpdateProviders(c *gin.Context) {
	var incoming config.AI
	if err := c.ShouldBindJSON(&incoming); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	updated, err := serviceOperations.UpdateProviders(c.Request.Context(), incoming)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(updated, "AI Provider 配置已更新", c)
}

func (operationsAPI) UsageSummary(c *gin.Context) {
	summary, err := serviceOperations.UsageSummary(c.Request.Context(), utils.GetUserID(c))
	if err != nil {
		response.FailWithMessage("获取 AI 用量失败", c)
		return
	}
	response.OkWithData(summary, c)
}

func (operationsAPI) Invocations(c *gin.Context) {
	var search aiService.InvocationSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceOperations.Invocations(c.Request.Context(), search)
	if err != nil {
		response.FailWithMessage("获取 AI 调用审计失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, "获取成功", c)
}

func (operationsAPI) Quotas(c *gin.Context) {
	list, err := serviceOperations.Quotas(c.Request.Context())
	if err != nil {
		response.FailWithMessage("获取 AI 配额失败", c)
		return
	}
	response.OkWithData(list, c)
}

func (operationsAPI) SaveQuota(c *gin.Context) {
	var quota ai.UsageQuota
	if err := c.ShouldBindJSON(&quota); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	saved, err := serviceOperations.SaveQuota(c.Request.Context(), quota)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(saved, "AI 配额已保存", c)
}

func (operationsAPI) Prompts(c *gin.Context) {
	list, err := serviceOperations.Prompts(c.Request.Context())
	if err != nil {
		response.FailWithMessage("获取 Prompt 模板失败", c)
		return
	}
	response.OkWithData(list, c)
}

func (operationsAPI) CreatePrompt(c *gin.Context) {
	var input aiService.PromptInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	prompt, err := serviceOperations.CreatePrompt(c.Request.Context(), input, utils.GetUserID(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(prompt, "Prompt 草稿已创建", c)
}

func (operationsAPI) ActivatePrompt(c *gin.Context) {
	var input aiService.PromptActivation
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceOperations.ActivatePrompt(c.Request.Context(), input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("Prompt 模板已激活", c)
}
