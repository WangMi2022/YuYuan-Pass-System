package api

import (
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	documentRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/document/model/request"
	documentService "github.com/flipped-aurora/gin-vue-admin/server/plugin/document/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Document = new(documentAPI)

type documentAPI struct{}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("ID 参数不正确", c)
		return 0, false
	}
	return uint(id), true
}

func (a *documentAPI) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("请选择要上传的文档", c)
		return
	}
	doc, err := serviceDocument.Upload(file)
	if err != nil {
		global.GVA_LOG.Error("上传文档失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(doc, "文档上传成功", c)
}

func (a *documentAPI) List(c *gin.Context) {
	var search documentRequest.DocumentSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceDocument.List(search)
	if err != nil {
		global.GVA_LOG.Error("获取文档列表失败", zap.Error(err))
		response.FailWithMessage("获取文档列表失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: search.Page, PageSize: search.PageSize,
	}, "获取成功", c)
}

func (a *documentAPI) File(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	doc, reader, contentType, err := serviceDocument.OpenFile(id)
	if err != nil {
		global.GVA_LOG.Error("读取文档原文件失败", zap.Error(err))
		response.FailWithMessage("读取文档原文件失败", c)
		return
	}
	defer reader.Close()
	filename := doc.OriginalName
	if filename == "" {
		filename = doc.Title
	}
	if filename == "" {
		filename = "document"
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "inline; filename*=UTF-8''"+url.QueryEscape(filename))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Stream(func(w io.Writer) bool {
		if _, err := io.Copy(w, reader); err != nil {
			global.GVA_LOG.Warn("输出文档原文件失败", zap.Error(err))
		}
		return false
	})
}

func (a *documentAPI) Detail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	doc, err := serviceDocument.Get(id)
	if err != nil {
		response.FailWithMessage("文档不存在", c)
		return
	}
	response.OkWithData(doc, c)
}

func (a *documentAPI) UpdateContent(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, documentService.MaxUpdateRequestBytes)
	var req documentRequest.UpdateContent
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	doc, err := serviceDocument.UpdateContent(req)
	if err != nil {
		global.GVA_LOG.Error("保存文档失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(doc, "文档保存成功", c)
}

func (a *documentAPI) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := serviceDocument.Delete(id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("文档删除成功", c)
}
