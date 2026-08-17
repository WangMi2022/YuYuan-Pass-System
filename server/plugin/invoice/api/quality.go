package api

import (
	"github.com/WangMi2022/mit-assets-admin/server/global"
	commonResponse "github.com/WangMi2022/mit-assets-admin/server/model/common/response"
	invoiceRequest "github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type qualityAPI struct{}

func bindQualitySearch(c *gin.Context) (invoiceRequest.QualitySearch, bool) {
	var search invoiceRequest.QualitySearch
	if err := c.ShouldBindQuery(&search); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return search, false
	}
	if err := search.Normalize(); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return search, false
	}
	return search, true
}

func (qualityAPI) Dashboard(c *gin.Context) {
	search, ok := bindQualitySearch(c)
	if !ok {
		return
	}
	data, err := serviceQuality.Dashboard(search, currentScope(c))
	if err != nil {
		global.GVA_LOG.Error("获取发票识别质量总览失败", zap.Error(err))
		commonResponse.FailWithMessage("获取质量总览失败", c)
		return
	}
	commonResponse.OkWithData(data, c)
}

func (qualityAPI) ProviderMetrics(c *gin.Context) {
	search, ok := bindQualitySearch(c)
	if !ok {
		return
	}
	data, err := serviceQuality.ProviderMetrics(search, currentScope(c))
	if err != nil {
		global.GVA_LOG.Error("获取发票 Provider 质量指标失败", zap.Error(err))
		commonResponse.FailWithMessage("获取 Provider 指标失败", c)
		return
	}
	commonResponse.OkWithData(data, c)
}

func (qualityAPI) FieldMetrics(c *gin.Context) {
	search, ok := bindQualitySearch(c)
	if !ok {
		return
	}
	data, err := serviceQuality.FieldMetrics(search, currentScope(c))
	if err != nil {
		global.GVA_LOG.Error("获取发票字段质量指标失败", zap.Error(err))
		commonResponse.FailWithMessage("获取字段指标失败", c)
		return
	}
	commonResponse.OkWithData(data, c)
}

func (qualityAPI) Failures(c *gin.Context) {
	search, ok := bindQualitySearch(c)
	if !ok {
		return
	}
	list, total, err := serviceQuality.Failures(search, currentScope(c))
	if err != nil {
		global.GVA_LOG.Error("获取发票识别失败明细失败", zap.Error(err))
		commonResponse.FailWithMessage("获取失败明细失败", c)
		return
	}
	commonResponse.OkWithDetailed(commonResponse.PageResult{
		List: list, Total: total, Page: search.Page, PageSize: search.PageSize,
	}, "获取成功", c)
}

func (qualityAPI) ClassificationMetrics(c *gin.Context) {
	search, ok := bindQualitySearch(c)
	if !ok {
		return
	}
	data, err := serviceQuality.ClassificationMetrics(search, currentScope(c))
	if err != nil {
		global.GVA_LOG.Error("获取发票分类质量指标失败", zap.Error(err))
		commonResponse.FailWithMessage("获取分类指标失败", c)
		return
	}
	commonResponse.OkWithData(data, c)
}
