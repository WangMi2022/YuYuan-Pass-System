package service

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/asset/model"
	"github.com/WangMi2022/mit-assets-admin/server/utils/upload"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func safeAssetRecognitionLocalPath(root, key string) (string, error) {
	cleanKey := filepath.Clean(strings.TrimLeft(strings.ReplaceAll(key, "\\", "/"), "/"))
	if cleanKey == "." || strings.HasPrefix(cleanKey, "..") || filepath.IsAbs(cleanKey) {
		return "", errors.New("非法资产识别图片路径")
	}
	if strings.TrimSpace(root) == "" {
		return "", errors.New("资产识别图片本地存储位置缺失")
	}
	base, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", err
	}
	fullPath, err := filepath.Abs(filepath.Join(base, cleanKey))
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(base, fullPath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("非法资产识别图片路径")
	}
	return fullPath, nil
}

func assetRecognitionMinioClient(job model.AssetRecognitionJob) (*minio.Client, string, error) {
	endpoint := strings.TrimSpace(job.StorageEndpoint)
	bucket := strings.TrimSpace(job.StorageBucket)
	if endpoint == "" || bucket == "" {
		return nil, "", errors.New("资产识别对象存储位置配置不完整")
	}
	if endpoint == global.GVA_CONFIG.Minio.Endpoint && bucket == global.GVA_CONFIG.Minio.BucketName && job.StorageUseSSL == global.GVA_CONFIG.Minio.UseSSL {
		client, err := upload.GetMinio(
			endpoint, global.GVA_CONFIG.Minio.AccessKeyId, global.GVA_CONFIG.Minio.AccessKeySecret,
			bucket, job.StorageUseSSL,
		)
		if err != nil {
			return nil, "", err
		}
		return client.Client, bucket, nil
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			global.GVA_CONFIG.Minio.AccessKeyId, global.GVA_CONFIG.Minio.AccessKeySecret, "",
		),
		Secure: job.StorageUseSSL,
	})
	return client, bucket, err
}

func openAssetRecognitionPhoto(ctx context.Context, job model.AssetRecognitionJob, key string) (io.ReadCloser, error) {
	switch job.StorageType {
	case "local":
		path, err := safeAssetRecognitionLocalPath(job.StorageRoot, key)
		if err != nil {
			return nil, err
		}
		return os.Open(path)
	case "minio":
		client, bucket, err := assetRecognitionMinioClient(job)
		if err != nil {
			return nil, err
		}
		object, err := client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
		if err != nil {
			return nil, err
		}
		if _, err = object.Stat(); err != nil {
			_ = object.Close()
			return nil, err
		}
		return object, nil
	default:
		return nil, errors.New("资产智能建档仅支持 local 或 minio 私有存储")
	}
}

func deleteAssetRecognitionPhoto(ctx context.Context, job model.AssetRecognitionJob, key string) error {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	switch job.StorageType {
	case "local":
		path, err := safeAssetRecognitionLocalPath(job.StorageRoot, key)
		if err != nil {
			return err
		}
		if err = os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	case "minio":
		client, bucket, err := assetRecognitionMinioClient(job)
		if err != nil {
			return err
		}
		deleteContext, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		return client.RemoveObject(deleteContext, bucket, key, minio.RemoveObjectOptions{})
	default:
		return errors.New("资产智能建档仅支持 local 或 minio 私有存储")
	}
}
