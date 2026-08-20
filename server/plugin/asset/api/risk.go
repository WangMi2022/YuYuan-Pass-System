package api

import (
	"errors"
	"fmt"
	"io"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/common/response"
	assetRequest "github.com/WangMi2022/mit-assets-admin/server/plugin/asset/model/request"
	"github.com/WangMi2022/mit-assets-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type riskAPI struct{}

func (a *riskAPI) Dashboard(c *gin.Context) {
	data, err := serviceRisk.Dashboard(c.Request.Context())
	if err != nil {
		global.GVA_LOG.Error("获取资产风险总览失败", zap.Error(err))
		response.FailWithMessage("获取资产风险总览失败", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *riskAPI) List(c *gin.Context) {
	var search assetRequest.RiskSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceRisk.List(c.Request.Context(), search)
	if err != nil {
		global.GVA_LOG.Error("获取资产风险列表失败", zap.Error(err))
		response.FailWithMessage("获取资产风险列表失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, "获取成功", c)
}

func (a *riskAPI) Detail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	detail, err := serviceRisk.Detail(c.Request.Context(), id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

func (a *riskAPI) Rules(c *gin.Context) {
	var search assetRequest.RiskRuleSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if search.Paged {
		rules, total, err := serviceRisk.RulePage(c.Request.Context(), &search.PageInfo)
		if err != nil {
			response.FailWithMessage("获取风险规则失败", c)
			return
		}
		response.OkWithDetailed(response.PageResult{List: rules, Total: total, Page: search.Page, PageSize: search.PageSize}, "获取成功", c)
		return
	}
	rules, err := serviceRisk.Rules(c.Request.Context())
	if err != nil {
		response.FailWithMessage("获取风险规则失败", c)
		return
	}
	response.OkWithData(rules, c)
}

func (a *riskAPI) UpdateRule(c *gin.Context) {
	var input assetRequest.RiskRuleUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage("风险规则参数不完整", c)
		return
	}
	rule, err := serviceRisk.UpdateRule(c.Request.Context(), input)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(rule, "风险规则已更新，新版本将在下次扫描生效", c)
}

func (a *riskAPI) StartScan(c *gin.Context) {
	var input assetRequest.StartRiskScan
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		response.FailWithMessage("扫描参数不正确", c)
		return
	}
	run, err := serviceRisk.StartScan("manual", utils.GetUserID(c), utils.GetUserName(c), input.RunID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(run, "资产风险扫描已启动", c)
}

func (a *riskAPI) ScanRuns(c *gin.Context) {
	var search assetRequest.RiskScanSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceRisk.ScanRuns(c.Request.Context(), search)
	if err != nil {
		response.FailWithMessage("获取扫描记录失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, "获取成功", c)
}

func (a *riskAPI) DeleteScanRuns(c *gin.Context) {
	var input assetRequest.RiskScanDelete
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage("扫描记录清理参数不正确", c)
		return
	}
	deleted, err := serviceRisk.DeleteScanRuns(c.Request.Context(), input)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"deleted": deleted}, fmt.Sprintf("已清理 %d 条扫描记录", deleted), c)
}

func (a *riskAPI) Acknowledge(c *gin.Context) { a.handleAction(c, "acknowledge") }
func (a *riskAPI) Resolve(c *gin.Context)     { a.handleAction(c, "resolve") }
func (a *riskAPI) Ignore(c *gin.Context)      { a.handleAction(c, "ignore") }
func (a *riskAPI) Reopen(c *gin.Context)      { a.handleAction(c, "reopen") }

func (a *riskAPI) handleAction(c *gin.Context, action string) {
	var input assetRequest.RiskAction
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage("风险处理参数不正确", c)
		return
	}
	actorID, actorName := utils.GetUserID(c), utils.GetUserName(c)
	var err error
	switch action {
	case "acknowledge":
		err = serviceRisk.Acknowledge(c.Request.Context(), input, actorID, actorName)
	case "resolve":
		err = serviceRisk.Resolve(c.Request.Context(), input, actorID, actorName)
	case "ignore":
		err = serviceRisk.Ignore(c.Request.Context(), input, actorID, actorName)
	case "reopen":
		err = serviceRisk.Reopen(c.Request.Context(), input, actorID, actorName)
	}
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("风险状态已更新", c)
}

func (a *riskAPI) Assign(c *gin.Context) {
	var input assetRequest.RiskAssignment
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage("风险分配参数不正确", c)
		return
	}
	if err := serviceRisk.Assign(c.Request.Context(), input, utils.GetUserID(c), utils.GetUserName(c)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("风险处理人已更新", c)
}
