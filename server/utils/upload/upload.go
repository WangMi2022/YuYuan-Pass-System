package upload

import (
	"mime/multipart"

	"github.com/WangMi2022/mit-assets-admin/server/global"
)

// OSS 对象存储接口
// Author [MiTo]
// Author [MiTo]
type OSS interface {
	UploadFile(file *multipart.FileHeader) (string, string, error)
	DeleteFile(key string) error
}

// NewOss OSS的实例化方法
// Author [MiTo]
// Author [MiTo]
func NewOss() OSS {
	switch global.MAA_CONFIG.System.OssType {
	case "local":
		return &Local{}
	case "qiniu":
		return &Qiniu{}
	case "tencent-cos":
		return &TencentCOS{}
	case "aliyun-oss":
		return &AliyunOSS{}
	case "huawei-obs":
		return HuaWeiObs
	case "aws-s3":
		return &AwsS3{}
	case "cloudflare-r2":
		return &CloudflareR2{}
	case "minio":
		minioClient, err := GetMinio(global.MAA_CONFIG.Minio.Endpoint, global.MAA_CONFIG.Minio.AccessKeyId, global.MAA_CONFIG.Minio.AccessKeySecret, global.MAA_CONFIG.Minio.BucketName, global.MAA_CONFIG.Minio.UseSSL)
		if err != nil {
			global.MAA_LOG.Warn("你配置了使用minio，但是初始化失败，请检查minio可用性或安全配置: " + err.Error())
			panic("minio初始化失败") // 建议这样做，用户自己配置了minio，如果报错了还要把服务开起来，使用起来也很危险
		}
		return minioClient
	default:
		return &Local{}
	}
}
