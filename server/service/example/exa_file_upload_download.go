package example

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"strings"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/example"
	"github.com/WangMi2022/mit-assets-admin/server/model/example/request"
	"github.com/WangMi2022/mit-assets-admin/server/utils/upload"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const mediaPreviewURLTTL = 12 * time.Hour

const (
	mediaPreviewPurposeRecord     = "media-preview"
	mediaPreviewPurposeConfigured = "configured-media-preview"
)

var (
	ErrInvalidMediaPreview = errors.New("媒体预览凭证无效")
	ErrMediaPreviewMissing = errors.New("媒体预览文件不存在")
)

type mediaPreviewURLSigner func(context.Context, string, time.Duration) (string, error)

type mediaPreviewClaims struct {
	Key     string `json:"key"`
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

func validMediaObjectKey(key string) bool {
	if key == "" || key != strings.TrimSpace(key) || len(key) > 1024 || strings.HasPrefix(key, "/") ||
		strings.Contains(key, "..") || strings.ContainsAny(key, "\\\x00\r\n") {
		return false
	}
	for _, part := range strings.Split(key, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}

func createMediaPreviewURL(_ context.Context, key string, expires time.Duration) (string, error) {
	return createMediaPreviewURLForPurpose(key, mediaPreviewPurposeRecord, expires)
}

func createMediaPreviewURLForPurpose(key, purpose string, expires time.Duration) (string, error) {
	key = strings.TrimSpace(key)
	secret := strings.TrimSpace(global.GVA_CONFIG.JWT.SigningKey)
	if !validMediaObjectKey(key) || secret == "" ||
		(purpose != mediaPreviewPurposeRecord && purpose != mediaPreviewPurposeConfigured) {
		return "", ErrInvalidMediaPreview
	}
	if expires <= 0 {
		expires = mediaPreviewURLTTL
	}
	now := time.Now()
	claims := mediaPreviewClaims{
		Key: key, Purpose: purpose,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "media-preview", Subject: key,
			IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(expires)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	query := url.Values{"key": []string{key}, "token": []string{token}}
	return "/api/fileUploadAndDownload/preview?" + query.Encode(), nil
}

func validateMediaPreviewToken(tokenString, key string) bool {
	_, ok := parseMediaPreviewToken(tokenString, key)
	return ok
}

func parseMediaPreviewToken(tokenString, key string) (*mediaPreviewClaims, bool) {
	secret := strings.TrimSpace(global.GVA_CONFIG.JWT.SigningKey)
	key = strings.TrimSpace(key)
	if secret == "" || tokenString == "" || !validMediaObjectKey(key) {
		return nil, false
	}
	claims := new(mediaPreviewClaims)
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidMediaPreview
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	validPurpose := claims.Purpose == mediaPreviewPurposeRecord || claims.Purpose == mediaPreviewPurposeConfigured
	if err != nil || !token.Valid || !validPurpose || claims.Key != key || claims.Subject != key {
		return nil, false
	}
	return claims, true
}

// ResolveConfiguredMediaPreviewURL signs a URL already trusted by an
// administrator-owned configuration record, such as the public brand logo.
// External URLs remain untouched.
func ResolveConfiguredMediaPreviewURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" || !strings.EqualFold(global.GVA_CONFIG.System.OssType, "minio") {
		return rawURL, nil
	}
	bucketURL := strings.TrimRight(strings.TrimSpace(global.GVA_CONFIG.Minio.BucketUrl), "/")
	prefix := bucketURL + "/"
	if bucketURL == "" || !strings.HasPrefix(rawURL, prefix) {
		return rawURL, nil
	}
	key := strings.TrimPrefix(rawURL, prefix)
	if !validMediaObjectKey(key) || rawURL != prefix+key {
		return "", ErrInvalidMediaPreview
	}
	return createMediaPreviewURLForPurpose(key, mediaPreviewPurposeConfigured, mediaPreviewURLTTL)
}

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
		signer = createMediaPreviewURL
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

// ResolveMediaPreviewURL returns a short-lived same-origin browser URL for a
// managed private MinIO object while leaving local and external URLs unchanged.
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

	files := []example.ExaFileUploadAndDownload{file}
	if err := attachMediaPreviewURLs(ctx, files, ossType, bucketURL, createMediaPreviewURL); err != nil {
		return rawURL, err
	}
	return files[0].PreviewURL, nil
}

// ResolveMediaPreviewObject validates the short-lived application token and
// resolves only a media record that belongs to the configured private bucket.
func (e *FileUploadAndDownloadService) ResolveMediaPreviewObject(ctx context.Context, key, token string) (example.ExaFileUploadAndDownload, error) {
	key = strings.TrimSpace(key)
	claims, validToken := parseMediaPreviewToken(token, key)
	if !strings.EqualFold(global.GVA_CONFIG.System.OssType, "minio") || !validToken {
		return example.ExaFileUploadAndDownload{}, ErrInvalidMediaPreview
	}
	if claims.Purpose == mediaPreviewPurposeConfigured {
		return example.ExaFileUploadAndDownload{Key: key}, nil
	}
	bucketURL := strings.TrimRight(strings.TrimSpace(global.GVA_CONFIG.Minio.BucketUrl), "/")
	if bucketURL == "" {
		return example.ExaFileUploadAndDownload{}, ErrInvalidMediaPreview
	}
	var file example.ExaFileUploadAndDownload
	err := global.GVA_DB.WithContext(ctx).Where("key = ? AND url = ?", key, bucketURL+"/"+key).First(&file).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return example.ExaFileUploadAndDownload{}, ErrMediaPreviewMissing
	}
	if err != nil {
		return example.ExaFileUploadAndDownload{}, fmt.Errorf("查询媒体预览文件: %w", err)
	}
	return file, nil
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
