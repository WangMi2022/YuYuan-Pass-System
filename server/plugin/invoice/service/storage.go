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
	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model"
	"github.com/WangMi2022/mit-assets-admin/server/utils/upload"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func safeInvoiceLocalPath(root, key string) (string, error) {
	cleanKey := filepath.Clean(strings.TrimLeft(strings.ReplaceAll(key, "\\", "/"), "/"))
	if cleanKey == "." || strings.HasPrefix(cleanKey, "..") || filepath.IsAbs(cleanKey) {
		return "", errors.New("非法发票文件路径")
	}
	if strings.TrimSpace(root) == "" {
		return "", errors.New("发票本地存储位置缺失")
	}
	base, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", err
	}
	fullPath, err := filepath.Abs(filepath.Join(base, cleanKey))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(base, fullPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("非法发票文件路径")
	}
	return fullPath, nil
}

func invoiceMinioClient(invoice model.Invoice) (*minio.Client, string, error) {
	endpoint := strings.TrimSpace(invoice.StorageEndpoint)
	bucket := strings.TrimSpace(invoice.StorageBucket)
	useSSL := invoice.StorageUseSSL
	if endpoint == "" || bucket == "" {
		return nil, "", errors.New("发票对象存储位置配置不完整")
	}
	if endpoint == global.GVA_CONFIG.Minio.Endpoint && bucket == global.GVA_CONFIG.Minio.BucketName && useSSL == global.GVA_CONFIG.Minio.UseSSL {
		client, err := upload.GetMinio(
			endpoint,
			global.GVA_CONFIG.Minio.AccessKeyId,
			global.GVA_CONFIG.Minio.AccessKeySecret,
			bucket,
			useSSL,
		)
		if err != nil {
			return nil, "", err
		}
		return client.Client, bucket, nil
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			global.GVA_CONFIG.Minio.AccessKeyId,
			global.GVA_CONFIG.Minio.AccessKeySecret,
			"",
		),
		Secure: useSSL,
	})
	return client, bucket, err
}

func openInvoiceObject(ctx context.Context, invoice model.Invoice) (io.ReadCloser, error) {
	switch invoice.StorageType {
	case "local":
		path, err := safeInvoiceLocalPath(invoice.StorageRoot, invoice.FileKey)
		if err != nil {
			return nil, err
		}
		return os.Open(path)
	case "minio":
		client, bucket, err := invoiceMinioClient(invoice)
		if err != nil {
			return nil, err
		}
		object, err := client.GetObject(ctx, bucket, invoice.FileKey, minio.GetObjectOptions{})
		if err != nil {
			return nil, err
		}
		if _, err = object.Stat(); err != nil {
			_ = object.Close()
			return nil, err
		}
		return object, nil
	default:
		return nil, errors.New("当前对象存储不支持私有发票读取")
	}
}

func deleteInvoiceObject(ctx context.Context, invoice model.Invoice) error {
	if strings.TrimSpace(invoice.FileKey) == "" {
		return nil
	}
	switch invoice.StorageType {
	case "local":
		path, err := safeInvoiceLocalPath(invoice.StorageRoot, invoice.FileKey)
		if err != nil {
			return err
		}
		if err = os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	case "minio":
		client, bucket, err := invoiceMinioClient(invoice)
		if err != nil {
			return err
		}
		deleteCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		return client.RemoveObject(deleteCtx, bucket, invoice.FileKey, minio.RemoveObjectOptions{})
	default:
		return errors.New("当前对象存储不支持私有发票删除")
	}
}
