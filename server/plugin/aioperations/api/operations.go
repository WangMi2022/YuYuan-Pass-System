package api

import (
	"github.com/WangMi2022/mit-assets-admin/server/ai"
	"github.com/WangMi2022/mit-assets-admin/server/config"
	commonRequest "github.com/WangMi2022/mit-assets-admin/server/model/common/request"
	"github.com/WangMi2022/mit-assets-admin/server/model/common/response"
	aiService "github.com/WangMi2022/mit-assets-admin/server/plugin/aioperations/service"
	systemService "github.com/WangMi2022/mit-assets-admin/server/service/system"
	"github.com/WangMi2022/mit-assets-admin/server/utils"
	"github.com/gin-gonic/gin"
)

type operationsAPI struct{}

func (operationsAPI) InvoiceRecognition(c *gin.Context) {
	response.OkWithData(systemService.SystemConfigServiceApp.GetInvoiceRecognitionConfig(), c)
}

func (operationsAPI) UpdateInvoiceRecognition(c *gin.Context) {
	if utils.GetUserAuthorityId(c) != 888 {
		response.FailWithMessage("仅超级管理员可维护智能识别配置", c)
		return
	}
	var incoming config.InvoiceRecognition
	if err := c.ShouldBindJSON(&incoming); err != nil {
		response.FailWithMessage("智能识别配置参数不正确", c)
		return
	}
	detections, err := systemService.SystemConfigServiceApp.SetInvoiceRecognitionConfig(c.Request.Context(), incoming)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(detections, "智能识别配置已保存", c)
}

func (operationsAPI) TestInvoiceRecognition(c *gin.Context) {
	if utils.GetUserAuthorityId(c) != 888 {
		response.FailWithMessage("仅超级管理员可测试智能识别服务", c)
		return
	}
	var request struct {
		Target string                    `json:"target" binding:"required,oneof=baidu public-ocr verification multimodal"`
		Config config.InvoiceRecognition `json:"config" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		response.FailWithMessage("连接测试参数不正确", c)
		return
	}
	detection, err := systemService.SystemConfigServiceApp.TestInvoiceRecognitionProvider(c.Request.Context(), request.Target, request.Config)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(detection, "连接测试成功", c)
}

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
	response.OkWithDetailed(updated, "模型接入配置已更新", c)
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
	var search struct {
		commonRequest.PageInfo
		Paged bool `form:"paged"`
	}
	_ = c.ShouldBindQuery(&search)
	if search.Paged {
		list, total, err := serviceOperations.QuotaPage(c.Request.Context(), &search.PageInfo)
		if err != nil {
			response.FailWithMessage("获取 AI 配额失败", c)
			return
		}
		response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, "获取成功", c)
		return
	}
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
	var search struct {
		commonRequest.PageInfo
		Paged bool `form:"paged"`
	}
	_ = c.ShouldBindQuery(&search)
	if search.Paged {
		list, total, err := serviceOperations.PromptPage(c.Request.Context(), &search.PageInfo)
		if err != nil {
			response.FailWithMessage("获取 Prompt 模板失败", c)
			return
		}
		response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, "获取成功", c)
		return
	}
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
