package api

import (
	"errors"
	"net/http"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	commonResponse "github.com/WangMi2022/mit-assets-admin/server/model/common/response"
	assetRequest "github.com/WangMi2022/mit-assets-admin/server/plugin/asset/model/request"
	"github.com/WangMi2022/mit-assets-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const maxAssetRecognitionRequestSize = 6*(10<<20) + (1 << 20)

type recognitionAPI struct{}

func currentAssetRecognitionActor(c *gin.Context) (uint, uint) {
	return utils.GetUserID(c), utils.GetUserAuthorityId(c)
}

func (recognitionAPI) Create(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAssetRecognitionRequestSize)
	form, err := c.MultipartForm()
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			commonResponse.FailWithMessage("每批资产图片总大小不能超过 61MB", c)
			return
		}
		commonResponse.FailWithMessage("请选择资产照片或铭牌图片", c)
		return
	}
	defer form.RemoveAll()
	files := form.File["files"]
	if len(files) == 0 {
		files = form.File["file"]
	}
	userID, authorityID := currentAssetRecognitionActor(c)
	job, err := serviceRecognition.Create(files, userID, authorityID)
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithDetailed(job, "智能建档任务已创建", c)
}

func (recognitionAPI) List(c *gin.Context) {
	var search assetRequest.AssetRecognitionSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	userID, authorityID := currentAssetRecognitionActor(c)
	list, total, err := serviceRecognition.List(search, userID, authorityID)
	if err != nil {
		global.GVA_LOG.Error("获取资产智能建档任务失败", zap.Error(err))
		commonResponse.FailWithMessage("获取识别任务失败", c)
		return
	}
	search.Normalize()
	commonResponse.OkWithDetailed(commonResponse.PageResult{
		List: list, Total: total, Page: search.Page, PageSize: search.PageSize,
	}, "获取成功", c)
}

func (recognitionAPI) Detail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	userID, authorityID := currentAssetRecognitionActor(c)
	job, err := serviceRecognition.Get(id, userID, authorityID)
	if err != nil {
		commonResponse.FailWithMessage("识别任务不存在", c)
		return
	}
	commonResponse.OkWithData(job, c)
}

func (recognitionAPI) Retry(c *gin.Context) {
	var input assetRequest.AssetRecognitionID
	if err := c.ShouldBindJSON(&input); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	userID, authorityID := currentAssetRecognitionActor(c)
	if err := serviceRecognition.Retry(input.ID, userID, authorityID); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithMessage("识别任务已重新排队", c)
}

func (recognitionAPI) Draft(c *gin.Context) {
	var input assetRequest.AssetRecognitionDraftUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	userID, authorityID := currentAssetRecognitionActor(c)
	job, err := serviceRecognition.SaveDraft(input, userID, authorityID)
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithDetailed(job, "草稿已保存", c)
}

func (recognitionAPI) Confirm(c *gin.Context) {
	var input assetRequest.AssetRecognitionID
	if err := c.ShouldBindJSON(&input); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	userID, authorityID := currentAssetRecognitionActor(c)
	asset, err := serviceRecognition.Confirm(input.ID, userID, authorityID)
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithDetailed(asset, "资产建档成功", c)
}

func (recognitionAPI) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	userID, authorityID := currentAssetRecognitionActor(c)
	if err := serviceRecognition.Delete(id, userID, authorityID); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithMessage("识别任务已删除", c)
}
