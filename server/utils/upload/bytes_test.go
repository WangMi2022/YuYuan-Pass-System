package upload

import (
	"errors"
	"io"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/require"
)

type recordingOSS struct {
	filename string
	content  []byte
	err      error
}

func (r *recordingOSS) UploadFile(file *multipart.FileHeader) (string, string, error) {
	if r.err != nil {
		return "", "", r.err
	}

	source, err := file.Open()
	if err != nil {
		return "", "", err
	}
	defer source.Close()

	r.filename = file.Filename
	r.content, err = io.ReadAll(source)
	if err != nil {
		return "", "", err
	}
	return "https://oss.example/excel/report.xlsx", "excel/report.xlsx", nil
}

func (*recordingOSS) DeleteFile(string) error { return nil }

func TestUploadBytes(t *testing.T) {
	storage := &recordingOSS{}
	content := []byte("generated excel content")

	fileURL, key, err := UploadBytes(storage, `../exports\\资产报表.xlsx`, content)

	require.NoError(t, err)
	require.Equal(t, "https://oss.example/excel/report.xlsx", fileURL)
	require.Equal(t, "excel/report.xlsx", key)
	require.Equal(t, "资产报表.xlsx", storage.filename)
	require.Equal(t, content, storage.content)
}

func TestUploadBytesRejectsEmptyFilename(t *testing.T) {
	_, _, err := UploadBytes(&recordingOSS{}, "", []byte("excel"))
	require.EqualError(t, err, "文件名不能为空")
}

func TestUploadBytesPropagatesStorageError(t *testing.T) {
	want := errors.New("oss unavailable")
	_, _, err := UploadBytes(&recordingOSS{err: want}, "report.xlsx", []byte("excel"))
	require.ErrorIs(t, err, want)
}
