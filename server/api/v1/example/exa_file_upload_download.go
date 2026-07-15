package example

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	exampleRes "github.com/flipped-aurora/gin-vue-admin/server/model/example/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FileUploadAndDownloadApi struct{}

// UploadFile
// @Tags      ExaFileUploadAndDownload
// @Summary   上传文件示例
// @Security  ApiKeyAuth
// @accept    multipart/form-data
// @Produce   application/json
// @Param     file  formData  file                                                           true  "上传文件示例"
// @Success   200   {object}  response.Response{data=exampleRes.ExaFileResponse,msg=string}  "上传文件示例,返回包括文件详情"
// @Router    /fileUploadAndDownload/upload [post]
func (b *FileUploadAndDownloadApi) UploadFile(c *gin.Context) {
	var file example.ExaFileUploadAndDownload
	noSave := c.DefaultQuery("noSave", "0")
	_, header, err := c.Request.FormFile("file")
	classId, _ := strconv.Atoi(c.DefaultPostForm("classId", "0"))
	if err != nil {
		global.GVA_LOG.Error("接收文件失败!", zap.Error(err))
		response.FailWithMessage("接收文件失败", c)
		return
	}
	if err = validateImageUpload(header); err != nil {
		global.GVA_LOG.Warn("上传文件类型不允许", zap.Error(err), zap.String("filename", header.Filename))
		response.FailWithMessage(err.Error(), c)
		return
	}
	file, err = fileUploadAndDownloadService.UploadFile(header, noSave, classId) // 文件上传后拿到文件路径
	if err != nil {
		global.GVA_LOG.Error("上传文件失败!", zap.Error(err))
		response.FailWithMessage("上传文件失败", c)
		return
	}
	response.OkWithDetailed(exampleRes.ExaFileResponse{File: file}, "上传成功", c)
}

func validateImageUpload(header *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !isAllowedImageExt(ext) {
		return errors.New("媒体库仅允许上传图片文件")
	}

	file, err := header.Open()
	if err != nil {
		return errors.New("读取上传文件失败")
	}
	defer file.Close()

	buffer := make([]byte, 4096)
	n, readErr := file.Read(buffer)
	if readErr != nil && readErr != io.EOF {
		return errors.New("读取上传文件失败")
	}
	sniff := buffer[:n]
	detected := strings.ToLower(http.DetectContentType(sniff))
	if strings.HasPrefix(detected, "image/") {
		return nil
	}
	if ext == ".svg" && bytes.Contains(bytes.ToLower(sniff), []byte("<svg")) {
		return nil
	}
	if ext == ".webp" && len(sniff) >= 12 && bytes.Equal(sniff[:4], []byte("RIFF")) && bytes.Equal(sniff[8:12], []byte("WEBP")) {
		return nil
	}
	if ext == ".avif" && bytes.Contains(sniff[:min(len(sniff), 32)], []byte("ftypavif")) {
		return nil
	}
	if ext == ".bmp" && len(sniff) >= 2 && bytes.Equal(sniff[:2], []byte("BM")) {
		return nil
	}
	return errors.New("媒体库仅允许上传图片文件")
}

func isAllowedImageExt(ext string) bool {
	allowedExt := map[string]struct{}{
		".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {}, ".bmp": {}, ".webp": {}, ".svg": {}, ".avif": {},
	}
	_, ok := allowedExt[strings.ToLower(ext)]
	return ok
}

func imageExtFromURL(rawURL string) string {
	clean := strings.Split(strings.Split(rawURL, "?")[0], "#")[0]
	return strings.ToLower(filepath.Ext(clean))
}

// EditFileName 编辑文件名或者备注
func (b *FileUploadAndDownloadApi) EditFileName(c *gin.Context) {
	var file example.ExaFileUploadAndDownload
	err := c.ShouldBindJSON(&file)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = fileUploadAndDownloadService.EditFileName(file)
	if err != nil {
		global.GVA_LOG.Error("编辑失败!", zap.Error(err))
		response.FailWithMessage("编辑失败", c)
		return
	}
	response.OkWithMessage("编辑成功", c)
}

// DeleteFile
// @Tags      ExaFileUploadAndDownload
// @Summary   删除文件
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     data  body      example.ExaFileUploadAndDownload  true  "传入文件里面id即可"
// @Success   200   {object}  response.Response{msg=string}     "删除文件"
// @Router    /fileUploadAndDownload/deleteFile [post]
func (b *FileUploadAndDownloadApi) DeleteFile(c *gin.Context) {
	var file example.ExaFileUploadAndDownload
	err := c.ShouldBindJSON(&file)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := fileUploadAndDownloadService.DeleteFile(file); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetFileList
// @Tags      ExaFileUploadAndDownload
// @Summary   分页文件列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.ExaAttachmentCategorySearch                                        true  "页码, 每页大小, 分类id"
// @Success   200   {object}  response.Response{data=response.PageResult,msg=string}  "分页文件列表,返回包括列表,总数,页码,每页数量"
// @Router    /fileUploadAndDownload/getFileList [post]
func (b *FileUploadAndDownloadApi) GetFileList(c *gin.Context) {
	var pageInfo request.ExaAttachmentCategorySearch
	err := c.ShouldBindJSON(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := fileUploadAndDownloadService.GetFileRecordInfoList(pageInfo)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}

// ImportURL
// @Tags      ExaFileUploadAndDownload
// @Summary   导入URL
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     data  body      example.ExaFileUploadAndDownload  true  "对象"
// @Success   200   {object}  response.Response{msg=string}     "导入URL"
// @Router    /fileUploadAndDownload/importURL [post]
func (b *FileUploadAndDownloadApi) ImportURL(c *gin.Context) {
	var file []example.ExaFileUploadAndDownload
	err := c.ShouldBindJSON(&file)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	for index := range file {
		ext := imageExtFromURL(file[index].Url)
		if !isAllowedImageExt(ext) {
			response.FailWithMessage("仅允许导入图片URL", c)
			return
		}
		file[index].Tag = strings.TrimPrefix(ext, ".")
	}
	if err := fileUploadAndDownloadService.ImportURL(&file); err != nil {
		global.GVA_LOG.Error("导入URL失败!", zap.Error(err))
		response.FailWithMessage("导入URL失败", c)
		return
	}
	response.OkWithMessage("导入URL成功", c)
}
