package service

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/document/model"
	documentRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/document/model/request"
	documentResponse "github.com/flipped-aurora/gin-vue-admin/server/plugin/document/model/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/upload"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

var Document = new(documentService)

type documentService struct{}

const (
	maxDocumentSize       = 100 << 20 // 100MB
	MaxEditableBytes      = 2 << 20   // 2MB
	MaxUpdateRequestBytes = MaxEditableBytes*6 + 64<<10
)

var allowedDocumentExt = map[string]struct{}{
	".txt": {}, ".md": {}, ".markdown": {}, ".html": {}, ".htm": {}, ".json": {}, ".xml": {}, ".csv": {},
	".yaml": {}, ".yml": {}, ".ini": {}, ".conf": {}, ".log": {}, ".doc": {}, ".docs": {}, ".docx": {}, ".pdf": {}, ".xls": {}, ".xlsx": {},
}

var textLikeExt = map[string]struct{}{
	".txt": {}, ".md": {}, ".markdown": {}, ".html": {}, ".htm": {}, ".json": {}, ".xml": {}, ".csv": {},
	".yaml": {}, ".yml": {}, ".ini": {}, ".conf": {}, ".log": {},
}

func normalizeExt(name string) string { return strings.ToLower(filepath.Ext(name)) }

func trimTitle(filename string) string {
	name := strings.TrimSpace(filename)
	if name == "" {
		return "未命名文档"
	}
	ext := filepath.Ext(name)
	base := strings.TrimSpace(strings.TrimSuffix(name, ext))
	if base == "" {
		return name
	}
	return base
}

func buildInitialContent(file *multipart.FileHeader, ext string) string {
	if _, ok := textLikeExt[ext]; !ok || file.Size > MaxEditableBytes {
		return ""
	}
	reader, err := file.Open()
	if err != nil {
		return ""
	}
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, MaxEditableBytes+1))
	if err != nil || len(data) == 0 {
		return ""
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return text
}

func safeLocalPath(storePath, key string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", errors.New("文件Key为空")
	}
	cleanKey := filepath.Clean(strings.TrimLeft(strings.ReplaceAll(key, "\\", "/"), "/"))
	if cleanKey == "." || strings.HasPrefix(cleanKey, "..") {
		return "", errors.New("非法文件路径")
	}
	base := filepath.Clean(storePath)
	fullPath := filepath.Join(base, cleanKey)
	rel, err := filepath.Rel(base, fullPath)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", errors.New("非法文件路径")
	}
	return fullPath, nil
}

func documentContentType(ext, fallback string) string {
	if fallback = strings.TrimSpace(fallback); fallback != "" && fallback != "application/octet-stream" {
		return fallback
	}
	switch strings.ToLower(strings.TrimPrefix(ext, ".")) {
	case "docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "doc", "docs":
		return "application/msword"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "xls":
		return "application/vnd.ms-excel"
	case "pdf":
		return "application/pdf"
	case "md", "markdown", "txt", "log", "ini", "conf", "yaml", "yml", "csv":
		return "text/plain; charset=utf-8"
	case "html", "htm":
		return "text/html; charset=utf-8"
	case "json":
		return "application/json; charset=utf-8"
	case "xml":
		return "application/xml; charset=utf-8"
	}
	if contentType := mime.TypeByExtension("." + strings.TrimPrefix(ext, ".")); contentType != "" {
		return contentType
	}
	return "application/octet-stream"
}

func (s *documentService) Upload(file *multipart.FileHeader) (model.Document, error) {
	var doc model.Document
	if file == nil {
		return doc, errors.New("请选择要上传的文档")
	}
	if file.Size <= 0 || file.Size > maxDocumentSize {
		return doc, errors.New("文档大小必须在 100MB 以内")
	}
	ext := normalizeExt(file.Filename)
	if _, ok := allowedDocumentExt[ext]; !ok {
		return doc, errors.New("仅支持 txt、md、html、json、csv、doc、docx、pdf、xls、xlsx 等文档格式")
	}
	content := buildInitialContent(file, ext)
	fileURL, key, err := upload.NewOss().UploadFile(file)
	if err != nil {
		return doc, err
	}
	doc = model.Document{
		Title:        trimTitle(file.Filename),
		OriginalName: file.Filename,
		FileExt:      strings.TrimPrefix(ext, "."),
		FileSize:     file.Size,
		MimeType:     file.Header.Get("Content-Type"),
		StorageType:  global.GVA_CONFIG.System.OssType,
		FileKey:      key,
		FileURL:      fileURL,
		Content:      content,
		Editable:     true,
	}
	if err := global.GVA_DB.Create(&doc).Error; err != nil {
		_ = upload.NewOss().DeleteFile(key)
		return doc, err
	}
	return doc, nil
}

func (s *documentService) List(search documentRequest.DocumentSearch) ([]documentResponse.DocumentListItem, int64, error) {
	var list []documentResponse.DocumentListItem
	var total int64
	db := global.GVA_DB.Model(&model.Document{})
	if keyword := strings.TrimSpace(search.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("title ILIKE ? OR original_name ILIKE ? OR file_ext ILIKE ? OR remarks ILIKE ?", like, like, like, like)
	}
	if ext := strings.Trim(strings.ToLower(search.FileExt), ". "); ext != "" {
		db = db.Where("file_ext = ?", ext)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Select(`id, created_at, updated_at, title, original_name, file_ext,
		file_size, mime_type, storage_type, editable`).
		Order("updated_at DESC, created_at DESC").Scopes(search.Paginate()).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *documentService) Get(id uint) (model.Document, error) {
	var doc model.Document
	err := global.GVA_DB.First(&doc, id).Error
	return doc, err
}

func (s *documentService) UpdateContent(req documentRequest.UpdateContent) (model.Document, error) {
	if req.ID == 0 {
		return model.Document{}, errors.New("缺少文档 ID")
	}
	if len(req.Content) > MaxEditableBytes {
		return model.Document{}, errors.New("在线文档内容不能超过 2MB")
	}
	var doc model.Document
	if err := global.GVA_DB.First(&doc, req.ID).Error; err != nil {
		return doc, err
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = doc.Title
	}
	updates := map[string]any{
		"title":   title,
		"content": req.Content,
		"remarks": strings.TrimSpace(req.Remarks),
	}
	if err := global.GVA_DB.Model(&doc).Updates(updates).Error; err != nil {
		return doc, err
	}
	return s.Get(req.ID)
}

func (s *documentService) OpenFile(id uint) (model.Document, io.ReadCloser, string, error) {
	doc, err := s.Get(id)
	if err != nil {
		return doc, nil, "", err
	}
	if strings.TrimSpace(doc.FileKey) == "" && strings.TrimSpace(doc.FileURL) == "" {
		return doc, nil, "", errors.New("文档原文件不存在")
	}
	contentType := documentContentType(doc.FileExt, doc.MimeType)
	storageType := strings.TrimSpace(doc.StorageType)
	if storageType == "" {
		storageType = global.GVA_CONFIG.System.OssType
	}

	switch storageType {
	case "local":
		fullPath, err := safeLocalPath(global.GVA_CONFIG.Local.StorePath, doc.FileKey)
		if err != nil {
			return doc, nil, "", err
		}
		file, err := os.Open(fullPath)
		if err != nil {
			return doc, nil, "", err
		}
		return doc, file, contentType, nil
	case "minio":
		client, err := upload.GetMinio(
			global.GVA_CONFIG.Minio.Endpoint,
			global.GVA_CONFIG.Minio.AccessKeyId,
			global.GVA_CONFIG.Minio.AccessKeySecret,
			global.GVA_CONFIG.Minio.BucketName,
			global.GVA_CONFIG.Minio.UseSSL,
		)
		if err != nil {
			return doc, nil, "", err
		}
		object, err := client.Client.GetObject(context.Background(), global.GVA_CONFIG.Minio.BucketName, doc.FileKey, minio.GetObjectOptions{})
		if err != nil {
			return doc, nil, "", err
		}
		if _, err = object.Stat(); err != nil {
			_ = object.Close()
			return doc, nil, "", err
		}
		return doc, object, contentType, nil
	}

	if strings.HasPrefix(doc.FileURL, "http://") || strings.HasPrefix(doc.FileURL, "https://") {
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Get(doc.FileURL)
		if err != nil {
			return doc, nil, "", err
		}
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			_ = resp.Body.Close()
			return doc, nil, "", errors.New("远程文档读取失败")
		}
		if headerContentType := resp.Header.Get("Content-Type"); contentType == "application/octet-stream" && headerContentType != "" {
			contentType = headerContentType
		}
		return doc, resp.Body, contentType, nil
	}

	return doc, nil, "", errors.New("当前存储类型暂不支持原文件读取")
}

func (s *documentService) Delete(id uint) error {
	if id == 0 {
		return errors.New("缺少文档 ID")
	}
	var doc model.Document
	if err := global.GVA_DB.First(&doc, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return err
	}
	if doc.FileKey != "" {
		_ = upload.NewOss().DeleteFile(doc.FileKey)
	}
	return global.GVA_DB.Delete(&doc).Error
}
