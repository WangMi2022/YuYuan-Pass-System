package api

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	commonResponse "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	invoiceRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/service"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type invoiceAPI struct{}

const (
	maxInvoiceUploadBatch       = 5
	maxInvoiceUploadRequestSize = maxInvoiceUploadBatch*(10<<20) + (1 << 20)
)

type invoiceUploadFailure struct {
	FileName string `json:"fileName"`
	Message  string `json:"message"`
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		commonResponse.FailWithMessage("ID 参数不正确", c)
		return 0, false
	}
	return uint(id), true
}

func currentScope(c *gin.Context) service.AccessScope {
	return service.ResolveAccessScope(utils.GetUserID(c), utils.GetUserAuthorityId(c))
}

func (invoiceAPI) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxInvoiceUploadRequestSize)
	form, formErr := c.MultipartForm()
	if formErr != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(formErr, &maxBytesError) {
			commonResponse.FailWithMessage("每批发票图片总大小不能超过 51MB", c)
			return
		}
		commonResponse.FailWithMessage("请选择发票图片", c)
		return
	}
	if len(form.File["files"]) > 0 {
		files := form.File["files"]
		if len(files) > maxInvoiceUploadBatch {
			commonResponse.FailWithMessage("每批最多上传 5 张发票图片", c)
			return
		}
		succeeded := make([]any, 0, len(files))
		failed := make([]invoiceUploadFailure, 0)
		for _, file := range files {
			invoice, err := serviceInvoice.Upload(file, utils.GetUserID(c), utils.GetUserAuthorityId(c))
			if err != nil {
				failed = append(failed, invoiceUploadFailure{FileName: file.Filename, Message: err.Error()})
				continue
			}
			succeeded = append(succeeded, invoice)
		}
		result := gin.H{"succeeded": succeeded, "failed": failed}
		if len(succeeded) == 0 {
			commonResponse.FailWithDetailed(result, "本批发票均上传失败", c)
			return
		}
		commonResponse.OkWithDetailed(result, "发票批量上传处理完成", c)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		commonResponse.FailWithMessage("请选择发票图片", c)
		return
	}
	invoice, err := serviceInvoice.Upload(file, utils.GetUserID(c), utils.GetUserAuthorityId(c))
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithDetailed(invoice, "发票已上传，正在进入识别队列", c)
}

func (invoiceAPI) List(c *gin.Context) {
	var search invoiceRequest.InvoiceSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceInvoice.List(search, currentScope(c))
	if err != nil {
		global.GVA_LOG.Error("获取发票列表失败", zap.Error(err))
		commonResponse.FailWithMessage("获取发票列表失败", c)
		return
	}
	commonResponse.OkWithDetailed(commonResponse.PageResult{
		List: list, Total: total, Page: search.Page, PageSize: search.PageSize,
	}, "获取成功", c)
}

func (invoiceAPI) Detail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	invoice, err := serviceInvoice.Get(id, currentScope(c))
	if err != nil {
		commonResponse.FailWithMessage("发票不存在或无权访问", c)
		return
	}
	commonResponse.OkWithData(invoice, c)
}

func (invoiceAPI) Update(c *gin.Context) {
	var request invoiceRequest.InvoiceUpdate
	if err := c.ShouldBindJSON(&request); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	invoice, err := serviceInvoice.Update(request, currentScope(c))
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithDetailed(invoice, "发票核对信息已保存", c)
}

func (invoiceAPI) Confirm(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	invoice, err := serviceInvoice.Confirm(id, currentScope(c))
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithDetailed(invoice, "发票已确认并纳入正式统计", c)
}

func (invoiceAPI) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := serviceInvoice.Delete(id, currentScope(c)); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithMessage("发票已删除", c)
}

func (invoiceAPI) Retry(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := serviceRecognition.Retry(id, currentScope(c)); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithMessage("已重新加入识别队列", c)
}

func (invoiceAPI) Recheck(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	result, err := serviceRecognition.Recheck(c.Request.Context(), id, currentScope(c))
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithDetailed(result, "模型核对完成", c)
}

func (invoiceAPI) TestProviderConnection(c *gin.Context) {
	if utils.GetUserAuthorityId(c) != 888 {
		commonResponse.FailWithMessage("仅超级管理员可测试识别服务连接", c)
		return
	}
	var request invoiceRequest.ProviderConnectionTest
	if err := c.ShouldBindJSON(&request); err != nil {
		commonResponse.FailWithMessage("连接测试参数不正确", c)
		return
	}
	protocol, err := serviceRecognition.TestProviderConnection(c.Request.Context(), request.Target, request.Config)
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithDetailed(gin.H{"protocol": protocol}, "连接测试成功", c)
}

func (invoiceAPI) Dashboard(c *gin.Context) {
	dashboard, err := serviceInvoice.Dashboard(currentScope(c))
	if err != nil {
		global.GVA_LOG.Error("获取流水统计失败", zap.Error(err))
		commonResponse.FailWithMessage("获取流水统计失败", c)
		return
	}
	commonResponse.OkWithData(dashboard, c)
}

func (invoiceAPI) File(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	invoice, reader, err := serviceInvoice.OpenFile(c.Request.Context(), id, currentScope(c))
	if err != nil {
		commonResponse.FailWithMessage("发票原图不存在或无权访问", c)
		return
	}
	defer reader.Close()
	contentType := invoice.MimeType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	filename := invoice.FileName
	if filename == "" {
		filename = "invoice"
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "inline; filename*=UTF-8''"+url.QueryEscape(filename))
	c.Header("Cache-Control", "private, max-age=300")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Stream(func(writer io.Writer) bool {
		if _, copyErr := io.Copy(writer, reader); copyErr != nil {
			global.GVA_LOG.Warn("输出发票原图失败", zap.Error(copyErr), zap.Uint("invoiceID", id))
		}
		return false
	})
}
