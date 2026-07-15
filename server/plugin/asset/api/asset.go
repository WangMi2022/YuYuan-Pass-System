package api

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	commonResponse "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
	assetRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/upload"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

var Asset = new(assetAPI)

type assetAPI struct{}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		commonResponse.FailWithMessage("ID 参数不正确", c)
		return 0, false
	}
	return uint(id), true
}

func (a *assetAPI) Create(c *gin.Context) {
	var asset model.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceAsset.Create(&asset); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithDetailed(asset, "资产创建成功", c)
}

func (a *assetAPI) Update(c *gin.Context) {
	var asset model.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceAsset.Update(&asset); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithMessage("资产更新成功", c)
}

func (a *assetAPI) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := serviceAsset.Delete(id); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithMessage("资产删除成功", c)
}

func (a *assetAPI) Detail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	asset, err := serviceAsset.Get(id)
	if err != nil {
		commonResponse.FailWithMessage("资产不存在", c)
		return
	}
	commonResponse.OkWithData(asset, c)
}

func (a *assetAPI) List(c *gin.Context) {
	var search assetRequest.AssetSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceAsset.List(search)
	if err != nil {
		global.GVA_LOG.Error("获取资产列表失败", zap.Error(err))
		commonResponse.FailWithMessage("获取资产列表失败", c)
		return
	}
	commonResponse.OkWithDetailed(commonResponse.PageResult{
		List: list, Total: total, Page: search.Page, PageSize: search.PageSize,
	}, "获取成功", c)
}

func (a *assetAPI) Dashboard(c *gin.Context) {
	data, err := serviceAsset.Dashboard()
	if err != nil {
		global.GVA_LOG.Error("获取资产统计失败", zap.Error(err))
		commonResponse.FailWithMessage("获取资产统计失败", c)
		return
	}
	commonResponse.OkWithData(data, c)
}

func (a *assetAPI) UploadPhoto(c *gin.Context) {
	header, err := c.FormFile("file")
	if err != nil {
		commonResponse.FailWithMessage("请选择要上传的图片", c)
		return
	}
	if header.Size <= 0 || header.Size > 10*1024*1024 {
		commonResponse.FailWithMessage("图片大小必须在 10MB 以内", c)
		return
	}
	file, err := header.Open()
	if err != nil {
		commonResponse.FailWithMessage("读取图片失败", c)
		return
	}
	buffer := make([]byte, 512)
	n, _ := file.Read(buffer)
	_ = file.Close()
	contentType := http.DetectContentType(buffer[:n])
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" && contentType != "image/gif" {
		commonResponse.FailWithMessage("仅支持 JPG、PNG、WebP、GIF 图片", c)
		return
	}
	_, key, err := upload.NewOss().UploadFile(header)
	if err != nil {
		global.GVA_LOG.Error("上传资产图片失败", zap.Error(err))
		commonResponse.FailWithMessage("上传到 RustFS 失败", c)
		return
	}
	photo := model.Photo{
		Name: header.Filename,
		Key:  key,
		URL:  "/api/asset/photo?key=" + url.QueryEscape(key),
	}
	commonResponse.OkWithDetailed(photo, "图片上传成功", c)
}

func validObjectKey(key string) bool {
	prefix := strings.Trim(strings.TrimSpace(global.GVA_CONFIG.Minio.BasePath), "/")
	if prefix == "" {
		prefix = "uploads"
	}
	return strings.HasPrefix(key, prefix+"/") && !strings.Contains(key, "..")
}

func (a *assetAPI) DeletePhoto(c *gin.Context) {
	key := c.Query("key")
	if !validObjectKey(key) {
		commonResponse.FailWithMessage("对象路径不正确", c)
		return
	}
	if err := upload.NewOss().DeleteFile(key); err != nil {
		commonResponse.FailWithMessage("删除图片失败", c)
		return
	}
	commonResponse.OkWithMessage("图片删除成功", c)
}

// Photo 从私有 RustFS 桶中代理读取图片，避免把存储凭证或内网地址暴露给浏览器。
func (a *assetAPI) Photo(c *gin.Context) {
	key := c.Query("key")
	if !validObjectKey(key) {
		c.Status(http.StatusBadRequest)
		return
	}
	client, err := upload.GetMinio(
		global.GVA_CONFIG.Minio.Endpoint,
		global.GVA_CONFIG.Minio.AccessKeyId,
		global.GVA_CONFIG.Minio.AccessKeySecret,
		global.GVA_CONFIG.Minio.BucketName,
		global.GVA_CONFIG.Minio.UseSSL,
	)
	if err != nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	object, err := client.Client.GetObject(c.Request.Context(), global.GVA_CONFIG.Minio.BucketName, key, minio.GetObjectOptions{})
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer object.Close()
	stat, err := object.Stat()
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	contentType := stat.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Cache-Control", "private, max-age=3600")
	c.DataFromReader(http.StatusOK, stat.Size, contentType, object, nil)
}
