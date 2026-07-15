package upload

import (
	"bytes"
	"errors"
	"mime/multipart"
	"path"
	"strings"
)

// UploadBytes uploads generated in-memory content through the configured OSS
// implementation without forcing callers to create a temporary local file.
func UploadBytes(storage OSS, filename string, data []byte) (fileURL, key string, err error) {
	if storage == nil {
		return "", "", errors.New("OSS存储实例不能为空")
	}

	filename = path.Base(strings.ReplaceAll(strings.TrimSpace(filename), "\\", "/"))
	if filename == "" || filename == "." || filename == "/" {
		return "", "", errors.New("文件名不能为空")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", "", err
	}
	if _, err = part.Write(data); err != nil {
		return "", "", err
	}
	if err = writer.Close(); err != nil {
		return "", "", err
	}

	reader := multipart.NewReader(&body, writer.Boundary())
	form, err := reader.ReadForm(int64(len(data)) + 1024*1024)
	if err != nil {
		return "", "", err
	}
	defer form.RemoveAll()

	files := form.File["file"]
	if len(files) == 0 {
		return "", "", errors.New("生成上传文件失败")
	}

	return storage.UploadFile(files[0])
}
