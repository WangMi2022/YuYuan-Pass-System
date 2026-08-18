package example

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/example"
	"github.com/WangMi2022/mit-assets-admin/server/model/example/request"
	"github.com/WangMi2022/mit-assets-admin/server/utils/upload"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const mediaPreviewURLTTL = 15 * time.Minute

type mediaPreviewURLSigner func(context.Context, string, time.Duration) (string, error)

//@author: [piexlmax](https://github.com/piexlmax)
//@function: Upload
//@description: 创建文件上传记录
//@param: file model.ExaFileUploadAndDownload
//@return: error

func (e *FileUploadAndDownloadService) Upload(file example.ExaFileUploadAndDownload) error {
	return global.GVA_DB.Create(&file).Error
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: FindFile
//@description: 查询文件记录
//@param: id uint
//@return: model.ExaFileUploadAndDownload, error

func (e *FileUploadAndDownloadService) FindFile(id uint) (example.ExaFileUploadAndDownload, error) {
	var file example.ExaFileUploadAndDownload
	err := global.GVA_DB.Where("id = ?", id).First(&file).Error
	return file, err
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: DeleteFile
//@description: 删除文件记录
//@param: file model.ExaFileUploadAndDownload
//@return: err error

func (e *FileUploadAndDownloadService) DeleteFile(file example.ExaFileUploadAndDownload) (err error) {
	var fileFromDb example.ExaFileUploadAndDownload
	fileFromDb, err = e.FindFile(file.ID)
	if err != nil {
		return
	}
	oss := upload.NewOss()
	if err = oss.DeleteFile(fileFromDb.Key); err != nil {
		return errors.New("文件删除失败")
	}
	err = global.GVA_DB.Where("id = ?", file.ID).Unscoped().Delete(&file).Error
	return err
}

// EditFileName 编辑文件名或者备注
func (e *FileUploadAndDownloadService) EditFileName(file example.ExaFileUploadAndDownload) (err error) {
	var fileFromDb example.ExaFileUploadAndDownload
	return global.GVA_DB.Where("id = ?", file.ID).First(&fileFromDb).Update("name", file.Name).Error
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: GetFileRecordInfoList
//@description: 分页获取数据
//@param: info request.ExaAttachmentCategorySearch
//@return: list interface{}, total int64, err error

func (e *FileUploadAndDownloadService) GetFileRecordInfoList(ctx context.Context, info request.ExaAttachmentCategorySearch) (list []example.ExaFileUploadAndDownload, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&example.ExaFileUploadAndDownload{})

	if len(info.Keyword) > 0 {
		db = db.Where("name LIKE ?", "%"+info.Keyword+"%")
	}

	if info.ClassId > 0 {
		db = db.Where("class_id = ?", info.ClassId)
	}

	if strings.EqualFold(info.FileType, "image") {
		db = db.Where("LOWER(tag) IN ?", []string{"jpg", "jpeg", "png", "gif", "bmp", "webp", "svg", "avif"})
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Scopes(info.Paginate()).Order("id desc").Find(&list).Error
	if err != nil {
		return
	}

	var signer mediaPreviewURLSigner
	if strings.EqualFold(global.GVA_CONFIG.System.OssType, "minio") {
		minioClient, minioErr := upload.GetMinio(
			global.GVA_CONFIG.Minio.Endpoint,
			global.GVA_CONFIG.Minio.AccessKeyId,
			global.GVA_CONFIG.Minio.AccessKeySecret,
			global.GVA_CONFIG.Minio.BucketName,
			global.GVA_CONFIG.Minio.UseSSL,
		)
		if minioErr != nil {
			err = fmt.Errorf("初始化媒体预览存储: %w", minioErr)
			return
		}
		signer = minioClient.PreviewURL
	}
	if previewErr := attachMediaPreviewURLs(ctx, list, global.GVA_CONFIG.System.OssType, global.GVA_CONFIG.Minio.BucketUrl, signer); previewErr != nil {
		global.GVA_LOG.Warn("部分媒体预览链接签发失败", zap.Error(previewErr))
	}
	return list, total, nil
}

func attachMediaPreviewURLs(ctx context.Context, files []example.ExaFileUploadAndDownload, ossType, bucketURL string, signer mediaPreviewURLSigner) error {
	bucketURL = strings.TrimRight(strings.TrimSpace(bucketURL), "/")
	var previewErrors []error
	for index := range files {
		files[index].PreviewURL = files[index].Url
		if !strings.EqualFold(ossType, "minio") || signer == nil || bucketURL == "" {
			continue
		}
		key := strings.TrimLeft(strings.TrimSpace(files[index].Key), "/")
		if key == "" || strings.TrimSpace(files[index].Url) != bucketURL+"/"+key {
			continue
		}
		previewURL, err := resolveMediaPreviewURL(ctx, files[index].Url, ossType, bucketURL, signer)
		if err != nil {
			previewErrors = append(previewErrors, fmt.Errorf("签发媒体预览链接 %q: %w", key, err))
			continue
		}
		files[index].PreviewURL = previewURL
	}
	return errors.Join(previewErrors...)
}

func resolveMediaPreviewURL(ctx context.Context, rawURL, ossType, bucketURL string, signer mediaPreviewURLSigner) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	bucketURL = strings.TrimRight(strings.TrimSpace(bucketURL), "/")
	if rawURL == "" || !strings.EqualFold(ossType, "minio") || bucketURL == "" || signer == nil {
		return rawURL, nil
	}

	prefix := bucketURL + "/"
	if !strings.HasPrefix(rawURL, prefix) {
		return rawURL, nil
	}
	key := strings.TrimLeft(strings.TrimSpace(strings.TrimPrefix(rawURL, prefix)), "/")
	if key == "" {
		return rawURL, nil
	}

	previewURL, err := signer(ctx, key, mediaPreviewURLTTL)
	if err != nil {
		return rawURL, err
	}
	return previewURL, nil
}

// ResolveMediaPreviewURL returns a short-lived browser URL for managed private
// MinIO objects while leaving local and external URLs unchanged.
func (e *FileUploadAndDownloadService) ResolveMediaPreviewURL(ctx context.Context, rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	ossType := global.GVA_CONFIG.System.OssType
	bucketURL := global.GVA_CONFIG.Minio.BucketUrl
	if rawURL == "" || !strings.EqualFold(ossType, "minio") || strings.TrimSpace(bucketURL) == "" {
		return rawURL, nil
	}

	var file example.ExaFileUploadAndDownload
	if err := global.GVA_DB.WithContext(ctx).Select("url", "key").Where("url = ?", rawURL).First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return rawURL, nil
		}
		return rawURL, fmt.Errorf("查询头像媒体记录: %w", err)
	}

	minioClient, err := upload.GetMinio(
		global.GVA_CONFIG.Minio.Endpoint,
		global.GVA_CONFIG.Minio.AccessKeyId,
		global.GVA_CONFIG.Minio.AccessKeySecret,
		global.GVA_CONFIG.Minio.BucketName,
		global.GVA_CONFIG.Minio.UseSSL,
	)
	if err != nil {
		return rawURL, fmt.Errorf("初始化头像预览存储: %w", err)
	}
	files := []example.ExaFileUploadAndDownload{file}
	if err := attachMediaPreviewURLs(ctx, files, ossType, bucketURL, minioClient.PreviewURL); err != nil {
		return rawURL, err
	}
	return files[0].PreviewURL, nil
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: UploadFile
//@description: 根据配置文件判断是文件上传到本地或者七牛云
//@param: header *multipart.FileHeader, noSave string
//@return: file model.ExaFileUploadAndDownload, err error

func (e *FileUploadAndDownloadService) UploadFile(header *multipart.FileHeader, noSave string, classId int) (file example.ExaFileUploadAndDownload, err error) {
	oss := upload.NewOss()
	filePath, key, uploadErr := oss.UploadFile(header)
	if uploadErr != nil {
		return file, uploadErr
	}
	s := strings.Split(header.Filename, ".")
	f := example.ExaFileUploadAndDownload{
		Url:     filePath,
		Name:    header.Filename,
		ClassId: classId,
		Tag:     s[len(s)-1],
		Key:     key,
	}
	if noSave == "0" {
		// 检查是否已存在相同key的记录
		var existingFile example.ExaFileUploadAndDownload
		err = global.GVA_DB.Where(&example.ExaFileUploadAndDownload{Key: key}).First(&existingFile).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return f, e.Upload(f)
		}
		return f, err
	}
	return f, nil
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: ImportURL
//@description: 导入URL
//@param: file model.ExaFileUploadAndDownload
//@return: error

func (e *FileUploadAndDownloadService) ImportURL(file *[]example.ExaFileUploadAndDownload) error {
	return global.GVA_DB.Create(&file).Error
}
