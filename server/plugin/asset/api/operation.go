package api

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	assetRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Operation = new(operationAPI)

type operationAPI struct{}

func operationID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("业务单 ID 不正确", c)
		return 0, false
	}
	return uint(id), true
}

func (a *operationAPI) Create(c *gin.Context) {
	var req assetRequest.SaveOperation
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("业务单信息不完整", c)
		return
	}
	order, err := serviceOperation.Create(req, utils.GetUserID(c), utils.GetUserName(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(order, "业务单创建成功", c)
}

func (a *operationAPI) Update(c *gin.Context) {
	var req assetRequest.SaveOperation
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("业务单信息不完整", c)
		return
	}
	order, err := serviceOperation.Update(req, utils.GetUserID(c), utils.GetUserName(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(order, "业务单更新成功", c)
}

func (a *operationAPI) Submit(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	if err := serviceOperation.Submit(id, utils.GetUserID(c), utils.GetUserName(c)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("业务单已提交，资产状态已更新", c)
}

func (a *operationAPI) Delete(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	if err := serviceOperation.Delete(id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("草稿已删除", c)
}

func (a *operationAPI) Detail(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	order, err := serviceOperation.Get(id)
	if err != nil {
		response.FailWithMessage("业务单不存在", c)
		return
	}
	response.OkWithData(order, c)
}

func (a *operationAPI) List(c *gin.Context) {
	var search assetRequest.OperationSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceOperation.List(search)
	if err != nil {
		global.GVA_LOG.Error("获取资产业务单失败", zap.Error(err))
		response.FailWithMessage("获取业务单失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: search.Page, PageSize: search.PageSize,
	}, "获取成功", c)
}

func (a *operationAPI) AssetOptions(c *gin.Context) {
	var search assetRequest.OperationAssetSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := serviceOperation.AssetOptions(search)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}
